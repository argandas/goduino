package goduino

import (
	"fmt"
	"time"
	"github.com/argandas/serial"
)

type Goduino struct {
	port serial.SerialPort
}

func New() *Goduino {
	goduino := Goduino {
		port = serial.New()
	}
}

func (ino *Goduino) Open(name string, baud int) error {
	return ino.port.Open(name, baud, time.Millisecond * 100)
}

func (ino *Goduino) Close() error {
	return ino.port.Close()
}

func (ino *Goduino) PinMode() {

}

func (ino *Goduino) DigitalWrite() {

}

func (ino *Goduino) DigitalRead() {

}

func (ino *Goduino) AnalogRead(){

}

func (ino *Goduino) AnalogWrite() {

}

