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
	"log"
	"net/http"
	"time"
)

const(
	dHttpTimeout = 10 * time.Second
	dMaxHeaderBytes = 1024
	dMaxTries = 5
	dTryDuration = time.Minute
	kAddresses = "listenAddresses"
	kHttpd = "httpd"
	kLocations = "locations"
	kPath = "path"
	kHttpTimeout = "timeout"
	kValues = "values"
)

// handler for HTTP requests
type handler struct {
	cc *commandConfig
	values config
}

// ServeHTTP processes a request
func ( h handler ) ServeHTTP( w http.ResponseWriter, r *http.Request ) {
	// FIXME
}

// setHandler sets the http handler for one location
func setHandler( locConf config, cc *commandConfig ) error {
	path, err := locConf.GetString( kPath )
	if err != nil {
		return fmt.Errorf( "Unable to extract path: %v", err )
	}
	values, err := locConf.GetSubConfig( kValues )
	if err != nil {
		return fmt.Errorf( "Unable to extract values: %v", err )
	}
	http.Handle( path, handler{ cc, values } )

	return nil
}

// setHandlers sets the http handlers
func setHandlers( httpdConf config, sched *scheduler ) error {
	// Get locations config
	locsConf, err := httpdConf.GetSubConfig( kLocations )
	if err != nil {
		return fmt.Errorf( "Unable to obtain locations configuration: %v", err )
	}

	// Set handler for each location
	for key, cc := range sched.commandMap {
		locConf, err := locsConf.GetSubConfig( key )
		if err != nil {
			return fmt.Errorf( "Unable to find location for command '%v': %v", key, err )
		}
		if err = setHandler( locConf, cc ); err != nil {
			return fmt.Errorf( "Unable to set handler for command '%v': %v", key, err )
		}
	}

	return nil
}

// runServer starts one HTTP server
func runServer( addr string, timeout time.Duration, errchan chan<- error ) {
	server := http.Server{
		Addr: addr,
		ReadTimeout: timeout,
		ReadHeaderTimeout: timeout,
		WriteTimeout: timeout,
		IdleTimeout: timeout,
		MaxHeaderBytes: dMaxHeaderBytes,
	}
	relevantTime := time.Now()
	relevantFailures := 0
	for relevantFailures < dMaxTries {
		relevantFailures++
		log.Printf( "HTTP server '%s' error %d of %d: %v", addr, relevantFailures, dMaxTries, server.ListenAndServe() )
		now := time.Now()
		if now.Sub( relevantTime ) > dTryDuration {
			log.Printf( "HTTP: resetting failure counter for server '%s'", addr )
			relevantTime = now
			relevantFailures = 0
		}
		time.Sleep( time.Second )
	}

	errchan <- fmt.Errorf( "Server '%s' had too many failures in too little time", addr )
}

// serveHttp starts all HTTP server(s)
func serveHttp( httpConf config ) error {
	addrList, err := httpConf.GetList( kAddresses )
	if err != nil {
		return fmt.Errorf( "Unable to obtain HTTP server address list: %v", err )
	}
	timeout, err := httpConf.GetDurationOrDefault( kHttpTimeout, dHttpTimeout )
	if err != nil {
		return fmt.Errorf( "Unable to obtain HTTP timeout: %v", err )
	}
	errchan := make( chan error )
	for _, item := range addrList {
		addr, ok := item.( string )
		if !ok {
			return fmt.Errorf( "HTTP server address value is not a string: %v", item )
		}
		go runServer( addr, timeout, errchan )
	}

	return <-errchan
}

func runHttpd( conf config, sched *scheduler ) error {
	// Get config
	httpdConf, err := conf.GetSubConfig( kHttpd )
	if err != nil {
		return fmt.Errorf( "Unable to obtain httpd configuration: %v", err )
	}

	// Install all the handlers
	if err = setHandlers( httpdConf, sched ); err != nil {
		return fmt.Errorf( "Unable to set httpd handlers: %v", err )
	}

	// Run the server
	return serveHttp( httpdConf )
}
