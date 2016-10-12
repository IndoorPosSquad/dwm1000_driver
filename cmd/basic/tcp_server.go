package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func tcpService() {
	service := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	request := make([]byte, 128)                          // set maxium request length to 128B to prevent flood attack
	defer conn.Close()                                    // close connection before exit
	if read_len, err := conn.Read(request); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(request[:read_len]))
	}
	for {
		result := <-PosChan
		str := fmt.Sprintf("%3.2f,%3.2f,%3.2f\n", result.X, result.Y, result.Z)
		if _, err := conn.Write([]byte(str)); err != nil {
			fmt.Println(err)
			break
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
