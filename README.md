# goduino
Go's package for Arduino

Goduino uses [Firmata](https://github.com/firmata/protocol) protocol for [Arduino](https://www.arduino.cc/) 

[![GoDoc](http://godoc.org/github.com/argandas/goduino?status.svg)](http://godoc.org/github.com/argandas/goduino)

## Prerequisites

1. Download and install the [Arduio IDE](http://www.arduino.cc/en/Main/Software)
2. Plug in your Arduino via USB
3. Open the Arduino IDE and open: `File > Examples > StandardFirmata`
4. Select Arduino´s board: `Tools > Board`
5. Select Arduino´s serial port: `Tools > Serial Port`
6. Click the Upload button

## Installation

```bash
	go get github.com/argandas/goduino
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/argandas/goduino"
	"time"
)

func main() {
	arduino := goduino.New("myArduino", "COM1")
	err := arduino.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer arduino.Disconnect()
	
	arduino.PinMode(13, goduino.Output)
	for {
		arduino.DigitalWrite(13, 1)
		arduino.Delay(time.Millisecond * 500)
		arduino.DigitalWrite(13, 0)
		arduino.Delay(time.Millisecond * 500)
	}
}
```

Note: For this example the selected serial port is `COM1`, be sure your Arduino is connected on this serial port.

## Stable versions

This package has been tested on Go v1.4.2 & Firmata v2.4
