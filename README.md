[![GoDoc](https://godoc.org/github.com/IndoorPosSquad/dwm1000_driver?status.svg)](http://godoc.org/github.com/IndoorPosSquad/dwm1000_driver)

# dwm1000_driver
PC端dwm1000的驱动程序，通过串口使用，搭配v2版底层固件。

目前已经实现了部分设置的API，以及Beacon()测距、SendTo()发送数据。

# 安装使用

```shell
go get github.com/IndoorPosSquad/dwm1000_driver
```

cmd下面包含了三个demo。Basic是定时Beacon测距，TX是发送数据，RX是接受并显示数据。

# 例子

```go
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
```

更多文档请参照godoc

# TODO 已经更多功能

* [ ] 实现驱动Daemon，应用使用socket和dwm1000交互
* [ ] 更多TODO请参照固件的README
