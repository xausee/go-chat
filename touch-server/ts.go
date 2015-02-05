package main

import (
	"flag"
	"fmt"
	"go-chat/server"
)

// 地址可以是域名，例如： www.sanwenjia.net:8000
// 也可以是IP地址，例如： 127.0.0.1:8000
func GetAddress() string {
	address := flag.String("address", "127.0.0.1:8000", "服务器地址")
	flag.Parse()

	if *address == "127.0.0.1:8000" {
		fmt.Println("未设置服务器地址， 使用测试地址： " + *address)
	} else {
		fmt.Println("使用指定的服务器地址： " + *address)
	}

	return *address
}

func main() {
	fmt.Println("初始化服务器... (按下Ctrl-C停止服务)")
	s := server.NewServer("tcp", GetAddress())
	defer s.Close()
	s.Listen()
	s.Run()
}
