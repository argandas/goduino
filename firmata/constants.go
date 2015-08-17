package firmata

import "fmt"

type FirmataCommand byte
type SysExCommand byte
type SerialPort byte

// Pin Modes
const (
	Input  = 0x00
	Output = 0x01
	Analog = 0x02
	Pwm    = 0x03
	Servo  = 0x04

	// SPIConfig SPISubCommand = 0x10
	// SPIComm   SPISubCommand = 0x20

	SPI_MODE0 = 0x00
	SPI_MODE1 = 0x04
	SPI_MODE2 = 0x08
	SPI_MODE3 = 0x0C

	SoftSerial  SerialPort = 0x00
	HardSerial1 SerialPort = 0x01
	HardSerial2 SerialPort = 0x02
	HardSerial3 SerialPort = 0x03

	I2CModeWrite          byte = 0x00
	I2CModeRead           byte = 0x01
	I2CModeContinuousRead byte = 0x02
	I2CModeStopReading    byte = 0x03

	// SerialConfig SerialSubCommand = 0x10
	// SerialComm   SerialSubCommand = 0x20
	// SerialFlush  SerialSubCommand = 0x30
	// SerialClose  SerialSubCommand = 0x40
)

// Firmata commands
const (
	DigitalMessage           FirmataCommand = 0x90
	DigitalMessageRangeStart FirmataCommand = 0x90
	DigitalMessageRangeEnd   FirmataCommand = 0x9F
	ReportAnalog             FirmataCommand = 0xC0
	ReportDigital            FirmataCommand = 0xD0
	AnalogMessage            FirmataCommand = 0xE0
	AnalogMessageRangeStart  FirmataCommand = 0xE0
	AnalogMessageRangeEnd    FirmataCommand = 0xEF
	StartSysex               FirmataCommand = 0xF0
	PinMode                  FirmataCommand = 0xF4
	EndSysex                 FirmataCommand = 0xF7
	ProtocolVersion          FirmataCommand = 0xF9
	SystemReset              FirmataCommand = 0xFF
)

// SysEx Commands
const (
	Serial                SysExCommand = 0x60
	AnalogMappingQuery    SysExCommand = 0x69
	AnalogMappingResponse SysExCommand = 0x6A
	CapabilityQuery       SysExCommand = 0x6B
	CapabilityResponse    SysExCommand = 0x6C
	PinStateQuery         SysExCommand = 0x6D
	PinStateResponse      SysExCommand = 0x6E
	ServoConfig           SysExCommand = 0x70
	StringData            SysExCommand = 0x71
	ShiftData             SysExCommand = 0x75 // a bitstream to/from a shift register
	I2CRequest            SysExCommand = 0x76
	I2CReply              SysExCommand = 0x77
	I2CConfig             SysExCommand = 0x78
	FirmwareQuery         SysExCommand = 0x79
	SamplingInterval      SysExCommand = 0x7A // set the poll rate of the main loop
	SysExNonRealtime      SysExCommand = 0x7E // MIDI Reserved for non-realtime messages
	SysExRealtime         SysExCommand = 0x7F // MIDI Reserved for realtime messages
	SysExSPI              SysExCommand = 0x80
)

func (c FirmataCommand) String() string {
	switch {
	case (c & 0xF0) == DigitalMessage:
		return fmt.Sprintf("DigitalMessage (0x%x)", uint8(c))
	case (c & 0xF0) == AnalogMessage:
		return fmt.Sprintf("AnalogMessage (0x%x)", uint8(c))
	case c == ReportAnalog:
		return fmt.Sprintf("ReportAnalog (0x%x)", uint8(c))
	case c == ReportDigital:
		return fmt.Sprintf("ReportDigital (0x%x)", uint8(c))
	case c == PinMode:
		return fmt.Sprintf("PinMode (0x%x)", uint8(c))
	case c == ProtocolVersion:
		return fmt.Sprintf("ProtocolVersion (0x%x)", uint8(c))
	case c == SystemReset:
		return fmt.Sprintf("SystemReset (0x%x)", uint8(c))
	case c == StartSysex:
		return fmt.Sprintf("StartSysex (0x%x)", uint8(c))
	case c == EndSysex:
		return fmt.Sprintf("EndSysex (0x%x)", uint8(c))
	}
	return fmt.Sprintf("Unexpected command (0x%x)", uint8(c))
}

func (c SysExCommand) String() string {
	switch {
	case c == ServoConfig:
		return fmt.Sprintf("ServoConfig (0x%x)", uint8(c))
	case c == StringData:
		return fmt.Sprintf("StringData (0x%x)", uint8(c))
	case c == ShiftData:
		return fmt.Sprintf("ShiftData (0x%x)", uint8(c))
	case c == I2CRequest:
		return fmt.Sprintf("I2CRequest (0x%x)", uint8(c))
	case c == I2CReply:
		return fmt.Sprintf("I2CReply (0x%x)", uint8(c))
	case c == I2CConfig:
		return fmt.Sprintf("I2CConfig (0x%x)", uint8(c))
	// case c == ExtendedAnalog:
	//	return fmt.Sprintf("ExtendedAnalog (0x%x)", uint8(c))
	case c == PinStateQuery:
		return fmt.Sprintf("PinStateQuery (0x%x)", uint8(c))
	case c == PinStateResponse:
		return fmt.Sprintf("PinStateResponse (0x%x)", uint8(c))
	case c == CapabilityQuery:
		return fmt.Sprintf("CapabilityQuery (0x%x)", uint8(c))
	case c == CapabilityResponse:
		return fmt.Sprintf("CapabilityResponse (0x%x)", uint8(c))
	case c == AnalogMappingQuery:
		return fmt.Sprintf("AnalogMappingQuery (0x%x)", uint8(c))
	case c == AnalogMappingResponse:
		return fmt.Sprintf("AnalogMappingResponse (0x%x)", uint8(c))
	case c == FirmwareQuery:
		return fmt.Sprintf("FirmwareQuery (0x%x)", uint8(c))
	case c == SamplingInterval:
		return fmt.Sprintf("SamplingInterval (0x%x)", uint8(c))
	case c == SysExNonRealtime:
		return fmt.Sprintf("SysExNonRealtime (0x%x)", uint8(c))
	case c == SysExRealtime:
		return fmt.Sprintf("SysExRealtime (0x%x)", uint8(c))
	case c == Serial:
		return fmt.Sprintf("Serial (0x%x)", uint8(c))
	case c == SysExSPI:
		return fmt.Sprintf("SPI (0x%x)", uint8(c))
	}
	return fmt.Sprintf("Unexpected SysEx command (0x%x)", uint8(c))
}
