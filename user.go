package main

import "net"

type User struct {
	Name string
	Addr string
	Chan chan string
	conn net.Conn
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.Chan
		this.conn.Write([]byte(msg + "\n"))
	}
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Chan: make(chan string),
		conn: conn,
	}
	go user.ListenMessage()
	return user
}
