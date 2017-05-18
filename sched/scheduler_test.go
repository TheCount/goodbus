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
	"testing"
	"time"
)

type TestCommand struct {
	Id int
	err error
	reportChan chan<- int
}

func ( c *TestCommand ) Execute() error {
	c.reportChan <- c.Id
	return c.err
}

func ( c *TestCommand ) Finalize() {
	close( c.reportChan )
}

func assertPanic( t *testing.T ) {
	r := recover()
	if r == nil {
		t.Error( "No panic" )
	}
}

func TestNewScheduler( t *testing.T ) {
	for i := 0; i < 100; i++ {
		s := NewScheduler( i )
		if s == nil {
			t.Error( "Got nil scheduler" )
		}
	}

	defer assertPanic( t )
	s := NewScheduler( -1 )
	t.Errorf( "Got %v instead of panic", s )
}

func TestStartScheduler( t *testing.T ) {
	for i := 0; i < 100; i++ {
		s := NewScheduler( 5 )
		errChan, err := s.Start( i )
		if err != nil {
			t.Errorf( "Unable to start scheduler with error backlog %d", i )
		}
		if errChan == nil {
			t.Error( "Returned error channel is nil" )
		}
		s.Stop()
	}

	s := NewScheduler( 5 )
	errChan, err := s.Start( -1 )
	if err == nil {
		t.Error( "Started scheduler with negative error backlog" )
	}
	errChan, err = s.Start( 5 )
	if err != nil {
		t.Error( "Unable to start scheduler" )
	}
	if errChan == nil {
		t.Error( "Returned error channel is nil" )
	}
	errChan, err = s.Start( 5 )
	if err == nil {
		t.Error( "Started scheduler twice" )
	}
	s.Stop()
	errChan, err = s.Start( 5 )
	if err != nil {
		t.Error( "Unable to start scheduler after stop" )
	}
	if errChan == nil {
		t.Error( "Returned error channel is nil" )
	}
}

func TestAddScheduler( t *testing.T ) {
	s := NewScheduler( 5 )
	_, err := s.Start( 5 )
	if err != nil {
		t.Error( "Unable to start scheduler" )
	}
	reportChan := make( chan int )

	err = s.Add( "test", &TestCommand{ 1, nil, reportChan }, Schedule{ Flags: ScheduleIdle, MinWait: 0, MaxWait: time.Second } )
	if err != nil {
		t.Error( "Unable to add command" )
	}

	err = s.Add( "test", &TestCommand{ 2, nil, reportChan }, Schedule{ Flags: ScheduleIdle, MinWait: 0, MaxWait: time.Second } )
	if err == nil {
		t.Error( "Successfully added command with the same name twice" )
	}

	result := <-reportChan
	if result != 1 {
		t.Error( "Command yielded wrong result" )
	}
}

func TestRepeat( t *testing.T ) {
	s := NewScheduler( 5 )
	_, err := s.Start( 5 )
	if err != nil {
		t.Error( "Unable to start scheduler" )
	}
	reportChan := make( chan int )

	err = s.Add( "test", &TestCommand{ 1, nil, reportChan }, Schedule{ Flags: ScheduleIdle | ScheduleRepeat, MinWait: 0, MaxWait: time.Second } )
	if err != nil {
		t.Error( "Unable to add command" )
	}
	result := <-reportChan
	if result != 1 {
		t.Error( "Command yielded wrong result" )
	}
	result = <-reportChan
	if result != 1 {
		t.Error( "Command yielded wrong result" )
	}
}
