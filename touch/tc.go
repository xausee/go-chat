package main

import (
	"flag"
	"fmt"
	"go-chat/client"
)

// 命令行参数表
var args map[string]string

// 获取所有的命令参数
func GetArgs() map[string]string {
	cmd := flag.String("cmd", "", "命令")
	lname := flag.String("lname", "", "一对一聊天，发起方的用户名")
	rname := flag.String("rname", "", "一对一聊天，接受方的用户名")
	addr := flag.String("addr", "", "一对一聊天，发起方的地址，该地址由响应方发送给服务器")
	address := flag.String("address", "127.0.0.1:8000", "要连接的服务器地址")

	flag.Parse()
	args = map[string]string{
		"cmd":     *cmd,
		"lname":   *lname,
		"rname":   *rname,
		"addr":    *addr,
		"address": *address,
	}
	return args
}

// 地址可以是域名，例如： www.sanwenjia.net:8000
// 也可以是IP地址，例如： 127.0.0.1:8000
func GetAddress() string {
	if args["address"] == "127.0.0.1:8000" {
		fmt.Println("未输入要连接的服务器地址， 使用测试地址： " + args["address"])
	}
	return args["address"]
}

func main() {
	GetArgs()
	c := client.NewClient("tcp", GetAddress())
	c.Connect()
	defer c.Disconnect()

	if !c.CheckArgs(args) {
		c.CreateName()
	}
	go c.ReadMsg()
	c.SendMsg()
}
