package goduino

type PinMode uint8
type PinState uint8

const (
	// Pin modes
	Input  PinMode = 0x00
	Output PinMode = 0x01
	Analog PinMode = 0x02
	PWM    PinMode = 0x03
	Servo  PinMode = 0x04
	Shift  PinMode = 0x05
	I2C    PinMode = 0x06
	SPI    PinMode = 0x07
	// Pin state
	HIGH PinState = 0x01
	LOW  PinState = 0x00
)

func (m PinMode) String() string {
	switch {
	case m == Input:
		return "INPUT"
	case m == Output:
		return "OUTPUT"
	case m == Analog:
		return "ANALOG"
	case m == PWM:
		return "PWM"
	case m == Servo:
		return "SERVO"
	case m == Shift:
		return "SHIFT"
	case m == I2C:
		return "I2C"
	}
	return "UNKNOWN"
}

func (s PinState) String() string {
	switch {
	case s == HIGH:
		return "HIGH"
	case s == LOW:
		return "LOW"
	}
	return "UNKNOWN"
}
