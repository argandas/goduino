package goduino

import (
	"fmt"
)

// PinMode configures the specified pin to behave either as an input or an output. 
func (ino *Goduino) PinMode(pin int, mode PinMode) error {
	if ino.pinModes[pin][mode] == nil {
		return fmt.Errorf("Pin mode %v not supported by pin %v", mode, pin)
	}
	cmd := []byte{byte(SetPinMode), (byte(pin) & 0x7F), byte(mode)}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	ino.Log.Printf("pinMode(%d, %s)\r\n", pin, mode)
	return nil
}

// DigitalWrite write a HIGH or a LOW value to a digital pin.
// 
// If the pin has been configured as an OUTPUT with pinMode(), 
// its voltage will be set to the corresponding value: 
// 5V (or 3.3V on 3.3V boards) for HIGH, 0V (ground) for LOW.
func (ino *Goduino) DigitalWrite(pin int, value PinState) error {
	if uint8(pin) < 0 || pin > len(ino.pinModes) && ino.pinModes[pin][Output] != nil {
		return fmt.Errorf("Invalid pin number %v\n", pin)
	}
	port := (uint8(pin) / 8) & 0x7F
	portData := &ino.digitalPinState[port]
	pinData := uint8(pin % 8)
	if value >= HIGH {
		value = HIGH
		*portData |= (1 << pinData)
	} else {
		value = LOW
		*portData &= ^(1 << pinData)
	}
	data := to7Bit(*portData)
	cmd := []byte{byte(DigitalMessage) | byte(port), data[0], data[1]}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	ino.Log.Printf("digitalWrite(%d, %s)\r\n", pin, value)
	return nil
}

// DigitalRead reads the value from a specified digital pin, either HIGH or LOW.
func (ino *Goduino) DigitalRead(pin int) (PinState, error) {
	state := LOW
	ino.Log.Printf("digitalRead(%d)\r\n", pin)
	return state, nil
}

func (v FirmataValue) GetDigitalValue() (port byte, val map[byte]interface{}, err error) {
	if v.IsAnalog() {
		err = fmt.Errorf("Cannot get digital value for analog pin")
		return
	}

	port = byte(v.valueType & ^DigitalMessage)
	val = make(map[byte]interface{})
	mask := 0x01
	for i := byte(0); i < 8; i++ {
		val[port*8+i] = ((v.value & mask) > 0)
		mask = mask * 2
	}
	return
}

// Specified if a digital Pin should be watched for input.
// Values will be streamed back over a channel which can be retrieved by the GetValues() call
func (ino *Goduino) EnableDigitalInput(pin uint, val bool) (err error) {
	if pin < 0 || pin > uint(len(ino.pinModes)) {
		err = fmt.Errorf("Invalid pin number %v\n", pin)
		return
	}
	port := (pin / 8) & 0x7F
	pin = pin % 8

	if val {
		cmd := []byte{byte(EnableDigitalInput) | byte(port), 0x01}
		err = ino.sendCommand(cmd)
	} else {
		cmd := []byte{byte(EnableDigitalInput) | byte(port), 0x00}
		err = ino.sendCommand(cmd)
	}

	return
}