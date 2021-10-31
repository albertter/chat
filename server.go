package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	ip        string
	port      int
	onlineMap map[string]*User
	mapLock   sync.RWMutex
	message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		ip:        ip,
		port:      port,
		onlineMap: make(map[string]*User),
		message:   make(chan string),
	}
	return server
}

// ListenMessage 监听message广播消息的channel的go程，一旦有消息就发送到所有到用户
func (s *Server) ListenMessage() {
	for {
		msg := <-s.message
		s.mapLock.Lock()
		for _, cli := range s.onlineMap {
			cli.c <- msg
		}
		s.mapLock.Unlock()
	}
}
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.addr + "]" + user.name + ":" + msg
	s.message <- sendMsg
}
func (s *Server) Handler(conn net.Conn) {
	fmt.Println("连接建立成功")
	user := NewUser(conn, s)
	user.Online()
	//监听用户是否活跃
	isLive := make(chan bool)
	// 接受客户端到消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err", err)
				return
			}
			msg := string(buf[:n-1])
			// 用户发送消息
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	//当前handler阻塞
	for {
		select {
		case <-isLive:
			//更新定时器
		case <-time.After(time.Minute * 10):
			// 已经超时
			// 将当前用户强制关闭
			user.sendMsg("你被T了\n")
			close(user.c)
			conn.Close()
			return

		}
	}

}

func (s *Server) start() {
	// listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		fmt.Println("listen", err)
		return
	}
	defer listener.Close()
	go s.ListenMessage()
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}
}
