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
	"github.com/TheCount/goodbus/builder"
	"time"
)

const(
	kBitfield = "bitfield"
	kBitmap = "bitmap"
	kInt16 = "int16"
	kLength = "length"
	kNumber = "number"
	kOffset = "offset"
	kScaler = "scaler"
	kTime = "time"
	kUInt16 = "uint16"
	kValue = "value"
)

// TypeInfo carries information about the types supported
type typeInfo struct {
	// size is the size of the type in bytes.
	// A value of zero means variable size.
	size uint

	// build builds a value for the type
	build func( data []byte, conf config, length uint ) ( builder.Dict, error )
}

var typeInfoMap = map[string]typeInfo{
	kBitfield: { 0, buildBitfield },
	kInt16: { 2, buildInt16 },
	kUInt16: { 2, buildUInt16 },
}

// buildBitfield builds a value for a bitfield
func buildBitfield( data []byte, conf config, length uint ) ( builder.Dict, error ) {
	result := builder.NewDict()
	result[kType] = kBitfield
	result[kLength] = builder.UInt( length )
	bitmap := builder.NewDict()
	list, err := conf.GetList( kBitmap )
	if err != nil {
		return nil, fmt.Errorf( "Unable to get bitmap for bitfield: %v", err )
	}
	for i, item := range list {
		if uint( i ) >= length {
			return nil, fmt.Errorf( "Bitmap entry out of bounds (length: %v)", length )
		}
		if item == nil {
			continue
		}
		name, ok := item.( string )
		if !ok {
			return nil, fmt.Errorf( "Bitmap entry name must be a string: %v", item )
		}
		bitmap[name] = builder.Bool( ( data[i / 8] & ( 1 << ( uint( i ) % 8 ) ) ) != 0 )
	}
	result[kValue] = bitmap

	return result, nil
}

// buildInt16 builds a 16 bit signed integer value
func buildInt16( data []byte, conf config, unused uint ) ( builder.Dict, error ) {
	result := builder.NewDict()
	result[kType] = kNumber
	scaler, err := conf.GetFloatOrDefault( kScaler, 1.0 )
	if err != nil {
		return nil, fmt.Errorf( "Unable to obtain scaler for 16 bit integer: %v", err )
	}
	hw := ( uint16( data[0] ) << 8 ) | uint16( data[1] )
	var value float64
	if ( hw & 0x8000 ) != 0 {
		value = -float64( ^hw ) - 1.0
	} else {
		value = float64( hw )
	}
	result[kValue] = builder.Float( value * scaler )

	return result, nil
}

// buildUInt16 builds a 16 bit unsigned integer value
func buildUInt16( data []byte, conf config, unused uint ) ( builder.Dict, error ) {
	result := builder.NewDict()
	result[kType] = kNumber
	scaler, err := conf.GetFloatOrDefault( kScaler, 1.0 )
	if err != nil {
		return nil, fmt.Errorf( "Unable to obtain scaler for 16 bit unsigned integer: %v", err )
	}
	value := float64( ( uint16( data[0] ) << 8 ) | uint16( data[1] ) )
	result[kValue] = builder.Float( value * scaler )

	return result, nil
}

// buildValue builds a single value dictionary
func buildValue( data []byte, valueConf config ) ( builder.Object, error ) {
	// copy config first because we're going to alter it
	conf := make( config )
	for key, value := range valueConf {
		conf[key] = value
	}

	// Extract type info
	offset, err := conf.GetUInt( kOffset )
	if err != nil {
		return nil, fmt.Errorf( "Unable to extract offset: %v", err )
	}
	typ, err := conf.GetString( kType )
	if err != nil {
		return nil, fmt.Errorf( "Unable to extract type: %v", err )
	}
	delete( conf, kOffset )
	delete( conf, kType )

	// check validity of type/offset
	info, ok := typeInfoMap[typ]
	if !ok {
		return nil, fmt.Errorf( "Unknown type '%v'", typ )
	}
	var length uint
	size := info.size
	if size == 0 {
		length, err = conf.GetUInt( kLength )
		if err != nil {
			return nil, fmt.Errorf( "Unable to extract mandatory length for type '%v': %v", typ, err )
		}
		delete( conf, kLength )
		size = ( length + 7 ) / 8 // size = length in bits as bytes, rounded up
	}
	if uint( len( data ) ) < 2 * offset + size {
		return nil, fmt.Errorf( "Offset %v and/or size %v out of bounds (data length: %v)", offset, size, len( data ) )
	}

	// Create and return result
	result, err := info.build( data[2 * offset : 2 * offset + size], conf, length )
	if err != nil {
		return nil, fmt.Errorf( "Unable to build value: %v", err )
	}
	for key, value := range conf {
		s, ok := value.( string )
		if !ok {
			return nil, fmt.Errorf( "Invalid value '%v' for key '%v'", value, key )
		}
		result[key] = builder.String( s )
	}

	return result, nil
}

// buildObject builds a generic object from time, binary data, and a value configuration.
func buildObject( time time.Time, data []byte, valueConf config ) ( builder.Object, error ) {
	result := builder.NewDict()
	result[kTime] = builder.Float( time.Unix() ) + 1e-9 * builder.Float( time.Nanosecond() )
	values := builder.NewDict()
	for valueName, _ := range valueConf {
		conf, err := valueConf.GetSubConfig( valueName )
		if err != nil {
			return nil, fmt.Errorf( "Unable to get value configuration '%v': %v", valueName, err )
		}
		value, err := buildValue( data, conf )
		if err != nil {
			return nil, fmt.Errorf( "Unable to build value '%v': %v", valueName, err )
		}
		values[valueName] = value
	}
	result[kValues] = values

	return result, nil
}
