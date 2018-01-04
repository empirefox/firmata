package firmata

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type Firmata struct {
	ProtocolVersion *Version
	FirmwareVersion *Version
	FirmwareName    string

	connecting  atomic.Value
	connected   atomic.Value
	connectedCh chan struct{}      // init from Dial
	closed      chan struct{}      // init from Dial
	c           io.ReadWriteCloser // init from Dial

	reportVersionOnce      sync.Once
	reportFirmwareOnce     sync.Once
	capabilityQueryOnce    sync.Once
	analogMappingQueryOnce sync.Once

	TotalPorts int
	pins       []*Pin
	analogPins []*Pin

	OnPinState       func(pin *Pin)
	OnConnected      func()
	OnAnalogMessage  func(pin *Pin)
	OnDigitalMessage func(pins []*Pin)
	OnI2cReply       func(reply *I2cReply)
	OnStringData     func(data []byte)
	OnSysexResponse  func(buf []byte)
	OnError          func(err error)
}

func NewFirmata() *Firmata {
	f := new(Firmata)
	f.connecting.Store(false)
	f.connected.Store(false)
	return f
}

// Close disconnects the Firmata
func (b *Firmata) Close() (err error) {
	b.connected.Store(false)
	if err = b.c.Close(); err != nil {
		return err
	}
	if b.closed != nil {
		<-b.closed
		b.closed = nil
	}
	return nil
}

// Connecting returns true when the Firmata is connecting
func (b *Firmata) Connecting() bool {
	return b.connecting.Load().(bool)
}

// Connected returns the current connection state of the Firmata
func (b *Firmata) Connected() bool {
	return b.connected.Load().(bool)
}

// Pins returns all available pins
func (b *Firmata) Pins() []*Pin {
	return b.pins
}

// Pins returns all available pins
func (b *Firmata) AnalogPins() []*Pin {
	return b.analogPins
}

// Reset sends the SystemReset sysex code.
func (b *Firmata) Reset() error {
	return b.write([]byte{SYSTEM_RESET})
}

// SetPinMode sets the pin to mode.
func (b *Firmata) SetPinMode(pin int, mode int) error {
	if pin >= len(b.pins) {
		return fmt.Errorf("SetPinMode pin(%d) not found", pin)
	}
	b.pins[pin].Mode = mode
	return b.write([]byte{SET_PIN_MODE, byte(pin), byte(mode)})
}

// SetDigitalPinValue sets the pin to value(0/1).
func (b *Firmata) SetDigitalPinValue(pin int, value int) error {
	if pin >= len(b.pins) {
		return fmt.Errorf("SetDigitalPinValue pin(%d) not found", pin)
	}
	b.pins[pin].Value = value
	return b.write([]byte{SET_DIGITAL_PIN_VALUE, byte(pin), byte(value)})
}

// SetDigitalPinHigh sets the pin to 1.
func (b *Firmata) SetDigitalPinHigh(pin int) error {
	return b.SetDigitalPinValue(pin, 1)
}

// SetDigitalPinLow sets the pin to 0.
func (b *Firmata) SetDigitalPinLow(pin int) error {
	return b.SetDigitalPinValue(pin, 0)
}

// DigitalWrite writes value to pin.
func (b *Firmata) DigitalWrite(port int, values *[8]int) error {
	portValue := byte(0)
	for i, v := range values {
		pid := 8*port + i
		if pid >= len(b.pins) {
			if i == 0 {
				return fmt.Errorf("DigitalWrite port(%d) out of range", port)
			}
			break
		}
		b.pins[pid].Value = v
		if v != 0 {
			portValue = portValue | (1 << uint(i))
		}
	}
	return b.write([]byte{DIGITAL_MESSAGE | byte(port), portValue & 0x7F, (portValue >> 7) & 0x7F})
}

// ServoConfig sets the min and max pulse width for servo PWM range
func (b *Firmata) ServoConfig(pin int, max int, min int) error {
	c := []byte{
		SERVO_CONFIG,
		byte(pin),
		byte(min & 0x7F),
		byte((min >> 7) & 0x7F),
		byte(max & 0x7F),
		byte((max >> 7) & 0x7F),
	}
	return b.WriteSysex(c)
}

// AnalogWrite writes value to pin.
func (b *Firmata) AnalogWrite(pin int, value int) error {
	if pin >= len(b.pins) {
		return fmt.Errorf("AnalogWrite pin(%d) not found", pin)
	}
	b.pins[pin].Value = value
	return b.write([]byte{ANALOG_MESSAGE | byte(pin), byte(value & 0x7F), byte((value >> 7) & 0x7F)})
}

// PinStateQuery sends a PinStateQuery for pin.
func (b *Firmata) PinStateQuery(pin int) error {
	return b.WriteSysex([]byte{PIN_STATE_QUERY, byte(pin)})
}

func (b *Firmata) reportVersion() error {
	return b.write([]byte{REPORT_VERSION})
}

func (b *Firmata) reportFirmware() error {
	return b.WriteSysex([]byte{REPORT_FIRMWARE})
}

func (b *Firmata) capabilitiesQuery() error {
	return b.WriteSysex([]byte{CAPABILITY_QUERY})
}

func (b *Firmata) analogMappingQuery() error {
	return b.WriteSysex([]byte{ANALOG_MAPPING_QUERY})
}

// ReportDigital enables or disables digital reporting for pin, a 0/1 value enables reporting
func (b *Firmata) ReportDigital(port int, value int) error {
	return b.write([]byte{REPORT_DIGITAL | byte(port), byte(value)})
}

// ReportAnalog enables or disables analog reporting for pin, a 0/1 value enables reporting
func (b *Firmata) ReportAnalog(pin int, value int) error {
	return b.write([]byte{REPORT_ANALOG | byte(pin), byte(value)})
}

// I2cRead reads numBytes from address once.
func (b *Firmata) I2cRead(address int, numBytes int) error {
	return b.WriteSysex([]byte{I2C_REQUEST, byte(address), byte(PIN_MODE_OUTPUT << 3),
		byte(numBytes) & 0x7F, (byte(numBytes) >> 7) & 0x7F})
}

// I2cWrite writes data to address.
func (b *Firmata) I2cWrite(address int, data []byte) error {
	req := []byte{I2C_REQUEST, byte(address), byte(PIN_MODE_INPUT << 3)}
	for _, val := range data {
		req = append(req, byte(val&0x7F))
		req = append(req, byte((val>>7)&0x7F))
	}
	return b.WriteSysex(req)
}

// I2cConfig configures the delay in which a register can be read from after it
// has been written to.
func (b *Firmata) I2cConfig(delay int) error {
	return b.WriteSysex([]byte{I2C_CONFIG, byte(delay & 0xFF), byte((delay >> 8) & 0xFF)})
}

// WriteSysex writes an arbitrary Sysex command to the microcontroller.
func (b *Firmata) WriteSysex(data []byte) (err error) {
	return b.write(append([]byte{START_SYSEX}, append(data, END_SYSEX)...))
}

func (b *Firmata) write(data []byte) (err error) {
	_, err = b.c.Write(data[:])
	if err != nil {
		b.connected.Store(false)
	}
	return
}

func (b *Firmata) read(n int) (buf []byte, err error) {
	buf = make([]byte, n)
	_, err = io.ReadFull(b.c, buf)
	if err != nil {
		b.connected.Store(false)
	}
	return
}
