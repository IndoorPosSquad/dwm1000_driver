package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/IndoorPosSquad/dwm1000_driver"
)

type RTLSService struct {
	sync.Mutex
	// data[id][slave_addr]
	data [256](map[string]SyncRecord)
	// tdoaData[id][anchor_addr]
	tdoaData [256](map[string]TDOARecord)
	index    byte
}

func (c *RTLSService) Report(data *SyncRecord, resp *byte) error {
	c.Lock()
	defer c.Unlock()
	// log.Printf("New record(%02x)\n", data.ID)
	addr := fmt.Sprintf("%02x", data.Slave)
	records := c.data[data.ID]
	if records == nil {
		data.Finished = 1
		c.data[data.ID] = map[string]SyncRecord{addr: *data}
		*resp = 0
		return nil
	}
	record := records[addr]
	if data.T1 == 0 && data.T2 == 0 && data.T4 == 0 {
		record.T3 = data.T3
		record.Slave = data.Slave
		record.ID = data.ID
		record.Finished++
		record.Timestamp = time.Now()
		records[addr] = record
	} else {
		record.T1 = data.T1
		record.T2 = data.T2
		record.T4 = data.T4
		record.Master = data.Master
		record.Slave = data.Slave
		record.ID = data.ID
		record.Finished++
		record.Timestamp = time.Now()
		records[addr] = record
	}
	if record.Finished%2 == 0 {
		record.Delta = ((record.T2 - record.T1) - (record.T4 - record.T3)) / 2
		if record.Delta < 0 {
			record.Delta += 0xffffffffff
		}
		records[addr] = record
		str := fmt.Sprintf("s %d %d %02x %d %d\n", record.Timestamp.UnixNano()/1000/1000, record.ID, record.Slave, record.T1, record.Delta)
		log.Print(str)
		file.WriteString(str)
	}
	*resp = 0
	return nil
}

func (s *RTLSService) TDOA(data *TDOARecord, resp *byte) error {
	s.Lock()
	defer s.Unlock()
	// log.Printf("New TDOA record(%02x)\n", data.ID)
	addr := fmt.Sprintf("%02x", data.Anchor)
	records := s.tdoaData[data.ID]
	data.Timestamp = time.Now()
	if records == nil {
		s.tdoaData[data.ID] = map[string]TDOARecord{addr: *data}
		*resp = 0
		return nil
	}
	s.tdoaData[data.ID][addr] = *data
	// str := fmt.Sprintf("t %d %d %02x %d\n", data.Timestamp.UnixNano()/1000/1000, data.ID, data.Anchor, data.TS)
	// log.Print(str)
	// file.WriteString(str)
	return nil
}

func (s *RTLSService) dump() {
	s.Lock()
	defer s.Unlock()
}

var (
	service = new(RTLSService)
	port    = flag.Int("p", 8888, "RPC port")
	logfile = flag.String("o", "rtls.log", "log file")
	file    *os.File
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

func main() {
	flag.Parse()
	var err error
	file, err = os.OpenFile(*logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0664)
	defer file.Close()
	if err != nil {
		log.Fatalln("Failed to open log, ", err)
	}
	rpc.Register(service)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalln(err)
	}
	go rpc.Accept(l)
	go http.Serve(l, nil)
	for {
		time.Sleep(10 * time.Second)
		service.dump()
	}
}
