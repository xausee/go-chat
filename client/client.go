package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"go-chat/constant"
)

// 客户端结构体
type Client struct {
	conn      net.Conn // 连接
	name      string   // 用户名
	network   string   // 网络协议
	address   string   // 地址
	Timestamp int      // 时间戳
}

func NewClient(network, address string) *Client {
	return &Client{nil, "", network, address, int(time.Now().Unix())}
}

func (this *Client) Connect() error {
	con, err := net.Dial(this.network, this.address)
	if err != nil {
		fmt.Printf("连接服务器失败: %s, 错误信息: %s\n", this.address, err)
		os.Exit(-1)
	}
	fmt.Println("连接服务器成功")
	this.conn = con
	return err
}

func (this *Client) Disconnect() {
	this.conn.Close()
}

// 解析用户输入信息，判断是否有命令需要执行
// 若是命令则执行对应操作，返回真值
// 若是一般对话信息，则返回假值
func (this *Client) Command(input string) bool {
	isCmd := false
	ins := strings.Split(input, " ")

	cmd := ins[0]
	switch cmd {
	case "help", "--help", "-h":
		this.Help()
		isCmd = true
	case "quit":
		fmt.Println("交流结束.")
		os.Exit(1)
	case "connect":
		if len(ins) != 2 {
			fmt.Printf("输入命令部不对， 参数例子：connect user\n\n")
		} else {
			rname := ins[1]
			this.TalkTo(rname)
		}
		isCmd = true
	default:
		isCmd = false
	}
	return isCmd
}

func (this *Client) Help() {
	fmt.Printf("\nhelp/--help/-h        查看可用命令及使用方法")
	fmt.Printf("\nlist                  查看当前在线用户")
	fmt.Printf("\nconnect               发起对指定用户的聊天。例如：connect user")
	fmt.Printf("\nquit                  结束聊天\n\n")
}

func (this *Client) TalkTo(userName string) {
	cmd := exec.Command("cmd", "/c", "start", os.Args[0], "-cmd=connect", "-lname="+this.name, "-rname="+userName)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
}

// 解析传入参数，确定是否是未注册用户名的连接
// 返回假值代表该连接未被注册用户名，需要在下一步创建用户名
// 返回真值代表该连接是由已存在的用户创建，用于其它操作，不再需要创建用户名
func (this *Client) CheckArgs(args map[string]string) bool {
	if args["cmd"] == "connect" && args["lname"] != "" && args["rname"] != "" {
		fmt.Println("已发起对 " + args["rname"] + " 的聊天请求，耐心等待回应...")
		this.name = args["lname"]
		this.conn.Write([]byte(constant.ConnectRequest + " " + args["lname"] + " " + args["rname"]))
		return true
	}

	if args["addr"] != "" && args["lname"] != "" && args["rname"] != "" {
		fmt.Println(args["rname"] + " 发起对你的聊天，请回应")
		this.name = args["lname"]
		this.conn.Write([]byte(constant.ConnectRespone + " " + args["lname"] + " " + args["rname"] + " " + args["addr"]))
		return true
	}

	return false
}

// 为新连接创建一个用户名
func (this *Client) CreateName() string {
	fmt.Println("给自己取一个名字，长度不大于5个中文字符或15个英文字符")
	fmt.Println("输入用户名：")
	name := this.GetInputName()

	for {
		if this.CheckNameWithServer(name) {
			this.name = name
			fmt.Println("您的用户名：", name, "  开始交谈吧!")
			break
		} else {
			fmt.Println(constant.UserExistingMsg)
			fmt.Println("请重新输入：")
			name = this.GetInputName()
		}
	}
	return name
}

// 从标准输入中获取用户名，并检查其长度，长度不符合要求则重新输入
func (this *Client) GetInputName() string {
	r := bufio.NewReader(os.Stdin)
	l, _, _ := r.ReadLine()
	n := string(l)
	for {
		if len(n) > 15 {
			fmt.Println("用户名过长")
			fmt.Println("请重新输入：")
			l, _, _ = r.ReadLine()
			n = string(l)
		} else {
			break
		}
	}
	return n
}

// 检查输入的名字是否已经被使用
// 返回真值代表用户名注册成功
// 返回假值代表用户名已经被使用
func (this *Client) CheckNameWithServer(name string) bool {
	msg := constant.NewUserRequest + " " + name
	in, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("发送消息到服务器失败: %d\n", in)
		os.Exit(0)
	}

	var readStr = make([]byte, 1024)
	length, err := this.conn.Read(readStr)
	if err != nil {
		fmt.Printf("获取服务器消息失败. 错误：%s\n", err)
		os.Exit(0)
	}

	res := string(readStr[:length])
	if res == constant.UserExistingMsg {
		return false
	} else if res == constant.CreateUserSuccess {
		return true
	}
	return false
}

// 解析服务器传回的消息，是否是需要执行的命令
// 是命令，则执行，返回真值
// 一般消息，不做任何操作，返回假值
func (this *Client) IsCmd(message string) bool {
	isCmd := false
	msgs := strings.Split(message, " ")

	if len(msgs) == 3 {
		cmd := msgs[0]
		requestUserName := msgs[1]
		addr := msgs[2]
		switch cmd {
		case constant.ConnectRequest:
			this.CreateNewConnect(requestUserName, addr)
			isCmd = true
		default:
			isCmd = false
		}
	}
	return isCmd
}

func (this *Client) ReadMsg() string {
	var readStr = make([]byte, 1024)

	for {
		length, err := this.conn.Read(readStr)
		if err != nil {
			fmt.Printf("获取服务器消息失败. 错误：%s\n", err)
			os.Exit(0)
		}
		response := string(readStr[:length])
		if this.IsCmd(response) {
			continue
		}
		fmt.Println(response)
	}
}

// 使用windows命令cmd start开启新的聊天窗口
func (this *Client) CreateNewConnect(requestUserName, addr string) {
	cmd := exec.Command("cmd", "/c", "start", os.Args[0], "-lname="+this.name, "-rname="+requestUserName, "-addr="+addr)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
}

func (this *Client) SendMsg() {
	var (
		reader   = bufio.NewReader(os.Stdin)
		writeStr = make([]byte, 1024)
	)

	for {
		writeStr, _, _ = reader.ReadLine()
		if this.Command(string(writeStr)) {
			continue
		}

		in, err := this.conn.Write([]byte(writeStr))
		if err != nil {
			fmt.Printf("发送消息到服务器失败: %d\n", in)
			os.Exit(0)
		}
	}
}
