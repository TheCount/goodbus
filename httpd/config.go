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
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
	"time"
)

type config map[string]interface{}

// getConfig reads in the configuration
// from the file name provided on the command line.
func getConfig() ( config, error ) {
	if len( os.Args ) <= 1 {
		return nil, errors.New( "Please provide a config file name on the command line" )
	}
	viper.SetConfigFile( os.Args[1] )
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf( "Unable to read configuration file %s: %v", os.Args[1], err )
	}

	return config( viper.AllSettings() ), nil
}

// GetSubConfig gets a subconfiguration from a config.
func ( c config ) GetSubConfig( name string ) ( config, error ) {
	item, ok := c[name]
	if !ok {
		return nil, fmt.Errorf( "Subconfiguration '%s' not found", name )
	}
	result, ok := item.( config )
	if !ok {
		return nil, fmt.Errorf( "Item '%s' is not a subconfiguration", name )
	}

	return result, nil
}

// GetBoolOrDefault gets a boolean value from the config,
// or a default value.
func ( c config ) GetBoolOrDefault( name string, dflt bool ) ( bool, error ) {
	item, ok := c[name]
	if !ok {
		return dflt, nil
	}
	result, ok := item.( bool )
	if !ok {
		return false, fmt.Errorf( "Item '%s' is not a boolean", name )
	}

	return result, nil
}

// GetInt gets an integer from a config.
func ( c config ) GetInt( name string ) ( int, error ) {
	item, ok := c[name]
	if !ok {
		return 0, fmt.Errorf( "Integer '%s' not found", name )
	}
	result, err := cast.ToIntE( item )
	if err != nil {
		return 0, fmt.Errorf( "Item '%s' is not an integer", name );
	}

	return result, nil
}

// GetIntOrDefault gets an integer or a default value from a config.
func ( c config ) GetIntOrDefault( name string, dflt int ) ( int, error ) {
	item, ok := c[name]
	if !ok {
		return dflt, nil
	}
	result, err := cast.ToIntE( item )
	if err != nil {
		return 0, fmt.Errorf( "Item '%s' is not an integer", name );
	}

	return result, nil
}

// GetUInt8OrDefault gets an unsigned 8-bit integer
// or a default value from a config.
func ( c config ) GetUInt8OrDefault( name string, dflt uint8 ) ( uint8, error ) {
	item, ok := c[name]
	if !ok {
		return dflt, nil
	}
	result, err := cast.ToUint8E( item )
	if err != nil {
		return 0, fmt.Errorf( "Item '%s' is not an unsigned 8-bit integer", name );
	}

	return result, nil
}

// GetUInt16 gets an unsigned 16-bit integer from a config.
func ( c config ) GetUInt16( name string ) ( uint16, error ) {
	item, ok := c[name]
	if !ok {
		return 0, fmt.Errorf( "Unsigned 16-bit integer '%s' not found", name )
	}
	result, err := cast.ToUint16E( item )
	if err != nil {
		return 0, fmt.Errorf( "Item '%s' is not an unsigned 16-bit integer", name )
	}

	return result, nil
}

// GetString gets a string from a config.
func ( c config ) GetString( name string ) ( string, error ) {
	item, ok := c[name]
	if !ok {
		return "<error>", fmt.Errorf( "String '%s' not found", name )
	}
	result, ok := item.( string )
	if !ok {
		return "<error>", fmt.Errorf( "Item '%s' is not a string", name )
	}

	return result, nil
}

// GetDurationOrDefault gets a duration from a config,
// or the specified default value if it is not found.
func ( c config ) GetDurationOrDefault( name string, dflt time.Duration ) ( time.Duration, error ) {
	item, ok := c[name]
	if !ok {
		return dflt, nil
	}
	durationString, ok := item.( string )
	if !ok {
		return 0, errors.New( "Duration must be a string" )
	}
	result, err := time.ParseDuration( durationString )
	if err != nil {
		return 0, fmt.Errorf( "Duration '%s' string '%s' invalid: %v", name, durationString, err )
	}

	return result, nil
}
