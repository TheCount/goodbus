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
	"math"
	"time"
	_ "unsafe" // for go:linkname
)

//go:linkname nanotime runtime.nanotime
func nanotime() int64

// monotonicTime describes a concept of time that is unaffected by clock
// adjustments.
// This should become unnecessary in future versions of go, see
// https://github.com/golang/go/issues/12914.
type monotonicTime int64

const(
	// Farthest monotonic time point in the future
	inTheFuture monotonicTime = math.MaxInt64
)

// monotonicNow returns the current monotonic time.
func monotonicNow() monotonicTime {
	return monotonicTime( nanotime() )
}

// Sub calculates the duration between two monotonic times.
// If T1 and T2 are results of a call to monotonicNow,
// and the call for T2 happened after the call for T1,
// then T2.Sub( T1 ) is guaranteed to be non-negative.
func ( t2 monotonicTime ) Sub( t1 monotonicTime ) time.Duration {
	return time.Duration( t2 - t1 )
}

// Add adds a duration to a monotonic time.
func ( t monotonicTime ) Add( d time.Duration ) monotonicTime {
	return t + monotonicTime( d )
}

// Before checks whether time instant t is before u.
func ( t monotonicTime ) Before( u monotonicTime ) bool {
	return t < u
}
