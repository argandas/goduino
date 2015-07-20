package goduino

import (
	"fmt"
)

// AnalogWrite writes an analog value (PWM wave) to a pin.
// Can be used to light a LED at varying brightnesses or drive a motor at various speeds.
// After a call to analogWrite(), the pin will generate a steady square wave of the specified duty cycle until the next call to analogWrite() (or a call to digitalRead() or digitalWrite() on the same pin).
// The frequency of the PWM signal on most pins is approximately 490 Hz.
// On the Uno and similar boards, pins 5 and 6 have a frequency of approximately 980 Hz.
func (ino *Goduino) AnalogWrite(pin, value int) error {
	if pin < 0 || pin > len(ino.pinModes) && ino.pinModes[pin][Analog] != nil {
		return fmt.Errorf("Invalid pin number %d\n", pin)
	}
	// Analog 14-bit data format
	//   [0]  analog pin, 0xE0-0xEF, (MIDI Pitch Wheel)
	//   [1]  analog 7 lsb
	//   [2]  analog 7 msb
	data := to7Bit(byte(value))
	cmd := []byte{byte(AnalogMessage) | byte(pin), data[0], data[1]}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	ino.Log.Printf("analogWrite(%d, %d)\r\n", pin, value)
	return nil
}
