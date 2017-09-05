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

package main

import(
	"fmt"
	"sync/atomic"
	"time"
)

type scratchpadType struct {
	// Time is the update time.
	Time time.Time

	// Data is the scratchpad data.
	Data []byte
}

type Scratchpad struct {
	atomic.Value

	// Size is the immutable size the scratchpad data should have
	Size int
}

// NewScratchpad creates a new scratchpad with the supposed size.
func NewScratchpad( size int ) *Scratchpad {
	return &Scratchpad{
		Size: size,
	}
}

func ( sp *Scratchpad ) Update( data []byte ) error {
	if len( data ) != sp.Size {
		return fmt.Errorf( "Expected scratchpad data size of %v, got %v", sp.Size, len( data ) )
	}
	newvalue := scratchpadType{
		Time: time.Now(),
		Data: data,
	}
	sp.Value.Store( newvalue )

	return nil
}

func ( sp *Scratchpad ) Get() ( time.Time, []byte ) {
	value := sp.Value.Load()
	if value == nil {
		return time.Time{}, nil
	}
	st := value.( scratchpadType )

	return st.Time, st.Data
}
