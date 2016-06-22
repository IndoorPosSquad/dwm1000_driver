package main

import (
	"log"
	"time"

	"github.com/IndoorPosSquad/dwm1000_driver"
)

func mc(data []byte, src *dw1000.Addr) {
	log.Printf("Got msg from %02x : %s\n", src, data)
}

func bc(distance float64, src *dw1000.Addr) {
	log.Printf("Distance to %02x : %3.2f\n", src, distance)
}

func ec(d *dw1000.DW1000, e1 error, e2 error) {
	log.Printf("Error: %v, %v\n", e1, e2)
	d.Close()
}

func main() {
	a := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF1}}
	dst := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF2}}
	c := &dw1000.Config{Address: a, AutoBeacon: false, SerialPort: "COM4"}
	d, err := dw1000.OpenDevice(c)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetCallbacks(bc, mc, ec)
	for {
		err := d.SendTo([]byte("hello"), dst)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second)
	}
}
