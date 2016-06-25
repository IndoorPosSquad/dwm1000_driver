/*
Package dw1000
dw1000是PC端dwm1000的驱动程序，通过串口使用，搭配v2版底层固件。
目前已经实现了部分设置的API，以及Beacon()测距、SendTo()发送数据。

安装：
	go get github.com/IndoorPosSquad/dwm1000_driver

cmd下面包含了三个demo。Basic是定时Beacon测距，TX是发送数据，RX是接受并显示数据。

使用：

	package main

	import (
		"log"
		"os"
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
		os.Exit(2)
	}

	func main() {
		a := &dw1000.Addr{PANID: []byte{0xCA, 0xDE}, MAC: []byte{0xFF, 0xF1}}
		c := &dw1000.Config{Address: a, AutoBeacon: true, SerialPort: "COM4"}
		d, err := dw1000.OpenDevice(c)
		if err != nil {
			log.Fatalln(err)
		}
		d.SetCallbacks(bc, mc, ec)
		for {
			// d.Beacon()
			time.Sleep(time.Second)
		}
	}

更多文档请参照godoc
*/

package dw1000

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

const (
	UsartMsg        = 0x00
	UsartBeacon     = 0x01
	UsartSetAddr    = 0x02
	UsartRST        = 0x03
	UsartAutoBeacon = 0x04
	UsartLog        = 0x05
)

var (
	db      = []byte{0xDB, 0xDC}
	newline = []byte{0xDB, 0xDD}
	// ErrRead 是在读串口发生错误时产生的
	ErrRead = fmt.Errorf("%s", "Failed to read from serial port")
	// ErrMsgTooLarge 是当发送的数据超过了114字节时产生
	ErrMsgTooLarge = fmt.Errorf("%s", "The payload is too large")
)

// DW1000 是驱动抽象对象
type DW1000 struct {
	sync.Mutex
	serialPort     *serial.Port
	buffer         *bufio.Reader
	callbackLock   sync.Mutex
	callbackset    bool
	beaconCallback func(distance float64, src *Addr)
	msgCallback    func(payload []byte, src *Addr)
	errCallback    func(d *DW1000, e1, e2 error)
	config         *Config
	close          chan interface{}
}

func (d *DW1000) write(data []byte) error {
	d.Lock()
	defer func() { time.AfterFunc(100*time.Millisecond, d.Unlock) }()
	_, err := d.serialPort.Write(data)
	return err
}

// Config 包含了初始化一个DW1000驱动抽象对象的所有参数
// 目前仅有串口地址、自动Beacon以及本机MAC地址几个选项
//
// 例如:
// c := &dw1000.Config{Address:addr, AutoBeacon: true, SerialPort: "COM4"}
type Config struct {
	Address    *Addr
	AutoBeacon bool
	SerialPort string
}

// Addr 是DW1000使用的地址
// 其中包含了PANID以及短地址
// 要注意这个数据结构是低位在前，高位在后
type Addr struct {
	MAC   []byte
	PANID []byte
}

// AddrFromString 从字符串初始化一个DW1000物理地址
// 注意高低位
//
// 例如:
// a := AddrFromString("YuKi")
// 产生的 PANID 是 uY
// 产生的 MAC 是 iK
// 写入DW1000的时候就按照这个顺序，对于应用层就不用考虑这么多了
func AddrFromString(s string) (a *Addr, err error) {
	if len(s) < 4 {
		return nil, fmt.Errorf("Address(%s) is too short", s)
	}
	mac := make([]byte, 2)
	panid := make([]byte, 2)
	mac[0] = s[3]
	mac[1] = s[2]
	panid[0] = s[1]
	panid[1] = s[0]
	a.MAC = mac
	a.PANID = panid
	return
}

func (a *Addr) String() string {
	tmp := make([]byte, 4)
	tmp[0] = a.PANID[1]
	tmp[1] = a.PANID[0]
	tmp[2] = a.MAC[1]
	tmp[3] = a.MAC[0]
	return string(tmp)
}

// SetAddr 设置DW1000的MAC地址
func (d *DW1000) SetAddr(addr *Addr) error {
	buf := append(make([]byte, 0, 6), UsartSetAddr, 4)
	buf = append(buf, addr.PANID[:2]...)
	buf = append(buf, addr.MAC[:2]...)
	return d.write(buf)
}

// Beacon 发射一次Beacon
func (d *DW1000) Beacon() error {
	buf := append(make([]byte, 0, 1), 0x01)
	return d.write(buf)
}

// AutoBeacon 开启DW1000的自动Beacon
func (d *DW1000) AutoBeacon(enable bool) error {
	buf := append(make([]byte, 0, 3), UsartAutoBeacon, 1)
	if enable {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	return d.write(buf)
}

// Reset 重置DW1000
func (d *DW1000) Reset() error {
	buf := append(make([]byte, 0, 1), UsartRST)
	return d.write(buf)
}

func (d *DW1000) rawSend(data []byte) error {
	buf := append(make([]byte, 0, 2+len(data)), UsartMsg, byte(len(data)))
	buf = append(buf, data...)
	return d.write(buf)
}

// SendTo 将数据发送到指定地址
// 这个data就是发送的数据包
//
// 例如:
// d.SendTo([]byte("hello"), dst)
func (d *DW1000) SendTo(data []byte, dst *Addr) error {
	if len(data) > 114 {
		return ErrMsgTooLarge
	}
	buf := make([]byte, 0, 9+len(data)+2)
	buf = append(buf, 0x41, 0x88, 0x00)
	buf = append(buf, dst.PANID[:2]...)
	buf = append(buf, dst.MAC[:2]...)
	buf = append(buf, 0xff, 0xff) // will be replaced by stm32 firmware
	buf = append(buf, data...)
	buf = append(buf, 0x00, 0x00) // FCS Place Holder
	return d.rawSend(buf)
}

// SetCallbacks 设置回调函数
// 三个回调函数分别处理接收数据、定位Beacon以及错误信息
func (d *DW1000) SetCallbacks(bc func(distance float64, src *Addr), mc func(payload []byte, src *Addr), ec func(d *DW1000, e1, e2 error)) {
	d.callbackLock.Lock()
	d.callbackset = true
	d.beaconCallback = bc
	d.msgCallback = mc
	d.errCallback = ec
	d.callbackLock.Unlock()
}

func (d *DW1000) run() {
	for {
		select {
		case <-d.close:
			return
		default:
			line, err := d.buffer.ReadString('\n')
			if err != nil {
				d.callbackLock.Lock()
				if d.callbackset {
					d.errCallback(d, ErrRead, err)
				}
				d.callbackLock.Unlock()
				time.Sleep(time.Microsecond)
				return
			}
			line = line[:len(line)-1]
			line = strings.Replace(line, string(db), string([]byte{0xDB}), -1)
			line = strings.Replace(line, string(newline), string([]byte{'\n'}), -1)
			bline := []byte(line)
			mtype := bline[0]
			switch mtype {
			case UsartMsg:
				a := &Addr{PANID: bline[5:7], MAC: bline[9:11]}
				d.callbackLock.Lock()
				if d.callbackset {
					d.msgCallback(bline[11:len(bline)-2], a)
				}
				d.callbackLock.Unlock()
			case UsartBeacon:
				a := &Addr{PANID: d.config.Address.PANID, MAC: bline[2:4]}
				dis := (int32(bline[4])) + (int32(bline[5]) << 8) + (int32(bline[6]) << 16) + (int32(bline[7]) << 24)
				distance := float64(dis) / 2.0 * 1.0 / 499.2e6 / 128.0 * 299702547.0
				d.callbackLock.Lock()
				if d.callbackset {
					d.beaconCallback(distance, a)
				}
				d.callbackLock.Unlock()
			case UsartLog:
				log.Printf("%s\n", bline[2:])
			default:
			}
		}
	}
}

// Close 安全地关闭一个DW1000对象
func (d *DW1000) Close() error {
	close(d.close)
	return d.serialPort.Close()
}

// OpenDevice 读取设置，初始胡一个DW1000对象
func OpenDevice(c *Config) (d *DW1000, err error) {
	d = new(DW1000)
	d.config = c
	d.close = make(chan interface{})
	serialConf := &serial.Config{Name: c.SerialPort, Baud: 115200, StopBits: serial.Stop1}
	if d.serialPort, err = serial.OpenPort(serialConf); err != nil {
		return nil, err
	}
	d.buffer = bufio.NewReader(d.serialPort)
	if err = d.SetAddr(c.Address); err != nil {
		return nil, err
	}
	if err = d.SetAddr(c.Address); err != nil {
		return nil, err
	}
	if err = d.AutoBeacon(c.AutoBeacon); err != nil {
		return nil, err
	}
	go d.run()
	return
}
