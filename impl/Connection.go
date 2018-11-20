package impl

import (
	"errors"
	"src/github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte
	mutex     sync.Mutex //锁
	isClose   bool       // conn是否关闭
}

func InitWebsocket(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:  wsConn,
		inChan:  make(chan []byte, 1000), //容纳1000个元素
		outChan: make(chan []byte, 1000),
	}
	// 读消息
	go conn.readLoop()
	// 写消息
	go conn.writeLoop()
	return conn, nil
}

/**
读取消息
 */
func (this *Connection) ReadMessage() (message []byte, err error) {
	select {
	case message = <-this.inChan:
	case <-this.closeChan:
		err = errors.New("connect is closed")

	}
	return message, err
}

/**
写消息
 */
func (this *Connection) WriteMessage(data []byte) (err error) {
	select {
	case this.outChan <- data:
	case <-this.closeChan:
		err = errors.New("connect is closed")
	}
	return err
}

func (this *Connection) Close() {
	// 线程安全，可重入
	this.wsConn.Close()
	//这里只能执行一次，需要保证线程安全
	this.mutex.Lock()
	if !this.isClose {
		close(this.closeChan)
		this.isClose = true
	}
	this.mutex.Unlock()
}
func (this *Connection) readLoop() {
	var (
		message []byte
		err     error
	)
	for {
		if _, message, err = this.wsConn.ReadMessage(); err != nil {
			goto STOP
		}
		// 如果容量不够，阻塞在此处，等待inChan有容量
		select {
		case this.inChan <- message:

		case <-this.closeChan:
			goto STOP
		}
		// 将读取到的消息放到inChan chanel中
	}
STOP:
	this.wsConn.Close()
}

func (this *Connection) writeLoop() {
	var (
		message []byte
		err     error
	)
	for {
		select {
		case message = <-this.outChan:
		case <-this.closeChan:
			goto STOP
		}

		if err = this.wsConn.WriteMessage(websocket.TextMessage, message); err != nil {
			goto STOP
		}
	}
STOP:
	this.wsConn.Close()
}
