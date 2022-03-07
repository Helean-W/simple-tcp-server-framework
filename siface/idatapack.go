//为解决TCP粘包问题
package siface

//封包、拆包 模块，直接面向TCP连接中的数据流，用于处理TCP粘包问题
type IDataPack interface {
	//获取包头长度
	GetHeadLen() uint32
	//封包
	Pack(msg IMessage) ([]byte, error)
	//拆包
	Unpack([]byte) IMessage
}
