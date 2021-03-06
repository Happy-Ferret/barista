// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package timing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func assertTriggered(t *testing.T, s Scheduler, msgAndArgs ...interface{}) {
	select {
	case <-s.Tick():
	case <-time.After(time.Second):
		require.Fail(t, "scheduler did not trigger", msgAndArgs...)
	}
}

func assertNotTriggered(t *testing.T, s Scheduler, msgAndArgs ...interface{}) {
	select {
	case <-s.Tick():
		require.Fail(t, "scheduler was triggered", msgAndArgs...)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestStop(t *testing.T) {
	ExitTestMode()

	sch := NewScheduler()
	assertNotTriggered(t, sch, "when not scheduled")

	sch.After(50 * time.Millisecond).Stop()
	assertNotTriggered(t, sch, "when stopped")

	sch.Every(50 * time.Millisecond).Stop()
	assertNotTriggered(t, sch, "when stopped")

	sch.At(Now().Add(50 * time.Millisecond)).Stop()
	assertNotTriggered(t, sch, "when stopped")

	sch.After(10 * time.Millisecond)
	assertTriggered(t, sch, "after interval elapses")

	sch.Stop()
	assertNotTriggered(t, sch, "when elapsed scheduler is stopped")

	sch.Stop()
	assertNotTriggered(t, sch, "when elapsed scheduler is stopped again")
}

func TestPauseResume(t *testing.T) {
	ExitTestMode()
	sch := NewScheduler()

	sch.At(Now().Add(5 * time.Millisecond))
	Pause()
	schWhilePaused := NewScheduler().After(2 * time.Millisecond)

	assertNotTriggered(t, sch, "when paused")
	assertNotTriggered(t, schWhilePaused, "scheduler created while paused")

	Resume()
	assertTriggered(t, sch, "when resumed")
	assertTriggered(t, schWhilePaused, "when resumed")

	Resume()
	assertNotTriggered(t, sch, "repeated resume is nop")
}

func TestRepeating(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real repeating test in short mode")
	}
	ExitTestMode()
	sch := NewScheduler()

	sch.Every(100 * time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	assertTriggered(t, sch, "after interval elapses")
	time.Sleep(100 * time.Millisecond)
	assertTriggered(t, sch, "after interval elapses")
	time.Sleep(100 * time.Millisecond)
	assertTriggered(t, sch, "after interval elapses")

	Pause()
	time.Sleep(100 * time.Millisecond)
	assertNotTriggered(t, sch, "when paused")
	time.Sleep(1 * time.Second) // > 2 intervals.
	Resume()

	assertTriggered(t, sch, "when resumed")
	assertNotTriggered(t, sch, "only once on resume")

	sch.After(20 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	assertTriggered(t, sch, "after delay elapses")
	assertNotTriggered(t, sch, "after first trigger")
}

func TestCoalescedUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real coalescing test in short mode")
	}
	ExitTestMode()
	sch := NewScheduler()
	sch.Every(300 * time.Millisecond)
	time.Sleep(3100 * time.Millisecond)
	assertTriggered(t, sch, "after multiple intervals")
	assertNotTriggered(t, sch, "multiple updates coalesced")
}

func TestPastTriggers(t *testing.T) {
	ExitTestMode()
	sch := NewScheduler()
	sch.After(-1 * time.Minute)
	assertTriggered(t, sch, "negative delay notifies immediately")
	sch.At(Now().Add(-1 * time.Minute))
	assertTriggered(t, sch, "past trigger notifies immediately")

	Pause()
	sch.After(-1 * time.Minute)
	assertNotTriggered(t, sch, "when paused")
	Resume()
	assertTriggered(t, sch, "on resume")

	Pause()
	sch.At(Now().Add(-1 * time.Minute))
	assertNotTriggered(t, sch, "when paused")
	Resume()
	assertTriggered(t, sch, "on resume")

	require.Panics(t, func() {
		sch.Every(-1 * time.Second)
	}, "negative repeating interval")
}
