package main

import (
	"log"
	"time"

	"github.com/karalabe/hid"
)

const (
	vendorID          = 0x093A
	productID         = 0x522C
	productIDCharging = 0x622C
	comPort           = "COM3" //stm32 serial port
	mouseStatusIdx    = 6
)

func presets(dpi int, pollingRate int) (int, int) {
	// DPI presets: 800, 1200, 1600, 2400
	// Polling rate presets: 125Hz, 250Hz, 500Hz, 1000Hz
	dpiPresets := []int{
		400,
		800,
		1200,
		1600,
		2400,
		3200,
		6400}
	pollingRatePresets := []int{
		1000,
		500,
		250,
		125}

	return dpiPresets[dpi], pollingRatePresets[pollingRate]

}
func main() {
	dev := hid.Enumerate(vendorID, productID)
	devCharging := hid.Enumerate(vendorID, productIDCharging)

	log.Printf("Found devices: %d", len(dev), len(devCharging))

	var TargetDeviceInfo *hid.DeviceInfo

	if len(dev) > mouseStatusIdx {
		TargetDeviceInfo = &dev[mouseStatusIdx]
	} else if len(devCharging) > mouseStatusIdx {
		TargetDeviceInfo = &devCharging[mouseStatusIdx]
	} else {
		log.Fatal("No target device found")
		return
	}

	device, err := TargetDeviceInfo.Open()

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
			pollingDpiData := buf[2]
			pollingRatePreset := int(pollingDpiData & 0x0F)
			dpiPreset := int((pollingDpiData & 0xF0) >> 4)
			Dpi, pollingRate := presets(dpiPreset, pollingRatePreset)

			log.Printf("Battery: %d", batteryLevel)
			log.Printf("Polling rate: %d", pollingRate)
			log.Printf("DPI: %d", Dpi)
		}

	}
}
