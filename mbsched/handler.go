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
	"github.com/goburrow/modbus"
	"time"
)

// handler is an interface wrapping the various handler concepts
// of the modbus package.
type handler interface {
	modbus.ClientHandler

	// MakeClient sets the slave ID for the next command
	// executed on this handler and returns an appropriate client.
	// The returned client is only good until the next call to
	// MakeClient() as the modbus is a shared resource.
	MakeClient( slaveId byte ) modbus.Client

	// Connect connects to the modbus.
	Connect() error

	// Close closes the connection to the modbus.
	Close() error
}

// asciiHandler is a wrapper around modbus.ASCIIClientHandler.
type asciiHandler struct {
	modbus.ASCIIClientHandler
}

// MakeClient returns a client for the asciiHandler.
func ( h *asciiHandler ) MakeClient( slaveId byte ) modbus.Client {
	h.SlaveId = slaveId
	return modbus.NewClient( h )
}

// newAsciiHandler creates a new modbus ASCII handler.
func newAsciiHandler( addr string, baudRate int, dataBits int, parity string, stopBits int, timeout time.Duration ) *asciiHandler {
	result := &asciiHandler{
		ASCIIClientHandler: *modbus.NewASCIIClientHandler( addr ),
	}
	result.BaudRate = baudRate
	result.DataBits = dataBits
	result.Parity = parity
	result.StopBits = stopBits
	result.Timeout = timeout

	return result
}

// rtuHandler is a wrapper around modbus.RTUClientHandler.
type rtuHandler struct {
	modbus.RTUClientHandler
}

// MakeClient resturns a client for the rtuHandler.
func ( h *rtuHandler ) MakeClient( slaveId byte ) modbus.Client {
	h.SlaveId = slaveId
	return modbus.NewClient( h )
}

// newRtuHandler creates a new modbus RTU handler.
func newRtuHandler( addr string, baudRate int, dataBits int, parity string, stopBits int, timeout time.Duration ) *rtuHandler {
	result := &rtuHandler{
		RTUClientHandler: *modbus.NewRTUClientHandler( addr ),
	}
	result.BaudRate = baudRate
	result.DataBits = dataBits
	result.Parity = parity
	result.StopBits = stopBits
	result.Timeout = timeout

	return result
}

// tcpHandler is a wrapper around modbus.TCPClientHandler.
type tcpHandler struct {
	modbus.TCPClientHandler
}

// MakeClient returns a client for the tcpHandler.
func ( h *tcpHandler ) MakeClient( slaveId byte ) modbus.Client {
	h.SlaveId = slaveId
	return modbus.NewClient( h )
}

// newTcpHandler creates a new modbus TCP handler.
func newTcpHandler( addr string, timeout time.Duration ) *tcpHandler {
	result := &tcpHandler{
		TCPClientHandler: *modbus.NewTCPClientHandler( addr ),
	}
	result.Timeout = timeout

	return result
}
