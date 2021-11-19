# Part implements of Firmata remote control

**Not tested yet!!!**

## Add to `ConfigurableFirmata.ino`

```c++
#define UD_PIN_NAMES_REQUEST 0x06
#define UD_PIN_NAMES_REPLY 0x07

void handlePinNames() {
  Firmata.write(START_SYSEX);
  Firmata.write(UD_PIN_NAMES_REPLY);
  for (byte i = 0; i < TOTAL_PINS; i++) {
    Firmata.sendValueAsTwo7bitBytes(digitalPin[i]);
  }
  Firmata.write(END_SYSEX);
}

void initFirmata()
  // add to end
  Firmata.attach(UD_PIN_NAMES_REQUEST, handlePinNames);
}
```