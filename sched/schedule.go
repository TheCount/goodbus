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
	"sync"
	"sync/atomic"
	"time"
)

const(
	// ScheduleRepeat indicates that the scheduled command
	// should not be removed from the scheduler after execution.
	ScheduleRepeat = 1 << iota

	// ScheduleIdle indicates that the scheduled command
	// should be executed only if the scheduler is idle.
	ScheduleIdle

	// ScheduleBurst indicates that if the system is currently too slow
	// to keep up executing the command, still commands should not be
	// skipped. This flag only makes sense together with ScheduleRepeat.
	ScheduleBurst

	// ScheduleRemoveOnError indicates that the command should be removed
	// from the scheduler if it returns an error.
	// This flag only makes sense together with ScheduleRepeat.
	ScheduleRemoveOnError

	// scheduleRemoved indicates that the command has been removed
	// from the scheduler. Internal use only.
	scheduleRemoved
)

// Schedule describes how a command should be scheduled
type Schedule struct {
	Flags uint32

	// MinWait is minimum duration to wait before the command
	// is scheduled (again).
	MinWait time.Duration

	// MaxWait is the maximum duration to wait before the command
	// is scheduled (again).
	//
	// It can still happen that it takes longer than MaxWait
	// to (re-)schedule the command, namely if the system is too busy or
	// if the difference between MinWait and MaxWait is too small for the
	// timing resolution of the system.
	MaxWait time.Duration

	// name is used internally to store the name of the schedule
	name string

	// triggerTime is used internally to store the time when
	// a MinWait or a MaxWait elapses.
	triggerTime monotonicTime

	// command is used internally to store the actual command.
	command Command

	// once is used to call the commands finalizer.
	once sync.Once
}

// getFlags atomically obtains the schedule flags.
func ( s *Schedule ) getFlags() uint32 {
	return atomic.LoadUint32( &s.Flags )
}

// markRemoved atomically marks a schedule as removed.
func ( s *Schedule ) markRemoved() {
	flags := s.getFlags() | scheduleRemoved
	atomic.StoreUint32( &s.Flags, flags )
}
