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

// command is a generic modbus command.
type command struct {
	// resultChan is the command's result channel.
	resultChan chan<- []byte

	// execFunc is the modbus function to be executed.
	execFunc func() ( []byte, error )
}

// Execute executes the command.
func ( c *command ) Execute() error {
	result, err := c.execFunc()
	if err != nil {
		return err
	}

	c.resultChan <- result

	return nil
}

// Finalize closes the command's result channel.
func ( c *command ) Finalize() {
	close( c.resultChan )
}

// newCommand creates a new command with nil execFunc.
// A channel with a buffer size of bufSize
// yielding the command's results is returned alongside.
// A negative buffer size will cause a panic.
func newCommand( bufSize int ) ( *command, <-chan []byte ) {
	resultChan := make( chan []byte, bufSize )
	return &command{
		resultChan: resultChan,
	}, resultChan
}

// newReadInputRegisters creates a new modbus read input registers command.
// A channel with a buffer size of bufSize
// yielding the command's results is returned alongside.
// A negative buffer size will cause a panic.
func newReadInputRegisters( bufSize int, handler handler, slaveId byte, address uint16, quantity uint16 ) ( *command, <-chan []byte ) {
	command, resultChan := newCommand( bufSize )
	command.execFunc = func() ( []byte, error ) {
		return handler.MakeClient( slaveId ).ReadInputRegisters( address, quantity )
	}

	return command, resultChan
}

// newReadHoldingRegisters creates a new modbus read holding registers command.
// A channel with a buffer size of bufSize
// yielding the command's results is returned alongside.
// A negative buffer size will cause a panic.
func newReadHoldingRegisters( bufSize int, handler handler, slaveId byte, address uint16, quantity uint16 ) ( *command, <-chan []byte ) {
	command, resultChan := newCommand( bufSize )
	command.execFunc = func() ( []byte, error ) {
		return handler.MakeClient( slaveId ).ReadHoldingRegisters( address, quantity )
	}

	return command, resultChan
}

// newWriteSingleRegister creates a new modbus write single register command.
// A channel with a buffer size of bufSize
// yielding the command's results is returned alongside.
// A negative buffer size will cause a panic.
func newWriteSingleRegister( bufSize int, handler handler, slaveId byte, address uint16, value uint16 ) ( *command, <-chan []byte ) {
	command, resultChan := newCommand( bufSize )
	command.execFunc = func() ( []byte, error ) {
		return handler.MakeClient( slaveId ).WriteSingleRegister( address, value )
	}

	return command, resultChan
}

// newWriteMultipleRegisters creates a new modbus write multiple registers command.
// The length of the values slice must be exactly twice the quantity.
// A channel with a buffer size of bufSize
// yielding the command's results is returned alongside.
// A negative buffer size will cause a panic.
func newWriteMultipleRegisters( bufSize int, handler handler, slaveId byte, address uint16, quantity uint16, values []byte ) ( *command, <-chan []byte ) {
	command, resultChan := newCommand( bufSize )
	command.execFunc = func() ( []byte, error ) {
		return handler.MakeClient( slaveId ).WriteMultipleRegisters( address, quantity, values )
	}

	return command, resultChan
}
