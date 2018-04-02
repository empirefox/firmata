package firmata

// Pin represents a pin on the firmata board
type Pin struct {
	ID            byte
	Modes         map[byte]byte
	Mode          byte
	Value         int
	State         int
	AnalogChannel byte
}

func (pin *Pin) Clone() *Pin {
	clone := *pin
	clone.Modes = make(map[byte]byte, len(pin.Modes))
	for k, v := range pin.Modes {
		clone.Modes[k] = v
	}
	return &clone
}

func ClonePins(pins []*Pin) []*Pin {
	clone := make([]*Pin, len(pins))
	for i, pin := range pins {
		clone[i] = pin.Clone()
	}
	return clone
}
