package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

func echoHandler(ws *websocket.Conn) {
	fmt.Println("Hello")
	defer ws.Close()
	for {
		result := <-PosChan
		msg := fmt.Sprintf("%3.2f,%3.2f,%3.2f\n", result.X, result.Y, result.Z)
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
