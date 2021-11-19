package firmata

import "github.com/empirefox/firmata/pkg/pb"

type PinName = pb.PinName

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
	Value_l uint32
	State_l uint32
}

func (pin *Pin) ToPb_l() *pb.Instance_Pin {
	modes := make([]*pb.Instance_SupportedMode, 0, len(pin.Modes))
	for i := PIN_MODE_INPUT; i < TOTAL_PIN_MODES; i++ {
		if v, ok := pin.Modes[i]; ok {
			modes = append(modes, &pb.Instance_SupportedMode{
				Mode:       pb.Mode(i),
				Resolution: uint32(v),
			})
		}
	}
	return &pb.Instance_Pin{
		Dx:    uint32(pin.Dx),
		Ax:    uint32(pin.Ax),
		Name:  pin.Name,
		Modes: modes,
		Mode:  pb.Mode(pin.Mode_l),
		Value: pin.Value_l,
		State: pin.State_l,
	}
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
