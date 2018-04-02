package firmata

// I2cReply represents the response from an I2cReply message
type I2cReply struct {
	Address  byte
	Register int
	Data     []byte
}
