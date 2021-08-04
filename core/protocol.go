package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	// 协议-结果
	protocolResultSuccess           = 0 // 成功，默认值
	protocolResultFail              = 1 // 失败
	protocolResultHeartBeat         = 2 // 心跳
	protocolResultFailToReceive     = 3 // 接收失败
	protocolResultFailToAuth        = 4 // 鉴权失败
	protocolResultVersionMismatch   = 5 // 版本不匹配
	protocolResultIllegalAccessPort = 6 // 访问端口不合法
	protocolResultPortIsOccupied    = 7 // 访问端口被占用

	// 协议发送超时时间
	protocolSendTimeout = 5 * time.Second
	// 协议接收超时时间
	protocolReceiveTimeout = heartBeatIntervalTime + 5*time.Second

	// 版本序列(单调递增)
	// 从右往左
	// 第1位为小版本号，用于修复BUG
	// 第2位为次版本号，用于增删功能
	// 第3位为主版本号，用于结构等大的升级
	Version = 140
)

// 协议格式
// 结果|版本号|访问端口|machineid|Key
// 1|2|13306|uuid|winshu

// 协议
type Protocol struct {
	Result  byte   // 结果：0 失败，1 成功
	Version uint32 // 版本号，单调递增
	Port    uint32 // 访问端口
	ID      string // 机器码
	Key     string // 身份验证
}

// 转字符串
func (p *Protocol) String() string {
	return fmt.Sprintf("%d|%d|%d|%s|%s", p.Result, p.Version, p.Port, p.ID, p.Key)
}

// 返回一个新结果
func (p *Protocol) NewResult(newResult byte) Protocol {
	return Protocol{
		Result:  newResult,
		Version: p.Version,
		Port:    p.Port,
		ID:      p.ID,
		Key:     p.Key,
	}
}

// 序列化
func (p *Protocol) Bytes() []byte {
	buffer := bytes.NewBuffer([]byte{})

	buffer.WriteByte(p.Result)
	_ = binary.Write(buffer, binary.BigEndian, p.Version)
	_ = binary.Write(buffer, binary.BigEndian, p.Port)
	buffer.WriteString(p.ID)
	buffer.WriteString(p.Key)
	return buffer.Bytes()
}

// 协议长度
func (p *Protocol) Len() byte {
	return byte(len(p.Bytes()))
}

// 是否成功
func (p *Protocol) Success() bool {
	return p.Result == protocolResultSuccess
}

// 检查是否是同一客户端
func (p *Protocol) IsSameID(other *Protocol) bool {
	return p.ID == other.ID
}

// 解析协议
func parseProtocol(body []byte) Protocol {
	// 检查 body 长度，是否合法
	if len(body) < 42 {
		return Protocol{Result: protocolResultFail}
	}
	return Protocol{
		Result:  body[0],
		Version: binary.BigEndian.Uint32(body[1:5]),
		Port:    binary.BigEndian.Uint32(body[5:9]),
		ID:      string(body[9:41]),
		Key:     string(body[41:]),
	}
}

// 发送协议
// 第一个字节为协议长度
// 协议长度只支持到255
func sendProtocol(conn net.Conn, req Protocol) bool {
	buffer := bytes.NewBuffer([]byte{})
	buffer.WriteByte(req.Len())
	buffer.Write(req.Bytes())

	// 设置写超时时间，避免连接断开的问题
	if err := conn.SetWriteDeadline(time.Now().Add(protocolSendTimeout)); err != nil {
		log.Println("Fail to set write deadline.", err.Error())
		return false
	}
	// 写协议内容
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		log.Printf("Send protocol failed. [%s] %s\n", req.String(), err.Error())
		return false
	}
	// 清空写超时设置
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		log.Println("Fail to clear write deadline.", err.Error())
		return false
	}
	//log.Println("Send protocol", req.String())
	return true
}

// 接收协议
// 第一个字节为协议长度
func receiveProtocol(conn net.Conn) Protocol {
	var err error
	var length byte

	// 设置读超时时间略大于心跳时间，避免连接断开的问题
	if err := conn.SetReadDeadline(time.Now().Add(protocolReceiveTimeout)); err != nil {
		log.Println("Fail to set read deadline.", err.Error())
		return Protocol{Result: protocolResultFailToReceive}
	}
	// 读取协议长度
	if err = binary.Read(conn, binary.BigEndian, &length); err != nil {
		return Protocol{Result: protocolResultFailToReceive}
	}
	// 读取协议内容
	body := make([]byte, length)
	if err = binary.Read(conn, binary.BigEndian, &body); err != nil {
		return Protocol{Result: protocolResultFailToReceive}
	}
	// 清空读超时设置
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		log.Println("Fail to clear read deadline.", err.Error())
		return Protocol{Result: protocolResultFailToReceive}
	}
	return parseProtocol(body)
}
