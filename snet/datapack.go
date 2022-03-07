package snet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/STS/siface"
	"github.com/STS/utils"
)

//封包拆包的具体模块
type DataPack struct{}

//拆包封包实例的初始化
func NewDataPack() *DataPack {
	return &DataPack{}
}

func (d *DataPack) GetHeadLen() uint32 {
	//DataLen uint32----4字节  DataId  uint32----4字节
	return 4 + 4
}

//封包  Len(4字节)|Id(4字节)|Data
func (d *DataPack) Pack(msg siface.IMessage) ([]byte, error) {
	//创建一个存放byte字节的缓冲
	dataBuff := bytes.NewBuffer([]byte(nil))

	//将len,id,data写入buffer中
	err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen())
	if err != nil {
		return nil, err
	}
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId())
	if err != nil {
		return nil, err
	}
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetData())
	if err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

//拆包(将包的Head信息读出来  之后再根据head信息里的data长度再进行一次读，读data在套接字里读就行)
func (d *DataPack) Unpack(binaryData []byte) (siface.IMessage, error) {
	//创建一个从输入二进制数据读的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只提取head信息，得到ID和Len

	msg := &Message{}

	//读dataLen  read方法根据第三个参数的类型决定读取多少字节放到第三个参数上
	err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen) //这里要加&，因为虽然msg是地址，但是msg.DataLen是结构体中的一个变量，而不是变量的地址
	if err != nil {
		return nil, err
	}
	err = binary.Read(dataBuff, binary.LittleEndian, &msg.ID)
	if err != nil {
		return nil, err
	}

	//判断datalen是否超出了MaxPackageLen
	if utils.GlobalObject.MaxPackageSize > 0 && msg.DataLen > utils.GlobalObject.MaxPackageSize {
		return nil, errors.New("too large msg dara recieve")
	}
	return msg, nil
}
