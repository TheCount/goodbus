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

package builder

import(
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Generic builder object
type Object interface {
}

// objectFromJSON constructs an object from a JSON value
func objectFromJSON( decoder *json.Decoder ) ( Object, error ) {
	var result Object
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf( "Unable to obtain JSON token while constructing object: %v", err )
	}
	switch x := token.( type ) {
	case json.Delim:
		if x == '[' {
			result, err = arrayFromJSON( decoder )
		} else if x == '{' {
			result, err = dictFromJSON( decoder )
		} else {
			// should not happen
			return nil, fmt.Errorf( "Bad delimiter: %v", x )
		}
		if err != nil {
			return nil, fmt.Errorf( "Unable to construct compound object from JSON: %v", err )
		}
		token, err = decoder.Token()
		if err != nil {
			return nil, fmt.Errorf( "Unable to get closing JSON delimiter for compound object: %v", err )
		}
		y, ok := token.( json.Delim )
		if !ok {
			// should not happen
			return nil, errors.New( "Undelimited compound object" )
		}
		if ( x == '[' && y != ']' ) || ( x == '{' && y != '}' ) {
			// should not happen
			return nil, errors.New( "Badly delimited compound object" )
		}
		return result, nil
	case bool:
		return Bool( x ), nil
	case json.Number:
		signed, err := x.Int64()
		if err == nil {
			return Int( signed ), nil
		}
		unsigned, err := strconv.ParseUint( string( x ), 10, 64 )
		if err == nil {
			return UInt( unsigned ), nil
		}
		flt, err := x.Float64()
		if err == nil {
			return Float( flt ), nil
		}
		return nil, fmt.Errorf( "Unable to parse number: %v", x )
	case string:
		return String( x ), nil
	case nil:
		return nil, nil
	default:
		// Should not happen
		return nil, fmt.Errorf( "Invalid token: %v", token )
	}
}
