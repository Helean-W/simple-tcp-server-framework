package snet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/STS/siface"
	"github.com/STS/utils"
)

//连接模块
type Connection struct {
	//当前Conn隶属于哪个Server
	TcpServer siface.IServer

	//当前连接的socket TCP套接字
	Conn *net.TCPConn

	//连接的ID
	ConnID uint32

	//当前连接状态
	isClosed bool

	//告知当前连接已经停止的channel(由Reader告知Writer退出的信号)
	ExitChan chan bool

	//无缓冲的管道，用于读、写gorountine之间的消息通信
	msgChan chan []byte

	//当前的Connection的消息管理模块，用来绑定MsgID和对应的API业务
	MsgHandle siface.IMsgHandle

	//连接属性的集合
	property map[string]interface{}

	//保护连接属性的锁
	propertyLock sync.RWMutex
}

//初始化连接模块的方法
func NewConnection(server siface.IServer, conn *net.TCPConn, connID uint32, msgHandle siface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer: server,
		Conn:      conn,
		ConnID:    connID,
		isClosed:  false,
		ExitChan:  make(chan bool, 1),
		msgChan:   make(chan []byte), //无缓冲，必须同时进行读写，一方会被阻塞等待另一方
		MsgHandle: msgHandle,
		property:  make(map[string]interface{}),
	}

	//将conn加入到ConnManager中
	server.GetConnMgr().Add(c)

	return c
}

//读消息的gorountine，专门从客户端读消息
func (c *Connection) StartReader() {
	fmt.Println("[reader gorountine is runing....]")
	defer fmt.Println(" [reader is exit]", " connID=", c.ConnID, " remote addr is ", c.RemoteAddr())
	defer c.Stop()

	for {
		//创建拆包对象
		dp := NewDataPack()

		//读取客户端的msg head 8bytes 二进制流
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("server read msg head error:", err)
			break
		}
		//拆包，得到id和len放在一个msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack msg head error:", err)
			break
		}
		//根据datalen 再次读取data，放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error:", err)
				break
			}
			msg.SetData(data)
		}
		//建立当前连接的Request
		req := Request{
			conn: c,
			msg:  msg,
		}

		//判断是否开启工作池
		if utils.GlobalObject.WorkPoolSize > 0 {
			//发送给消息队列
			c.MsgHandle.SendMsgToTaskQueue(&req)
		} else {
			//直接开个go处理
			go c.MsgHandle.DoMsgHandler(&req)
		}

	}
}

//写消息的gorountine，专门发送给客户端消息
func (c *Connection) StartWriter() {
	fmt.Println("[Writter gorountine is runing....]")
	defer fmt.Println(" [writer is exit]", " connID=", c.ConnID, "remote addr is ", c.RemoteAddr())

	//不断阻塞等到channel的消息，有消息就会写客户端
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error :", err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，此时writer也要退出
			return
		}
	}
}

//启动连接 让当前连接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start()...ConnID=", c.ConnID)

	//启动从当前连接的读数据的业务
	go c.StartReader()
	//启动从当前连接写数据的业务
	go c.StartWriter()

	//按照开发者传递进来的创建连接之后需要调用的处理业务，执行对应Hook函数
	c.TcpServer.CallOnConnStart(c)
}

//停止连接 结束当前连接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...ConnID=", c.ConnID)

	if c.isClosed {
		return
	}

	c.isClosed = true

	//调用开发者注册的销毁连接之前需要执行的业务Hook
	c.TcpServer.CallOnConnStop(c)

	//关闭socket连接
	c.Conn.Close()

	//将当前连接从ConnMgr中摘除
	c.TcpServer.GetConnMgr().Remove(c)

	c.ExitChan <- true //非必要？
	close(c.ExitChan)  //Reader关闭channel后，本来Writer在读取空的channel被阻塞，关闭后则不会阻塞，会读到空值，所以Writer的select的此case得以顺利执行，Writer关闭

	close(c.msgChan)
}

//获取当前连接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//获取当前连接模块的连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//获取远程客户端的TCP状态 IP Port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, msgData []byte) error {
	if c.isClosed {
		return errors.New("connecting closed when send message")
	}
	//将data进行封包
	dp := NewDataPack()

	TlvMsg, err := dp.Pack(NewMessage(msgId, msgData))
	if err != nil {
		fmt.Println("pack error, message id = ", msgId)
		return errors.New("pack error when send message")
	}
	//将数据发送给客户端
	c.msgChan <- TlvMsg
	return nil
}

//设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	//添加一个连接属性
	c.property[key] = value
}

//获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	//删除一个连接属性
	delete(c.property, key)
}
