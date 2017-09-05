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
	"github.com/TheCount/goodbus/sched"
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
	vDefaultMinWait = 0
	vDefaultMaxWait = time.Second
	vDefaultSlaveId = 255
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

// Constants
const(
	// Upper maximum quantity of registers in modbus commands.
	// Actual maximum allowed quantity may be smaller depending on command.
	MaxModbusQuantity = 255
)

type commandConfig struct {
	scratchpad *Scratchpad

	// Command launcher for one-shot commands.
	// Nil for repeated commands.
	launcher func() error
}

// IsReadCommand returns true if and only if the underlying modbus command
// is one of the read commands.
func ( c *commandConfig ) IsReadCommand() bool {
	// Right now, read command status is equivalent with there not being a launcher.
	// This might change in the future, though.
	return c.launcher == nil
}

type scheduler struct {
	mbsched.Scheduler

	// commandMap maps modbus command names to command configurations
	commandMap map[string]*commandConfig
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
		result = &scheduler{
			Scheduler: *mbsched.NewModbusAsciiScheduler( bufsize, addr, baudRate, dataBits, parity, stopBits, timeout ),
		}
	case vModbusRTU:
		if serialErr != nil {
			return nil, serialErr
		}
		result = &scheduler{
			Scheduler: *mbsched.NewModbusRtuScheduler( bufsize, addr, baudRate, dataBits, parity, stopBits, timeout ),
		}
	case vModbusTCP:
		result = &scheduler{
			Scheduler: *mbsched.NewModbusTcpScheduler( bufsize, addr, timeout ),
		}
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

// getScheduleConf obtains the configuration for a schedule.
func getScheduleConf( conf config ) ( *sched.Schedule, error ) {
	repeat, err := conf.GetBoolOrDefault( kRepeat, false )
	if err != nil {
		return nil, fmt.Errorf( "Unable to read repeat setting: %v", err )
	}
	onlyOnIdle, err := conf.GetBoolOrDefault( kIdle, false )
	if err != nil {
		return nil, fmt.Errorf( "Unable to read idle setting: %v", err )
	}
	minWait, err := conf.GetDurationOrDefault( kMinWait, vDefaultMinWait )
	if err != nil {
		return nil, fmt.Errorf( "Unable to read minimum wait duration: %v", err )
	}
	maxWait, err := conf.GetDurationOrDefault( kMaxWait, vDefaultMaxWait )
	if err != nil {
		return nil, fmt.Errorf( "Unable to read maximum wait duration: %v", err )
	}

	result := &sched.Schedule{
		MinWait: minWait,
		MaxWait: maxWait,
	}
	if repeat {
		result.Flags |= sched.ScheduleRepeat
	}
	if onlyOnIdle {
		result.Flags |= sched.ScheduleIdle
	}

	return result, nil
}

// getCommandAddress obtains the address for a modbus command,
// including the slave ID.
func getCommandAddress( conf config ) ( uint8, uint16, error ) {
	slaveId, err := conf.GetUInt8OrDefault( kSlaveId, vDefaultSlaveId )
	if err != nil {
		return 0, 0, fmt.Errorf( "Unable to read slave ID: %v", err )
	}
	address, err := conf.GetUInt16( kAddress )
	if err != nil {
		return 0, 0, fmt.Errorf( "Unable to read address: %v", err )
	}

	return slaveId, address, nil
}

func watchResultChan( scratchpad *Scratchpad, rChan <-chan []byte ) {
	for result := range rChan {
		scratchpad.Update( result )
	}
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
	quantity, qErr := conf.GetUInt16( kQuantity )
	if ( quantity > MaxModbusQuantity ) {
		return fmt.Errorf( "Register quantity %v out of bounds for command '%s'", quantity, name )
	}
	cc := &commandConfig{
		scratchpad: NewScratchpad( 2 * int( quantity ) ),
		launcher: nil,
	}
	switch ( typeString ) {
	case vReadHoldingRegisters:
		if qErr != nil {
			return fmt.Errorf( "read holding registers: %v", qErr )
		}
		rChan, err := s.AddReadHoldingRegisters( name, *schedule, vSchedulerBufsize, slaveId, addr, quantity )
		if err != nil {
			return fmt.Errorf( "Unable to create read holding registers schedule: %v", err )
		}
		go watchResultChan( cc.scratchpad, rChan )
	case vReadInputRegisters:
		if qErr != nil {
			return fmt.Errorf( "read input registers: %v", qErr )
		}
		rChan, err := s.AddReadInputRegisters( name, *schedule, vSchedulerBufsize, slaveId, addr, quantity )
		if err != nil {
			return fmt.Errorf( "Unable to create read input registers schedule: %v", err )
		}
		go watchResultChan( cc.scratchpad, rChan )
	case vWriteSingleRegister:
		if schedule.Flags & sched.ScheduleRepeat != 0 {
			return fmt.Errorf( "Repeat '%s' not supported for write single register", name )
		}
		cc.launcher = func() error {
			_, data := cc.scratchpad.Get()
			if data == nil {
				log.Panicf( "Internal error: data for '%s' not set", name )
			}
			if len( data ) < 2 {
				log.Panicf( "Internal error: data for '%s' too short", name )
			}
			rChan, err := s.AddWriteSingleRegister( name, *schedule, vSchedulerBufsize, slaveId, addr, ( uint16( data[0] ) << 8 ) | uint16( data[1] ) )
			if err != nil {
				return fmt.Errorf( "Unable to add write single register command: %v", err )
			}
			result, ok := <-rChan
			if !ok {
				return fmt.Errorf( "No data from result channel for '%s'", name )
			}
			cc.scratchpad.Update( result )
			_, ok = <-rChan
			if ok {
				return fmt.Errorf( "Result channel did not close for '%s'", name )
			}

			return nil
		}
	case vWriteMultipleRegisters:
		if qErr != nil {
			return fmt.Errorf( "write multiple registers: %v", qErr )
		}
		if schedule.Flags & sched.ScheduleRepeat != 0 {
			return fmt.Errorf( "Repeat '%s' not supported for write multiple registers", name )
		}
		cc.launcher = func() error {
			_, data := cc.scratchpad.Get()
			if data == nil {
				log.Panicf( "Internal error: data for '%s' not set", name )
			}
			if quantity > 256 || len( data ) < 2 * int( quantity ) {
				log.Panicf( "Internal error: data for '%s' too short", name )
			}
			rChan, err := s.AddWriteMultipleRegisters( name, *schedule, vSchedulerBufsize, slaveId, addr, quantity, data )
			if err != nil {
				return fmt.Errorf( "Unable to add write multiple registers command: %v", err )
			}
			result, ok := <-rChan
			if !ok {
				return fmt.Errorf( "No data from result channel for '%s'", name )
			}
			cc.scratchpad.Update( result )
			_, ok = <-rChan
			if ok {
				return fmt.Errorf( "Result channel did not close for '%s'", name )
			}

			return nil
		}
	default:
		return fmt.Errorf( "Modbus command type '%s' not supported", typeString )
	}
	s.commandMap[name] = cc

	return nil
}

// fillCommands fills in the configured commands
// for the scheduler.
func ( s *scheduler ) fillCommands( conf config ) error {
	s.commandMap = make( map[string]*commandConfig )
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
