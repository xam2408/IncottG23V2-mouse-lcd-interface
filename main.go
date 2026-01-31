package main

import (
	"log"
	"time"

	"github.com/karalabe/hid"
)

const (
	vendorID       = 0x093A
	productID      = 0x522C
	comPort        = "COM3" //stm32 serial port
	mouseStatusIdx = 6
)

func main() {
	dev := hid.Enumerate(vendorID, productID)
	log.Printf("Found devices: %d", len(dev))

	if len(dev) == 0 {
		log.Fatal("No device found") //later data for stm32
	}

	device, err := dev[mouseStatusIdx].Open()
	if err != nil {
		log.Fatal("Failed to open device:", err)
	}
	defer device.Close()

	log.Println("Device opened successfully")
	buf := make([]byte, 64)
	for {
		n, err := device.Read(buf)
		if err != nil {
			log.Fatal("Failed to read from device:", err)
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if n > 0 && buf[0] == 0x09 {
			log.Printf("Received data: % X", buf[:n])
			batteryLevel := int16(buf[1])
			pollingRate := buf[2]

			log.Printf("Battery level: %d", batteryLevel)
			log.Printf("Polling rate: %d", pollingRate)
		}

	}
}
