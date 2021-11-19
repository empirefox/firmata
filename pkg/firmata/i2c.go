package firmata

// I2cReply represents the response from an I2cReply message
type I2cReply struct {
	Address  int
	Register int
	Data     []byte
}
