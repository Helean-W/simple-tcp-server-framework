package siface

//定义服务器接口
type IServer interface {
	Start()
	Stop()
	Serve()
	//路由功能：给当前的服务注册一个路由方法，供客户端的连接处理使用
	AddRouter(MsgID uint32, router IRouter)
	//获取当前server的连接管理器
	GetConnMgr() IConnManager
	//注册OnConnStart
	SetOnConnStart(func(IConnection))
	//注册OnConnStop
	SetOnConnStop(func(IConnection))
	//调用OnConnStart
	CallOnConnStart(IConnection)
	//调用OnConnStop
	CallOnConnStop(IConnection)
}
