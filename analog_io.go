package goduino

// AnalogRead retrieves value from analog pin.
// Returns -1 if the response from the board has timed out
func (ino *Goduino) AnalogRead(pin int) (val int, err error) {
	pin = ino.digitalPin(pin)
	// Check if pin is configured as analog
	if ino.board.Pins()[pin].Mode != Analog {
		if err = ino.PinMode(pin, Analog); err != nil {
			return
		}
	}
	return ino.board.Pins()[pin].Value, nil
}
