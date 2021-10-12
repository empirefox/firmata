package firmata

import (
	"fmt"
	"strconv"
)

const (
	PortA = iota
	PortB
	PortC
	PortD
	PortE
	PortF
	PortG
	PortH
	PortI
	PortJ
	PortK
	PortZ
)

type PinName byte

// Pin name definition
const (
	PA0  PinName = (PortA << 4) + 0x00
	PA1  PinName = (PortA << 4) + 0x01
	PA2  PinName = (PortA << 4) + 0x02
	PA3  PinName = (PortA << 4) + 0x03
	PA4  PinName = (PortA << 4) + 0x04
	PA5  PinName = (PortA << 4) + 0x05
	PA6  PinName = (PortA << 4) + 0x06
	PA7  PinName = (PortA << 4) + 0x07
	PA8  PinName = (PortA << 4) + 0x08
	PA9  PinName = (PortA << 4) + 0x09
	PA10 PinName = (PortA << 4) + 0x0A
	PA11 PinName = (PortA << 4) + 0x0B
	PA12 PinName = (PortA << 4) + 0x0C
	PA13 PinName = (PortA << 4) + 0x0D
	PA14 PinName = (PortA << 4) + 0x0E
	PA15 PinName = (PortA << 4) + 0x0F

	PB0  PinName = (PortB << 4) + 0x00
	PB1  PinName = (PortB << 4) + 0x01
	PB2  PinName = (PortB << 4) + 0x02
	PB3  PinName = (PortB << 4) + 0x03
	PB4  PinName = (PortB << 4) + 0x04
	PB5  PinName = (PortB << 4) + 0x05
	PB6  PinName = (PortB << 4) + 0x06
	PB7  PinName = (PortB << 4) + 0x07
	PB8  PinName = (PortB << 4) + 0x08
	PB9  PinName = (PortB << 4) + 0x09
	PB10 PinName = (PortB << 4) + 0x0A
	PB11 PinName = (PortB << 4) + 0x0B
	PB12 PinName = (PortB << 4) + 0x0C
	PB13 PinName = (PortB << 4) + 0x0D
	PB14 PinName = (PortB << 4) + 0x0E
	PB15 PinName = (PortB << 4) + 0x0F

	PC0  PinName = (PortC << 4) + 0x00
	PC1  PinName = (PortC << 4) + 0x01
	PC2  PinName = (PortC << 4) + 0x02
	PC3  PinName = (PortC << 4) + 0x03
	PC4  PinName = (PortC << 4) + 0x04
	PC5  PinName = (PortC << 4) + 0x05
	PC6  PinName = (PortC << 4) + 0x06
	PC7  PinName = (PortC << 4) + 0x07
	PC8  PinName = (PortC << 4) + 0x08
	PC9  PinName = (PortC << 4) + 0x09
	PC10 PinName = (PortC << 4) + 0x0A
	PC11 PinName = (PortC << 4) + 0x0B
	PC12 PinName = (PortC << 4) + 0x0C
	PC13 PinName = (PortC << 4) + 0x0D
	PC14 PinName = (PortC << 4) + 0x0E
	PC15 PinName = (PortC << 4) + 0x0F

	PD0  PinName = (PortD << 4) + 0x00
	PD1  PinName = (PortD << 4) + 0x01
	PD2  PinName = (PortD << 4) + 0x02
	PD3  PinName = (PortD << 4) + 0x03
	PD4  PinName = (PortD << 4) + 0x04
	PD5  PinName = (PortD << 4) + 0x05
	PD6  PinName = (PortD << 4) + 0x06
	PD7  PinName = (PortD << 4) + 0x07
	PD8  PinName = (PortD << 4) + 0x08
	PD9  PinName = (PortD << 4) + 0x09
	PD10 PinName = (PortD << 4) + 0x0A
	PD11 PinName = (PortD << 4) + 0x0B
	PD12 PinName = (PortD << 4) + 0x0C
	PD13 PinName = (PortD << 4) + 0x0D
	PD14 PinName = (PortD << 4) + 0x0E
	PD15 PinName = (PortD << 4) + 0x0F

	PE0  PinName = (PortE << 4) + 0x00
	PE1  PinName = (PortE << 4) + 0x01
	PE2  PinName = (PortE << 4) + 0x02
	PE3  PinName = (PortE << 4) + 0x03
	PE4  PinName = (PortE << 4) + 0x04
	PE5  PinName = (PortE << 4) + 0x05
	PE6  PinName = (PortE << 4) + 0x06
	PE7  PinName = (PortE << 4) + 0x07
	PE8  PinName = (PortE << 4) + 0x08
	PE9  PinName = (PortE << 4) + 0x09
	PE10 PinName = (PortE << 4) + 0x0A
	PE11 PinName = (PortE << 4) + 0x0B
	PE12 PinName = (PortE << 4) + 0x0C
	PE13 PinName = (PortE << 4) + 0x0D
	PE14 PinName = (PortE << 4) + 0x0E
	PE15 PinName = (PortE << 4) + 0x0F

	PF0  PinName = (PortF << 4) + 0x00
	PF1  PinName = (PortF << 4) + 0x01
	PF2  PinName = (PortF << 4) + 0x02
	PF3  PinName = (PortF << 4) + 0x03
	PF4  PinName = (PortF << 4) + 0x04
	PF5  PinName = (PortF << 4) + 0x05
	PF6  PinName = (PortF << 4) + 0x06
	PF7  PinName = (PortF << 4) + 0x07
	PF8  PinName = (PortF << 4) + 0x08
	PF9  PinName = (PortF << 4) + 0x09
	PF10 PinName = (PortF << 4) + 0x0A
	PF11 PinName = (PortF << 4) + 0x0B
	PF12 PinName = (PortF << 4) + 0x0C
	PF13 PinName = (PortF << 4) + 0x0D
	PF14 PinName = (PortF << 4) + 0x0E
	PF15 PinName = (PortF << 4) + 0x0F

	PG0  PinName = (PortG << 4) + 0x00
	PG1  PinName = (PortG << 4) + 0x01
	PG2  PinName = (PortG << 4) + 0x02
	PG3  PinName = (PortG << 4) + 0x03
	PG4  PinName = (PortG << 4) + 0x04
	PG5  PinName = (PortG << 4) + 0x05
	PG6  PinName = (PortG << 4) + 0x06
	PG7  PinName = (PortG << 4) + 0x07
	PG8  PinName = (PortG << 4) + 0x08
	PG9  PinName = (PortG << 4) + 0x09
	PG10 PinName = (PortG << 4) + 0x0A
	PG11 PinName = (PortG << 4) + 0x0B
	PG12 PinName = (PortG << 4) + 0x0C
	PG13 PinName = (PortG << 4) + 0x0D
	PG14 PinName = (PortG << 4) + 0x0E
	PG15 PinName = (PortG << 4) + 0x0F

	PH0  PinName = (PortH << 4) + 0x00
	PH1  PinName = (PortH << 4) + 0x01
	PH2  PinName = (PortH << 4) + 0x02
	PH3  PinName = (PortH << 4) + 0x03
	PH4  PinName = (PortH << 4) + 0x04
	PH5  PinName = (PortH << 4) + 0x05
	PH6  PinName = (PortH << 4) + 0x06
	PH7  PinName = (PortH << 4) + 0x07
	PH8  PinName = (PortH << 4) + 0x08
	PH9  PinName = (PortH << 4) + 0x09
	PH10 PinName = (PortH << 4) + 0x0A
	PH11 PinName = (PortH << 4) + 0x0B
	PH12 PinName = (PortH << 4) + 0x0C
	PH13 PinName = (PortH << 4) + 0x0D
	PH14 PinName = (PortH << 4) + 0x0E
	PH15 PinName = (PortH << 4) + 0x0F

	PI0  PinName = (PortI << 4) + 0x00
	PI1  PinName = (PortI << 4) + 0x01
	PI2  PinName = (PortI << 4) + 0x02
	PI3  PinName = (PortI << 4) + 0x03
	PI4  PinName = (PortI << 4) + 0x04
	PI5  PinName = (PortI << 4) + 0x05
	PI6  PinName = (PortI << 4) + 0x06
	PI7  PinName = (PortI << 4) + 0x07
	PI8  PinName = (PortI << 4) + 0x08
	PI9  PinName = (PortI << 4) + 0x09
	PI10 PinName = (PortI << 4) + 0x0A
	PI11 PinName = (PortI << 4) + 0x0B
	PI12 PinName = (PortI << 4) + 0x0C
	PI13 PinName = (PortI << 4) + 0x0D
	PI14 PinName = (PortI << 4) + 0x0E
	PI15 PinName = (PortI << 4) + 0x0F

	PJ0  PinName = (PortJ << 4) + 0x00
	PJ1  PinName = (PortJ << 4) + 0x01
	PJ2  PinName = (PortJ << 4) + 0x02
	PJ3  PinName = (PortJ << 4) + 0x03
	PJ4  PinName = (PortJ << 4) + 0x04
	PJ5  PinName = (PortJ << 4) + 0x05
	PJ6  PinName = (PortJ << 4) + 0x06
	PJ7  PinName = (PortJ << 4) + 0x07
	PJ8  PinName = (PortJ << 4) + 0x08
	PJ9  PinName = (PortJ << 4) + 0x09
	PJ10 PinName = (PortJ << 4) + 0x0A
	PJ11 PinName = (PortJ << 4) + 0x0B
	PJ12 PinName = (PortJ << 4) + 0x0C
	PJ13 PinName = (PortJ << 4) + 0x0D
	PJ14 PinName = (PortJ << 4) + 0x0E
	PJ15 PinName = (PortJ << 4) + 0x0F

	PK0  PinName = (PortK << 4) + 0x00
	PK1  PinName = (PortK << 4) + 0x01
	PK2  PinName = (PortK << 4) + 0x02
	PK3  PinName = (PortK << 4) + 0x03
	PK4  PinName = (PortK << 4) + 0x04
	PK5  PinName = (PortK << 4) + 0x05
	PK6  PinName = (PortK << 4) + 0x06
	PK7  PinName = (PortK << 4) + 0x07
	PK8  PinName = (PortK << 4) + 0x08
	PK9  PinName = (PortK << 4) + 0x09
	PK10 PinName = (PortK << 4) + 0x0A
	PK11 PinName = (PortK << 4) + 0x0B
	PK12 PinName = (PortK << 4) + 0x0C
	PK13 PinName = (PortK << 4) + 0x0D
	PK14 PinName = (PortK << 4) + 0x0E
	PK15 PinName = (PortK << 4) + 0x0F

	PZ0  PinName = (PortZ << 4) + 0x00
	PZ1  PinName = (PortZ << 4) + 0x01
	PZ2  PinName = (PortZ << 4) + 0x02
	PZ3  PinName = (PortZ << 4) + 0x03
	PZ4  PinName = (PortZ << 4) + 0x04
	PZ5  PinName = (PortZ << 4) + 0x05
	PZ6  PinName = (PortZ << 4) + 0x06
	PZ7  PinName = (PortZ << 4) + 0x07
	PZ8  PinName = (PortZ << 4) + 0x08
	PZ9  PinName = (PortZ << 4) + 0x09
	PZ10 PinName = (PortZ << 4) + 0x0A
	PZ11 PinName = (PortZ << 4) + 0x0B
	PZ12 PinName = (PortZ << 4) + 0x0C
	PZ13 PinName = (PortZ << 4) + 0x0D
	PZ14 PinName = (PortZ << 4) + 0x0E
	PZ15 PinName = (PortZ << 4) + 0x0F

	// PX means unknown pin
	PX PinName = 0xFF
)

// non-gpio pins
const (
	P_3V3 PinName = iota + PZ15 + 1
	P_5V
	P_GND
	P_RESET
	P_VBAT
	P_NONE
)

var pinNames map[string]PinName = make(map[string]PinName, PZ15+1)

func init() {
	for i := PA0; i <= PZ15; i++ {
		pinNames[i.String()] = i
	}
}

func FindPinName(s string) (PinName, error) {
	n, ok := pinNames[s]
	if !ok {
		return PX, fmt.Errorf("PinName not found for `%s`", s)
	}
	return n, nil
}

func (n PinName) IsGpio() bool {
	return n <= PZ15
}

func (n PinName) IsNonGpio() bool {
	return n >= P_3V3 && n <= P_NONE
}

func (n PinName) IsUnknown() bool {
	return n > P_NONE
}

func (n PinName) String() string {
	if n < PZ0 {
		return "P" + string('A'+n>>4) + strconv.Itoa(int(n&0x0F))
	}
	if n <= PZ15 {
		return "PZ" + strconv.Itoa(int(n&0x0F))
	}
	if n <= P_NONE {
		switch n {
		case P_3V3:
			return "3V3"
		case P_5V:
			return "5V"
		case P_GND:
			return "GND"
		case P_RESET:
			return "RESET"
		case P_VBAT:
			return "VBAT"
		case P_NONE:
			return "NONE"
		}
	}
	return "P_0x" + strconv.FormatInt(int64(n), 16)
}
