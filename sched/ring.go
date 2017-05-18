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
	"container/ring"
)

// ScheduleRing is a circular list of schedules.
// Its zero value is the empty list.
type ScheduleRing struct {
	ring *ring.Ring
}

// IsEmpty checks whether the ScheduleRing ring is empty.
func ( sr ScheduleRing ) IsEmpty() bool {
	return ( sr.ring == nil )
}

// Insert inserts a schedule into the ScheduleRing.
func ( sr *ScheduleRing ) Insert( x *Schedule ) {
	r := &ring.Ring{
		Value: x,
	}
	r.Link( sr.ring )
	sr.ring = r
}

// Next obtains the next schedule in the ring.
// If the ring is empty, nil is returned.
func ( sr *ScheduleRing ) Next() *Schedule {
	if sr.ring == nil {
		return nil
	} else {
		sr.ring = sr.ring.Next()
		return sr.ring.Value.( *Schedule )
	}
}

// Remove removes the element last returned by Next from the ring.
// Remove works only if the previous call on the specified ring
// was actually a call to Next. Otherwise, the behaviour is undefined.
func ( sr *ScheduleRing ) Remove() {
	if sr.ring != nil {
		previous := sr.ring.Prev()
		if previous == sr.ring {
			sr.ring = nil
		} else {
			next := sr.ring.Next()
			previous.Link( next )
			sr.ring = previous
		}
	}
}
