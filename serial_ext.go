package goduino

type SerialSubCommand byte

// Configure a builtin or soft serial port. This command must be called before sending serial data.
// Set txPin and rxPin to 0x00 for builtin serial ports.
func (c *Goduino) SerialConfig(port SerialPort, baud int, txPin byte, rxPin byte) (err error) {
	baudBytes := intto7Bit(baud)
	bufferSize := intto7Bit(1024)
	termChar := to7Bit('\n')
	c.serialChan = make(chan string, 10)

	err = c.sendSysEx(Serial, byte(SerialConfig)|byte(port),
		baudBytes[0], baudBytes[1], baudBytes[2],
		bufferSize[0], bufferSize[1], bufferSize[2],
		termChar[0], termChar[1])
	return
}

// Get channel for incoming serial data
func (c *Goduino) GetSerialData() <-chan string {
	return c.serialChan
}

func (c *Goduino) parseSerialResponse(data7bit []byte) {

	data := make([]byte, 0)
	for i := 1; i < len(data7bit); i = i + 2 {
		data = append(data, byte(from7Bit(data7bit[i], data7bit[i+1])))
	}
	select {
	case c.serialChan <- string(data):
	default:
		c.Log.Print("Serial data buffer overflow. No listener?")
	}
}
