package firmata

import (
	"fmt"
	"io"
	"math"
)

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
	return fr.write([]byte{SET_PIN_MODE, byte(pin), byte(mode)})
}

func (fr *WriteFramer) SetDigitalPinValue(pin byte, value byte) error {
	return fr.write([]byte{SET_DIGITAL_PIN_VALUE, byte(pin), byte(value)})
}

func (fr *WriteFramer) SetDigitalPinHigh(pin byte) error {
	return fr.SetDigitalPinValue(pin, 1)
}

func (fr *WriteFramer) SetDigitalPinLow(pin byte) error {
	return fr.SetDigitalPinValue(pin, 0)
}

func (fr *WriteFramer) DigitalWrite(port byte, values *[8]byte) error {
	var portValue byte
	for i, v := range values {
		if v != 0 {
			portValue = portValue | (1 << uint(i))
		}
	}
	return fr.write([]byte{
		DIGITAL_MESSAGE | byte(port),
		portValue & 0x7F,
		(portValue >> 7) & 0x7F},
	)
}

func (fr *WriteFramer) ServoConfig(pin byte, max int, min int) error {
	return fr.write([]byte{
		START_SYSEX,
		SERVO_CONFIG,
		byte(pin),
		byte(min & 0x7F),
		byte((min >> 7) & 0x7F),
		byte(max & 0x7F),
		byte((max >> 7) & 0x7F),
		END_SYSEX,
	})
}

func (fr *WriteFramer) AnalogWrite(pin byte, value int) error {
	return fr.write([]byte{ANALOG_MESSAGE | byte(pin), byte(value & 0x7F), byte((value >> 7) & 0x7F)})
}

func (fr *WriteFramer) PinStateQuery(pin byte) error {
	return fr.write([]byte{START_SYSEX, PIN_STATE_QUERY, byte(pin), END_SYSEX})
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

func (fr *WriteFramer) ReportDigital(port byte, value byte) error {
	return fr.write([]byte{REPORT_DIGITAL | byte(port), byte(value)})
}

func (fr *WriteFramer) ReportAnalog(pin byte, value byte) error {
	return fr.write([]byte{REPORT_ANALOG | byte(pin), byte(value)})
}

func (fr *WriteFramer) I2cRead(address int, numBytes byte) error {
	return fr.write([]byte{
		START_SYSEX,
		I2C_REQUEST,
		byte(address),
		byte(PIN_MODE_OUTPUT << 3),
		byte(numBytes) & 0x7F,
		(byte(numBytes) >> 7) & 0x7F,
		END_SYSEX,
	})
}

func (fr *WriteFramer) I2cWrite(address int, data []byte) error {
	rs := len(data)*2 + 5
	if rs > MAX_DATA_BYTES {
		return fmt.Errorf("MAX_DATA_BYTES is %d, but data len is %d", MAX_DATA_BYTES, rs)
	}

	fr.bufI2C[2], fr.bufI2C[3], fr.bufI2C[rs-1] = byte(address), byte(PIN_MODE_INPUT<<3), END_SYSEX
	buf := fr.bufI2C[4:]
	for i, val := range data {
		i = i * 2
		buf[i] = byte(val & 0x7F)
		buf[i+1] = byte((val >> 7) & 0x7F)
	}
	return fr.write(fr.bufI2C[:rs])
}

func (fr *WriteFramer) I2cConfig(delay uint) error {
	return fr.write([]byte{
		START_SYSEX,
		I2C_CONFIG,
		byte(delay & 0xFF),
		byte((delay >> 8) & 0xFF),
		END_SYSEX,
	})
}

func (fr *WriteFramer) write(b []byte) (err error) {
	_, err = fr.w.Write(b)
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
	Values [8]byte
}

type CapabilityFrameData struct {
	TotalPorts byte
	Pins       []*Pin
}

type PinStateFrameData struct {
	Pin   byte
	Mode  byte
	State int
}

type FirmwareFrameData struct {
	FirmwareVersion *Version
	FirmwareName    []byte
}

type ReadFramer struct {
	r   io.Reader
	buf [MaxRecvSize]byte
	cur int

	pipes []io.WriteCloser

	cacheReportversion  []byte
	cacheQueryfirmware  []byte
	cacheAnalogMappings []byte
	cacheCapability     []byte
}

func NewReadFramer(r io.Reader) *ReadFramer {
	return &ReadFramer{r: r, pipes: make([]io.WriteCloser, 0, 32)}
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
		fr.cacheReportversion = []byte{fr.buf[0], fr.buf[1], fr.buf[2]}
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
		fr.proxyPipesWrite()
	case DIGITAL_MESSAGE <= messageType && messageType <= 0x9F:
		data := DigitalPinValueFrameData{
			Port: messageType & 0x0F,
		}
		// D7----D0
		portValue := fr.buf[1] | fr.buf[2]<<7
		var i byte
		for i = 0; i < 8; i++ {
			data.Values[i] = portValue >> i & 0x01
		}
		f = &ReadFrame{
			Type: DIGITAL_MESSAGE,
			Data: &data,
		}
		fr.proxyPipesWrite()
	case START_SYSEX == messageType:
		err = fr.readUtil(END_SYSEX)
		if err != nil {
			return
		}

		switch fr.buf[1] {
		case CAPABILITY_RESPONSE:
			fr.cacheCapability = fr.copyBuf()
			pins := make([]*Pin, fr.cur/2-1)
			modes := make(map[byte]byte, TOTAL_PIN_MODES)
			var pid byte
			n := 0

			for i, val := range fr.buf[2 : fr.cur-1] {
				if val == 127 {
					pins[pid] = &Pin{ID: pid, Modes: modes, Mode: PIN_MODE_OUTPUT}
					pid++
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
					TotalPorts: byte(math.Floor(float64(pid)/8) + 1),
					Pins:       pins[:pid],
				},
			}
		case ANALOG_MAPPING_RESPONSE:
			fr.cacheAnalogMappings = fr.copyBuf()
			data := make([]byte, fr.cur-3)
			copy(data, fr.buf[2:])
			f = &ReadFrame{
				Type: ANALOG_MAPPING_RESPONSE,
				Data: data,
			}
		case PIN_STATE_RESPONSE:
			state := uint(fr.buf[4])
			if fr.cur > 6 {
				state = uint(state) | uint(fr.buf[5])<<7
			}
			if fr.cur > 7 {
				state = uint(state) | uint(fr.buf[6])<<14
			}
			f = &ReadFrame{
				Type: PIN_STATE_RESPONSE,
				Data: &PinStateFrameData{
					Pin:   fr.buf[2],
					Mode:  fr.buf[3],
					State: int(state),
				},
			}
			fr.proxyPipesWrite()
		case I2C_REPLY:
			data := make([]byte, fr.cur/2-3)
			data[0] = byte(fr.buf[6]) | byte(fr.buf[7])<<7
			ds := 1
			for i := 8; i < fr.cur; i = i + 2 {
				if fr.buf[i] == byte(0xF7) {
					break
				}
				if i+2 > fr.cur {
					break
				}
				data[ds] = byte(fr.buf[i]) | byte(fr.buf[i+1])<<7
				ds++
			}
			f = &ReadFrame{
				Type: I2C_REPLY,
				Data: &I2cReply{
					Address:  byte(fr.buf[2]) | byte(fr.buf[3])<<7,
					Register: int(byte(fr.buf[4]) | byte(fr.buf[5])<<7),
					Data:     data,
				},
			}
			// TODO standalone conn pipe
			// fr.proxyPipesWrite()
		case REPORT_FIRMWARE:
			fr.cacheQueryfirmware = fr.copyBuf()
			name := make([]byte, fr.cur-5)
			ns := 0
			for _, val := range fr.buf[4 : fr.cur-1] {
				if val != 0 {
					name[ns] = val
					ns++
				}
			}
			f = &ReadFrame{
				Type: REPORT_FIRMWARE,
				Data: &FirmwareFrameData{
					FirmwareVersion: NewFirmwareVersion(fr.buf[2], fr.buf[3]),
					FirmwareName:    name[:ns],
				},
			}
		case STRING_DATA:
			data := make([]byte, fr.cur-3)
			copy(data, fr.buf[2:])
			f = &ReadFrame{
				Type: STRING_DATA,
				Data: data,
			}
			fr.proxyPipesWrite()
		default:
			data := make([]byte, fr.cur-2)
			copy(data, fr.buf[1:])
			f = &ReadFrame{
				Type: START_SYSEX,
				Data: data,
			}
			fr.proxyPipesWrite()
		}
	default:
		err = fmt.Errorf("Unsupported firmata type: %X", fr.buf[0])
	}
	return
}

func (fr *ReadFramer) copyBuf() []byte {
	clone := make([]byte, fr.cur)
	copy(clone, fr.buf[:])
	return clone
}

func (fr *ReadFramer) proxyPipesWrite() {
	var i int
	var err error
	for _, w := range fr.pipes {
		_, err = w.Write(fr.buf[:fr.cur])
		if err != nil {
			w.Close()
			continue
		}
		fr.pipes[i] = w
		i++
	}
	fr.pipes = fr.pipes[:i]
}

type proxyReadFramer struct {
	f    *Firmata
	r    io.Reader
	w    io.Writer
	buf  [MAX_DATA_BYTES]byte
	cur  int
	done chan error
}

func (fr *proxyReadFramer) readSendNextFrame() (err error) {
	fr.cur = 0
	err = fr.read(1)
	if err != nil {
		return
	}

	messageType := fr.buf[0]
	if messageType < 0xF0 {
		messageType = fr.buf[0] & 0xF0
	}
	switch messageType {
	case REPORT_VERSION:
		_, err = fr.w.Write(fr.f.reader.cacheReportversion)
	case SYSTEM_RESET:
		err = fr.send()
	case REPORT_ANALOG, REPORT_DIGITAL:
		err = fr.readSend(1)
	case ANALOG_MESSAGE, DIGITAL_MESSAGE, SET_PIN_MODE, SET_DIGITAL_PIN_VALUE:
		err = fr.readSend(2)
	case START_SYSEX:
		err = fr.readUtil(END_SYSEX)
		if err != nil {
			return
		}
		switch fr.buf[1] {
		case REPORT_FIRMWARE:
			_, err = fr.w.Write(fr.f.reader.cacheQueryfirmware)
		case CAPABILITY_QUERY:
			_, err = fr.w.Write(fr.f.reader.cacheCapability)
		case ANALOG_MAPPING_QUERY:
			_, err = fr.w.Write(fr.f.reader.cacheAnalogMappings)
		default:
			err = fr.send()
		}
	default:
		err = fmt.Errorf("Unsupported firmata type: %X", fr.buf[0])
	}
	return
}

func (fr *proxyReadFramer) readSend(n int) error {
	if err := fr.read(n); err != nil {
		return err
	}
	return fr.send()
}

func (fr *proxyReadFramer) read(n int) error {
	_, err := io.ReadFull(fr.r, fr.buf[fr.cur:fr.cur+n])
	if err == nil {
		fr.cur += n
	}
	return err
}

func (fr *proxyReadFramer) readUtil(end byte) error {
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

func (fr *proxyReadFramer) send() error {
	fr.f.Loop(func() { fr.done <- fr.f.writer.write(fr.buf[:fr.cur]) })
	return <-fr.done
}
