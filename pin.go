package firmata

// Pin represents a pin on the firmata board
type Pin struct {
	ID            int
	Modes         map[int]byte
	Mode          int
	Value         int
	State         int
	AnalogChannel int
}
