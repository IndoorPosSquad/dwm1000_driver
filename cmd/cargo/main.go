package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/IndoorPosSquad/dwm1000_driver"
	"github.com/lixin9311/icli-go"
	"github.com/tarm/serial"
)

var (
	lock   sync.Mutex
	sent   bool
	gline  string
	d      *dw1000.DW1000
	target *dw1000.Addr
)

type Record struct {
	Quantity int
	ID       string
}

var data = map[string]int{
	"cargo_1": 10,
	"cargo_2": 20,
}

func mc(data []byte, src *dw1000.Addr) {
	fmt.Printf("Got msg from %02x : %s\n", src, data)
	ssrc := fmt.Sprintf("%02x", src)
	if string(data) == "send" {
		fmt.Println("准备好发送货物信息到", ssrc)
		lock.Lock()
		target = src
		sent = false
		lock.Unlock()
	}
}

func bc(distance float64, src *dw1000.Addr) {
	return
}

func ec(d *dw1000.DW1000, e1 error, e2 error) {
	fmt.Printf("Error: %v, %v\n", e1, e2)
	d.Close()
	os.Exit(2)
}

func run(args ...string) error {
	var err error
	flagset := flag.NewFlagSet(args[0], flag.ContinueOnError)
	// os.Stdout is redirected.
	flagset.SetOutput(os.Stdout)
	dev := flagset.String("d", "COM1", "COM端口或者TTY串口设备")
	if err = flagset.Parse(args[1:]); err != nil {
		return fmt.Errorf("处理参数失败: %v", err)
	}
	a := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF2}}
	c := &dw1000.Config{Address: a, AutoBeacon: false, SerialPort: *dev}
	d, err = dw1000.OpenDevice(c)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	d.SetCallbacks(bc, mc, ec)
	return nil
}

func add(args ...string) error {
	if len(args) != 2 {
		return fmt.Errorf("处理参数失败")
	}
	handleCode(args[1])
	return nil
}

func send(args ...string) error {
	lock.Lock()
	defer lock.Unlock()
	if sent {
		fmt.Println("货物信息已经发送，请勿重复操作!")
		return nil
	}
	for k, v := range data {
		r := &Record{Quantity: v, ID: k}
		msg, err := json.Marshal(r)
		if err != nil {
			fmt.Println("json序列化错误 ", err)
			return err
		}
		if err := d.SendTo(msg, target); err != nil {
			fmt.Println("发送数据出错 ", err)
			return err
		}
		delete(data, k)
	}
	sent = true
	return nil
}

func exit(args ...string) error {
	return icli.ExitIcli
}

// should also process nil error
func errorhandler(e error) error {
	if e != nil {
		fmt.Println(e)
		return icli.ExitIcli
	}
	return e
}

var boundStatus = "inbound"

var exitCode = "6901236341582"

// statusChange 定义某个条形码是改变进出方向
var statusChange = "6907992101330"

// startBarcode 是启动二维码扫描的进程
func startBarcode(args ...string) error {
	// sub flag set.
	var err error
	flagset := flag.NewFlagSet(args[0], flag.ContinueOnError)
	// os.Stdout is redirected.
	flagset.SetOutput(os.Stdout)
	dev := flagset.String("d", "COM1", "COM端口或者TTY串口设备")
	if err = flagset.Parse(args[1:]); err != nil {
		return fmt.Errorf("处理参数失败: %v", err)
	}

	c := &serial.Config{Name: *dev, Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		return fmt.Errorf("打开扫码器设备出错: %v", err)
	}
	buffer := bufio.NewReader(s)
	for {
		result, err := buffer.ReadString(13)
		result = result[:len(result)-1]
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("扫码结果: ", result)
		if result == statusChange {
			if boundStatus == "inbound" {
				boundStatus = "outbound"
			} else {
				boundStatus = "inbound"
			}
			fmt.Println("录入状态改变，当前状态: ", boundStatus)
		} else if result == exitCode {
			fmt.Println("exit")
			return nil
		} else {
			handleCode(result)
		}
	}
}

func printall(args ...string) error {
	lock.Lock()
	for k, v := range data {
		fmt.Printf("key: %s, val: %d\n", k, v)
	}
	lock.Unlock()
	return nil
}

// handleCode 插入一条二维码数据
func handleCode(code string) {
	lock.Lock()
	data[code]++
	lock.Unlock()
}

func main() {
	icli.AddCmd([]icli.CommandOption{
		{"run", "启动主程序", run},
		{"send", "发送货物信息", send},
		{"exit", "exit", exit},
		{"scan", "scan", startBarcode},
		{"add", "add", add},
		{"print", "print", printall},
	})
	// utf-8 safe
	icli.SetPromt("输入命令 >")
	icli.Start(errorhandler)
}
