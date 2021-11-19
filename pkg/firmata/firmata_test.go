package firmata

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/empirefox/firmata/pkg/pb"
	"gobot.io/x/gobot/gobottest"
)

type readWriteCloser struct {
	writeDataMutex sync.Mutex
	readDataMutex  sync.Mutex
	testReadData   []byte
	testWriteData  bytes.Buffer
}

func (rwc *readWriteCloser) Write(p []byte) (int, error) {
	rwc.writeDataMutex.Lock()
	defer rwc.writeDataMutex.Unlock()
	return rwc.testWriteData.Write(p)
}

func setTestReadData(f *Firmata, d []byte) {
	rwc := f.closer.(*readWriteCloser)
	rwc.readDataMutex.Lock()
	defer rwc.readDataMutex.Unlock()
	rwc.testReadData = d
}

func (rwc *readWriteCloser) Read(b []byte) (int, error) {
	rwc.readDataMutex.Lock()
	defer rwc.readDataMutex.Unlock()

	size := len(b)
	if len(rwc.testReadData) < size {
		size = len(rwc.testReadData)
	}
	copy(b, []byte(rwc.testReadData)[:size])
	rwc.testReadData = rwc.testReadData[size:]

	return size, nil
}

func (rwc *readWriteCloser) Close() error {
	return nil
}

func testProtocolResponse() []byte {
	// arduino uno r3 protocol response "2.3"
	return []byte{249, 2, 3}
}

func testFirmwareResponse() []byte {
	// arduino uno r3 firmware response "StandardFirmata.ino"
	return []byte{240, 121, 2, 3, 83, 0, 116, 0, 97, 0, 110, 0, 100, 0, 97,
		0, 114, 0, 100, 0, 70, 0, 105, 0, 114, 0, 109, 0, 97, 0, 116, 0, 97, 0, 46,
		0, 105, 0, 110, 0, 111, 0, 247}
}

func testCapabilitiesResponse() []byte {
	// arduino uno r3 capabilities response
	return []byte{240, 108, 127, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 3,
		8, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 3, 8, 4, 14, 127, 0, 1,
		1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0,
		1, 1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 3, 8, 4, 14, 127, 0, 1, 1, 1, 3, 8,
		4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 4, 14, 127, 0, 1, 1, 1, 2,
		10, 127, 0, 1, 1, 1, 2, 10, 127, 0, 1, 1, 1, 2, 10, 127, 0, 1, 1, 1, 2, 10,
		127, 0, 1, 1, 1, 2, 10, 6, 1, 127, 0, 1, 1, 1, 2, 10, 6, 1, 127, 247}
}

func testAnalogMappingResponse() []byte {
	// arduino uno r3 analog mapping response
	return []byte{240, 106, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127,
		127, 127, 127, 127, 0, 1, 2, 3, 4, 5, 247}
}

func testPinStateReply() [][]byte {
	// mock pin state reply
	return [][]byte{
		// D0
		{START_SYSEX, PIN_STATE_RESPONSE,
			0, PIN_MODE_IGNORE, 0,
			END_SYSEX},
		// D1
		{START_SYSEX, PIN_STATE_RESPONSE,
			1, PIN_MODE_IGNORE, 0,
			END_SYSEX},
		// D2
		{START_SYSEX, PIN_STATE_RESPONSE,
			2, PIN_MODE_INPUT, 1,
			END_SYSEX},
		// D3
		{START_SYSEX, PIN_STATE_RESPONSE,
			3, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D4
		{START_SYSEX, PIN_STATE_RESPONSE,
			4, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D5
		{START_SYSEX, PIN_STATE_RESPONSE,
			5, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D6
		{START_SYSEX, PIN_STATE_RESPONSE,
			6, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D7
		{START_SYSEX, PIN_STATE_RESPONSE,
			7, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D8
		{START_SYSEX, PIN_STATE_RESPONSE,
			8, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D9
		{START_SYSEX, PIN_STATE_RESPONSE,
			9, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D10
		{START_SYSEX, PIN_STATE_RESPONSE,
			10, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D11
		{START_SYSEX, PIN_STATE_RESPONSE,
			11, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D12
		{START_SYSEX, PIN_STATE_RESPONSE,
			12, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// D13
		{START_SYSEX, PIN_STATE_RESPONSE,
			13, PIN_MODE_OUTPUT, 1,
			END_SYSEX},
		// A0
		{START_SYSEX, PIN_STATE_RESPONSE,
			14, PIN_MODE_ANALOG, 1,
			END_SYSEX},
		// A1
		{START_SYSEX, PIN_STATE_RESPONSE,
			15, PIN_MODE_ANALOG, 1,
			END_SYSEX},
		// A2
		{START_SYSEX, PIN_STATE_RESPONSE,
			16, PIN_MODE_ANALOG, 1,
			END_SYSEX},
		// A3
		{START_SYSEX, PIN_STATE_RESPONSE,
			17, PIN_MODE_ANALOG, 1,
			END_SYSEX},
		// A4
		{START_SYSEX, PIN_STATE_RESPONSE,
			18, PIN_MODE_ANALOG, 1,
			END_SYSEX},
		// A5
		{START_SYSEX, PIN_STATE_RESPONSE,
			19, PIN_MODE_ANALOG, 1,
			END_SYSEX},
	}
}

var testPinNames = [20]PinName{
	pb.PinName_PB9,  //D0
	pb.PinName_PB8,  //D1
	pb.PinName_PB7,  //D2
	pb.PinName_PB6,  //D3
	pb.PinName_PB5,  //D4
	pb.PinName_PB4,  //D5
	pb.PinName_PB3,  //D6
	pb.PinName_PA15, //D7
	pb.PinName_PA12, //D8
	pb.PinName_PA11, //D9
	pb.PinName_PA10, //D10
	pb.PinName_PA9,  //D11
	pb.PinName_PA8,  //D12
	pb.PinName_PB15, //D13
	pb.PinName_PB14, //D14
	pb.PinName_PB13, //D15
	pb.PinName_PB12, //D16
	pb.PinName_PC13, //D17
	pb.PinName_PC14, //D18
	pb.PinName_PC15, //D19
}

func testPinNamesReply() []byte {
	// mock pin names reply
	// 0  START_SYSEX                  (0xF0)
	// 1  UD_PIN_NAMES_REPLY           (0x07)
	// 2  pin0 PA0-PZ15(191) bits 0-6  (least significant byte)
	// 4  pin0 PA0-PZ15(191) bits 7-13 (most significant byte)
	// ... pin1 and more
	// N  END_SYSEX                    (0xF7)
	return []byte{
		START_SYSEX, UD_PIN_NAMES_REPLY,
		// pb.PinName_PB9  //D0
		byte(pb.PinName_PB9) & 0x7F, byte(pb.PinName_PB9) >> 7,
		// pb.PinName_PB8  //D1
		byte(pb.PinName_PB8) & 0x7F, byte(pb.PinName_PB8) >> 7,
		// pb.PinName_PB7  //D2
		byte(pb.PinName_PB7) & 0x7F, byte(pb.PinName_PB7) >> 7,
		// pb.PinName_PB6  //D3
		byte(pb.PinName_PB6) & 0x7F, byte(pb.PinName_PB6) >> 7,
		// pb.PinName_PB5  //D4
		byte(pb.PinName_PB5) & 0x7F, byte(pb.PinName_PB5) >> 7,
		// pb.PinName_PB4  //D5
		byte(pb.PinName_PB4) & 0x7F, byte(pb.PinName_PB4) >> 7,
		// pb.PinName_PB3  //D6
		byte(pb.PinName_PB3) & 0x7F, byte(pb.PinName_PB3) >> 7,
		// pb.PinName_PA15 //D7
		byte(pb.PinName_PA15) & 0x7F, byte(pb.PinName_PA15) >> 7,
		// pb.PinName_PA12 //D8
		byte(pb.PinName_PA12) & 0x7F, byte(pb.PinName_PA12) >> 7,
		// pb.PinName_PA11 //D9
		byte(pb.PinName_PA11) & 0x7F, byte(pb.PinName_PA11) >> 7,
		// pb.PinName_PA10 //D10
		byte(pb.PinName_PA10) & 0x7F, byte(pb.PinName_PA10) >> 7,
		// pb.PinName_PA9  //D11
		byte(pb.PinName_PA9) & 0x7F, byte(pb.PinName_PA9) >> 7,
		// pb.PinName_PA8  //D12
		byte(pb.PinName_PA8) & 0x7F, byte(pb.PinName_PA8) >> 7,
		// pb.PinName_PB15 //D13
		byte(pb.PinName_PB15) & 0x7F, byte(pb.PinName_PB15) >> 7,
		// pb.PinName_PB14 //A0
		byte(pb.PinName_PB14) & 0x7F, byte(pb.PinName_PB14) >> 7,
		// pb.PinName_PB13 //A1
		byte(pb.PinName_PB13) & 0x7F, byte(pb.PinName_PB13) >> 7,
		// pb.PinName_PB12 //A2
		byte(pb.PinName_PB12) & 0x7F, byte(pb.PinName_PB12) >> 7,
		// pb.PinName_PC13 //A3
		byte(pb.PinName_PC13) & 0x7F, byte(pb.PinName_PC13) >> 7,
		// pb.PinName_PC14 //A4
		byte(pb.PinName_PC14) & 0x7F, byte(pb.PinName_PC14) >> 7,
		// pb.PinName_PC15 //A5
		byte(pb.PinName_PC15) & 0x7F, byte(pb.PinName_PC15) >> 7,
		END_SYSEX,
	}
}

func processFrame(f *Firmata) error {
	frame, err := f.reader.ReadFrame()
	if err != nil {
		return err
	}
	return f.proccessFrame(frame)
}

func initTestFirmata() (*Firmata, error) {
	f := NewFirmata(new(readWriteCloser), nil)

	bss := [][]byte{
		testProtocolResponse(),
		testFirmwareResponse(),
		testCapabilitiesResponse(),
		testAnalogMappingResponse(),
	}
	bss = append(bss, testPinStateReply()...)
	bss = append(bss, testPinNamesReply())

	for _, s := range bss {
		setTestReadData(f, s)
		frame, err := f.reader.ReadFrame()
		if err != nil {
			return nil, err
		}
		err = f.proccessFrame(frame)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func TestInit(t *testing.T) {
	b, err := initTestFirmata()
	if err != nil {
		t.Fatalf("initTestFirmata should ok, but got %v", err)
	}

	gobottest.Assert(t, b.ProtocolVersion.Server.Name, FirmataProtocolName)
	gobottest.Assert(t, b.FirmwareVersion.Server.Name, "StandardFirmata.ino")
	gobottest.Assert(t, b.TotalPorts, byte(3))
	gobottest.Assert(t, b.TotalPins, byte(20))
	gobottest.Assert(t, b.TotalAnalogPins, byte(6))
	gobottest.Assert(t, b.PortConfigInputs_l, [16]byte{0b00000100})

	for n, d := range b.DxByName {
		gobottest.Assert(t, n, b.Pins[d].Name)
	}
}

func TestReportVersion(t *testing.T) {
	b, _ := initTestFirmata()
	//test if functions executes
	gobottest.Assert(t, b.reportInit_l(), nil)
}

func TestProcessAnalogRead0(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b, []byte{0xE0, 0x23, 0x05})

	sem := make(chan bool, 1)
	b.Config.OnAnalogMessage = func(f *Firmata, pin *Pin) {
		gobottest.Assert(t, pin.Value_l, uint32(675))
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("AnalogRead0 was not published")
	}
}

func TestProcessAnalogRead1(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b, []byte{0xE1, 0x23, 0x06})

	sem := make(chan bool, 1)
	b.Config.OnAnalogMessage = func(f *Firmata, pin *Pin) {
		gobottest.Assert(t, pin.Value_l, uint32(803))
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("AnalogRead1 was not published")
	}
}

func TestProcessDigitalRead2(t *testing.T) {
	b, _ := initTestFirmata()
	b.handlePinMode_l(2, PIN_MODE_INPUT)
	setTestReadData(b, []byte{0x90, 0x04, 0x00})

	sem := make(chan bool, 1)
	b.Config.OnDigitalMessage = func(f *Firmata, port byte, pins byte, values byte) {
		gobottest.Assert(t, port, byte(0))
		gobottest.Assert(t, pins, byte(0b00000100))
		gobottest.Assert(t, f.Pins[2].Value_l, uint32(1))
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("DigitalRead2 was not published")
	}
}

func TestProcessDigitalRead4(t *testing.T) {
	b, _ := initTestFirmata()
	b.handlePinMode_l(4, PIN_MODE_INPUT)
	setTestReadData(b, []byte{DIGITAL_MESSAGE, 0b00010000, 0x00})

	sem := make(chan bool, 1)
	b.Config.OnDigitalMessage = func(f *Firmata, port byte, pins byte, values byte) {
		gobottest.Assert(t, port, byte(0))
		gobottest.Assert(t, pins, byte(0b00010000))
		gobottest.Assert(t, f.Pins[4].Value_l, uint32(1))
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("DigitalRead4 was not published")
	}
}

func TestDigitalWrite(t *testing.T) {
	b, _ := initTestFirmata()
	_, err := b.DigitalWrite_l(2, 0b00001111)
	gobottest.Assert(t, err, nil)
}

func TestSetPinMode(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.SetPinMode_l(13, PIN_MODE_OUTPUT), nil)
}

func TestAnalogWrite(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.AnalogWrite_l(0, 128), nil)
}

func TestReportAnalog(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.ReportAnalog_l(0, true), nil)
	gobottest.Assert(t, b.ReportAnalog_l(0, false), nil)
}

func TestProcessPinState13(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b, []byte{240, 110, 13, 1, 1, 247})

	sem := make(chan bool, 1)
	b.Config.OnPinState = func(f *Firmata, pin *Pin) {
		gobottest.Assert(t, *pin, Pin{
			Dx:   13,
			Name: pb.PinName_PB15,
			Modes: map[byte]byte{
				0: 127,
				1: 1,
				4: 1,
			},
			Mode_l:  1,
			Value_l: 0,
			State_l: 1,
			Ax:      127,
		})
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("PinState13 was not published")
	}
}

func TestI2cConfig(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.I2cConfig_l(100), nil)
}

func TestI2cWrite(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.I2cWrite_l(0x00, []byte{0x01, 0x02}), nil)
}

func TestI2cRead(t *testing.T) {
	b, _ := initTestFirmata()
	gobottest.Assert(t, b.I2cRead_l(0x00, false, false, 10), nil)
}

func TestProcessI2cReply(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b, []byte{240, 119, 9, 0, 0, 0, 24, 1, 1, 0, 26, 1, 247})

	sem := make(chan bool, 1)
	b.Config.OnI2cReply = func(f *Firmata, reply *I2cReply) {
		gobottest.Assert(t, *reply, I2cReply{
			Address:  9,
			Register: 0,
			Data:     []byte{152, 1, 154},
		})
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("I2cReply was not published")
	}
}

func TestProcessStringData(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b,
		append([]byte{240, 0x71},
			append(To14bits([]byte("Hello Firmata!")), 247)...))

	sem := make(chan bool, 1)
	b.Config.OnStringData = func(f *Firmata, buf []byte) {
		gobottest.Assert(t, string(buf), "Hello Firmata!")
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("StringData was not published")
	}
}

func TestServoConfig(t *testing.T) {
	b, _ := initTestFirmata()

	tests := []struct {
		description string
		arguments   [3]int32
		expected    []byte
		result      error
	}{
		{
			description: "Min values for min & max",
			arguments:   [3]int32{9, 0, 0},
			expected:    []byte{0xF0, 0x70, 9, 0, 0, 0, 0, 0xF7},
		},
		{
			description: "Max values for min & max",
			arguments:   [3]int32{9, 0x3FFF, 0x3FFF},
			expected:    []byte{0xF0, 0x70, 9, 0x7F, 0x7F, 0x7F, 0x7F, 0xF7},
		},
		{
			description: "Clipped max values for min & max",
			arguments:   [3]int32{9, 0xFFFF, 0xFFFF},
			expected:    []byte{0xF0, 0x70, 9, 0x7F, 0x7F, 0x7F, 0x7F, 0xF7},
		},
	}

	rwc := b.closer.(*readWriteCloser)

	for _, test := range tests {
		rwc.writeDataMutex.Lock()
		rwc.testWriteData.Reset()
		rwc.writeDataMutex.Unlock()
		err := b.ServoConfig_l(byte(test.arguments[0]), test.arguments[1], test.arguments[2])
		rwc.writeDataMutex.Lock()
		gobottest.Assert(t, rwc.testWriteData.Bytes(), test.expected)
		gobottest.Assert(t, err, test.result)
		rwc.writeDataMutex.Unlock()
	}
}

func TestProcessSysexData(t *testing.T) {
	b, _ := initTestFirmata()
	setTestReadData(b, []byte{240, 17, 1, 2, 3, 247})

	sem := make(chan bool, 1)
	b.Config.OnSysexResponse = func(f *Firmata, buf []byte) {
		gobottest.Assert(t, buf, []byte{17, 1, 2, 3})
		sem <- true
	}

	err := processFrame(b)
	if err != nil {
		t.Fatalf("processFrame should ok, but got %v", err)
	}

	select {
	case <-sem:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("SysexResponse was not published")
	}
}
