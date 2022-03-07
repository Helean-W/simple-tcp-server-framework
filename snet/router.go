package snet

import "github.com/STS/siface"

//实现router时， 先嵌入这个BaseRouter基类，然后根据需要对这个基类的方法进行重写就好了,
//如果是实现IRouter接口，则需要对三个方法都进行实现
type BaseRouter struct {
}

//后续可以直接继承这个基类，重写某一个方法就可以了，不用全部实现
//在处理conn业务之前的钩子
func (br *BaseRouter) PreHandle(request siface.IRequest) {}

//在处理conn业务的主方法hook
func (br *BaseRouter) Handle(request siface.IRequest) {}

//在处理conn业务之后的钩子
func (br *BaseRouter) PostHandle(request siface.IRequest) {}
