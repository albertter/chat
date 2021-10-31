package main

import (
	"net"
	"strings"
)

type User struct {
	name string
	addr string
	c    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		name:   userAddr,
		addr:   userAddr,
		c:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}
func (u *User) ListenMessage() {
	for {
		msg := <-u.c
		u.conn.Write([]byte(msg + "\n"))
	}
}

//给当前用户对应客户端发送消息
func (u User) sendMsg(msg string) {
	u.conn.Write([]byte(msg))
}
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.onlineMap {
			onlineMsg := "[" + user.addr + "]" + user.name + ":" + "在线\n"
			u.sendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.onlineMap[newName]
		if ok {
			u.sendMsg("当前用户名已被使用\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.onlineMap, u.name)
			u.server.onlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.name = newName
			u.sendMsg("已经更新用户名：" + u.name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式：to｜张三｜消息内容

		//获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.sendMsg("消息格式不正确\n")
			return
		}
		//根据用户名查询user对象
		remoteUser, ok := u.server.onlineMap[remoteName]
		if !ok {
			u.sendMsg("该用户名不存在\n")
			return
		}
		//获取消息内容 发送
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.sendMsg("消息格式不正确\n")
			return
		}
		remoteUser.sendMsg(u.name + "说:" + content+"\n")

	} else {
		u.server.BroadCast(u, msg)
	}
}
func (u *User) Online() {

	// 用户上线，将用户加入到onlinemap中
	u.server.mapLock.Lock()
	u.server.onlineMap[u.name] = u
	u.server.mapLock.Unlock()
	// 广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}
func (u *User) Offline() {
	// 用户上线，将用户从onlinemap中删除
	u.server.mapLock.Lock()
	delete(u.server.onlineMap, u.name)
	u.server.mapLock.Unlock()
	// 广播当前用户上线消息
	u.server.BroadCast(u, "下线")
}
