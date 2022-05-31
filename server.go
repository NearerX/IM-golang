package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)
	user.Online()

	isLive := make(chan bool)
	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			len, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println("Conn read error:", err)
				return
			}
			if len == 0 {
				user.Offline()
				return
			}
			msg := string(buf[:len-1])
			user.DoMessage(msg)
			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 600):
			user.SendMsg("you ara out\n")
			close(user.Chan)
			conn.Close()
			return
		}
	}
}

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.Chan <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}

	defer listen.Close()

	go this.ListenMessage()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listener err:", err)
			continue
		}
		go this.Handler(conn)
	}
}
