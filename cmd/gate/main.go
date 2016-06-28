package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/IndoorPosSquad/dwm1000_driver"
)

var (
	dev     = flag.String("d", "COM1", "COM端口或者TTY串口设备")
	httpapi = flag.String("p", "http://127.0.0.1:8889/", "默认api地址")
	data    = new(dataStorage)
	d       *dw1000.DW1000
)

type Record struct {
	Quantity int
	ID       string
}

func decode(msg []byte, src *dw1000.Addr) {
	r := new(Record)
	if err := json.Unmarshal(msg, r); err != nil {
		log.Println("解析Json出错! ", err)
		return
	}
	log.Printf("货物ID: %s 数量: %d", r.ID, r.Quantity)
	resp, err := http.Get(fmt.Sprintf("%sdata/%02x/%s/%s?Quantity=%d", *httpapi, src, "inbound", r.ID, r.Quantity))
	if err != nil {
		log.Println("发送http请求错误: ", err)
		return
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Println("读取返回错误: ", err)
	} else {
		log.Println(string(body))
	}
	resp.Body.Close()
}

type dataStorage struct {
	sync.Mutex
	data map[string]string
}

func (d *dataStorage) Get(key string) (val string, ok bool) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		return
	}
	val, ok = d.data[key]
	return
}

func (d *dataStorage) Put(key, val string) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		d.data = map[string]string{key: val}
		return
	}
	d.data[key] = val
	return
}

func (d *dataStorage) Del(key string) {
	d.Lock()
	defer d.Unlock()
	if d.data == nil {
		return
	}
	delete(d.data, key)
	return
}

func mc(data []byte, src *dw1000.Addr) {
	log.Printf("Got msg from %02x : %s\n", src, data)
	decode(data, src)
}

func bc(distance float64, src *dw1000.Addr) {
	log.Printf("Distance to %02x : %3.2f\n", src, distance)
	key := fmt.Sprintf("%02x", src)
	if distance > 2.5 {
		data.Del(key)
		return
	}
	if _, ok := data.Get(key); !ok {
		data.Put(key, "y")
		time.AfterFunc(time.Minute, func() { data.Del(key) })
		d.SendTo([]byte("send"), src)
	}
}

func ec(d *dw1000.DW1000, e1 error, e2 error) {
	log.Printf("Error: %v, %v\n", e1, e2)
	d.Close()
	os.Exit(2)
}

func main() {
	var err error
	flag.Parse()
	a := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF1}}
	c := &dw1000.Config{Address: a, AutoBeacon: false, SerialPort: *dev}
	d, err = dw1000.OpenDevice(c)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetCallbacks(bc, mc, ec)
	for {
		d.Beacon()
		time.Sleep(time.Second)
	}
}
