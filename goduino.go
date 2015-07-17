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

// Sets the Pin mode (input, output, etc.) for the Arduino pin
func (ino *Goduino) PinMode(pin int, mode PinMode) error {
	if ino.pinModes[pin][mode] == nil {
		return fmt.Errorf("Pin mode %v not supported by pin %v", mode, pin)
	}
	cmd := []byte{byte(SetPinMode), (uint8(pin) & 0x7F), byte(mode)}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	ino.Log.Printf("pinMode(%d, %s)\r\n", pin, mode)
	return nil
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

// Set the value of a digital pin
func (ino *Goduino) DigitalWrite(pin int, val PinState) error {
	if uint8(pin) < 0 || uint8(pin) > uint8(len(ino.pinModes)) && ino.pinModes[pin][Output] != nil {
		return fmt.Errorf("Invalid pin number %v\n", pin)
	}
	port := (uint8(pin) / 8) & 0x7F
	portData := &ino.digitalPinState[port]
	pinData := uint8(pin % 8)
	if val >= HIGH {
		val = HIGH
		(*portData) = (*portData) | (1 << pinData)
	} else {
		val = LOW
		(*portData) = (*portData) & ^(1 << pinData)
	}
	data := to7Bit(*(portData))
	cmd := []byte{byte(DigitalMessage) | byte(port), data[0], data[1]}
	if err := ino.sendCommand(cmd); err != nil {
		return err
	}
	ino.Log.Printf("digitalWrite(%d, %s)\r\n", pin, val)
	return nil
}

// Specified if a analog Pin should be watched for input.
// Values will be streamed back over a channel which can be retrieved by the GetValues() call
func (ino *Goduino) EnableAnalogInput(pin uint, val bool) (err error) {
	if pin < 0 || pin > uint(len(ino.pinModes)) && ino.pinModes[pin][Analog] != nil {
		err = fmt.Errorf("Invalid pin number %v\n", pin)
		return
	}

	ch := byte(ino.analogPinsChannelMap[int(pin)])
	ino.Log.Printf("Enable analog inout on pin %v channel %v", pin, ch)
	if val {
		cmd := []byte{byte(EnableAnalogInput) | ch, 0x01}
		err = ino.sendCommand(cmd)
	} else {
		cmd := []byte{byte(EnableAnalogInput) | ch, 0x00}
		err = ino.sendCommand(cmd)
	}

	return
}

// Set the value of a analog pin
func (ino *Goduino) AnalogWrite(pin uint, pinData byte) (err error) {
	if pin < 0 || pin > uint(len(ino.pinModes)) && ino.pinModes[pin][Analog] != nil {
		err = fmt.Errorf("Invalid pin number %v\n", pin)
		return
	}

	data := to7Bit(pinData)
	cmd := []byte{byte(AnalogMessage) | byte(pin), data[0], data[1]}
	err = ino.sendCommand(cmd)
	return
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
func (ino *Goduino) GetValues() <-chan FirmataValue {
	return ino.valueChan
}
