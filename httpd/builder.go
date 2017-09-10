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
	"errors"
	"fmt"
	"github.com/TheCount/goodbus/builder"
	"io"
	"math"
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

// buildFunc is a function type for functions that build values
// from binary data and configurations.
type buildFunc func( data []byte, conf config ) ( builder.Dict, error )

// serialiseFunc is a function type for functions that serialise
// objects into serial data.
type serialiseFunc func( data []byte, obj builder.Object, conf config ) error

// extractNumber extracts a numeric value out of an object
func extractNumber( value builder.Object, conf config ) ( float64, error ) {
	// Get scaler
	scaler, err := conf.GetFloatOrDefault( kScaler, 1.0 )
	if err != nil {
		return 0, fmt.Errorf( "Unable to obtain scaler for 16 bit integer: %v", err )
	}
	if scaler == 0 {
		return 0, errors.New( "Scaler must not be zero" )
	}

	// Get value
	var scaled float64
	switch v := value.( type ) {
	case builder.UInt:
		scaled = float64( v ) / scaler // FIXME: special considerations for large v
	case builder.Int:
		scaled = float64( v ) / scaler // FIXME: special considerations for large v
	case builder.Float:
		scaled = float64( v ) / scaler
	default:
		return 0, fmt.Errorf( "Unable to serialise value from unknown object type: %+v", v )
	}

	return scaled, nil
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

// serialiseBitfield serialises a bitfield
func serialiseBitfield( data []byte, value builder.Object, conf config, length uint ) error {
	// Sanity check
	size := ( length + 7 ) / 8
	if uint( len( data ) ) < size {
		return fmt.Errorf( "Data length %v too short to hold bitfield of size %v", len( data ), length )
	}
	origdict, ok := value.( builder.Dict )
	if !ok {
		return errors.New( "Bitfield value must be a dictionary" )
	}

	// Store bits
	dict := make( builder.Dict )
	for key, value := range origdict {
		dict[key] = value
	}
	for i := uint( 0 ); i != size; i++ {
		data[i] = 0
	}
	list, err := conf.GetList( kBitmap )
	if err != nil {
		return fmt.Errorf( "Unable to get bitmap configuration for %+v: %v", dict, err )
	}
	for i, item := range list {
		if uint( i ) >= length {
			return fmt.Errorf( "Bitmap entry out of bounds (length: %v)", length )
		}
		if item == nil {
			continue
		}
		name, ok := item.( string )
		if !ok {
			return fmt.Errorf( "Bitmap entry name must be a string: %v", item )
		}
		entry, ok := dict[name]
		if !ok {
			return fmt.Errorf( "Mandatory bitmap entry '%v' not found", name )
		}
		boolv, ok := entry.( builder.Bool )
		if !ok {
			return fmt.Errorf( "Bitmap entry '%v' must have a boolean value", name )
		}
		if boolv {
			data[i / 8] |= 1 << ( uint( i ) % 8 )
		}
		delete( dict, name )
	}
	if len( dict ) > 0 {
		return fmt.Errorf( "Unable to assign unsupported entries: %+v", dict )
	}

	return nil
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

// serialiseInt16 serialises a value into a 16 bit 2's complement integer value
func serialiseInt16( data []byte, value builder.Object, conf config, unused uint ) error {
	// Sanity check
	if len( data ) < 2 {
		return fmt.Errorf( "Invalid data length for 16 bit integer: %v", len( data ) )
	}

	// Serialise
	scaled, err := extractNumber( value, conf )
	if err != nil {
		return fmt.Errorf( "Unable to extract numeric value from %+v: %v", value, err )
	}
	if scaled >= math.MaxInt16 + 0.5 || scaled <= math.MinInt16 - 0.5 {
		return fmt.Errorf( "Scaled value out of bounds: %v", scaled )
	}
	var intv int16
	if scaled >= 0 {
		intv := int16( scaled + 0.5 )
	} else {
		intv := int16( scaled - 0.5 )
	}
	uintv := uint16( intv )
	data[0] = byte( uintv >> 8 )
	data[1] = byte( uintv & 0xFF )

	return nil
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

// serialiseUInt16 serialises a value into a 16 bit unsigned integer value
func serialiseUInt16( data []byte, value builder.Object, conf config, unused uint ) error {
	// Sanity check
	if len( data ) < 2 {
		return fmt.Errorf( "Invalid data length for 16 bit unsigned integer: %v", len( data ) )
	}

	// Serialise
	scaled, err := extractNumber( value, conf )
	if err != nil {
		return fmt.Errorf( "Unable to extract numeric value from %+v: %v", value, err )
	}
	if scaled >= math.MaxUint16 + 0.5 || scaled <= -0.5 {
		return fmt.Errorf( "Scaled value out of bounds: %v", scaled )
	}
	uintv := uint16( scaled + 0.5 )
	data[0] = byte( uintv >> 8 )
	data[1] = byte( uintv & 0xFF )

	return nil
}

// extractInfo extracts offset and type information from a values configuration
func extractInfo( conf config ) ( uint, uint, buildFunc, serialiseFunc, error ) {
	// TypeInfo carries information about the types supported
	type typeInfo struct {
		// size is the size of the type in bytes.
		// A value of zero means variable size.
		size uint

		// build builds a value for the type
		build func( data []byte, conf config, length uint ) ( builder.Dict, error )

		// serialise serialises a value for the type
		serialise func( data []byte, obj builder.Object, conf config, length uint ) error
	}

	var typeInfoMap = map[string]typeInfo{
		kBitfield: { 0, buildBitfield, serialiseBitfield },
		kInt16: { 2, buildInt16, serialiseInt16 },
		kUInt16: { 2, buildUInt16, serialiseUInt16 },
	}

	offset, err := conf.GetUInt( kOffset )
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf( "Unable to extract offset: %v", err )
	}
	typ, err := conf.GetString( kType )
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf( "Unable to extract type: %v", err )
	}

	// check validity of type/offset
	info, ok := typeInfoMap[typ]
	if !ok {
		return 0, 0, nil, nil, fmt.Errorf( "Unknown type '%v'", typ )
	}
	var length uint
	size := info.size
	if size == 0 {
		length, err = conf.GetUInt( kLength )
		if err != nil {
			return 0, 0, nil, nil, fmt.Errorf( "Unable to extract mandatory length for type '%v': %v", typ, err )
		}
		size = ( length + 7 ) / 8 // size = length in bits as bytes, rounded up
	}

	return offset, size, func( data []byte, conf config ) ( builder.Dict, error ) {
		return info.build( data, conf, length )
	}, func( data []byte, obj builder.Object, conf config ) error {
		return info.serialise( data, obj, conf, length )
	}, nil
}

// buildValue builds a single value dictionary
func buildValue( data []byte, valueConf config ) ( builder.Object, error ) {
	// get info
	offset, size, build, _, err := extractInfo( valueConf );
	if err != nil {
		return nil, fmt.Errorf( "Unable to extract info from config: %v", err )
	}

	// create altered config
	conf := make( config )
	for key, value := range valueConf {
		conf[key] = value
	}
	delete( conf, kOffset )
	delete( conf, kType )
	delete( conf, kLength )

	// Create and return result
	if uint( len( data ) ) < 2 * offset + size {
		return nil, fmt.Errorf( "Offset %v and/or size %v out of bounds (data length: %v)", offset, size, len( data ) )
	}
	result, err := build( data[2 * offset : 2 * offset + size], conf )
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

// serialiseValue serialises an object value to binary according to a value subconfiguration
func serialiseValue( data []byte, object builder.Object, valueConf config ) error {
	// get info
	offset, size, _, serialise, err := extractInfo( valueConf )
	if err != nil {
		return fmt.Errorf( "Unable to extract info from config: %v", err )
	}

	// Serialise
	if uint( len( data ) ) < 2 * offset + size {
		return fmt.Errorf( "Offset %v and/or size %v out of bounds (data length: %v)", offset, size, len( data ) )
	}
	err = serialise( data[2 * offset : 2 * offset + size], object, valueConf )
	if err != nil {
		return fmt.Errorf( "Unable to serialise value: %v", err )
	}

	return nil
}

// buildData builds binary data from JSON according to a value configuration.
func buildData( r io.Reader, valueConf config, size int ) ( []byte, error ) {
	// Sanity check
	if ( size <= 0 ) {
		return nil, fmt.Errorf( "Data size must be positive: %v", size );
	}
	data := make( []byte, size )

	// Get object
	obj, err := builder.FromJSON( r )
	if err != nil {
		return nil, fmt.Errorf( "Unable to build object: %v", err );
	}
	dict, ok := obj.( *builder.Dict )
	if !ok {
		return nil, errors.New( "Not a JSON object" )
	}

	// Traverse configuration to build data
	for key, _ := range valueConf {
		entry, ok := dict[key]
		if !ok {
			return nil, fmt.Errorf( "Entry for value '%s' missing", key )
		}
		subConf, err := valueConf.GetSubConfig( key )
		if err != nil {
			return nil, fmt.Errorf( "Invalid subconfiguration '%s': %v", key, err )
		}
		err = serialiseValue( data, entry, subConf )
		if err != nil {
			return nil, fmt.Errorf( "Unable to serialise value '%s': %v", key, err )
		}
	}

	return data, nil
}
