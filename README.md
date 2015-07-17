# go-firmata
A Golang wrapper for [Firmata](https://www.arduino.cc/en/reference/firmata) on [Arduino](https://www.arduino.cc/) 

[![GoDoc](http://godoc.org/github.com/kraman/go-firmata?status.svg)](http://godoc.org/github.com/kraman/go-firmata)

## Installation

### Go's go-firmata package:

```bash
	go get github.com/kraman/go-firmata
```

### Install firmata firmware to Arduino:

Open arduino IDE and open:
```bach
File > Examples > Firmata > StandardFirmata
```
Select the appriate port for your arduino and click upload. Wait for the upload to finish and you should be ready to start using firmata with your arduino.

## Hardware support

- [Arduino Uno R3](https://www.arduino.cc/en/Main/arduinoBoardUno)

## Usage

```go
package main

import (
	"github.com/kraman/go-firmata"
	"time"
)

var led uint8 = 13
var myDelay = time.Millisecond * 500

func main() {
	arduino, err := firmata.NewClient("COM1", 57600)
	defer arduino.Close()
	if err != nil {
		panic(err)
	}
	// Set led pin as output
	arduino.SetPinMode(led, firmata.Output)
	// Blink led 10 times
	for x := 0; x < 10; x++ {
		// Turn ON led
		arduino.DigitalWrite(led, true)
		arduino.Delay(myDelay)
		// Turn OFF led
		arduino.DigitalWrite(led, false)
		arduino.Delay(myDelay)
	}
}
```