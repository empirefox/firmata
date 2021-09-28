package firmata

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type Firmata struct {
	reader *ReadFramer
	writer *WriteFramer
	closer io.Closer

	connectedOnce sync.Once
	doneOnce      sync.Once
	handshakeOnce sync.Once
	handshakeOK   chan struct{}
	doneServing   chan struct{}        // closed when Firmata.serve ends
	readFrameCh   chan readFrameResult // written by serverConn.readFrames
	loopCh        chan func()

	// ClosedError if non-nil it is the reason why serve loop stopped.
	ClosedError error

	// VersionInfo Only valid after every OnConnected called, but invalid after reset.
	VersionInfo

	TotalPorts_l byte
	Pins_l       []*Pin
	AnalogPins_l []*Pin

	OnConnected      func()
	OnAnalogMessage  func(pin *Pin)
	OnDigitalMessage func(pins []*Pin)
	OnPinState       func(pin *Pin)
	OnI2cReply       func(reply *I2cReply)
	OnStringData     func(data []byte)
	OnSysexResponse  func(buf []byte)
}

type Config struct {
	OnConnected      func()
	OnAnalogMessage  func(pin *Pin)
	OnDigitalMessage func(pins []*Pin)
	OnPinState       func(pin *Pin)
	OnI2cReply       func(reply *I2cReply)
	OnStringData     func(data []byte)
	OnSysexResponse  func(buf []byte)
}

func Connect(ctx context.Context, c io.ReadWriteCloser, config *Config) (*Firmata, error) {
	f := NewFirmata(c)
	if config != nil {
		f.OnConnected = config.OnConnected
		f.OnAnalogMessage = config.OnAnalogMessage
		f.OnDigitalMessage = config.OnDigitalMessage
		f.OnPinState = config.OnPinState
		f.OnI2cReply = config.OnI2cReply
		f.OnSysexResponse = config.OnSysexResponse
	}

	if config != nil && config.OnStringData != nil {
		f.OnStringData = func(data []byte) {
			config.OnStringData(From14bits(data))
		}
	}

	return f, f.Handshake(ctx)
}

func NewFirmata(c io.ReadWriteCloser) *Firmata {
	return &Firmata{
		reader: NewReadFramer(c),
		writer: NewWriteFramer(c),
		closer: c,

		handshakeOK: make(chan struct{}),
		doneServing: make(chan struct{}),
		readFrameCh: make(chan readFrameResult),
		loopCh:      make(chan func(), 32),
	}
}

func (f *Firmata) Handshake(ctx context.Context) (err error) {
	go f.serve()
	go f.readFrames()

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	done := make(chan error)
	f.Loop(func() { done <- f.reportVersion() })
	err = <-done
	if err != nil {
		return
	}

	select {
	case <-f.handshakeOK:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}

func (f *Firmata) AddProxyPipe_l(c io.ReadWriteCloser) {
	f.reader.pipes = append(f.reader.pipes, c)
	go func() {
		var err error
		fr := proxyReadFramer{
			f:    f,
			r:    c,
			w:    c,
			done: make(chan error, 1),
		}
		for {
			err = fr.readSendNextFrame()
			if err != nil {
				c.Close()
				return
			}
		}
	}()
}

func (f *Firmata) SnapshotPin(pin byte) (p *Pin, ok bool) {
	done := make(chan struct{})
	f.Loop(func() {
		ok = pin < byte(len(f.Pins_l))
		if ok {
			p = f.Pins_l[pin]
		}
		close(done)
	})
	<-done
	return
}

func (f *Firmata) SnapshotAnalogPin(pin byte) (p *Pin, ok bool) {
	done := make(chan struct{})
	f.Loop(func() {
		ok = pin < byte(len(f.AnalogPins_l))
		if ok {
			p = f.AnalogPins_l[pin]
		}
		close(done)
	})
	<-done
	return
}

// Pins returns all available Pins
func (f *Firmata) SnapshotPins() []*Pin {
	done := make(chan []*Pin)
	f.Loop(func() { done <- ClonePins(f.Pins_l) })
	return <-done
}

// Pins returns all available AnalogPins
func (f *Firmata) SnapshotAnalogPins() []*Pin {
	done := make(chan []*Pin)
	f.Loop(func() { done <- ClonePins(f.AnalogPins_l) })
	return <-done
}

// Reset sends the SystemReset sysex code.
func (f *Firmata) Reset_l() (err error) {
	err = f.writer.Reset()
	if err != nil {
		return
	}
	err = f.reportVersion()
	if err != nil {
		return
	}
	f.Pins_l = nil
	f.AnalogPins_l = nil
	f.connectedOnce = sync.Once{}
	return nil
}

// SetPinMode sets the pin to mode.
func (f *Firmata) SetPinMode_l(pin byte, mode byte) error {
	if pin >= byte(len(f.Pins_l)) {
		return fmt.Errorf("SetPinMode pin out of index: %d", pin)
	}
	if f.Pins_l[pin].Mode == mode {
		return nil
	}
	err := f.writer.SetPinMode(pin, mode)
	if err == nil {
		f.Pins_l[pin].Mode = mode
	}
	return err
}

// SetDigitalPinValue sets the pin to value(0/1).
func (f *Firmata) SetDigitalPinValue_l(pin byte, value byte) error {
	if pin >= byte(len(f.Pins_l)) {
		return fmt.Errorf("SetDigitalPinValue pin out of index: %d", pin)
	}
	if f.Pins_l[pin].Value == int(value) {
		return nil
	}
	err := f.writer.SetDigitalPinValue(pin, value)
	if err == nil {
		f.Pins_l[pin].Value = int(value)
	}
	return err
}

// SetDigitalPinHigh sets the pin to 1.
func (f *Firmata) SetDigitalPinHigh_l(pin byte) error {
	return f.SetDigitalPinValue_l(pin, 1)
}

// SetDigitalPinLow sets the pin to 0.
func (f *Firmata) SetDigitalPinLow_l(pin byte) error {
	return f.SetDigitalPinValue_l(pin, 0)
}

// DigitalWrite writes value to pin.
func (f *Firmata) DigitalWrite_l(port byte, values *[8]byte) error {
	if port > f.TotalPorts_l {
		return fmt.Errorf("DigitalWrite port out of index: %d", port)
	}
	err := f.writer.DigitalWrite(port, values)
	if err == nil {
		f.setDigitalPortValues(port, values)
	}
	return err
}

// ServoConfig sets the min and max pulse width for servo PWM range
func (f *Firmata) ServoConfig_l(pin byte, max int, min int) error {
	// TODO save to Firmata
	return f.writer.ServoConfig(pin, max, min)
}

// AnalogWrite writes value to pin.
func (f *Firmata) AnalogWrite_l(pin byte, value int) (err error) {
	if pin >= byte(len(f.Pins_l)) {
		return fmt.Errorf("AnalogWrite pin(%d) not found", pin)
	}

	if pin > 15 || value >= 0x4000 {
		err = f.writer.ExtendedAnalogWrite(pin, value)
	} else {
		err = f.writer.AnalogWrite(pin, value)
	}
	if err == nil {
		f.Pins_l[pin].Value = value
	}
	return err
}

// PinStateQuery sends a PinStateQuery for pin.
func (f *Firmata) PinStateQuery_l(pin byte) error {
	return f.writer.PinStateQuery(pin)
}

// PinState_l sends a PinStateQuery for pin.
func (f *Firmata) PinState_l(pin byte) (int, error) {
	if pin >= byte(len(f.Pins_l)) {
		return 0, fmt.Errorf("PinStateQuery pin(%d) not found", pin)
	}
	return f.Pins_l[pin].State, nil
}

func (f *Firmata) reportVersion() error {
	if f.ProtocolVersion == nil {
		return f.writer.ReportVersion()
	}
	if f.FirmwareVersion == nil {
		return f.writer.ReportFirmware()
	}
	if f.Pins_l == nil {
		return f.writer.CapabilitiesQuery()
	}
	if f.AnalogPins_l == nil {
		return f.writer.AnalogMappingQuery()
	}
	return nil
}

// ReportDigital enables or disables digital reporting for pin, a 0/1 value enables reporting
func (f *Firmata) ReportDigital_l(port byte, value byte) error {
	return f.writer.ReportDigital(port, value)
}

// ReportAnalog enables or disables analog reporting for pin, a 0/1 value enables reporting
func (f *Firmata) ReportAnalog_l(pin byte, value byte) error {
	return f.writer.ReportAnalog(pin, value)
}

// I2cWrite writes data to address.
func (f *Firmata) I2cWrite_l(address int, data []byte) error {
	return f.writer.I2cWrite(address, data)
}

// I2cRead reads numBytes from address once or continuous.
func (f *Firmata) I2cRead_l(address int, autoRestartTransmission bool, continuous bool, numBytes byte) error {
	return f.writer.I2cRead(address, autoRestartTransmission, continuous, numBytes)
}

func (f *Firmata) I2cStopReading_l(address int) error {
	return f.writer.I2cStopReading(address)
}

// I2cConfig configures the delay in which a register can be read from after it
// has been written to.
func (f *Firmata) I2cConfig_l(delay uint) error {
	return f.writer.I2cConfig(delay)
}

// SamplingInterval sets how often analog data and i2c data is reported to the
// client. The default for the arduino implementation is 19ms. This means that
// every 19ms analog data will be reported and any i2c devices with read
// continuous mode will be read.
func (f *Firmata) SamplingInterval_l(ms uint) error {
	return f.writer.SamplingInterval(ms)
}
