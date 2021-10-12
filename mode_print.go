package firmata

import "strconv"

var pinModeSimpleViews = [TOTAL_PIN_MODES]string{
	// pin modes
	"I",     // = 0x00 // INPUT is defined in Arduino.h, but may conflict with other uses
	"O",     // = 0x01 // OUTPUT is defined in Arduino.h. Careful: OUTPUT is defined as 2 on ESP32! therefore OUTPUT and OUTPU not the same!
	"A",     // = 0x02 // analog pin in analogInput mode
	"PWM",   // = 0x03 // digital pin in PWM output mode
	"SERVO", // = 0x04 // digital pin in Servo output mode
	"IxO",   // = 0x05 // shiftIn/shiftOut mode
	"I2C",   // = 0x06 // pin included in I2C setup
	"W1",    // = 0x07 // pin configured for 1-wire
	"SM",    // = 0x08 // pin configured for stepper motor
	"RE",    // = 0x09 // pin configured for rotary encoders
	"UART",  // = 0x0A // pin configured for serial communication
	"PU",    // = 0x0B // enable internal pull-up resistor for pin
	// Extensions under development
	"SPI",   // = 0x0C // pin configured for SPI
	"SONAR", // = 0x0D // pin configured for HC-SR04
	"TONE",  // = 0x0E // pin configured for tone
	"DHT",   // = 0x0F // pin configured for DHT
	"FREQ",  // = 0x10 // pin configured for frequency measurement
	// "X",     // = 0x7F // pin configured to be ignored by digitalWrite and capabilityResponse
}

func PinModeSimpleView(mode byte) string {
	if mode == 0x7F {
		return "X"
	}
	if mode > 0x10 {
		return unknownModeView(mode)
	}
	return pinModeSimpleViews[mode]
}

var pinModeNormalViews = [TOTAL_PIN_MODES]string{
	// pin modes
	"INPUT",   // = 0x00 // INPUT is defined in Arduino.h, but may conflict with other uses
	"OUTPUT",  // = 0x01 // OUTPUT is defined in Arduino.h. Careful: OUTPUT is defined as 2 on ESP32! therefore OUTPUT and OUTPU not the same!
	"ANALOG",  // = 0x02 // analog pin in analogInput mode
	"PWM",     // = 0x03 // digital pin in PWM output mode
	"SERVO",   // = 0x04 // digital pin in Servo output mode
	"SHIFT",   // = 0x05 // shiftIn/shiftOut mode
	"I2C",     // = 0x06 // pin included in I2C setup
	"ONEWIRE", // = 0x07 // pin configured for 1-wire
	"STEPPER", // = 0x08 // pin configured for stepper motor
	"ENCODER", // = 0x09 // pin configured for rotary encoders
	"SERIAL",  // = 0x0A // pin configured for serial communication
	"PULLUP",  // = 0x0B // enable internal pull-up resistor for pin
	// Extensions under development
	"SPI",       // = 0x0C // pin configured for SPI
	"SONAR",     // = 0x0D // pin configured for HC-SR04
	"TONE",      // = 0x0E // pin configured for tone
	"DHT",       // = 0x0F // pin configured for DHT
	"FREQUENCY", // = 0x10 // pin configured for frequency measurement
	// "IGNORE", // = 0x7F // pin configured to be ignored by digitalWrite and capabilityResponse
}

func PinModeNormalView(mode byte) string {
	if mode == 0x7F {
		return "IGNORE"
	}
	if mode > 0x10 {
		return unknownModeView(mode)
	}
	return pinModeNormalViews[mode]
}

func unknownModeView(mode byte) string {
	return "M_0x" + strconv.FormatInt(int64(mode), 16) + "?"
}
