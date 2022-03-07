package utils

import (
	"encoding/json"
	"io/ioutil"

	"github.com/STS/siface"
)

//存储一切有关服务器框架的全局参数，供其他模块使用
//一些参数可以通过json由用户进行配置

type GlobalObj struct {
	//server
	TcpServer siface.IServer //当前全局的server对象
	Host      string         //当前服务器主机监听的IP
	TcpPort   int            //当前服务器主机监听的端口号
	Name      string         //当前服务器名称
	//STS
	Version        string //当前STS的版本号
	MaxConn        int    //当前服务器主机允许的最大连接数
	MaxPackageSize uint32 //当前框架数据包的最大值
	WorkPoolSize   uint32 //当前业务工作Worker池的Goroutine数量
	MaxWorkTaskLen uint32 //每个Worker消息队列的最大长度
}

//定义一个全局的对外GlobalObj
var GlobalObject *GlobalObj

//提供一个init方法，初始化当前的GlobalObject
//init方法导入此包的时候会被调用
func init() {
	//如果配置文件没有加载，默认的值
	GlobalObject = &GlobalObj{
		Name:           "ServerApp",
		Version:        "V1.0",
		TcpPort:        8999,
		Host:           "0.0.0.0",
		MaxConn:        1000,
		MaxPackageSize: 4096,
		WorkPoolSize:   10,
		MaxWorkTaskLen: 1024,
	}

	//尝试从sts.json去加载一些用户自定义数据
	GlobalObject.Reload()
}

//从sts.json去加载用于自定义的参数
func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile("conf/sts.json")
	if err != nil {
		panic(err)
	}
	//将json文件数据解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}
