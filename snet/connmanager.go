package snet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/STS/siface"
)

/* 连接管理模块 */

type ConnManager struct {
	connections map[uint32]siface.IConnection //管理的连接信息集合
	connLock    sync.RWMutex                  //保护连接集合的读写锁
}

//创建当前连接管理的方法
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]siface.IConnection),
	}
}

//添加连接
func (connMgr *ConnManager) Add(conn siface.IConnection) {
	//保护共享资源，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn加入到ConnManager中
	connMgr.connections[conn.GetConnID()] = conn
	fmt.Printf("connection ID = %d add to ConnManager successfully: conn len = %d\n", conn.GetConnID(), connMgr.Count())
}

//删除连接
func (connMgr *ConnManager) Remove(conn siface.IConnection) {
	//保护共享资源，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())
	fmt.Printf("connection ID = %d remove successfully: conn len = %d\n", conn.GetConnID(), connMgr.Count())
}

//根据connID获取连接
func (connMgr *ConnManager) Get(connID uint32) (siface.IConnection, error) {
	//保护共享资源，加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not FOUND")
	}
}

//得到当前连接总数
func (connMgr *ConnManager) Count() int {
	return len(connMgr.connections)
}

//清除并终止所有的连接
func (connMgr *ConnManager) ClearConn() {
	// //保护共享资源，加写锁
	// connMgr.connLock.Lock()
	// defer connMgr.connLock.Unlock()

	//删除conn并停止conn的工作
	for _, conn := range connMgr.connections {
		//停止
		conn.Stop()
		// //删除
		// connMgr.Remove(conn)
	}

	fmt.Println("Clear All connections succ!, conn num = ", connMgr.Count())
}
