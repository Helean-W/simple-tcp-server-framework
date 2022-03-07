package snet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

//负责测试datapack拆包封包  单元测试
func TestDataPack(t *testing.T) {
	/* 模拟的服务器 */
	//1创建socketTCP
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err:", err)
		return
	}

	go func() {
		//2从客户端读取数据，拆包处理
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept error", err)
			}
			go func(conn net.Conn) {
				dp := NewDataPack()
				for {
					headData := make([]byte, dp.GetHeadLen())
					if _, err := io.ReadFull(conn, headData); err != nil {
						fmt.Println("read head error", err)
						break
					}
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("unpack head error", err)
						return
					}
					if msgHead.GetMsgLen() > 0 {
						//msg有数据，需要进行第二次读取
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())

						//根据datalen的长度再次在io流中读取
						if _, err := io.ReadFull(conn, msg.Data); err != nil {
							fmt.Println("read data error", err)
							return
						}

						//完整的消息已经读完
						fmt.Println("--->Recv MsgID:", msg.ID, "datalen: ", msg.DataLen, "data = ", string(msg.Data))
					}

				}
			}(conn)
		}
	}()

	/* 模拟的客户端 */
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client dial err: ", err)
		return
	}

	dp := NewDataPack()

	//模拟粘包过程，封装两个msg一同发送
	//封装第一个
	msg1 := &Message{
		ID:      1,
		DataLen: 3,
		Data:    []byte{'S', 'T', 'S'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err: ", err)
		return
	}
	//封装第二个
	msg2 := &Message{
		ID:      2,
		DataLen: 7,
		Data:    []byte{'S', 'T', 'S', 'B', 'A', 'I', 'L'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg1 err: ", err)
		return
	}
	//将两个包粘在一起
	sendData1 = append(sendData1, sendData2...)
	//一次性发送给服务端
	conn.Write(sendData1)
	//客户端阻塞
	select {}
}
