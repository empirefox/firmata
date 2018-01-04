package firmata

import (
	"fmt"
	"math"
)

func (b *Firmata) unmashalReportVersion(buf []byte) (err error) {
	b.reportVersionOnce.Do(func() {
		b.ProtocolVersion = NewProtocalVersion(buf[0], buf[1])
		err = b.reportFirmware()
	})
	return err
}

func (b *Firmata) unmashalReportFirmware(currentBuffer []byte) (err error) {
	b.reportFirmwareOnce.Do(func() {
		if err = b.capabilitiesQuery(); err != nil {
			return
		}
		name := []byte{}
		for _, val := range currentBuffer[4:(len(currentBuffer) - 1)] {
			if val != 0 {
				name = append(name, val)
			}
		}
		b.FirmwareVersion = NewFirmwareVersion(currentBuffer[2], currentBuffer[3])
		b.FirmwareName = string(name)
	})
	return err
}

func (b *Firmata) unmashalCapabilityResponse(currentBuffer []byte) (err error) {
	b.capabilityQueryOnce.Do(func() {
		if err = b.analogMappingQuery(); err != nil {
			return
		}
		b.pins = make([]*Pin, 0)
		pid := 0
		modes := make(map[int]byte)
		n := 0

		for i, val := range currentBuffer[2 : len(currentBuffer)-1] {
			if val == 127 {
				b.pins = append(b.pins, &Pin{ID: pid, Modes: modes, Mode: PIN_MODE_OUTPUT})
				pid++
				modes = make(map[int]byte)
			} else {
				if n == 0 {
					i++
					modes[int(val)] = currentBuffer[i]
				}
				n ^= 1
			}
		}
		b.TotalPorts = int(math.Floor(float64(len(b.pins))/8) + 1)
	})
	return err
}

func (b *Firmata) unmashalAnalogMappingResponse(currentBuffer []byte) (err error) {
	b.analogMappingQueryOnce.Do(func() {
		pinIndex := 0
		b.analogPins = make([]*Pin, 0)

		for _, val := range currentBuffer[2 : len(currentBuffer)-1] {
			if pinIndex >= len(b.pins) {
				err = fmt.Errorf("ANALOG_MAPPING_RESPONSE pin index out of range: %d", pinIndex)
				return
			}
			pin := b.pins[pinIndex]
			pin.AnalogChannel = int(val)

			if val != 127 {
				b.analogPins = append(b.analogPins, pin)
			}
			pinIndex++
		}
		b.connecting.Store(false)
		b.connected.Store(true)
		b.connectedCh <- struct{}{}
	})
	return err
}

func (b *Firmata) unmashal() (err error) {
	msgBuf, err := b.read(1)
	if err != nil {
		return err
	}
	messageType := msgBuf[0]
	switch {
	case REPORT_VERSION == messageType:
		buf, err := b.read(2)
		if err != nil {
			return err
		}
		if err = b.unmashalReportVersion(buf); err != nil {
			b.connecting.Store(false)
			return err
		}
	case ANALOG_MESSAGE <= messageType && messageType <= 0xEF:
		buf, err := b.read(2)
		if err != nil {
			return err
		}
		value := uint(buf[0]) | uint(buf[1])<<7
		pid := int(messageType & 0x0F)
		if pid >= len(b.analogPins) {
			return fmt.Errorf("ANALOG_MESSAGE pin not found: %d", pid)
		}
		pin := b.analogPins[pid]
		pin.Value = int(value)
		if b.OnAnalogMessage != nil {
			b.OnAnalogMessage(pin)
		}
	case DIGITAL_MESSAGE <= messageType && messageType <= 0x9F:
		buf, err := b.read(2)
		if err != nil {
			return err
		}
		port := messageType & 0x0F
		portValue := buf[0] | (buf[1] << 7)

		var pins []*Pin
		for i := 0; i < 8; i++ {
			pinNumber := int((8*byte(port) + byte(i)))
			if pinNumber >= len(b.pins) {
				if i == 0 {
					return fmt.Errorf("DIGITAL_MESSAGE port not found: %d", port)
				}
				break
			}
			pin := b.pins[pinNumber]
			if pin.Mode == PIN_MODE_INPUT {
				pin.Value = int((portValue >> (byte(i) & 0x07)) & 0x01)
				pins = append(pins, pin)
			}
		}
		if b.OnDigitalMessage != nil {
			b.OnDigitalMessage(pins)
		}
	case START_SYSEX == messageType:
		buf, err := b.read(2)
		if err != nil {
			return err
		}

		currentBuffer := append(msgBuf, buf[0], buf[1])
		for {
			buf, err = b.read(1)
			if err != nil {
				return err
			}
			currentBuffer = append(currentBuffer, buf[0])
			if buf[0] == END_SYSEX {
				break
			}
		}
		command := currentBuffer[1]
		switch command {
		case CAPABILITY_RESPONSE:
			if err = b.unmashalCapabilityResponse(currentBuffer); err != nil {
				b.connecting.Store(false)
				return err
			}
		case ANALOG_MAPPING_RESPONSE:
			if err = b.unmashalAnalogMappingResponse(currentBuffer); err != nil {
				b.connecting.Store(false)
				return err
			}
		case PIN_STATE_RESPONSE:
			pid := int(currentBuffer[2])
			if pid >= len(b.pins) {
				return fmt.Errorf("PIN_STATE_RESPONSE pin not found: %d", pid)
			}
			pin := b.pins[pid]
			pin.Mode = int(currentBuffer[3])
			pin.State = int(currentBuffer[4])

			if len(currentBuffer) > 6 {
				pin.State = int(uint(pin.State) | uint(currentBuffer[5])<<7)
			}
			if len(currentBuffer) > 7 {
				pin.State = int(uint(pin.State) | uint(currentBuffer[6])<<14)
			}
			if b.OnPinState != nil {
				b.OnPinState(pin)
			}
		case I2C_REPLY:
			if b.OnI2cReply != nil {
				reply := I2cReply{
					Address:  int(byte(currentBuffer[2]) | byte(currentBuffer[3])<<7),
					Register: int(byte(currentBuffer[4]) | byte(currentBuffer[5])<<7),
					Data:     []byte{byte(currentBuffer[6]) | byte(currentBuffer[7])<<7},
				}
				for i := 8; i < len(currentBuffer); i = i + 2 {
					if currentBuffer[i] == byte(0xF7) {
						break
					}
					if i+2 > len(currentBuffer) {
						break
					}
					reply.Data = append(reply.Data, byte(currentBuffer[i])|byte(currentBuffer[i+1])<<7)
				}
				b.OnI2cReply(&reply)
			}
		case REPORT_FIRMWARE:
			if err = b.unmashalReportFirmware(currentBuffer); err != nil {
				b.connecting.Store(false)
				return err
			}
		case STRING_DATA:
			if b.OnStringData != nil {
				b.OnStringData(currentBuffer[2 : len(currentBuffer)-1])
			}
		default:
			if b.OnSysexResponse != nil {
				b.OnSysexResponse(currentBuffer)
			}
		}
	}
	return
}
