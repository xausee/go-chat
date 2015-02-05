package server

import "net"

// 服务器端连接结构体， 对net.Conn的封装
type Connect struct {
	name     string   // 用户名
	room     string   // 所在房间
	AConnect net.Conn // A端连接：发起通信端的连接
	BConnect net.Conn // B端连接：一对一时的远端连接，可为空
}

func NewConnect(name, room string, aConnect, bConnect net.Conn) *Connect {
	connect := &Connect{name, room, aConnect, bConnect}
	return connect
}
