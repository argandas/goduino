package goduino

type SPISubCommand byte

// Enable SPI communication for selected chip-select pin
func (c *Goduino) SPIConfig(csPin byte, spiMode byte) (err error) {
	csPinBytes := to7Bit(csPin)
	spiModeBytes := to7Bit(spiMode)
	c.spiChan = make(chan []byte)

	err = c.sendSysEx(SysExSPI, byte(SPIConfig),
		csPinBytes[0], csPinBytes[1],
		spiModeBytes[0], spiModeBytes[1])
	return
}

// Read and write data to SPI device
func (c *Goduino) SPIReadWrite(csPin byte, data []byte) (dataOut []byte, err error) {
	csPinBytes := to7Bit(csPin)
	data7Bit := []byte{byte(SPIComm)}

	data7Bit = append(data7Bit, csPinBytes...)
	for i := 0; i < len(data); i++ {
		bytes := to7Bit(data[i])
		data7Bit = append(data7Bit, bytes...)
	}

	err = c.sendSysEx(SysExSPI, data7Bit...)
	dataOut = <-c.spiChan
	return
}

func (c *Goduino) parseSPIResponse(data7bit []byte) {
	data := make([]byte, 0)
	for i, _ := range data7bit {
		if i >= 3 && i%2 != 0 {
			data = append(data, from7Bit(data7bit[i], data7bit[i+1]))
		}
	}
	c.spiChan <- data
}
