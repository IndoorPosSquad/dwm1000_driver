package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/IndoorPosSquad/dwm1000_driver"
)

type SyncRecord struct {
	Master         *dw1000.Addr
	Slave          *dw1000.Addr
	ID             byte
	T1, T2, T3, T4 int64
	Delta          int64
	Finished       byte
	Timestamp      time.Time
}

type TDOARecord struct {
	Anchor    *dw1000.Addr
	Tag       *dw1000.Addr
	ID        byte
	TS        int64
	Timestamp time.Time
}

var (
	d         = flag.String("d", "COM4", "Serial port.")
	port      = flag.Int("p", 8888, "port")
	a         *dw1000.Addr
	rpcclient *rpc.Client
	role      = flag.String("r", "master", "role, master or slave")
	logfile   = flag.String("o", "rtls.log", "log file")
	file      *os.File
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
	os.Exit(2)
}

func sc(src *dw1000.Addr, id byte, T1, T2, T3, T4 uint64, master bool) {
	data := new(SyncRecord)
	data.ID = id
	data.T1 = int64(T1)
	data.T2 = int64(T2)
	data.T3 = int64(T3)
	data.T4 = int64(T4)
	data.Master = a
	data.Slave = src
	var reply byte
	err := rpcclient.Call("RTLSService.Report", data, &reply)
	if err != nil {
		log.Println("Fail to report, ", err)
	}
}

func tc(src *dw1000.Addr, id byte, ts uint64) {
	data := new(TDOARecord)
	data.ID = id
	data.TS = int64(ts)
	data.Tag = src
	data.Anchor = a
	data.Timestamp = time.Now()
	// var reply byte
	// err := rpcclient.Call("RTLSService.TDOA", data, &reply)
	// if err != nil {
	// 	log.Println("Fail to call TDOA, ", err)
	// }
	str := fmt.Sprintf("t %d %d %02x %d\n", data.Timestamp.UnixNano()/1000/1000, data.ID, data.Anchor, data.TS)
	log.Print(str)
	file.WriteString(str)
}

func main() {
	flag.Parse()
	var err error
	var id byte
	file, err = os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("role:", *role)
	if *role == "1" {
		a = &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF1}}
	} else if *role == "2" {
		a = &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF2}}
	} else if *role == "3" {
		a = &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF3}}
	} else if *role == "4" {
		a = &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF4}}
	}
	dst2 := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF2}}
	dst3 := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF3}}
	dst4 := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF4}}
	c := &dw1000.Config{Address: a, AutoBeacon: false, SerialPort: *d}
	d, err := dw1000.OpenDevice(c)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("remoteaddr: ", fmt.Sprintf("127.0.0.1:%d", *port))
	rpcclient, err = rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	d.SetCallbacks(bc, mc, ec)
	d.SetTDOACallback(sc, tc)
	for {
		if *role == "1" {
			err := d.ClcSync(dst2, id)
			if err != nil {
				log.Fatalln(err)
			}
			err = d.ClcSync(dst3, id)
			if err != nil {
				log.Fatalln(err)
			}
			err = d.ClcSync(dst4, id)
			if err != nil {
				log.Fatalln(err)
			}
			if id == 0xff {
				id = 0
			} else {
				id++
			}
			time.Sleep(150 * time.Millisecond)
		} else {
			time.Sleep(time.Hour)
		}
	}
}
