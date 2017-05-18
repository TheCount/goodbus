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
	"container/heap"
)

// scheduleStack implements heap.Interface and hold schedules.
type scheduleStack []*Schedule

// Len returns the number of elements currently stored in the scheduleStack
func ( s scheduleStack ) Len() int {
	return len( s )
}

// Less defines an order on schedules by ascending trigger time.
func ( s scheduleStack ) Less( i, j int ) bool {
	return s[i].triggerTime.Before( s[j].triggerTime )
}

// Swap exchanges two elements in the scheduleStack
func ( s scheduleStack ) Swap( i, j int ) {
	s[i], s[j] = s[j], s[i]
}

// Push adds an element to the scheduleStack.
// If x is not a pointer to a Schedule, Push() panics.
func ( s *scheduleStack ) Push( x interface{} ) {
	*s = append( *s, x.( *Schedule ) )
}

// Pop removes an element from the scheduleStack.
// The type of the returned element is pointer to Schedule.
// If the scheduleStack is empty, Pop() panics.
func ( s *scheduleStack ) Pop() interface{} {
	oldstack := *s
	index := len( oldstack ) - 1
	x := oldstack[index]
	*s = oldstack[0 : index]
	return x
}

// ScheduleQueue is a priority queue of schedules.
type ScheduleQueue struct {
	scheduleStack
}

// Push inserts a schedule into the priority queue.
func ( sq *ScheduleQueue ) Push( x *Schedule ) {
	heap.Push( &sq.scheduleStack, x )
}

// Pop removes a schedule from the priority queue.
// If the queue is empty, nil is returned.
func ( sq *ScheduleQueue ) Pop() *Schedule {
	if sq.Len() == 0 {
		return nil
	}

	return heap.Pop( &sq.scheduleStack ).( *Schedule )
}

// Peek returns the minimal element from the priority queue
// without removing it.
// If the queue is empty, nil is returned.
func ( sq ScheduleQueue ) Peek() *Schedule {
	if len( sq.scheduleStack ) == 0 {
		return nil
	}
	return sq.scheduleStack[0]
}

// NewScheduleQueue creates a new schedule queue.
// reserve is the initial number of slots reserved for schedules
// in the new queue. A negative number will cause a panic.
func NewScheduleQueue( reserve int ) *ScheduleQueue {
	return &ScheduleQueue{
		make( scheduleStack, 0, reserve ),
	}
}
