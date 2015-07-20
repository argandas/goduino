package goduino

import (
	"bufio"
	"fmt"
)

type FirmataValue struct {
	valueType            FirmataCommand
	value                int
	analogChannelPinsMap map[byte]int
}

func (v FirmataValue) IsAnalog() bool {
	return (v.valueType & 0xF0) == AnalogMessage
}

func (v FirmataValue) GetAnalogValue() (pin int, val int, err error) {
	if !v.IsAnalog() {
		err = fmt.Errorf("Cannot get analog value for digital pin")
		return
	}
	pin = v.analogChannelPinsMap[byte(v.valueType & ^AnalogMessage)]
	val = v.value
	return
}

func (v FirmataValue) String() string {
	if v.IsAnalog() {
		p, v, _ := v.GetAnalogValue()
		return fmt.Sprintf("Analog value %v = %v", p, v)
	} else {
		p, v, _ := v.GetAnalogValue()
		return fmt.Sprintf("Digital port %v = %b", p, v)
	}
}

func (c *Goduino) replyReader() {
	r := bufio.NewReader(*c.conn)
	c.valueChan = make(chan FirmataValue)
	var init bool
	for {
		b, err := (r.ReadByte())
		if err != nil {
			c.Log.Print(err)
			return
		}
		cmd := FirmataCommand(b)
		if c.Verbose {
			c.Log.Printf("Incoming cmd %v", cmd)
		}
		if !init {
			if cmd != ReportVersion {
				if c.Verbose {
					c.Log.Printf("Discarding unexpected command byte %0d (not initialized)\n", b)
				}
				continue
			} else {
				init = true
			}
		}

		switch {
		case cmd == ReportVersion:
			c.protocolVersion = make([]byte, 2)
			c.protocolVersion[0], err = r.ReadByte()
			c.protocolVersion[1], err = r.ReadByte()
			c.Log.Printf("Protocol version: %d.%d", c.protocolVersion[0], c.protocolVersion[1])
		case cmd == StartSysEx:
			var sysExData []byte
			sysExData, err = r.ReadSlice(byte(EndSysEx))
			if err == nil {
				c.parseSysEx(sysExData[0 : len(sysExData)-1])
			}
		case (cmd&DigitalMessage) > 0 || byte(cmd&AnalogMessage) > 0:
			b1, _ := r.ReadByte()
			b2, _ := r.ReadByte()
			select {
			case c.valueChan <- FirmataValue{cmd, int(from7Bit(b1, b2)), c.analogChannelPinsMap}:
			}
		default:
			if c.Verbose {
				c.Log.Printf("Discarding unexpected command byte %0d\n", b)
			}
		}
		if err != nil {
			c.Log.Print(err)
			return
		}
	}
}
