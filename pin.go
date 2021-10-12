package firmata

// Pin represents a pin on the firmata board
type Pin struct {
	// Dx is the firmata digital pin.
	Dx byte
	// Ax is the firmata analog pin.
	Ax byte
	// Name of pin.
	Name PinName
	// Modes: modes/resolutions.
	Modes   map[byte]byte
	Mode_l  byte
	Value_l uint
	State_l uint
}

func (pin *Pin) SupportMode(mode byte) (ok bool) {
	_, ok = pin.Modes[mode]
	return
}

func (pin *Pin) IsDigital_l() bool {
	return pin.Mode_l != PIN_MODE_SERIAL
}

func (pin *Pin) IsSerial_l() bool {
	return pin.Mode_l == PIN_MODE_SERIAL
}

func (pin *Pin) IsAnalog() bool {
	return pin.Ax != 127
}

func (pin *Pin) Clone_l() *Pin {
	clone := *pin
	clone.Modes = make(map[byte]byte, len(pin.Modes))
	for k, v := range pin.Modes {
		clone.Modes[k] = v
	}
	return &clone
}

func ClonePins_l(pins []*Pin) []*Pin {
	clone := make([]*Pin, len(pins))
	for i, pin := range pins {
		clone[i] = pin.Clone_l()
	}
	return clone
}
