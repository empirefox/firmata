package firmata

const (
	FIRMATA_PROTOCOL_MAJOR_VERSION  byte = 2 // for non-compatible changes
	FIRMATA_PROTOCOL_MINOR_VERSION  byte = 6 // for backwards compatible changes
	FIRMATA_PROTOCOL_BUGFIX_VERSION byte = 0 // for bugfix releases

	//  Version numbers for the Firmata library.
	//  ConfigurableFirmata 2.10.1 implements version 2.6.0 of the Firmata protocol.
	//  The firmware version will not always equal the protocol version going forward.
	//  Query using the REPORT_FIRMWARE message.
	FIRMATA_FIRMWARE_MAJOR_VERSION  byte = 2  // for non-compatible changes
	FIRMATA_FIRMWARE_MINOR_VERSION  byte = 10 // for backwards compatible changes
	FIRMATA_FIRMWARE_BUGFIX_VERSION byte = 1  // for bugfix releases

	MAX_DATA_BYTES int = 64 // max number of data bytes in incoming messages

	// message command bytes (128-255/0x80-0xFF)
	DIGITAL_MESSAGE byte = 0x90 // send data for a digital pin
	ANALOG_MESSAGE  byte = 0xE0 // send data for an analog pin (or PWM)
	REPORT_ANALOG   byte = 0xC0 // enable analog input by pin #
	REPORT_DIGITAL  byte = 0xD0 // enable digital input by port pair
	//
	SET_PIN_MODE          byte = 0xF4 // set a pin to INPUT/OUTPUT/PWM/etc
	SET_DIGITAL_PIN_VALUE byte = 0xF5 // set value of an individual digital pin
	//
	REPORT_VERSION byte = 0xF9 // report protocol version
	SYSTEM_RESET   byte = 0xFF // reset from MIDI
	//
	START_SYSEX byte = 0xF0 // start a MIDI Sysex message
	END_SYSEX   byte = 0xF7 // end a MIDI Sysex message

	// extended command set using sysex (0-127/0x00-0x7F)
	/* 0x00-0x0F reserved for user-defined commands */
	SERIAL_MESSAGE          byte = 0x60 // communicate with serial devices, including other boards
	ENCODER_DATA            byte = 0x61 // reply with encoders current positions
	ACCELSTEPPER_DATA       byte = 0x62 // control a stepper motor
	SERVO_CONFIG            byte = 0x70 // set max angle, minPulse, maxPulse, freq
	STRING_DATA             byte = 0x71 // a string message with 14-bits per char
	STEPPER_DATA            byte = 0x72 // control a stepper motor
	ONEWIRE_DATA            byte = 0x73 // send an OneWire read/write/reset/select/skip/search request
	SHIFT_DATA              byte = 0x75 // a bitstream to/from a shift register
	I2C_REQUEST             byte = 0x76 // send an I2C read/write request
	I2C_REPLY               byte = 0x77 // a reply to an I2C read request
	I2C_CONFIG              byte = 0x78 // config I2C settings such as delay times and power pins
	EXTENDED_ANALOG         byte = 0x6F // analog write (PWM, Servo, etc) to any pin
	PIN_STATE_QUERY         byte = 0x6D // ask for a pin's current mode and value
	PIN_STATE_RESPONSE      byte = 0x6E // reply with pin's current mode and value
	CAPABILITY_QUERY        byte = 0x6B // ask for supported modes and resolution of all pins
	CAPABILITY_RESPONSE     byte = 0x6C // reply with supported modes and resolution
	ANALOG_MAPPING_QUERY    byte = 0x69 // ask for mapping of analog to pin numbers
	ANALOG_MAPPING_RESPONSE byte = 0x6A // reply with mapping info
	REPORT_FIRMWARE         byte = 0x79 // report name and version of the firmware
	SAMPLING_INTERVAL       byte = 0x7A // set the poll rate of the main loop
	SCHEDULER_DATA          byte = 0x7B // send a createtask/deletetask/addtotask/schedule/querytasks/querytask request to the scheduler
	SYSEX_NON_REALTIME      byte = 0x7E // MIDI Reserved for non-realtime messages
	SYSEX_REALTIME          byte = 0x7F // MIDI Reserved for realtime messages

	// pin modes
	PIN_MODE_INPUT   byte = 0x00 // defined in Arduino.h
	PIN_MODE_OUTPUT  byte = 0x01 // defined in Arduino.h
	PIN_MODE_ANALOG  byte = 0x02 // analog pin in analogInput mode
	PIN_MODE_PWM     byte = 0x03 // digital pin in PWM output mode
	PIN_MODE_SERVO   byte = 0x04 // digital pin in Servo output mode
	PIN_MODE_SHIFT   byte = 0x05 // shiftIn/shiftOut mode
	PIN_MODE_I2C     byte = 0x06 // pin included in I2C setup
	PIN_MODE_ONEWIRE byte = 0x07 // pin configured for 1-wire
	PIN_MODE_STEPPER byte = 0x08 // pin configured for stepper motor
	PIN_MODE_ENCODER byte = 0x09 // pin configured for rotary encoders
	PIN_MODE_SERIAL  byte = 0x0A // pin configured for serial communication
	PIN_MODE_PULLUP  byte = 0x0B // enable internal pull-up resistor for pin
	PIN_MODE_IGNORE  byte = 0x7F // pin configured to be ignored by digitalWrite and capabilityResponse
	TOTAL_PIN_MODES  byte = 13

	MaxRecvSize int = 1373
)
