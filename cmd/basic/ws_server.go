package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"
)

func echoHandler(ws *websocket.Conn) {
	fmt.Println("Hello")
	defer ws.Close()
	go func() {
		defer ws.Close()
		buf := make([]byte, 128)
		for {
			n, err := ws.Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("incoming: ", string(buf[:n]))
			str := strings.Split(string(buf[:n]), ",")
			if str[0] == "COORD" {
				x, _ := strconv.Atoi(str[0+1])
				A.X = float64(x) / 100.0
				x, _ = strconv.Atoi(str[1+1])
				A.Y = float64(x) / 100.0
				x, _ = strconv.Atoi(str[2+1])
				A.Z = float64(x) / 100.0

				x, _ = strconv.Atoi(str[0+3+1])
				B.X = float64(x) / 100.0
				x, _ = strconv.Atoi(str[1+3+1])
				B.Y = float64(x) / 100.0
				x, _ = strconv.Atoi(str[2+3+1])
				B.Z = float64(x) / 100.0

				x, _ = strconv.Atoi(str[0+6+1])
				C.X = float64(x) / 100.0
				x, _ = strconv.Atoi(str[1+6+1])
				C.Y = float64(x) / 100.0
				x, _ = strconv.Atoi(str[2+6+1])
				C.Z = float64(x) / 100.0

				x, _ = strconv.Atoi(str[0+9+1])
				D.X = float64(x) / 100.0
				x, _ = strconv.Atoi(str[1+9+1])
				D.Y = float64(x) / 100.0
				x, _ = strconv.Atoi(str[2+9+1])
				D.Z = float64(x) / 100.0

				fmt.Printf("Anchor updated:\n%v\n%v\n%v\n%v\n", A, B, C, D)
			} else if str[0] == "CALIB" {
				x, _ := strconv.Atoi(str[0+1])
				y, _ := strconv.Atoi(str[1+1])
				z, _ := strconv.Atoi(str[2+1])

				xx, _ := strconv.Atoi(str[0+3+1])
				yy, _ := strconv.Atoi(str[1+3+1])
				zz, _ := strconv.Atoi(str[2+3+1])
				实测的点 := &Node{X: float64(x) / 100.0, Y: float64(y) / 100.0, Z: float64(z) / 100.0}
				卡尔曼滤波的点 := &Node{X: float64(xx) / 100.0, Y: float64(yy) / 100.0, Z: float64(zz) / 100.0}
				OffetA = distance(实测的点, A) - distance(卡尔曼滤波的点, A)
				OffetB = distance(实测的点, B) - distance(卡尔曼滤波的点, B)
				OffetC = distance(实测的点, C) - distance(卡尔曼滤波的点, C)
				fmt.Printf("Offset 更新: (%3.2f, %3.2f, %3.2f)\n", OffetA, OffetB, OffetC)
			}
		}
	}()
	for {
		result := <-PosChan
		msg := fmt.Sprintf("%3.2f,%3.2f,%3.2f\n", result.X, result.Y, result.Z)
		fmt.Println("msg:", msg)
		if _, err := ws.Write([]byte(msg)); err != nil {
			fmt.Println(err)
			break
		}
	}
}

func wsService() {
	http.Handle("/pos", websocket.Handler(echoHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
