package firmata

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"
)

// Errors
var ErrConnected = errors.New("client is already connected")

// Firmata represents a client connection to a firmata board
type Firmata struct {
	pins              []Pin
	FirmwareName      string
	ProtocolVersion   string
	connected         bool
	connection        io.ReadWriteCloser
	analogPins        []int
	ready             bool
	analogMappingDone bool
	capabilityDone    bool
	logger            *log.Logger
}

// Pin represents a pin on the firmata board
type Pin struct {
	SupportedModes []int
	Mode           int
	Value          int
	State          int
	AnalogChannel  int
}

// I2cReply represents the response from an I2cReply message
type I2cReply struct {
	Address  int
	Register int
	Data     []byte
}

// New returns a new Firmata
func New() *Firmata {
	c := &Firmata{
		ProtocolVersion: "",
		FirmwareName:    "",
		connection:      nil,
		pins:            []Pin{},
		analogPins:      []int{},
		connected:       false,
		logger:          log.New(os.Stdout, "[firmata] ", log.Ltime),
	}

	return c
}

// Disconnect disconnects the Firmata
func (f *Firmata) Disconnect() (err error) {
	f.connected = false
	return f.connection.Close()
}

// Connected returns the current connection state of the Firmata
func (f *Firmata) Connected() bool {
	return f.connected
}

// Pins returns all available pins
func (f *Firmata) Pins() []Pin {
	return f.pins
}

// Connect connects to the Firmata given conn. It first resets the firmata board
// then continuously polls the firmata board for new information when it's
// available.
func (f *Firmata) Connect(conn io.ReadWriteCloser) (err error) {
	if f.connected {
		return ErrConnected
	}

	f.connection = conn

	// Start threads
	go f.process()

	// Reset device
	f.Reset()

	// Wait for device to response
	t := time.NewTicker(time.Second)
	for !f.connected {
		select {
		case <-t.C:
			// Do nothing
		case <-time.After(time.Second * 15):
			f.logger.Print("No response in 15 seconds. Resetting device")
			f.Reset()
		case <-time.After(time.Second * 30):
			// Close connections
			f.connection.Close()
			return errors.New("Unable to initialize connection")
		}
	}

	// Firmata creation successful
	f.logger.Print("Firmata ready to use")
	return nil
}

// Reset sends the SystemReset sysex code.
func (f *Firmata) Reset() error {
	return f.write([]byte{byte(SystemReset)})
}

// SetPinMode sets the pin to mode.
func (f *Firmata) SetPinMode(pin int, mode int) error {
	f.pins[byte(pin)].Mode = mode
	return f.sendCommand([]byte{byte(PinMode), byte(pin), byte(mode)})
}

// DigitalWrite writes value to pin.
func (f *Firmata) DigitalWrite(pin int, value int) error {
	port := byte(math.Floor(float64(pin) / 8))
	portValue := byte(0)
	f.pins[pin].Value = value
	// Build command
	for i := byte(0); i < 8; i++ {
		if f.pins[8*port+i].Value != 0 {
			portValue = portValue | (1 << i)
		}
	}
	return f.sendCommand([]byte{byte(DigitalMessage) | port, portValue & 0x7F, (portValue >> 7) & 0x7F})
}

// ServoConfig sets the min and max pulse width for servo PWM range
func (f *Firmata) ServoConfig(pin int, max int, min int) error {
	ret := []byte{
		byte(ServoConfig),
		byte(pin),
		byte(max & 0x7F),
		byte((max >> 7) & 0x7F),
		byte(min & 0x7F),
		byte((min >> 7) & 0x7F),
	}
	return f.writeSysex(ret)
}

// AnalogWrite writes value to pin.
func (f *Firmata) AnalogWrite(pin int, value int) error {
	f.pins[pin].Value = value
	return f.write([]byte{byte(AnalogMessage) | byte(pin), byte(value & 0x7F), byte((value >> 7) & 0x7F)})
}

// FirmwareQuery sends the FirmwareQuery sysex code.
func (f *Firmata) FirmwareQuery() error {
	return f.writeSysex([]byte{byte(FirmwareQuery)})
}

// PinStateQuery sends a PinStateQuery for pin.
func (f *Firmata) PinStateQuery(pin int) error {
	return f.writeSysex([]byte{byte(PinStateQuery), byte(pin)})
}

// ProtocolVersionQuery sends the ProtocolVersion sysex code.
func (f *Firmata) ProtocolVersionQuery() error {
	return f.write([]byte{byte(ProtocolVersion)})
}

// CapabilitiesQuery sends the CapabilityQuery sysex code.
func (f *Firmata) CapabilitiesQuery() error {
	return f.writeSysex([]byte{byte(CapabilityQuery)})
}

// AnalogMappingQuery sends the AnalogMappingQuery sysex code.
func (f *Firmata) AnalogMappingQuery() error {
	return f.writeSysex([]byte{byte(AnalogMappingQuery)})
}

// ReportDigital enables or disables digital reporting for pin, a non zero
// state enables reporting
func (f *Firmata) ReportDigital(pin int, state int) error {
	return f.togglePinReporting(pin, state, byte(ReportDigital))
}

// ReportAnalog enables or disables analog reporting for pin, a non zero
// state enables reporting
func (f *Firmata) ReportAnalog(pin int, state int) error {
	return f.togglePinReporting(pin, state, byte(ReportAnalog))
}

// I2cRead reads numBytes from address once.
func (f *Firmata) I2cRead(address int, numBytes int) error {
	return f.writeSysex([]byte{byte(I2CRequest), byte(address), (I2CModeRead << 3),
		byte(numBytes) & 0x7F, (byte(numBytes) >> 7) & 0x7F})
}

// I2cWrite writes data to address.
func (f *Firmata) I2cWrite(address int, data []byte) error {
	ret := []byte{byte(I2CRequest), byte(address), (I2CModeWrite << 3)}
	for _, val := range data {
		ret = append(ret, byte(val&0x7F))
		ret = append(ret, byte((val>>7)&0x7F))
	}
	return f.writeSysex(ret)
}

// I2cConfig configures the delay in which a register can be read from after it
// has been written to.
func (f *Firmata) I2cConfig(delay int) error {
	return f.writeSysex([]byte{byte(I2CConfig), byte(delay & 0xFF), byte((delay >> 8) & 0xFF)})
}

func (f *Firmata) togglePinReporting(pin int, state int, mode byte) error {
	if state != 0 {
		state = 1
	} else {
		state = 0
	}

	if err := f.write([]byte{byte(mode) | byte(pin), byte(state)}); err != nil {
		return err
	}

	return nil

}

func (f *Firmata) writeSysex(data []byte) (err error) {
	frame := append([]byte{byte(StartSysex)}, append(data, byte(EndSysex))...)
	f.printSysExData("SysEx Tx", SysExCommand(frame[1]), frame)
	return f.write(frame)
}

func (f *Firmata) write(data []byte) (err error) {
	_, err = f.connection.Write(data[:])
	return
}

func (f *Firmata) sendCommand(cmd []byte) error {
	f.printByteArray("Command send", cmd)
	_, err := f.connection.Write(cmd)
	return err
}

func (f *Firmata) read(length int) (buf []byte, err error) {
	i := 0
	for length > 0 {
		tmp := make([]byte, length)
		if i, err = f.connection.Read(tmp); err != nil {
			if err.Error() != "EOF" {
				return
			}
			<-time.After(5 * time.Millisecond)
		}
		if i > 0 {
			buf = append(buf, tmp...)
			length = length - i
		}
	}
	return
}

func (f *Firmata) process() {
	r := bufio.NewReader(f.connection)
	var init bool
	for {
		b, err := r.ReadByte()
		if err != nil {
			f.logger.Panic(err)
			return
		}
		cmd := FirmataCommand(b)
		f.logger.Printf("Incoming cmd %v", cmd)

		// First received byte must be ReportVersion command
		if !init {
			if cmd != ProtocolVersion {
				f.logger.Printf("Discarding unexpected command byte %0d (not initialized)\n", b)
				continue
			} else {
				init = true
			}
		}

		switch {
		case ProtocolVersion == cmd:
			buf, err := f.read(2)
			if err != nil {
				f.logger.Panic(err)
				return
			}
			f.ProtocolVersion = fmt.Sprintf("%v.%v", buf[0], buf[1])
			f.logger.Printf("Protocol version: %s", f.ProtocolVersion)
			f.FirmwareQuery()
		case AnalogMessageRangeStart <= cmd && AnalogMessageRangeEnd >= cmd:
			buf, err := f.read(2)
			if err != nil {
				f.logger.Panic(err)
				return
			}

			value := uint(buf[0]) | uint(buf[1])<<7
			pin := int((cmd & 0x0F))

			if len(f.analogPins) > pin {
				if len(f.pins) > f.analogPins[pin] {
					f.pins[f.analogPins[pin]].Value = int(value)
					f.logger.Printf("AnalogRead%v", pin)
				}
			}
		case DigitalMessageRangeStart <= cmd && DigitalMessageRangeEnd >= cmd:
			buf, err := f.read(2)
			if err != nil {
				f.logger.Panic(err)
				return
			}
			port := cmd & 0x0F
			portValue := buf[1] | (buf[2] << 7)
			for i := 0; i < 8; i++ {
				pinNumber := int((8*byte(port) + byte(i)))
				if len(f.pins) > pinNumber {
					if f.pins[pinNumber].Mode == Input {
						f.pins[pinNumber].Value = int((portValue >> (byte(i) & 0x07)) & 0x01)
						f.logger.Printf("DigitalRead%v", pinNumber)
					}
				}
			}
		case StartSysex == cmd:
			sysExData, err := r.ReadSlice(byte(EndSysex))
			if err != nil {
				f.logger.Panic(err)
				break
			}
			// Remove EndSysEx byte
			f.parseSysEx(sysExData[:len(sysExData)-1])
		}
	}
}

func (f *Firmata) parseSysEx(data []byte) {

	// ino.printSysExData("SysEx Rx", cmd, data)

	cmd := SysExCommand(data[0])
	data = data[1:]
	f.printSysExData("SysEx Rx", cmd, data)

	switch cmd {
	case CapabilityResponse:
		f.pins = []Pin{}
		supportedModes := 0
		n := 0
		for _, val := range data[:(len(data) - 5)] {
			if val == 127 {
				modes := []int{}
				for _, mode := range []int{Input, Output, Analog, Pwm, Servo} {
					if (supportedModes & (1 << byte(mode))) != 0 {
						modes = append(modes, mode)
					}
				}

				f.pins = append(f.pins, Pin{SupportedModes: modes, Mode: Output})
				supportedModes = 0
				n = 0
				continue
			}

			if n == 0 {
				supportedModes = supportedModes | (1 << val)
			}
			n ^= 1
		}
		f.logger.Printf("Total pins: %v\n", len(f.pins))
		f.AnalogMappingQuery()
	case AnalogMappingResponse:
		f.analogPins = []int{}
		for index, val := range data[:len(f.pins)-1] {
			f.pins[index].AnalogChannel = int(val)
			if val != 127 {
				f.analogPins = append(f.analogPins, index)
			}
			// fmt.Println(index, ":", f.pins[index].AnalogChannel, ":", val)
		}
		f.logger.Printf("pin -> channel: %v\n", f.analogPins)
		f.connected = true
	case PinStateResponse:
		pin := data[0]
		f.pins[pin].Mode = int(data[1])
		f.pins[pin].State = int(data[2])

		if len(data) > 3 {
			f.pins[pin].State = int(uint(f.pins[pin].State) | uint(data[2])<<7)
		}
		if len(data) > 4 {
			f.pins[pin].State = int(uint(f.pins[pin].State) | uint(data[4])<<14)
		}
		f.logger.Printf("PinState%v", pin)
	case I2CReply:
		reply := I2cReply{
			Address:  int(byte(data[0]) | byte(data[1])<<7),
			Register: int(byte(data[2]) | byte(data[3])<<7),
			Data:     []byte{byte(data[4]) | byte(data[5])<<7},
		}
		for i := 8; i < len(data); i = i + 2 {
			if data[i] == byte(0xF7) {
				break
			}
			if i+2 > len(data) {
				break
			}
			reply.Data = append(reply.Data,
				byte(data[i])|byte(data[i+1])<<7,
			)
		}
		f.logger.Printf("I2cReply%v", reply)
	case FirmwareQuery:
		name := []byte{}
		for _, val := range data[2:(len(data) - 1)] {
			if val != 0 {
				name = append(name, val)
			}
		}
		f.FirmwareName = string(name[:])
		f.logger.Printf("Firmware: %s", f.FirmwareName)
		f.CapabilitiesQuery()
	case StringData:
		str := data[:]
		f.logger.Printf("StringData%v", string(str[:len(str)-1]))
	}
}

func (f *Firmata) printByteArray(title string, data []uint8) {
	fmt.Println()
	f.logger.Println(title)
	str := ""
	for index, b := range data {
		str += fmt.Sprintf("0x%02X ", b)
		if (index+1)%8 == 0 || index == len(data)-1 {
			f.logger.Println(str)
			str = ""
		}
	}
	fmt.Println()
}

func (f *Firmata) printSysExData(title string, cmd SysExCommand, data []uint8) {
	fmt.Println()
	f.logger.Println(title, "-", cmd)
	str := ""
	for index, b := range data {
		str += fmt.Sprintf("0x%02X ", b)
		if (index+1)%8 == 0 || index == len(data)-1 {
			f.logger.Println(str)
			str = ""
		}
	}
	fmt.Println()
}
