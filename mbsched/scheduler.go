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

package mbsched

import(
	"fmt"
	"github.com/TheCount/goodbus/sched"
	"time"
)

// Scheduler schedules modbus commands for one bus
// according to specific schedules.
type Scheduler struct {
	sched.Scheduler

	// handler is the modbus handler
	handler handler
}

// NewModbusAsciiScheduler creates a new modbus ASCII scheduler.
func NewModbusAsciiScheduler( scheduleBufferSize int, addr string, baudRate int, dataBits int, parity string, stopBits int, timeout time.Duration ) *Scheduler {
	return &Scheduler{
		Scheduler: *sched.NewScheduler( scheduleBufferSize ),
		handler: newAsciiHandler( addr, baudRate, dataBits, parity, stopBits, timeout ),
	}
}

// NewModbusRtuScheduler creates a new modbus RTU scheduler.
func NewModbusRtuScheduler( scheduleBufferSize int, addr string, baudRate int, dataBits int, parity string, stopBits int, timeout time.Duration ) *Scheduler {
	return &Scheduler{
		Scheduler: *sched.NewScheduler( scheduleBufferSize ),
		handler: newRtuHandler( addr, baudRate, dataBits, parity, stopBits, timeout ),
	}
}

// NewModbusTcpScheduler creates a new modbus TCP scheduler.
func NewModbusTcpScheduler( scheduleBufferSize int, addr string, timeout time.Duration ) *Scheduler {
	return &Scheduler{
		Scheduler: *sched.NewScheduler( scheduleBufferSize ),
		handler: newTcpHandler( addr, timeout ),
	}
}

// AddReadInputRegisters adds a modbus read input registers command
// to a running scheduler.
// On success, it returns a channel with buffer size bufSize
// yielding the read data.
func ( s *Scheduler ) AddReadInputRegisters( name string, schedule sched.Schedule, bufSize int, slaveId byte, address uint16, quantity uint16 ) ( <-chan []byte, error ) {
	command, resultChan := newReadInputRegisters( bufSize, s.handler, slaveId, address, quantity )
	err := s.Scheduler.Add( name, command, schedule )

	return resultChan, err
}

// Start starts the scheduler.
// A channel reporting scheduler errors is returned.
// The buffer size of this channel is given by error backlog.
// On success, the second return value is nil.
// Otherwise, it is an appropriate error message.
func ( s *Scheduler ) Start( errorBacklog int ) ( <-chan error, error ) {
	err := s.handler.Connect()
	if err != nil {
		return nil, err
	}

	result, err := s.Scheduler.Start( errorBacklog )
	if err != nil {
		err2 := s.handler.Close()
		if err2 != nil {
			err = fmt.Errorf( "Error '%v' starting scheduler followed by error '%v' closing handler", err, err2 )
		}
	}

	return result, err
}

// WaitStop waits for the scheduler to stop
// after a call to SignalStop.
// Without a call to SignalStop,
// WaitStop will wait forever.
func ( s *Scheduler ) WaitStop() {
	s.Scheduler.WaitStop()
	s.handler.Close()
}

// Stop stops the scheduler.
// It combines SignalStop and WaitStop into one method.
func ( s *Scheduler ) Stop() {
	s.SignalStop()
	s.WaitStop()
}

// AddReadHoldingRegisters adds a modbus read holding registers command
// to a running scheduler.
// On success, it returns a channel with buffer size bufSize
// yielding the read data.
func ( s *Scheduler ) AddReadHoldingRegisters( name string, schedule sched.Schedule, bufSize int, slaveId byte, address uint16, quantity uint16 ) ( <-chan []byte, error ) {
	command, resultChan := newReadHoldingRegisters( bufSize, s.handler, slaveId, address, quantity )
	err := s.Scheduler.Add( name, command, schedule )

	return resultChan, err
}

// AddWriteSingleRegister adds a modbus write single register command
// to a running scheduler.
// On success, it returns a channel with buffer size bufSize
// yielding the value written.
func ( s *Scheduler ) AddWriteSingleRegister( name string, schedule sched.Schedule, bufSize int, slaveId byte, address uint16, value uint16 ) ( <-chan []byte, error ) {
	command, resultChan := newWriteSingleRegister( bufSize, s.handler, slaveId, address, value )
	err := s.Scheduler.Add( name, command, schedule )

	return resultChan, err
}

// AddWriteMultipleRegisters adds a modbus write multiple registers command
// to a running scheduler.
// On success, it returns a channel with buffer size bufSize
// yielding the quantity of values written.
func ( s *Scheduler ) AddWriteMultipleRegisters( name string, schedule sched.Schedule, bufSize int, slaveId byte, address uint16, quantity uint16, values []byte ) ( <-chan []byte, error ) {
	command, resultChan := newWriteMultipleRegisters( bufSize, s.handler, slaveId, address, quantity, values )
	err := s.Scheduler.Add( name, command, schedule )

	return resultChan, err
}
