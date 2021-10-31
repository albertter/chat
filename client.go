package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       100,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil
	}
	client.conn = conn
	return client

}
func (client *Client) DealResponse() {
	io.Copy(os.Stdin, client.conn)
}
func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		return false
	}
	return true

}
func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println("请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println(err)
				break
			}

		}
		chatMsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(err)
		return
	}

}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteNmae string
	var chatMsg string
	client.SelectUsers()
	fmt.Println("请输入私聊对象，exit退出")
	fmt.Scanln(&remoteNmae)
	for remoteNmae != "exit" {
		fmt.Println("请输入私聊内容，exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteNmae + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println(err)
					break
				}

			}
			chatMsg = ""
			fmt.Println("请输入聊天内容，exit退出")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println("请输入私聊对象，exit退出")
		fmt.Scanln(&remoteNmae)
	}
}
func (client *Client) Run() {
	for client.flag != 0 {
		for !client.menu() {
		}
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		return false
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip")
	flag.IntVar(&serverPort, "port", 9988, "设置端口默认9988")
}
func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("客户端连接服务器失败")
		return
	}
	go client.DealResponse()

	fmt.Println("连接服务器成功")
	client.Run()
}
