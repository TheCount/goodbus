/*
Copyright (c) 2017 Alexander Klauer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package sched

import(
	"errors"
	"fmt"
	"sync"
	"time"
)

// Scheduler schedules commands within a single goroutine
// according to specific schedules.
type Scheduler struct {
	// errChan is a channel used by the scheduler to report errors.
	errChan chan<- error

	// scheduleChan is a channel used by the scheduler to communicate
	// new schedules.
	scheduleChan chan *Schedule

	// waitGroup is used to control scheduler start/stop.
	waitGroup sync.WaitGroup

	// scheduleMap maps schedule names to actual schedules.
	scheduleMap map[string]*Schedule

	// mapMutex protects access to scheduleMap.
	// Future versions of go provide sync.Map,
	// and that should be used instead
	// once it becomes available.
	mapMutex sync.Mutex

	// waitingQueue is the queue of commands waiting to become
	// pending.
	waitingQueue *ScheduleQueue

	// pendingQueue is the queue of commands waiting to be executed.
	pendingQueue *ScheduleQueue

	// idleRing is the ring of commands executed only when the scheduler
	// is otherwise idle.
	idleRing ScheduleRing

	// isRunning indicates whether the scheduler is currently running.
	isRunning bool
}

// NewScheduler creates a new scheduler.
// scheduleBufferSize is a performance parameter.
// Higher values need more memory but can avoid unnecessary goroutine stalls.
// In most situations, a value around 10 is sufficient.
// Negative values cause a panic.
func NewScheduler( scheduleBufferSize int ) *Scheduler {
	return &Scheduler{
		scheduleChan: make( chan *Schedule, scheduleBufferSize ),
	}
}

// execute executes a command
// unless the corresponding schedule is marked as removed.
// Returns whether the command should be removed.
// If it should be removed, it is also marked as such and
// already removed from the schedule map.
func ( s *Scheduler ) execute( candidate *Schedule ) ( shouldRemove bool ) {
	shouldRemove = false
	flags := candidate.getFlags()
	if flags & scheduleRemoved == 0 {
		err := candidate.command.Execute()
		if err != nil {
			s.errChan <- err
			if flags & ScheduleRemoveOnError != 0 {
				shouldRemove = true
			}
		}
	} else {
		shouldRemove = true
	}
	if flags & ScheduleRepeat == 0 {
		shouldRemove = true
	}
	if shouldRemove {
		candidate.markRemoved()
		s.mapMutex.Lock()
		delete( s.scheduleMap, candidate.name )
		s.mapMutex.Unlock()
		candidate.once.Do( candidate.command.Finalize )
	}

	return
}

// doSomeWork does a bit of scheduling.
// Normally, true is returned.
// If the schedule channel has been closed, false is returned.
func ( s *Scheduler ) doSomeWork() bool {
	now := monotonicNow()

	// Move a schedule from the waiting queue to the pending queue if appropriate
	candidate := s.waitingQueue.Peek()
	if candidate != nil && candidate.triggerTime <= now {
		s.waitingQueue.Pop()
		candidate.triggerTime = candidate.triggerTime.Add( candidate.MaxWait - candidate.MinWait )
		s.pendingQueue.Push( candidate )
	}

	// Execute a command from the pending queue if appropriate
	var didExecute bool
	candidate = s.pendingQueue.Peek()
	if candidate != nil && candidate.triggerTime.Before( now ) {
		s.pendingQueue.Pop()
		shouldRemove := s.execute( candidate )
		didExecute = true
		if !shouldRemove {
			candidate.triggerTime = candidate.triggerTime.Add( candidate.MinWait )
			flags := candidate.getFlags()
			if candidate.triggerTime < now && ( flags & ScheduleBurst ) == 0 {
				candidate.triggerTime = now.Add( candidate.MinWait )
			}
			s.waitingQueue.Push( candidate )
		}
	}

	// Execute a command from the idle ring if we haven't done so yet
	if !( didExecute || s.idleRing.IsEmpty() ) {
		candidate = s.idleRing.Next()
		shouldRemove := s.execute( candidate )
		didExecute = true
		if shouldRemove {
			s.idleRing.Remove()
		}
	}

	// Get a new schedule
	chanStillOpen := true
	later := inTheFuture
	candidate = s.waitingQueue.Peek()
	if candidate != nil {
		later = candidate.triggerTime
	}
	candidate = s.pendingQueue.Peek()
	if candidate != nil && later > candidate.triggerTime {
		later = candidate.triggerTime
	}
	if didExecute || later.Before( now ) {
		select {
		case candidate, chanStillOpen = <-s.scheduleChan:

		default:
			candidate = nil
		}
	} else {
		select {
		case candidate, chanStillOpen = <-s.scheduleChan:
			now = monotonicNow()

		case <-time.After( later.Sub( now ) ):
			candidate = nil
		}
	}

	// Insert new schedule
	if candidate != nil {
		flags := candidate.getFlags()
		if flags & ScheduleIdle == 0 {
			candidate.triggerTime = now.Add( candidate.MinWait )
			s.waitingQueue.Push( candidate )
		} else {
			s.idleRing.Insert( candidate )
		}
	}

	return chanStillOpen
}

// run runs the scheduler.
// This is meant to be called as a new goroutine.
func ( s *Scheduler ) run() {
	defer s.waitGroup.Done()

	for s.doSomeWork() {
	}

	close( s.errChan )
}

// Start starts the scheduler.
// A channel reporting scheduler errors is returned.
// The buffer size of this channel is given by error backlog.
// On success, the second return value is nil.
// Otherwise, it is an appropriate error message.
func ( s *Scheduler ) Start( errorBacklog int ) ( <-chan error, error ) {
	if s.isRunning {
		return nil, errors.New( "Scheduler already running" )
	}
	if errorBacklog < 0 {
		return nil, fmt.Errorf( "Bad error backlog: %d", errorBacklog )
	}
	scheduleBufferSize := cap( s.scheduleChan )
	errChan := make( chan error, errorBacklog )
	s.errChan = errChan
	s.scheduleChan = make( chan *Schedule, scheduleBufferSize )
	s.waitGroup = sync.WaitGroup{}
	s.scheduleMap = make( map[string]*Schedule )
	s.waitingQueue = NewScheduleQueue( scheduleBufferSize )
	s.pendingQueue = NewScheduleQueue( scheduleBufferSize )
	s.idleRing = ScheduleRing{}
	s.waitGroup.Add( 1 )
	go s.run()
	s.isRunning = true

	return errChan, nil
}

// SignalStop signals the scheduler to stop,
// but does not wait for it to actually stop.
// Use the WaitStop method for that.
func ( s *Scheduler ) SignalStop() {
	if s.isRunning {
		close( s.scheduleChan )
	}
}

// WaitStop waits for the scheduler to stop
// after a call to SignalStop.
// Without a call to SignalStop,
// WaitStop will wait forever.
func ( s *Scheduler ) WaitStop() {
	if s.isRunning {
		s.waitGroup.Wait()
		s.isRunning = false
	}
}

// Stop stops the scheduler.
// It combines SignalStop and WaitStop into one method.
func ( s *Scheduler ) Stop() {
	s.SignalStop()
	s.WaitStop()
}

// Add adds a schedule to the scheduler.
// Can only be called on a started scheduler,
// but is otherwise goroutine-safe.
func ( s *Scheduler ) Add( name string, command Command, schedule Schedule ) error {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	oldSchedule, ok := s.scheduleMap[name]
	if ok && ( oldSchedule.getFlags() & scheduleRemoved == 0 ) {
		return fmt.Errorf( "A schedule named '%s' already exists", name )
	}
	if command == nil {
		return fmt.Errorf( "Attempting to add schedule '%s' with nil command", name )
	}
	if schedule.MinWait < 0 {
		return fmt.Errorf( "Negative minimum wait: %d ns", schedule.MinWait )
	}
	if schedule.MaxWait < schedule.MinWait {
		return fmt.Errorf( "Maximum wait %d ns smaller than minimum wait %d ns", schedule.MaxWait, schedule.MinWait )
	}
	schedule.name = name
	schedule.command = command
	s.scheduleMap[name] = &schedule
	s.scheduleChan <- &schedule

	return nil
}

// Remove removes a schedule from the scheduler.
// Can only be called on a started scheduler,
// but is otherwise goroutine-safe.
func ( s *Scheduler ) Remove( name string ) error {
	s.mapMutex.Lock()
	schedule, ok := s.scheduleMap[name]
	s.mapMutex.Unlock()
	if !ok || ( schedule.getFlags() & scheduleRemoved != 0 ) {
		return fmt.Errorf( "A schedule named '%s' does not exist", name )
	}
	schedule.markRemoved()

	return nil
}
