package snet

import (
	"fmt"
	"net"

	"github.com/STS/siface"
	"github.com/STS/utils"
)

//IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//名称
	Name string
	//IP版本
	IPVersion string
	//监听IP
	IP string
	//监听端口
	Port int
	//当前的Server的消息管理模块，用来绑定MsgID和对应的API业务
	MsgHandle siface.IMsgHandle
	//当前server的连接管理器
	ConnMgr siface.IConnManager
	//当前server创建连接之后自动调用的hook函数
	OnConnStart func(siface.IConnection)
	//当前server销毁连接之前自动调用的hook函数
	OnConnStop func(siface.IConnection)
}

func (s *Server) Start() {
	fmt.Printf("[STS] Server Name: %s, listening at ip: %s, Port: %d is starting\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[STS]Version is : %s, MaxConn: %d, MaxPackSize is %d\n", utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)
	fmt.Printf("[Start] Server Listener at IP: %s, Port: %d, is starting\n", s.IP, s.Port)

	go func() {
		//0 开启消息队列及worker工作池(用不用go都行，里面开几个go并不占用时间)
		s.MsgHandle.StartWorkerPool()

		//1.获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve TCP ADDR err:", err)
			return
		}

		//2.监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("ListenTCP err:", err)
			return
		}
		fmt.Println("start server ", s.Name, " success, Listening...")

		var cid uint32 = 0

		//3.阻塞的等待客户端连接，处理客户端连接业务（读写）
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("AcceptTCP err:", err)
				continue
			}

			//设置最大连接个数的判断  如果超过最大连接的数量那么关闭此新的连接
			if s.ConnMgr.Count() >= utils.GlobalObject.MaxConn {
				//TODO 给客户端响应一个超出最大连接的错误包
				fmt.Println("==========>too many connections<============")
				conn.Close()
				continue
			}

			//将处理新连接的业务方法和conn进行绑定 得到我们的连接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandle)
			cid++

			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	//将一些服务器的资源，状态或者一些已经开辟的连接信息进行停止或者回收
	//清空连接管理器中的连接资源
	fmt.Println("[STOP] STS server name ", s.Name)
	s.MsgHandle.StopWorkerPool()
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()
	//TODO 做一些启动服务器之后的额外业务

	//阻塞状态，不阻塞的话start中的gorountine就会在函数返回时被关闭，无法一直循环等待连接
	select {}
}

//获取连接管理模块（供连接模块调用，向连接管理模块中添加记录）
func (s *Server) GetConnMgr() siface.IConnManager {
	return s.ConnMgr
}

//路由功能：给当前的服务注册一个路由方法，供客户端的连接处理使用
func (s *Server) AddRouter(MsgID uint32, router siface.IRouter) {
	s.MsgHandle.AddRouter(MsgID, router)
	fmt.Println("Add Router Success!!!")
}

//初始化Server模块
func NewServer(name string) siface.IServer {
	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}
	return s
}

//注册OnConnStart
func (s *Server) SetOnConnStart(hookFunc func(siface.IConnection)) {
	s.OnConnStart = hookFunc
}

//注册OnConnStop
func (s *Server) SetOnConnStop(hookFunc func(siface.IConnection)) {
	s.OnConnStop = hookFunc
}

//调用OnConnStart
func (s *Server) CallOnConnStart(conn siface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---->Call OnConnStart()<----")
		s.OnConnStart(conn)
	}
}

//调用OnConnStop
func (s *Server) CallOnConnStop(conn siface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---->Call OnConnStop()<----")
		s.OnConnStop(conn)
	}
}
