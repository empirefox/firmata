package firmata

import "testing"

func TestPinNameString(t *testing.T) {
	if PA0.String() != "PA0" {
		t.Errorf("String of PA0 error, got: %s", PA0.String())
	}
	if PZ15.String() != "PZ15" {
		t.Errorf("String of PZ15 error, got: %s", PZ15.String())
	}
	if PinName(P_NONE+1).String() != "P_0xc6" {
		t.Errorf("String of PinName(P_NONE+1) error, got: %s", PinName(P_NONE+1).String())
	}
}
