package goduino

import (
	"fmt"
	"github.com/tarm/serial"
	"io"
	"log"
	"os"
	"time"
)

// Arduino Firmata client for golang
type Goduino struct {
	serialDev string
	baud      int
	conn      *io.ReadWriteCloser
	Log       *log.Logger

	protocolVersion []byte
	firmwareVersion []int
	firmwareName    string

	ready             bool
	analogMappingDone bool
	capabilityDone    bool

	digitalPinState [8]byte

	analogPinsChannelMap map[int]byte
	analogChannelPinsMap map[byte]int
	pinModes             []map[PinMode]interface{}

	valueChan  chan FirmataValue
	serialChan chan string
	spiChan    chan []byte

	Verbose bool
}

// Creates a new Goduino object and connects to the Arduino board
// over specified serial port. This function blocks till a connection is
// succesfullt established and pin mappings are retrieved.
func New(dev string, baud int) (client *Goduino, err error) {
	var conn io.ReadWriteCloser

	c := &serial.Config{Name: dev, Baud: baud}
	conn, err = serial.OpenPort(c)
	if err != nil {
		return nil, err
	}

	client = &Goduino{
		serialDev: dev,
		baud:      baud,
		conn:      &conn,
		Log:       log.New(os.Stdout, "[goduino] ", log.Ltime),
	}
	go client.replyReader()

	conn.Write([]byte{byte(SystemReset)})
	t := time.NewTicker(time.Second)

	for !(client.ready && client.analogMappingDone && client.capabilityDone) {
		select {
		case <-t.C:
			//no-op
		case <-time.After(time.Second * 15):
			client.Log.Print("No response in 30 seconds. Resetting arduino")
			conn.Write([]byte{byte(SystemReset)})
		case <-time.After(time.Second * 30):
			client.Log.Print("Unable to initialize connection")
			conn.Close()
			client = nil
		}
	}

	client.Log.Print("Client ready to use")

	return
}

// PinMode configures the specified pin to behave either as an input or an output.
func (ino *Goduino) PinMode(pin int, mode PinMode) error {
	if ino.pinModes[pin][mode] == nil {
		return fmt.Errorf("Pin mode %v not supported by pin %v", mode, pin)
	}
	cmd := []byte{byte(SetPinMode), (byte(pin) & 0x7F), byte(mode)}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	switch mode {
	case Input:
		ino.EnableDigitalInput(uint(pin), true)
	case Analog:
		ino.EnableAnalogInput(uint(pin), true)
	}
	ino.Log.Printf("pinMode(%d, %s)\r\n", pin, mode)
	return nil
}

// Close the serial connection to properly clean up after ourselves
// Usage: defer client.Close()
func (ino *Goduino) Delay(duration time.Duration) {
	time.Sleep(duration)
}

// Close the serial connection to properly clean up after ourselves
// Usage: defer client.Close()
func (ino *Goduino) Close() {
	(*ino.conn).Close()
}

func (ino *Goduino) sendCommand(cmd []byte) (err error) {
	bStr := ""
	for _, b := range cmd {
		bStr = bStr + fmt.Sprintf(" %#2x", b)
	}

	if ino.Verbose {
		ino.Log.Printf("Command send%v\n", bStr)
	}

	_, err = (*ino.conn).Write(cmd)
	return
}

// Sets the polling interval in milliseconds for analog pin samples
func (ino *Goduino) SetAnalogSamplingInterval(ms byte) (err error) {
	data := to7Bit(ms)
	err = ino.sendSysEx(SamplingInterval, data[0], data[1])
	return
}

// Get the channel to retrieve analog and digital pin values
func (ino *Goduino) getValues() <-chan FirmataValue {
	return ino.valueChan
}
