package firmata

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"gobot.io/x/gobot/gobottest"
)

type readWriteCloser struct{}

func (readWriteCloser) Write(p []byte) (int, error) {
	writeDataMutex.Lock()
	defer writeDataMutex.Unlock()
	return testWriteData.Write(p)
}

var writeDataMutex sync.Mutex
var readDataMutex sync.Mutex
var testReadData = []byte{}
var testWriteData = bytes.Buffer{}

func SetTestReadData(d []byte) {
	readDataMutex.Lock()
	defer readDataMutex.Unlock()
	testReadData = d
}

func (readWriteCloser) Read(b []byte) (int, error) {
	readDataMutex.Lock()
	defer readDataMutex.Unlock()

	size := len(b)
	if len(testReadData) < size {
		size = len(testReadData)
	}
	copy(b, []byte(testReadData)[:size])
	testReadData = testReadData[size:]

	return size, nil
}

func (readWriteCloser) Close() error {
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

func processFrame(f *Firmata) error {
	frame, err := f.reader.ReadFrame()
	if err != nil {
		return err
	}
	return f.proccessFrame(frame)
}

func initTestFirmata() (*Firmata, error) {
	f := NewFirmata(readWriteCloser{})

	for _, fn := range []func() []byte{
		testProtocolResponse,
		testFirmwareResponse,
		testCapabilitiesResponse,
		testAnalogMappingResponse,
	} {
		SetTestReadData(fn())
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

	gobottest.Assert(t, b.ProtocolVersion.Server.Name, "v2.3")
	gobottest.Assert(t, b.FirmwareVersion.Server.Name, "v2.3")
	gobottest.Assert(t, string(b.FirmwareName), "StandardFirmata.ino")
	gobottest.Assert(t, len(b.Pins_l), 20)
	gobottest.Assert(t, len(b.AnalogPins_l), 6)
}

func TestReportVersion(t *testing.T) {
	b, _ := initTestFirmata()
	//test if functions executes
	gobottest.Assert(t, b.reportVersion(), nil)
}

func TestProcessAnalogRead0(t *testing.T) {
	b, _ := initTestFirmata()
	SetTestReadData([]byte{0xE0, 0x23, 0x05})

	sem := make(chan bool, 1)
	b.OnAnalogMessage = func(pin *Pin) {
		gobottest.Assert(t, pin.Value, 675)
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
	SetTestReadData([]byte{0xE1, 0x23, 0x06})

	sem := make(chan bool, 1)
	b.OnAnalogMessage = func(pin *Pin) {
		gobottest.Assert(t, pin.Value, 803)
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
	b.Pins_l[2].Mode = PIN_MODE_INPUT
	SetTestReadData([]byte{0x90, 0x04, 0x00})

	sem := make(chan bool, 1)
	b.OnDigitalMessage = func(pins []*Pin) {
		gobottest.Assert(t, len(pins), 1)
		gobottest.Assert(t, pins[0].ID, byte(2))
		gobottest.Assert(t, pins[0].Value, 1)
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
	b.Pins_l[4].Mode = PIN_MODE_INPUT
	SetTestReadData([]byte{0x90, 0x16, 0x00})

	sem := make(chan bool, 1)
	b.OnDigitalMessage = func(pins []*Pin) {
		gobottest.Assert(t, len(pins), 1)
		gobottest.Assert(t, pins[0].ID, byte(4))
		gobottest.Assert(t, pins[0].Value, 1)
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
	values := [8]byte{1, 1, 1, 1, 0, 0, 0, 0}
	gobottest.Assert(t, b.DigitalWrite_l(2, &values), nil)
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
	gobottest.Assert(t, b.ReportAnalog_l(0, 1), nil)
	gobottest.Assert(t, b.ReportAnalog_l(0, 0), nil)
}

func TestProcessPinState13(t *testing.T) {
	b, _ := initTestFirmata()
	SetTestReadData([]byte{240, 110, 13, 1, 1, 247})

	sem := make(chan bool, 1)
	b.OnPinState = func(pin *Pin) {
		gobottest.Assert(t, *pin, Pin{
			ID: 13,
			Modes: map[byte]byte{
				0: 127,
				1: 1,
				4: 1,
			},
			Mode:          1,
			Value:         0,
			State:         1,
			AnalogChannel: 127,
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
	gobottest.Assert(t, b.I2cRead_l(0x00, 10), nil)
}

func TestProcessI2cReply(t *testing.T) {
	b, _ := initTestFirmata()
	SetTestReadData([]byte{240, 119, 9, 0, 0, 0, 24, 1, 1, 0, 26, 1, 247})

	sem := make(chan bool, 1)
	b.OnI2cReply = func(reply *I2cReply) {
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
	SetTestReadData(append([]byte{240, 0x71}, append([]byte("Hello Firmata!"), 247)...))

	sem := make(chan bool, 1)
	b.OnStringData = func(buf []byte) {
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
		arguments   [3]int
		expected    []byte
		result      error
	}{
		{
			description: "Min values for min & max",
			arguments:   [3]int{9, 0, 0},
			expected:    []byte{0xF0, 0x70, 9, 0, 0, 0, 0, 0xF7},
		},
		{
			description: "Max values for min & max",
			arguments:   [3]int{9, 0x3FFF, 0x3FFF},
			expected:    []byte{0xF0, 0x70, 9, 0x7F, 0x7F, 0x7F, 0x7F, 0xF7},
		},
		{
			description: "Clipped max values for min & max",
			arguments:   [3]int{9, 0xFFFF, 0xFFFF},
			expected:    []byte{0xF0, 0x70, 9, 0x7F, 0x7F, 0x7F, 0x7F, 0xF7},
		},
	}

	for _, test := range tests {
		writeDataMutex.Lock()
		testWriteData.Reset()
		writeDataMutex.Unlock()
		err := b.ServoConfig_l(byte(test.arguments[0]), test.arguments[1], test.arguments[2])
		writeDataMutex.Lock()
		gobottest.Assert(t, testWriteData.Bytes(), test.expected)
		gobottest.Assert(t, err, test.result)
		writeDataMutex.Unlock()
	}
}

func TestProcessSysexData(t *testing.T) {
	b, _ := initTestFirmata()
	SetTestReadData([]byte{240, 17, 1, 2, 3, 247})

	sem := make(chan bool, 1)
	b.OnSysexResponse = func(buf []byte) {
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
