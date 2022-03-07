package snet

import (
	"github.com/STS/siface"
)

type Request struct {
	//已经和客户端建立好的连接
	conn siface.IConnection
	//客户端请求的数据
	msg siface.IMessage
}

func (r *Request) GetConnection() siface.IConnection {
	return r.conn
}

func (r *Request) GetMsgData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgId() uint32 {
	return r.msg.GetMsgId()
}

func (r *Request) GetMsgLen() uint32 {
	return r.msg.GetMsgLen()
}
