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
	"fmt"
)

// Dict is the builder dictionary type
type Dict map[string]Object

// NewDict creates a new dictionary
func NewDict() Dict {
	return make( Dict )
}

// dictFromJSON constructs a dictionary from a JSON object
func dictFromJSON( decoder *json.Decoder ) ( Dict, error ) {
	result := NewDict()
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf( "Unable to parse dictionary key from JSON: %v", err )
		}
		key, ok := token.( string )
		if !ok {
			return nil, fmt.Errorf( "Dictionary key '%v' should be a string", token )
		}
		_, ok = result[key]
		if ok {
			return nil, fmt.Errorf( "Duplicate dictionary key '%v'", key )
		}
		value, err := objectFromJSON( decoder )
		if err != nil {
			return nil, fmt.Errorf( "Unable to parse dictionary value for key '%v' from JSON: %v", key, err )
		}
		result[key] = value
	}

	return result, nil
}
