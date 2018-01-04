package firmata

const (
	FIRMWARE_MAJOR_VERSION  byte = 2
	FIRMWARE_MINOR_VERSION       = 5
	FIRMWARE_BUGFIX_VERSION      = 7

	PROTOCOL_MAJOR_VERSION  = 2 // for non-compatible changes
	PROTOCOL_MINOR_VERSION  = 5 // for backwards compatible changes
	PROTOCOL_BUGFIX_VERSION = 1 // for bugfix releases

	MAX_DATA_BYTES = 64 // max number of data bytes in incoming messages

	// message command bytes (128-255/0x80-0xFF)

	DIGITAL_MESSAGE = 0x90 // send data for a digital port (collection of 8 pins)
	ANALOG_MESSAGE  = 0xE0 // send data for an analog pin (or PWM)
	REPORT_ANALOG   = 0xC0 // enable analog input by pin #
	REPORT_DIGITAL  = 0xD0 // enable digital input by port pair
	//
	SET_PIN_MODE          = 0xF4 // set a pin to INPUT/OUTPUT/PWM/etc
	SET_DIGITAL_PIN_VALUE = 0xF5 // set value of an individual digital pin
	//
	REPORT_VERSION = 0xF9 // report protocol version
	SYSTEM_RESET   = 0xFF // reset from MIDI
	//
	START_SYSEX = 0xF0 // start a MIDI Sysex message
	END_SYSEX   = 0xF7 // end a MIDI Sysex message

	// extended command set using sysex (0-127/0x00-0x7F)
	//* 0x00-0x0F reserved for user-defined commands *//

	SERIAL_DATA             = 0x60 // communicate with serial devices, including other boards
	ENCODER_DATA            = 0x61 // reply with encoders current positions
	SERVO_CONFIG            = 0x70 // set max angle, minPulse, maxPulse, freq
	STRING_DATA             = 0x71 // a string message with 14-bits per char
	STEPPER_DATA            = 0x72 // control a stepper motor
	ONEWIRE_DATA            = 0x73 // send an OneWire read/write/reset/select/skip/search request
	SHIFT_DATA              = 0x75 // a bitstream to/from a shift register
	I2C_REQUEST             = 0x76 // send an I2C read/write request
	I2C_REPLY               = 0x77 // a reply to an I2C read request
	I2C_CONFIG              = 0x78 // config I2C settings such as delay times and power pins
	REPORT_FIRMWARE         = 0x79 // report name and version of the firmware
	EXTENDED_ANALOG         = 0x6F // analog write (PWM, Servo, etc) to any pin
	PIN_STATE_QUERY         = 0x6D // ask for a pin's current mode and value
	PIN_STATE_RESPONSE      = 0x6E // reply with pin's current mode and value
	CAPABILITY_QUERY        = 0x6B // ask for supported modes and resolution of all pins
	CAPABILITY_RESPONSE     = 0x6C // reply with supported modes and resolution
	ANALOG_MAPPING_QUERY    = 0x69 // ask for mapping of analog to pin numbers
	ANALOG_MAPPING_RESPONSE = 0x6A // reply with mapping info
	SAMPLING_INTERVAL       = 0x7A // set the poll rate of the main loop
	SCHEDULER_DATA          = 0x7B // send a createtask/deletetask/addtotask/schedule/querytasks/querytask request to the scheduler
	SYSEX_NON_REALTIME      = 0x7E // MIDI Reserved for non-realtime messages
	SYSEX_REALTIME          = 0x7F // MIDI Reserved for realtime messages

	// pin modes
	PIN_MODE_INPUT   int = 0x00 // same as INPUT defined in Arduino.h
	PIN_MODE_OUTPUT      = 0x01 // same as OUTPUT defined in Arduino.h
	PIN_MODE_ANALOG      = 0x02 // analog pin in analogInput mode
	PIN_MODE_PWM         = 0x03 // digital pin in PWM output mode
	PIN_MODE_SERVO       = 0x04 // digital pin in Servo output mode
	PIN_MODE_SHIFT       = 0x05 // shiftIn/shiftOut mode
	PIN_MODE_I2C         = 0x06 // pin included in I2C setup
	PIN_MODE_ONEWIRE     = 0x07 // pin configured for 1-wire
	PIN_MODE_STEPPER     = 0x08 // pin configured for stepper motor
	PIN_MODE_ENCODER     = 0x09 // pin configured for rotary encoders
	PIN_MODE_SERIAL      = 0x0A // pin configured for serial communication
	PIN_MODE_PULLUP      = 0x0B // enable internal pull-up resistor for pin
	PIN_MODE_IGNORE      = 0x7F // pin configured to be ignored by digitalWrite and capabilityResponse

	TOTAL_PIN_MODES = 13
)
