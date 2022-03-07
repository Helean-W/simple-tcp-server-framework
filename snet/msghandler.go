package snet

import (
	"fmt"
	"strconv"

	"github.com/STS/siface"
	"github.com/STS/utils"
)

/* 消息处理模块的实现 */

type MsgHandle struct {
	//存放每个msgID所对应的处理方法
	Apis map[uint32]siface.IRouter
	//负责Worker取任务的消息队列
	TaskQueue []chan siface.IRequest
	//业务工作Worker池的Worker数量
	WorkPoolSize uint32
}

//提供一个创建MsgHandler的方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:         make(map[uint32]siface.IRouter),
		WorkPoolSize: utils.GlobalObject.WorkPoolSize, //尝试从全局配置中获取
		TaskQueue:    make([]chan siface.IRequest, utils.GlobalObject.WorkPoolSize),
	}
}

//调度/执行对应的router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request siface.IRequest) {
	//1 从request中找到msgID
	handler, ok := mh.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgId(), " is not found, need register")
		return
	}
	//2 根据msgID调度对应router业务
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

//为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router siface.IRouter) {
	//1判断当前msg绑定的API是够已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//id已注册
		panic("repeat api, message id = " + strconv.Itoa(int(msgID)))
	}
	//2添加msg与API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("add API msgID = ", msgID, " success!")
}

//启动工作池(开启工作池的动作只能发生一次，一个框架只能有一个worker工作池)
func (mh *MsgHandle) StartWorkerPool() {
	//根据WorkerPoolSize 分别开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkPoolSize); i++ {
		//一个Worker被启动
		//1 给当前的Worker对应的channel消息队列 开辟空间 第0个Worker就用第0个channel。。。。
		mh.TaskQueue[i] = make(chan siface.IRequest, utils.GlobalObject.MaxWorkTaskLen)
		//2 启动当前的Worker， 阻塞等待消息从channel传递过来
		go mh.startOneWorker(i, mh.TaskQueue[i])
	}
}

func (mh *MsgHandle) StopWorkerPool() {
	for _, c := range mh.TaskQueue {
		close(c)
	}
}

//启动一个工作流程
func (mh *MsgHandle) startOneWorker(workerID int, taskChan chan siface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is starter...")

	//不断阻塞等待对应队列的消息
	for {
		select {
		//如果有消息过来，出列的就是一个客户端的Request，执行当前Requset所绑定的业务
		case request := <-taskChan:
			mh.DoMsgHandler(request)
		}
	}
}

//将消息交给TaskQueue，由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request siface.IRequest) {
	//1 将消息平均分配给不同的Worker
	//request目前没有ID，根据客户端建立的ConnID进行分配
	workerID := request.GetConnection().GetConnID() % mh.WorkPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(), " request MsgID = ", request.GetMsgId(), " to WorkerID = ", workerID)

	//2 将消息发送给Worker的TaskQueue
	mh.TaskQueue[workerID] <- request
}
