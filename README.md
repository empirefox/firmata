# Part implements of Firmata remote control

**Not tested yet!!!**

## Add to `ConfigurableFirmata.ino`

```c++
#define UD_PIN_NAMES_REQUEST 0x06
#define UD_PIN_NAMES_REPLY 0x07

class PinNames: public FirmataFeature
{
  public:
    void handleCapability(byte pin) {}
    boolean handlePinMode(byte pin, int mode) { return false; }
    boolean handleSysex(byte command, byte argc, byte* argv) {
      if (command == UD_PIN_NAMES_REQUEST) {
        Firmata.write(START_SYSEX);
        Firmata.write(UD_PIN_NAMES_REPLY);
        for (byte i = 0; i < TOTAL_PINS; i++) {
          Firmata.sendValueAsTwo7bitBytes(digitalPin[i]);
        }
        Firmata.write(END_SYSEX);
        return true;
      }
      return false;
    }
    void reset() {}
};
PinNames pinNames;

void initFirmata()
  // add to end
  firmataExt.addFeature(pinNames);
}
```