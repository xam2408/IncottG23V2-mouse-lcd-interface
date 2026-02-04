package main

import (
	"fmt"
	"log"
	"time"

	"github.com/karalabe/hid"
	"go.bug.st/serial"
)

const (
	vendorID          = 0x093A
	productID         = 0x522C
	productIDCharging = 0x622C
	mouseStatusIdx    = 6
	comPortName       = "COM3"
)

func presets(dpi int, pollingRate int) (int, int) {
	dpiPresets := []int{400, 800, 1200, 1600, 2400, 3200, 6400}
	pollingRatePresets := []int{1000, 500, 250, 125}
	d := 800
	if dpi >= 0 && dpi < len(dpiPresets) {
		d = dpiPresets[dpi]
	}

	p := 1000
	if pollingRate >= 0 && pollingRate < len(pollingRatePresets) {
		p = pollingRatePresets[pollingRate]
	}

	return d, p
}

func findDevice() *hid.DeviceInfo {
	dev := hid.Enumerate(vendorID, productID)
	devCharging := hid.Enumerate(vendorID, productIDCharging)

	if len(dev) > mouseStatusIdx {
		log.Println("Found device (Wireless/Normal)")
		return &dev[mouseStatusIdx]
	}
	if len(devCharging) > mouseStatusIdx {
		log.Println("Found device (Charging/Wired)")
		return &devCharging[mouseStatusIdx]
	}
	return nil
}

func sendUartFrame(port serial.Port, data []byte) {
	if port == nil {
		return
	}

	_, err := port.Write(data)
	if err != nil {
		log.Printf("Błąd wysyłania UART: %v", err)
	}
}

func handleDevicesConnection(info *hid.DeviceInfo, uartPort serial.Port) {
	device, err := info.Open()
	if err != nil {
		log.Printf("Device read error: %v", err)
		return
	}
	defer device.Close()

	log.Println("Device connected... Reading data...")
	buf := make([]byte, 64)

	for {
		n, err := device.Read(buf)
		if err != nil {
			log.Printf("Lost connection (read error): %v", err)
			return
		}

		if n > 0 && buf[0] == 0x09 {
			batteryLevel := int16(buf[1])
			pollingDpiData := buf[2]
			pollingRatePreset := int(pollingDpiData & 0x0F)
			dpiPreset := int((pollingDpiData & 0xF0) >> 4)

			Dpi, pollingRate := presets(dpiPreset, pollingRatePreset)

			if batteryLevel > 100 {
				batteryLevel = batteryLevel - 128
			}

			log.Printf("Bat: %d%% | DPI: %d | Hz: %d", batteryLevel, Dpi, pollingRate)

			msg := fmt.Sprintf("$%d,%d,%d\n", batteryLevel, Dpi, pollingRate)
			sendUartFrame(uartPort, []byte(msg))
		}
	}
}

func main() {
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	log.Println("Attempting to open UART port...")

	// Próba otwarcia portu
	uartPort, err := serial.Open(comPortName, mode)

	if err != nil {
		log.Printf("Error opening UART port %s: %v", comPortName, err)
		uartPort = nil
	} else {
		log.Println("Success Uart connected.")
		defer uartPort.Close()
	}

	log.Println("Mouse service started...")

	for {
		targetInfo := findDevice()
		if targetInfo == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		handleDevicesConnection(targetInfo, uartPort)

		log.Println("Restart search process...")
		time.Sleep(1 * time.Second)
	}
}
