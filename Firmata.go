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

	DxByName map[PinName]byte

	TotalPorts      byte
	TotalPins       byte
	TotalAnalogPins byte
	Pins            []*Pin
	AnalogPins      []*Pin

	// TODO report?
	PortConfigInputs [16]byte

	OnConnected      func(f *Firmata)
	OnAnalogMessage  func(f *Firmata, pin *Pin)
	OnDigitalMessage func(f *Firmata, port byte, pins byte)
	OnPinState       func(f *Firmata, pin *Pin)
	OnI2cReply       func(f *Firmata, reply *I2cReply)
	OnStringData     func(f *Firmata, data []byte)
	OnSysexResponse  func(f *Firmata, buf []byte)
}

type Config struct {
	OnConnected      func(f *Firmata)
	OnAnalogMessage  func(f *Firmata, pin *Pin)
	OnDigitalMessage func(f *Firmata, port byte, pins byte)
	OnPinState       func(f *Firmata, pin *Pin)
	OnI2cReply       func(f *Firmata, reply *I2cReply)
	OnStringData     func(f *Firmata, data []byte)
	OnSysexResponse  func(f *Firmata, buf []byte)
}

func Connect(ctx context.Context, c io.ReadWriteCloser, config *Config) (*Firmata, error) {
	f := NewFirmata(c)
	if config != nil {
		f.OnConnected = config.OnConnected
		f.OnAnalogMessage = config.OnAnalogMessage
		f.OnDigitalMessage = config.OnDigitalMessage
		f.OnPinState = config.OnPinState
		f.OnI2cReply = config.OnI2cReply
		f.OnStringData = config.OnStringData
		f.OnSysexResponse = config.OnSysexResponse
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
	f.Loop(func() { done <- f.reportInit_l() })
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

func (f *Firmata) SnapshotPin(pin byte) (p *Pin, ok bool) {
	done := make(chan struct{})
	f.Loop(func() {
		ok = pin < f.TotalPins
		if ok {
			p = f.Pins[pin]
		}
		close(done)
	})
	<-done
	return
}

func (f *Firmata) SnapshotAnalogPin(pin byte) (p *Pin, ok bool) {
	done := make(chan struct{})
	f.Loop(func() {
		ok = pin < f.TotalAnalogPins
		if ok {
			p = f.AnalogPins[pin]
		}
		close(done)
	})
	<-done
	return
}

// Pins returns all available Pins
func (f *Firmata) SnapshotPins() []*Pin {
	done := make(chan []*Pin)
	f.Loop(func() { done <- ClonePins_l(f.Pins) })
	return <-done
}

// Pins returns all available AnalogPins
func (f *Firmata) SnapshotAnalogPins() []*Pin {
	done := make(chan []*Pin)
	f.Loop(func() { done <- ClonePins_l(f.AnalogPins) })
	return <-done
}

// Reset sends the SystemReset sysex code.
func (f *Firmata) Reset_l() (err error) {
	err = f.writer.Reset()
	if err != nil {
		return
	}
	err = f.reportInit_l()
	if err != nil {
		return
	}
	f.ClosedError = nil
	f.VersionInfo = VersionInfo{}
	f.DxByName = nil
	f.Pins = nil
	f.AnalogPins = nil
	f.TotalPorts = 0
	f.TotalPins = 0
	f.TotalAnalogPins = 0
	f.PortConfigInputs = [16]byte{}
	f.connectedOnce = sync.Once{}
	return nil
}

// SetPinMode sets the pin to mode.
func (f *Firmata) SetPinMode_l(pin byte, mode byte) error {
	if pin >= f.TotalPins {
		return fmt.Errorf("SetPinMode pin out of index: %d", pin)
	}
	if !f.Pins[pin].SupportMode(mode) {
		return fmt.Errorf("unsupported mode, pin: %d", pin)
	}
	if f.Pins[pin].Mode_l == mode {
		return nil
	}
	err := f.writer.SetPinMode(pin, mode)
	if err != nil {
		return err
	}
	f.handlePinMode_l(pin, mode)
	return nil
}

// SetDigitalPinValue sets the pin to value(0/1).
func (f *Firmata) SetDigitalPinValue_l(pin byte, value byte) error {
	if pin >= f.TotalPins {
		return fmt.Errorf("SetDigitalPinValue pin out of index: %d", pin)
	}
	if f.Pins[pin].Value_l == uint(value) {
		return nil
	}
	err := f.writer.SetDigitalPinValue(pin, value)
	if err == nil {
		f.Pins[pin].Value_l = uint(value)
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
func (f *Firmata) DigitalWrite_l(port byte, values byte) (pins byte, err error) {
	if port > f.TotalPorts {
		return 0, fmt.Errorf("DigitalWrite port out of index: %d", port)
	}
	err = f.writer.DigitalWrite(port, values)
	if err != nil {
		return 0, err
	}
	return f.localDigitalOutputPortValues_l(port, values), nil
}

func (f *Firmata) PortPins_l(port byte) []*Pin {
	start, end := f.portRange(port)
	return f.Pins[start:end]
}

func (f *Firmata) portRange(port byte) (start byte, end byte) {
	start = 8 * port
	end = start + 8
	if end > f.TotalPins {
		end = f.TotalPins
	}
	return
}

// src/DigitalOutputFirmata.cpp
func (f *Firmata) localDigitalOutputPortValues_l(port byte, values byte) (pins byte) {
	if port >= f.TotalPorts {
		return
	}

	var pinValue uint
	var mask byte
	for i, pin := range f.PortPins_l(port) {
		if pin.IsDigital_l() {
			if pin.Mode_l == PIN_MODE_OUTPUT || pin.Mode_l == PIN_MODE_INPUT {
				mask = 1 << i
				if values&mask == 0 {
					pinValue = 0
				} else {
					pinValue = 1
				}

				if pin.Mode_l == PIN_MODE_OUTPUT && pinValue != pin.Value_l {
					pins |= mask
					pin.Value_l = pinValue
					pin.State_l = pinValue
				} else if pin.Mode_l == PIN_MODE_INPUT && pinValue == 1 &&
					byte(pin.State_l) != 1 {
					pins |= mask
					pin.Mode_l = PIN_MODE_PULLUP
					pin.State_l = pinValue
				}
			}
		}
	}
	return
}

// ServoConfig sets the min and max pulse width for servo PWM range
func (f *Firmata) ServoConfig_l(pin byte, max int, min int) error {
	// TODO save to Firmata
	return f.writer.ServoConfig(pin, max, min)
}

// AnalogWrite writes value to pin.
func (f *Firmata) AnalogWrite_l(pin byte, value uint) (err error) {
	if pin >= f.TotalPins {
		return fmt.Errorf("AnalogWrite pin(%d) not found", pin)
	}

	if pin > 15 || value >= 0x4000 {
		err = f.writer.ExtendedAnalogWrite(pin, value)
	} else {
		err = f.writer.AnalogWrite(pin, value)
	}
	if err == nil {
		f.Pins[pin].Value_l = value
	}
	return err
}

// PinStateQuery sends a PinStateQuery for pin.
func (f *Firmata) PinStateQuery_l(pin byte) error {
	return f.writer.PinStateQuery(pin)
}

// PinState_l sends a PinStateQuery for pin.
func (f *Firmata) PinState_l(pin byte) (uint, error) {
	if pin >= f.TotalPins {
		return 0, fmt.Errorf("PinStateQuery pin(%d) not found", pin)
	}
	return f.Pins[pin].State_l, nil
}

func (f *Firmata) reportInit_l() error {
	if f.ProtocolVersion == nil {
		return f.writer.ReportVersion()
	}
	if f.FirmwareVersion == nil {
		return f.writer.ReportFirmware()
	}
	if f.Pins == nil {
		return f.writer.CapabilitiesQuery()
	}
	if f.AnalogPins == nil {
		return f.writer.AnalogMappingQuery()
	}
	if f.DxByName == nil {
		return f.writer.PinNamesRequest()
	}
	return nil
}

func (f *Firmata) IsDigital_l(pin byte) bool {
	return pin < f.TotalPins && f.Pins[pin].IsDigital_l()
}

func (f *Firmata) handlePinMode_l(pin byte, mode byte) {
	if !f.IsDigital_l(pin) {
		return
	}

	p := f.Pins[pin]
	// FirmataClass::setPinMode
	p.State_l = 0
	p.Mode_l = mode
	// features below

	// src/DigitalInputFirmata.cpp
	if mode == PIN_MODE_INPUT || mode == PIN_MODE_PULLUP {
		f.PortConfigInputs[pin/8] |= (1 << (pin & 7))
		if mode == PIN_MODE_PULLUP {
			p.State_l = 1
		}
	} else {
		f.PortConfigInputs[pin/8] &^= (1 << (pin & 7))
	}

	switch mode {
	// case PIN_MODE_INPUT, PIN_MODE_PULLUP:
	// handled

	case PIN_MODE_OUTPUT:
		// src/DigitalOutputFirmata.cpp
		if p.Mode_l != PIN_MODE_IGNORE {
			p.Value_l = 0 // disable PWM
		}

	// case PIN_MODE_ANALOG:
	// src/AnalogInputFirmata.cpp
	// return

	// case PIN_MODE_SERIAL, PIN_MODE_I2C:
	// src/SerialFirmata.cpp
	// nothing else to do here since the mode is set in SERIAL_CONFIG

	// src/I2CFirmata.cpp
	// the user must call I2C_CONFIG to enable I2C for a device
	// return

	case PIN_MODE_PWM, PIN_MODE_STEPPER:
		// src/AnalogOutputFirmata.cpp
		// src/StepperFirmata.cpp
		p.Value_l = 0
		return
	}
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
	if len(data) > MaxI2cDataBytes {
		return fmt.Errorf("MaxI2cDataBytes is %d, but data len is %d",
			MaxI2cDataBytes, len(data))
	}
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

func (f *Firmata) StringWrite_l(data []byte) error {
	if len(data) > MaxStringDataBytes {
		return fmt.Errorf("MaxStringDataBytes is %d, but data len is %d",
			MaxStringDataBytes, len(data))
	}
	return f.writer.StringWrite(data)
}

// SamplingInterval sets how often analog data and i2c data is reported to the
// client. The default for the arduino implementation is 19ms. This means that
// every 19ms analog data will be reported and any i2c devices with read
// continuous mode will be read.
func (f *Firmata) SamplingInterval_l(ms uint) error {
	return f.writer.SamplingInterval(ms)
}
