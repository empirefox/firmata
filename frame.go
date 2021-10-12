package firmata

import (
	"fmt"
	"io"
)

var endSysex = []byte{END_SYSEX}

type WriteFramer struct {
	w      io.Writer
	bufI2C [MAX_DATA_BYTES]byte
}

func NewWriteFramer(w io.Writer) *WriteFramer {
	return &WriteFramer{
		w:      w,
		bufI2C: [MAX_DATA_BYTES]byte{START_SYSEX, I2C_REQUEST},
	}
}

func (fr *WriteFramer) Reset() error {
	return fr.write([]byte{SYSTEM_RESET})
}

func (fr *WriteFramer) SetPinMode(pin byte, mode byte) error {
	return fr.write([]byte{SET_PIN_MODE, pin, byte(mode)})
}

func (fr *WriteFramer) SetDigitalPinValue(pin byte, value byte) error {
	return fr.write([]byte{SET_DIGITAL_PIN_VALUE, pin, byte(value)})
}

func (fr *WriteFramer) SetDigitalPinHigh(pin byte) error {
	return fr.SetDigitalPinValue(pin, 1)
}

func (fr *WriteFramer) SetDigitalPinLow(pin byte) error {
	return fr.SetDigitalPinValue(pin, 0)
}

func (fr *WriteFramer) DigitalWrite(port byte, portValue byte) error {
	return fr.write([]byte{
		DIGITAL_MESSAGE | byte(port&0x0F),
		portValue & 0x7F,
		(portValue >> 7) & 0x01},
	)
}

func (fr *WriteFramer) ServoConfig(pin byte, max int, min int) error {
	return fr.write([]byte{
		START_SYSEX,
		SERVO_CONFIG,
		pin,
		byte(min & 0x7F),
		byte((min >> 7) & 0x7F),
		byte(max & 0x7F),
		byte((max >> 7) & 0x7F),
		END_SYSEX,
	})
}

func (fr *WriteFramer) AnalogWrite(pin byte, value uint) error {
	return fr.write([]byte{ANALOG_MESSAGE | pin, byte(value & 0x7F), byte((value >> 7) & 0x7F)})
}

func (fr *WriteFramer) ExtendedAnalogWrite(pin byte, value uint) error {
	b0 := byte(value & 0x7F)
	b1 := byte((value >> 7) & 0x7F)
	b2 := byte((value >> 14) & 0x7F)
	b3 := byte((value >> 21) & 0x7F)
	b4 := byte((value >> 28) & 0x7F)
	if b4 != 0 {
		return fr.write([]byte{START_SYSEX,
			EXTENDED_ANALOG,
			pin & 0x7F,
			b0, b1, b2, b3, b4,
			END_SYSEX,
		})
	}
	if b3 != 0 {
		return fr.write([]byte{START_SYSEX,
			EXTENDED_ANALOG,
			pin & 0x7F,
			b0, b1, b2, b3,
			END_SYSEX,
		})
	}
	if b2 != 0 {
		return fr.write([]byte{START_SYSEX,
			EXTENDED_ANALOG,
			pin & 0x7F,
			b0, b1, b2,
			END_SYSEX,
		})
	}
	if b1 != 0 {
		return fr.write([]byte{START_SYSEX,
			EXTENDED_ANALOG,
			pin & 0x7F,
			b0, b1,
			END_SYSEX,
		})
	}
	return fr.write([]byte{START_SYSEX,
		EXTENDED_ANALOG,
		pin & 0x7F,
		b0,
		END_SYSEX,
	})
}

func (fr *WriteFramer) PinStateQuery(pin byte) error {
	return fr.write([]byte{START_SYSEX, PIN_STATE_QUERY, pin, END_SYSEX})
}

func (fr *WriteFramer) ReportVersion() error {
	return fr.write([]byte{REPORT_VERSION})
}
func (fr *WriteFramer) ReportFirmware() error {
	return fr.write([]byte{START_SYSEX, REPORT_FIRMWARE, END_SYSEX})
}
func (fr *WriteFramer) CapabilitiesQuery() error {
	return fr.write([]byte{START_SYSEX, CAPABILITY_QUERY, END_SYSEX})
}
func (fr *WriteFramer) AnalogMappingQuery() error {
	return fr.write([]byte{START_SYSEX, ANALOG_MAPPING_QUERY, END_SYSEX})
}
func (fr *WriteFramer) PinNamesRequest() error {
	return fr.write([]byte{START_SYSEX, UD_PIN_NAMES_REQUEST, END_SYSEX})
}

func (fr *WriteFramer) ReportDigital(port byte, value byte) error {
	return fr.write([]byte{REPORT_DIGITAL | byte(port), byte(value)})
}

func (fr *WriteFramer) ReportAnalog(pin byte, value byte) error {
	return fr.write([]byte{REPORT_ANALOG | pin, byte(value)})
}

func (fr *WriteFramer) I2cWrite(address int, data []byte) error {
	rs := len(data)*2 + 5
	fr.bufI2C[2], fr.bufI2C[rs-1] = byte(address&0x7F), END_SYSEX

	byte3 := 0b00000000
	addr3 := 0b00000111 & (address >> 7)
	if addr3 != 0 {
		// 10bit address
		byte3 = 0b00100000 | addr3
	}
	fr.bufI2C[3] = byte(byte3)

	buf := fr.bufI2C[4:]
	for i, val := range data {
		i = i * 2
		buf[i] = byte(val & 0x7F)
		buf[i+1] = byte((val >> 7) & 0x7F)
	}
	return fr.write(fr.bufI2C[:rs])
}

func (fr *WriteFramer) I2cRead(address int, autoRestartTransmission bool, continuous bool, numBytes byte) error {
	byte3 := 0b00001000
	if continuous {
		byte3 = 0b00010000
	}
	addr3 := 0b00000111 & (address >> 7)
	if addr3 != 0 {
		// 10bit address
		byte3 |= 0b00100000 | addr3
	}
	if autoRestartTransmission {
		byte3 |= 0b01000000
	}

	return fr.write([]byte{
		START_SYSEX,
		I2C_REQUEST,
		byte(address),
		byte(byte3),
		byte(numBytes) & 0x7F,
		(byte(numBytes) >> 7) & 0x7F,
		END_SYSEX,
	})
}

func (fr *WriteFramer) I2cStopReading(address int) error {
	byte3 := 0b00011000
	addr3 := 0b00000111 & (address >> 7)
	if addr3 != 0 {
		// 10bit address
		byte3 = 0b00100000 | addr3
	}

	return fr.write([]byte{
		START_SYSEX,
		I2C_REQUEST,
		byte(address),
		byte(byte3),
		END_SYSEX,
	})
}

func (fr *WriteFramer) I2cConfig(delay uint) error {
	return fr.write([]byte{
		START_SYSEX,
		I2C_CONFIG,
		byte(delay & 0x7F),
		byte((delay >> 8) & 0x7F),
		END_SYSEX,
	})
}

func (fr *WriteFramer) StringWrite(s []byte) error {
	return fr.writeAll(
		[]byte{START_SYSEX, STRING_DATA},
		To14bits(s),
		endSysex,
	)
}

func (fr *WriteFramer) SamplingInterval(ms uint) error {
	return fr.write([]byte{
		START_SYSEX,
		SAMPLING_INTERVAL,
		byte(ms & 0x7F),
		byte((ms >> 7) & 0x7F),
		END_SYSEX,
	})
}

func (fr *WriteFramer) write(b []byte) (err error) {
	_, err = fr.w.Write(b)
	return
}

func (fr *WriteFramer) writeAll(bs ...[]byte) (err error) {
	for _, b := range bs {
		_, err = fr.w.Write(b)
		if err != nil {
			return
		}
	}
	return
}

type ReadFrame struct {
	Type byte
	Data interface{}
}

type AnalogPinValueFrameData struct {
	Pin   byte
	Value uint
}

type DigitalPinValueFrameData struct {
	Port   byte
	Values byte // D7----D0
}

type CapabilityFrameData struct {
	TotalPorts byte
	Pins       []*Pin
}

type PinStateFrameData struct {
	Pin   byte
	Mode  byte
	State uint
}

type FirmwareFrameData struct {
	FirmwareVersion *Version
	FirmwareName    []byte
}

type ReadFramer struct {
	r   io.Reader
	buf [MaxRecvSize]byte
	cur int
}

func NewReadFramer(r io.Reader) *ReadFramer {
	return &ReadFramer{r: r}
}

func (fr *ReadFramer) readStart3() error {
	_, err := io.ReadFull(fr.r, fr.buf[:3])
	if err == nil {
		fr.cur = 3
	}
	return err
}

func (fr *ReadFramer) readUtil(end byte) error {
	for {
		next := fr.cur + 1
		_, err := io.ReadFull(fr.r, fr.buf[fr.cur:next])
		if err != nil {
			return err
		}
		if fr.buf[fr.cur] == end {
			fr.cur = next
			return nil
		}
		fr.cur = next
	}
}

func (fr *ReadFramer) ReadFrame() (f *ReadFrame, err error) {
	fr.cur = 0
	err = fr.readStart3()
	if err != nil {
		return
	}

	messageType := fr.buf[0]
	switch {
	case REPORT_VERSION == messageType:
		f = &ReadFrame{
			Type: REPORT_VERSION,
			Data: NewProtocalVersion(fr.buf[1], fr.buf[2]),
		}
	case ANALOG_MESSAGE <= messageType && messageType <= 0xEF:
		f = &ReadFrame{
			Type: ANALOG_MESSAGE,
			Data: &AnalogPinValueFrameData{
				Pin:   messageType & 0x0F,
				Value: uint(fr.buf[1]) | uint(fr.buf[2])<<7,
			},
		}
	case DIGITAL_MESSAGE <= messageType && messageType <= 0x9F:
		data := DigitalPinValueFrameData{
			Port: messageType & 0x0F,
			// D7----D0
			Values: fr.buf[1] | fr.buf[2]<<7,
		}
		f = &ReadFrame{
			Type: DIGITAL_MESSAGE,
			Data: &data,
		}
	case START_SYSEX == messageType:
		err = fr.readUtil(END_SYSEX)
		if err != nil {
			return
		}

		// fr.cur == full message size

		switch fr.buf[1] {
		case CAPABILITY_RESPONSE:
			pins := make([]*Pin, (fr.cur-3)/2)
			modes := make(map[byte]byte, TOTAL_PIN_MODES)
			var dx byte
			n := 0

			for i, val := range fr.buf[2 : fr.cur-1] {
				if val == 0x7F {
					pins[dx] = &Pin{
						Dx:     dx,
						Name:   PX, // unkown name
						Modes:  modes,
						Mode_l: PIN_MODE_OUTPUT,
					}
					dx++
					modes = make(map[byte]byte)
				} else {
					if n == 0 {
						i++
						modes[val] = fr.buf[i]
					}
					n ^= 1
				}
			}

			f = &ReadFrame{
				Type: CAPABILITY_RESPONSE,
				Data: &CapabilityFrameData{
					TotalPorts: (dx + 7) / 8,
					Pins:       pins[:dx],
				},
			}
		case ANALOG_MAPPING_RESPONSE:
			data := make([]byte, fr.cur-3)
			copy(data, fr.buf[2:])
			f = &ReadFrame{
				Type: ANALOG_MAPPING_RESPONSE,
				Data: data,
			}
		case UD_PIN_NAMES_REPLY:
			// 0  START_SYSEX                  (0xF0)
			// 1  UD_PIN_NAMES_REPLY           (0x07)
			// 2  pin0 PA0-PZ15(191) bits 0-6  (least significant byte)
			// 4  pin0 PA0-PZ15(191) bits 7-13 (most significant byte)
			// ... pin1 and more
			// N  END_SYSEX                    (0xF7)
			f = &ReadFrame{
				Type: UD_PIN_NAMES_REPLY,
				Data: From14bits(fr.buf[2 : fr.cur-1]),
			}

		case PIN_STATE_RESPONSE:
			state := uint(fr.buf[4])
			if fr.cur > 6 {
				state = state | uint(fr.buf[5])<<7
			}
			if fr.cur > 7 {
				state = state | uint(fr.buf[6])<<14
			}
			if fr.cur > 8 {
				state = state | uint(fr.buf[7])<<21
			}
			f = &ReadFrame{
				Type: PIN_STATE_RESPONSE,
				Data: &PinStateFrameData{
					Pin:   fr.buf[2],
					Mode:  fr.buf[3],
					State: state,
				},
			}
		case I2C_REPLY:
			f = &ReadFrame{
				Type: I2C_REPLY,
				Data: &I2cReply{
					Address:  int(fr.buf[2]) | int(fr.buf[3])<<7,
					Register: int(fr.buf[4]) | int(fr.buf[5])<<7,
					Data:     From14bits(fr.buf[6 : fr.cur-1]),
				},
			}
		case REPORT_FIRMWARE:
			f = &ReadFrame{
				Type: REPORT_FIRMWARE,
				Data: &FirmwareFrameData{
					FirmwareVersion: NewFirmwareVersion(fr.buf[2], fr.buf[3]),
					FirmwareName:    From14bits(fr.buf[4 : fr.cur-1]),
				},
			}
		case STRING_DATA:
			f = &ReadFrame{
				Type: STRING_DATA,
				Data: From14bits(fr.buf[2 : fr.cur-1]),
			}
		default:
			data := make([]byte, fr.cur-2)
			copy(data, fr.buf[1:])
			f = &ReadFrame{
				Type: START_SYSEX,
				Data: data,
			}
		}
	default:
		err = fmt.Errorf("unsupported firmata type: %X", fr.buf[0])
	}
	return
}
