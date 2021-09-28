package firmata

func Make14bits(len14 int) []byte {
	if len14%2 != 0 {
		// the last one equals 0
		// TODO log here?
		len14++
	}
	return make([]byte, len14)
}

func From14bits(in []byte) (out []byte) {
	if len(in)%2 != 0 {
		in = append(in, 0)
	}

	lenOut := len(in) / 2
	out = make([]byte, lenOut)
	for i := 0; i < lenOut; i++ {
		j := i * 2
		out[i] = (in[j] & 0x7F) | ((in[j+1] & 0x7F) << 7)
	}

	return out
}
