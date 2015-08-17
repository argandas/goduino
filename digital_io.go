package goduino

// DigitalWrite write a HIGH or a LOW value to a digital pin.
//
// If the pin has been configured as an OUTPUT with pinMode(),
// its voltage will be set to the corresponding value:
// 5V (or 3.3V on 3.3V boards) for HIGH, 0V (ground) for LOW.
func (ino *Goduino) DigitalWrite(pin, value int) error {
	// Check if pin is configured as analog
	if ino.board.Pins()[pin].Mode != Output {
		if err := ino.PinMode(pin, Output); err != nil {
			return err
		}
	}
	ino.logger.Printf("digitalWrite(%d, %d)\r\n", pin, value)
	return ino.board.DigitalWrite(pin, value)
}

// DigitalRead reads the value from a specified digital pin, either HIGH or LOW.
func (ino *Goduino) DigitalRead(pin int) (value int, err error) {
	pin = ino.digitalPin(pin)
	// Check if pin is configured as input
	if ino.board.Pins()[pin].Mode != Input {
		if err = ino.PinMode(pin, Input); err != nil {
			return
		}
	}
	return ino.board.Pins()[pin].Value, nil
}
