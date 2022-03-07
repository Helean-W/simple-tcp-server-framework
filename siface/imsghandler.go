package siface

/* 消息管理抽象层 */
type IMsgHandle interface {
	//调度/执行对应的router消息处理方法
	DoMsgHandler(IRequest)
	//为消息添加具体的处理逻辑
	AddRouter(msgIID uint32, router IRouter)
	//启动Worker工作池
	StartWorkerPool()
	//关闭Worker工作池
	StopWorkerPool()
	//将消息交给消息任务队列处理
	SendMsgToTaskQueue(IRequest)
}
