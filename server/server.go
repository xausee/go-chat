package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"go-chat/constant"
)

// 服务器结构
type Server struct {
	connects  []*Connect          // 所有连接
	users     []string            // 所有用户
	lobby     map[string][]string // 大厅
	rooms     map[string][]string // 聊天室
	listen    net.Listener        // 服务器监听器
	network   string              // 网络协议
	address   string              // 地址以及端口
	Timestamp int                 // 时间戳
}

func NewServer(network, address string) *Server {
	return &Server{
		make([]*Connect, 0),             // 所有连接
		make([]string, 0),               // 所有用户
		make(map[string][]string, 1024), // 大厅
		make(map[string][]string, 1024), // 聊天室
		nil,                    // 服务器监听器
		network,                // 网络协议
		address,                // 地址以及端口
		int(time.Now().Unix())} // 时间戳
}

func (this *Server) Listen() error {
	lis, err := net.Listen(this.network, this.address)
	if err != nil {
		fmt.Printf("监听端口失败: %s, 错误信息: %s\n", this.address, err)
		os.Exit(-1)
	}
	this.listen = lis
	return err
}

func (this *Server) Close() {
	this.listen.Close()
}

func (this *Server) Run() {
	for {
		var res string
		conn, err := this.listen.Accept()
		if err != nil {
			fmt.Println("接受客户端失败: ", err.Error())
			os.Exit(0)
		}

		go func(con net.Conn) {
			connect, err := this.InitConnect(con)
			if err != nil {
				fmt.Printf("创建新用户失败，关闭链接： %v\n", connect.AConnect.RemoteAddr())
				connect.AConnect.Close()
				return
			}
			// 开始传递消息
			var data = make([]byte, 1024)
			for {
				length, err := connect.AConnect.Read(data)
				if err != nil {
					this.Disconnect(connect)
					return
				}

				res = string(data[:length])
				if this.ParseMsg(res, connect.AConnect) {
					continue
				}

				nameInfo := constructUserName(connect.name, constant.SpreadUserNameLength, true)
				sprdMsg := nameInfo + "： " + res
				fmt.Println(sprdMsg)
				this.DispatchMsg(connect, sprdMsg)
			}
		}(conn)
	}
}

// 分析用户发过来的消息，检查是否有命令需要执行，返回布尔值
// 如果执行了命令，返回真值，否则返回假
func (this *Server) ParseMsg(res string, con net.Conn) bool {
	args := strings.Split(res, " ")
	switch args[0] {
	case "list":
		this.ListUser(res, con)
		return true
	//case constant.ConnectRequest:
	//	this.NewTalk(res, con)
	//	return true
	//case constant.ConnectRespone:
	//	//rname := args[1]
	//	address := args[2]
	//	this.MatchConnects(address, con)
	//	return true
	default:
		return false
	}
}

// 获取用户列表信息，每三个为一行， 将结果发送给用户
func (this *Server) ListUser(res string, con net.Conn) {
	number, nameList := 1, "当前在线用户：\n\n"
	for _, userName := range this.users {
		n := constructUserName(userName, constant.UserNameListLength, false)
		if number%3 == 0 {
			nameList = nameList + n + "  " + "\n"
		} else {
			nameList = nameList + n + "  "
		}

		number++
	}
	nameList = nameList + "\n"
	fmt.Println(nameList)
	con.Write([]byte(nameList))
}

// 创建一对一对话
func (this *Server) NewTalk(res string, con net.Conn) {
	args := strings.Split(res, " ")
	//command := args[0]
	reQuestUserName := args[1]
	remoteUserName := args[2]
	for _, connect := range this.connects {
		if connect.name == remoteUserName {
			connect.AConnect.Write([]byte(constant.ConnectRequest + " " + reQuestUserName + " " + con.RemoteAddr().String()))
			break
		}
	}
}

// 根据地址，找发出连接请求的连接并将他们配对
func (this *Server) MatchConnects(address string, con net.Conn) {
	for _, connect := range this.connects {
		if connect.AConnect.RemoteAddr().String() == address {
			connect.BConnect = con
		}
	}
}

// 获取新加入的用户, 保存用户名，保存创建成功的连接
// 处理私聊的请求和回应， 创建私聊连接对
func (this *Server) InitConnect(con net.Conn) (connect *Connect, err error) {
	var (
		created = false
		length  int
		data    = make([]byte, 1024)
	)
	fmt.Println("新连接: ", con.RemoteAddr())
	for {
		length, err = con.Read(data)
		if err != nil {
			connect = NewConnect("", "", con, nil)
			this.connects = append(this.connects, connect)
			return
		}

		msg := string(data[:length])
		msgs := strings.Split(msg, " ")
		request := msgs[0]

		switch request {
		case constant.ConnectRequest:
			return this.AddReqConnect(con, msg), err
		case constant.ConnectRespone:
			return this.AddResConnect(con, msg), err
		case constant.NewUserRequest:
			connect, created = this.CreateConnect(con, msg)
			if created {
				return
			}
		default:
			fmt.Println("无法识别的请求： " + request)
		}
	}
	return
}

// 处理私聊请求
func (this *Server) AddReqConnect(con net.Conn, msg string) (connect *Connect) {
	msgs := strings.Split(msg, " ")
	request, lname, rname := msgs[0], msgs[1], msgs[2]
	fmt.Println(request, lname, rname)

	connect = NewConnect(lname, "", con, nil)
	this.connects = append(this.connects, connect)

	this.NewTalk(msg, con)
	return
}

// 处理私聊请求的回应
func (this *Server) AddResConnect(con net.Conn, msg string) (connect *Connect) {
	msgs := strings.Split(msg, " ")
	request, lname, rname, address := msgs[0], msgs[1], msgs[2], msgs[3]
	fmt.Println(request, lname, rname, address)

	for _, c := range this.connects {
		if c.AConnect.RemoteAddr().String() == address {
			connect = NewConnect(lname, "", con, c.AConnect)
			this.connects = append(this.connects, connect)

			c.BConnect = con
			this.MatchConnects(address, con)
			return
		}
	}
	return
}

// 核对用户名，创建新用户和连接
func (this *Server) CreateConnect(con net.Conn, msg string) (connect *Connect, created bool) {
	msgs := strings.Split(msg, " ")
	request, name := msgs[0], msgs[1]
	fmt.Println(request, name)

	if !hasValue(this.users, name) && name != constant.SystemName {
		fmt.Println("用户名创建成功：", name)
		con.Write([]byte(constant.CreateUserSuccess))
		this.users = append(this.users, name)
		connect = NewConnect(name, "lobby", con, nil)
		this.connects = append(this.connects, connect)

		// 广播新用户加入的消息
		sysInfo := constructUserName(constant.SystemName, constant.SpreadUserNameLength, true)
		comeInStr := sysInfo + "： " + connect.name + "   进入了房间"
		fmt.Println(comeInStr)
		this.Broadcast(comeInStr)
		created = true
	} else {
		fmt.Println(constant.UserExistingMsg)
		con.Write([]byte(constant.UserExistingMsg))
		fmt.Printf("该用户已经存在 %v\n", con.RemoteAddr())
		created = false
	}
	return
}

// 广播的形式通知所有用户
func (this *Server) Broadcast(msg string) {
	for _, connect := range this.connects {
		if connect.room == "lobby" {
			connect.AConnect.Write([]byte(msg))
		}
	}
}

// 分发消息
// 如果B端连接存在，则说明是私聊，否则广播消息
func (this *Server) DispatchMsg(con *Connect, msg string) {
	if con.BConnect != nil {
		con.BConnect.Write([]byte(msg))
	} else {
		this.Broadcast(msg)
	}
}

// 清除关掉的用户并通知其他用户
func (this *Server) Disconnect(connect *Connect) {
	fmt.Printf("关闭客户端 %s\n", connect.name)
	connect.AConnect.Close()

	for i, c := range this.connects {
		if c == connect {
			this.connects = append(this.connects[:i], this.connects[i+1:]...)
			break
		}
	}

	exist := false
	for _, c := range this.connects {
		if c.name == connect.name {
			exist = true
			break
		}
	}

	if !exist {
		for i, u := range this.users {
			if u == connect.name {
				this.users = append(this.users[:i], this.users[i+1:]...)
				break
			}
		}
	}

	if connect.room == "lobby" {
		delete(this.lobby, connect.name)
	}

	sysInfo := constructUserName(constant.SystemName, constant.SpreadUserNameLength, true)
	if connect.BConnect != nil {
		msg := sysInfo + "： " + connect.name + "   终止了对话"
		fmt.Println(msg)
		connect.BConnect.Write([]byte(msg))
	} else {
		msg := sysInfo + "： " + connect.name + "   离开了房间"
		fmt.Println(msg)
		this.Broadcast(msg)
	}
}
