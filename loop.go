package firmata

import (
	"errors"
	"fmt"
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
			return f.writer.ReportFirmware()
		}
	case REPORT_FIRMWARE:
		if f.FirmwareVersion == nil {
			data := frame.Data.(*FirmwareFrameData)
			f.FirmwareVersion = data.FirmwareVersion
			f.FirmwareName = data.FirmwareName
			return f.writer.CapabilitiesQuery()
		}
	case CAPABILITY_RESPONSE:
		if f.Pins_l == nil {
			data := frame.Data.(*CapabilityFrameData)
			f.Pins_l = data.Pins
			f.TotalPorts_l = data.TotalPorts
			return f.writer.AnalogMappingQuery()
		}
	case ANALOG_MAPPING_RESPONSE:
		if f.Pins_l == nil {
			return f.reportVersion()
		}
		if f.AnalogPins_l == nil {
			data := frame.Data.([]byte)
			if len(data) > len(f.Pins_l) {
				return fmt.Errorf("ANALOG_MAPPING_RESPONSE pins more than total: %d > %d", len(data), len(f.Pins_l))
			}

			aps := 0
			analogPins := make([]*Pin, len(f.Pins_l))

			for pid, val := range data {
				pin := f.Pins_l[pid]
				pin.AnalogChannel = val

				if val != 127 {
					analogPins[aps] = pin
					aps++
				}
			}
			f.AnalogPins_l = analogPins[:aps]
			f.connectedOnce.Do(f.onConnected)
		}
	default:
		if f.AnalogPins_l == nil {
			return f.reportVersion()
		}

		switch frame.Type {
		case ANALOG_MESSAGE:
			data := frame.Data.(*AnalogPinValueFrameData)
			if data.Pin >= byte(len(f.AnalogPins_l)) {
				return fmt.Errorf("ANALOG_MESSAGE pin out of index: %d", data.Pin)
			}
			pin := f.AnalogPins_l[data.Pin]
			pin.Value = int(data.Value)
			if f.OnAnalogMessage != nil {
				f.OnAnalogMessage(f, pin)
			}
		case DIGITAL_MESSAGE:
			data := frame.Data.(*DigitalPinValueFrameData)
			if data.Port > f.TotalPorts_l {
				return fmt.Errorf("DIGITAL_MESSAGE port out of index: %d", data.Port)
			}
			pins := f.setDigitalPortValues(data.Port, &data.Values)
			if f.OnDigitalMessage != nil {
				f.OnDigitalMessage(f, pins)
			}
		case PIN_STATE_RESPONSE:
			data := frame.Data.(*PinStateFrameData)
			if data.Pin >= byte(len(f.Pins_l)) {
				return fmt.Errorf("PIN_STATE_RESPONSE pin out of index: %d", data.Pin)
			}
			pin := f.Pins_l[data.Pin]
			pin.Mode = data.Mode
			pin.State = data.State
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

func (f *Firmata) setDigitalPortValues(port byte, values *[8]byte) []*Pin {
	start := 8 * int(port)
	end := start + 8
	if end > len(f.Pins_l) {
		end = len(f.Pins_l)
	}

	pins := make([]*Pin, end-start)
	ps := 0
	for i, pin := range f.Pins_l[start:end] {
		if pin.Mode == PIN_MODE_INPUT {
			pin.Value = int(values[i])
			pins[ps] = pin
			ps++
		}
	}
	return pins[:ps]
}

func (f *Firmata) Close() {
	f.doneOnce.Do(func() {
		for _, w := range f.reader.pipes {
			w.Close()
		}
		f.reader.pipes = nil
		f.closer.Close()
		close(f.doneServing)
	})
	<-f.doneServing
}

func (f *Firmata) CloseNotify() <-chan struct{} {
	return f.doneServing
}
