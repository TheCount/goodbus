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
	"github.com/TheCount/goodbus/mbsched"
	"log"
	"time"
)

// configuration keys
const (
	kAddress = "address"
	kBaudRate = "baudrate"
	kBufferSize = "buffersize"
	kCommands = "commands"
	kDataBits = "databits"
	kIdle = "onlyonidle"
	kMaxWait = "maxwait"
	kMinWait = "minwait"
	kParity = "parity"
	kQuantity = "quantity"
	kRepeat = "repeat"
	kScheduler = "scheduler"
	kSlaveId = "slaveid"
	kStopBits = "stopbits"
	kTimeout = "timeout"
	kType = "type"
)

// configuration values
const (
	vErrorBacklog = 5
	vModbusAscii = "ModbusASCII"
	vModbusRTU = "ModbusRTU"
	vModbusTCP = "ModbusTCP"
	vReadHoldingRegisters = "readHoldingRegisters"
	vReadInputRegisters = "readInputRegisters"
	vSchedulerTimeout = 5 * time.Second
	vSchedulerBufsize = 5
	vWriteSingleRegister = "writeSingleRegister"
	vWriteMultipleRegisters = "writeMultipleRegisters"
)

type scheduler struct {
	mbsched.Scheduler
}

// getAddrTimeoutBufsizeConf gets configuration common to
// all modbus types.
func getAddrTimeoutBufsizeConf( conf config ) ( string, time.Duration, int, error ) {
	addr, err := conf.GetString( kAddress )
	if err != nil {
		return "<error>", 0, 0, fmt.Errorf( "Unable to read modbus address: %v", err )
	}
	timeout, err := conf.GetDurationOrDefault( kTimeout, vSchedulerTimeout )
	if err != nil {
		return "<error>", 0, 0, fmt.Errorf( "Unable to read scheduler timeout: %v", err )
	}
	bufsize, err := conf.GetIntOrDefault( kBufferSize, vSchedulerBufsize )
	if err != nil {
		return "<error>", 0, 0, fmt.Errorf( "Unable to read scheduler buffer size: %v", err )
	}

	return addr, timeout, bufsize, nil
}

// getSerialConf gets configuration for the
// serial modbus types.
func getSerialConf( conf config ) ( int, int, string, int, error ) {
	baudRate, err := conf.GetInt( kBaudRate )
	if err != nil {
		return 0, 0, "<error>", 0, fmt.Errorf( "Unable to read baud rate: %v", err )
	}
	dataBits, err := conf.GetInt( kDataBits )
	if err != nil {
		return 0, 0, "<error>", 0, fmt.Errorf( "Unable to read data bits: %v", err )
	}
	parity, err := conf.GetString( kParity )
	if err != nil {
		return 0, 0, "<error>", 0, fmt.Errorf( "Unable to read parity: %v", err )
	}
	stopBits, err := conf.GetInt( kStopBits )
	if err != nil {
		return 0, 0, "<error>", 0, fmt.Errorf( "Unable to read stop bits: %v", err )
	}

	return baudRate, dataBits, parity, stopBits, nil
}

// watchSchedulerErrors logs scheduler errors
// and exits the program if too many errors occur in too short a time.
func watchSchedulerErrors( errchan <-chan error ) {
	const timeout = 5 * time.Minute
	const maxErrCount = 5
	lastCountReset := time.Now()
	errCount := 0
	for err := range errchan {
		log.Printf( "Scheduler error: %v", err )
		now := time.Now()
		if now.Sub( lastCountReset ) > timeout {
			errCount = 1
			lastCountReset = now
		} else {
			errCount++
		}
		if errCount > maxErrCount {
			log.Fatal( "Too many scheduler errors in too short a time" )
		}
	}
}

// startEmptyScheduler starts an empty scheduler
// according to a configuration.
func startEmptyScheduler( conf config ) ( *scheduler, error ) {
	// get scheduler type
	schedType, err := conf.GetString( kType )
	if err != nil {
		return nil, fmt.Errorf( "Scheduler type not found: %v", kType )
	}

	// configure scheduler according to type
	var result *scheduler
	addr, timeout, bufsize, err := getAddrTimeoutBufsizeConf( conf )
	if err != nil {
		return nil, err
	}
	baudRate, dataBits, parity, stopBits, serialErr := getSerialConf( conf )
	switch ( schedType ) {
	case vModbusAscii:
		if serialErr != nil {
			return nil, serialErr
		}
		result.Scheduler = *mbsched.NewModbusAsciiScheduler( bufsize, addr, baudRate, dataBits, parity, stopBits, timeout )
	case vModbusRTU:
		if serialErr != nil {
			return nil, serialErr
		}
		result.Scheduler = *mbsched.NewModbusRtuScheduler( bufsize, addr, baudRate, dataBits, parity, stopBits, timeout )
	case vModbusTCP:
		result.Scheduler = *mbsched.NewModbusTcpScheduler( bufsize, addr, timeout )
	default:
		return nil, fmt.Errorf( "Unknown scheduler type: %s", schedType )
	}

	// Start scheduler
	errChan, err := result.Start( vErrorBacklog )
	if err != nil {
		return nil, fmt.Errorf( "Error starting scheduler: %v", err )
	}
	go watchSchedulerErrors( errChan )

	return result, nil
}

// fillCommand fills in one configured command
// for the scheduler.
func ( s *scheduler ) fillCommand( name string, conf config ) error {
	schedule, err := getScheduleConf( conf )
	if err != nil {
		return fmt.Errorf( "Unable to get schedule configuration for command '%s': %v", name, err )
	}
	slaveId, addr, err := getCommandAddress( conf )
	if err != nil {
		return fmt.Errorf( "Unable to get address information for command '%s': %v", name, err )
	}
	typeString, err := conf.GetString( kType )
	if err != nil {
		return fmt.Errorf( "Unable to get type of command '%s': %v", name, err )
	}
	// FIXME
}

// fillCommands fills in the configured commands
// for the scheduler.
func ( s *scheduler ) fillCommands( conf config ) error {
	for name, _ := range conf {
		commandConf, err := conf.GetSubConfig( name )
		if err != nil {
			return fmt.Errorf( "Command configuration error: %v", err )
		}
		if err = s.fillCommand( name, commandConf ); err != nil {
			return err
		}
	}

	return nil
}

// startScheduler starts a scheduler
// according to a configuration.
func startScheduler( conf config ) ( *scheduler, error ) {
	// get configuration
	schedConf, err := conf.GetSubConfig( kScheduler )
	if err != nil {
		return nil, fmt.Errorf( "Unable to get scheduler configuration: %v", err )
	}

	// Start empty scheduler and add commands
	result, err := startEmptyScheduler( schedConf )
	if err != nil {
		return nil, fmt.Errorf( "Unable to start empty scheduler: %v", err )
	}
	commandConf, err := conf.GetSubConfig( kCommands )
	if err != nil {
		return nil, fmt.Errorf( "Unable to get commands configuration in scheduler configuration '%s': %v", kScheduler, err )
	}
	if err = result.fillCommands( commandConf ); err != nil {
		return nil, fmt.Errorf( "Unable to fill scheduler with commands: %v", err )
	}

	return result, nil
}
