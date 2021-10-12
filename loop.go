package firmata

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrClosed = errors.New("Firmata closed")
)

func (f *Firmata) onConnected() {
	f.handshakeOnce.Do(func() { close(f.handshakeOK) })
	if f.OnConnected != nil {
		f.OnConnected(f)
	}
}

func (f *Firmata) serve() {
	defer f.Close()

	for {
		select {
		case fn := <-f.loopCh:
			fn()
		case res := <-f.readFrameCh:
			if res.err != nil {
				f.ClosedError = res.err
				return
			}
			err := f.proccessFrame(res.frame)
			if err != nil {
				f.ClosedError = err
				return
			}
		case <-f.doneServing:
			return
		}
	}
}

func (f *Firmata) Loop(fn func()) error {
	select {
	case f.loopCh <- fn:
		return nil
	case <-f.doneServing:
		return ErrClosed
	}
}

type readFrameResult struct {
	frame *ReadFrame
	err   error
}

func (f *Firmata) readFrames() {
	for {
		frame, err := f.reader.ReadFrame()
		select {
		case f.readFrameCh <- readFrameResult{frame, err}:
		case <-f.doneServing:
			return
		}
		if err != nil {
			return
		}
	}
}

func (f *Firmata) proccessFrame(frame *ReadFrame) (err error) {
	switch frame.Type {
	case REPORT_VERSION:
		if f.ProtocolVersion == nil {
			f.ProtocolVersion = frame.Data.(*Version)
			return f.reportInit_l()
		}
	case REPORT_FIRMWARE:
		if f.ProtocolVersion == nil {
			return f.reportInit_l()
		}
		if f.FirmwareVersion == nil {
			data := frame.Data.(*FirmwareFrameData)
			f.FirmwareVersion = data.FirmwareVersion
			f.FirmwareName = data.FirmwareName
			return f.reportInit_l()
		}
	case CAPABILITY_RESPONSE:
		if f.FirmwareVersion == nil {
			return f.reportInit_l()
		}
		if f.Pins == nil {
			data := frame.Data.(*CapabilityFrameData)
			totalPins := len(data.Pins)
			if totalPins == 0 || totalPins > math.MaxUint8 {
				return fmt.Errorf("CAPABILITY_RESPONSE invalid pin size: %d",
					totalPins)
			}

			f.Pins = data.Pins
			f.TotalPins = byte(totalPins)
			f.TotalPorts = data.TotalPorts
			return f.reportInit_l()
		}
	case ANALOG_MAPPING_RESPONSE:
		if f.Pins == nil {
			return f.reportInit_l()
		}
		if f.AnalogPins == nil {
			data := frame.Data.([]byte)
			if len(data) > int(f.TotalPins) {
				return fmt.Errorf("ANALOG_MAPPING_RESPONSE pins more than total: %d > %d", len(data), f.TotalPins)
			}

			var aps byte
			analogPins := make([]*Pin, f.TotalPins)

			for dx, val := range data {
				pin := f.Pins[dx]
				pin.Ax = val

				if val != 127 {
					analogPins[aps] = pin
					aps++
				}
			}
			f.AnalogPins = analogPins[:aps]
			f.TotalAnalogPins = aps

			for i := byte(0); i < f.TotalPins; i++ {
				err := f.writer.PinStateQuery(i)
				if err != nil {
					return err
				}
			}
			return f.reportInit_l()
		}
	case UD_PIN_NAMES_REPLY:
		if f.AnalogPins == nil {
			return f.reportInit_l()
		}
		if f.DxByName == nil {
			data := frame.Data.([]byte)
			if len(data) != int(f.TotalPins) {
				return fmt.Errorf("PIN_NAMES size must be %d, but got %d",
					f.TotalPins, len(data))
			}

			f.DxByName = make(map[PinName]byte, f.TotalPins)
			for i, name := range data {
				n := PinName(name)
				if _, ok := f.DxByName[n]; ok {
					return fmt.Errorf("PIN_NAMES got duplicated name: %s", n)
				}
				f.DxByName[n] = byte(i)
				f.Pins[i].Name = n
			}

			f.connectedOnce.Do(f.onConnected)
		}
	default:
		if f.AnalogPins == nil || (f.DxByName == nil && frame.Type != PIN_STATE_RESPONSE) {
			return f.reportInit_l()
		}

		switch frame.Type {
		case ANALOG_MESSAGE:
			data := frame.Data.(*AnalogPinValueFrameData)
			if data.Pin >= f.TotalAnalogPins {
				return fmt.Errorf("ANALOG_MESSAGE pin out of index: %d", data.Pin)
			}
			pin := f.AnalogPins[data.Pin]
			pin.Value_l = data.Value
			if f.OnAnalogMessage != nil {
				f.OnAnalogMessage(f, pin)
			}
		case DIGITAL_MESSAGE:
			data := frame.Data.(*DigitalPinValueFrameData)
			if data.Port > f.TotalPorts {
				return fmt.Errorf("DIGITAL_MESSAGE port out of index: %d", data.Port)
			}

			inputs := f.PortConfigInputs[data.Port]
			if data.Values&^inputs != 0 {
				return fmt.Errorf("DIGITAL_MESSAGE portConfigInputs error: %#v", data)
			}

			var pins byte
			var mask byte
			var inValue uint
			for i, pin := range f.PortPins_l(data.Port) {
				mask = 1 << i
				inValue = uint(data.Values>>i) & 1
				if inputs&mask != 0 && inValue != pin.Value_l {
					pins |= mask
					pin.Value_l = inValue
				}
			}
			if f.OnDigitalMessage != nil {
				f.OnDigitalMessage(f, data.Port, pins)
			}
		case PIN_STATE_RESPONSE:
			data := frame.Data.(*PinStateFrameData)
			if data.Pin >= f.TotalPins {
				return fmt.Errorf("PIN_STATE_RESPONSE pin out of index: %d", data.Pin)
			}
			pin := f.Pins[data.Pin]
			// pin.Mode = data.Mode
			f.handlePinMode_l(data.Pin, data.Mode)
			if data.Mode == PIN_MODE_PULLUP && data.State != pin.State_l {
				return fmt.Errorf("PIN_STATE_RESPONSE pin(%d) state error, got %d",
					data.Pin, data.State)
			}
			pin.State_l = data.State
			if f.OnPinState != nil {
				f.OnPinState(f, pin)
			}
		case I2C_REPLY:
			if f.OnI2cReply != nil {
				f.OnI2cReply(f, frame.Data.(*I2cReply))
			}
		case STRING_DATA:
			if f.OnStringData != nil {
				f.OnStringData(f, frame.Data.([]byte))
			}
		case START_SYSEX:
			if f.OnSysexResponse != nil {
				f.OnSysexResponse(f, frame.Data.([]byte))
			}

		default:
			return fmt.Errorf("unexpected firmata type: %X", frame.Type)
		}
	}
	return nil
}

func (f *Firmata) Close() {
	f.doneOnce.Do(func() {
		f.closer.Close()
		close(f.doneServing)
	})
	<-f.doneServing
}

func (f *Firmata) CloseNotify() <-chan struct{} {
	return f.doneServing
}
