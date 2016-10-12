package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IndoorPosSquad/dwm1000_driver"
)

var PosChan chan *Node
var d = flag.String("d", "COM4", "Serial port.")
var data = map[int]float64{}

func mc(data []byte, src *dw1000.Addr) {
	log.Printf("Got msg from %02x : %s\n", src, data)
}

func bc(distance float64, src *dw1000.Addr) {
	log.Printf("Distance to %02x : %3.2f\n", src, distance)
	data[int(src.MAC[0])] = distance
}

func ec(d *dw1000.DW1000, e1 error, e2 error) {
	log.Printf("Error: %v, %v\n", e1, e2)
	d.Close()
	os.Exit(2)
}

func main() {
	flag.Parse()
	a := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF0}}
	dst := []*dw1000.Addr{{MAC: []byte{0xF1, 0xFF}}, {MAC: []byte{0xF2, 0xFF}}, {MAC: []byte{0xF3, 0xFF}}}
	c := &dw1000.Config{Address: a, AutoBeacon: false, SerialPort: *d}
	d, err := dw1000.OpenDevice(c)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetCallbacks(bc, mc, ec)
	PosChan = make(chan *Node)
	go wsService()
	for i := 0; i < 3; i++ {
		d.Distance(dst[i])
		// d.Beacon()
		time.Sleep(10 * time.Millisecond)
		if i == 2 {
			i = -1
			str := fmt.Sprintf("%d,%d,%d\n", int(data[0xf1]*100), int(data[0xf2]*100), int(data[0xf3]*100))
			fmt.Println(str)
			result := Solve(data[0xf1], data[0xf2], data[0xf3])
			fmt.Printf("解算坐标(%3.2f, %3.2f, %3.2f)\n", result.X, result.Y, result.Z)
			PosChan <- result
			time.Sleep(30 * time.Millisecond)
		}
	}
}
