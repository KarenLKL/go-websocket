package main

import (
	"fmt"
	"go-websocket/impl"
	"log"
	"net/http"
	"src/github.com/gorilla/websocket"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		wsConn *websocket.Conn
		msg    []byte
		conn   *impl.Connection
	)
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		log.Fatal("connect websocket fatal...")
		return
	}
	if conn, err = impl.InitWebsocket(wsConn); err != nil {
		goto STOP
	}
	// 测试法消息
	go func() {
		for {
			if err := conn.WriteMessage([]byte("test heart..")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		if msg, err = conn.ReadMessage(); err != nil {
			goto STOP
		}
		fmt.Printf("get a message from cient:%s", string(msg))
		fmt.Println("")
		if err = conn.WriteMessage([]byte("form server:" + string(msg))); err != nil {
			goto STOP
		}
	}
STOP:
	conn.Close()
}

func main() {
	http.HandleFunc("/ws", websocketHandler)
	http.ListenAndServe("0.0.0.0:8888", nil)
}
