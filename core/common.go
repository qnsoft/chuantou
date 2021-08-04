package core

import (
	"chuantou/config"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// 重连间隔时间
	retryIntervalTime = 5

	// 心跳间隔时间
	heartBeatIntervalTime = 60 * time.Second

	// 最大重试次数
	maxRetryTimes = 24 * 60 * 60 / retryIntervalTime
)

var bufferPool *sync.Pool

func init() {
	log.Println("Init copy buffer pool...")
	bufferPool = &sync.Pool{}
	bufferPool.New = func() interface{} {
		return make([]byte, 32*1024)
	}
}

// 使用 bufferPool 重写 copy 函数， 避免反复 gc，提升性能
func copyWithPool(dst io.Writer, src io.Reader) (written int64, err error) {
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

// 连接数据复制
func connCopy(dist, source net.Conn, wg *sync.WaitGroup) {
	if _, err := copyWithPool(dist, source); err != nil {
		//log.Println("Connection interrupted", err)
	}
	_ = dist.Close()
	wg.Done()
}

// 连接转发
func forward(conn1, conn2 net.Conn) {
	//log.Printf("Forward channel [%s/%s] <-> [%s/%s]\n",
	//	conn1.RemoteAddr(), conn1.LocalAddr(), conn2.RemoteAddr(), conn2.LocalAddr())

	var wg sync.WaitGroup
	// wait tow goroutines
	wg.Add(2)
	go connCopy(conn1, conn2, &wg)
	go connCopy(conn2, conn1, &wg)
	//blocking when the wg is locked
	wg.Wait()
}

// 关闭连接
func closeConn(connections ...net.Conn) {
	for _, conn := range connections {
		if conn != nil {
			_ = conn.Close()
		}
	}
}

// 拨号
func dial(targetAddr config.NetAddress /*目标地址*/, maxRedialTimes int /*最大重拨次数*/) net.Conn {
	redialTimes := 0
	for {
		conn, err := net.Dial("tcp", targetAddr.String())
		if err == nil {
			//log.Printf("Dial to [%s] success.\n", targetAddr)
			return conn
		}
		redialTimes++
		if maxRedialTimes < 0 || redialTimes < maxRedialTimes {
			// 重连模式，每5秒一次
			log.Printf("Dial to [%s] failed, redial(%d) after %d seconeds.", targetAddr.String(), redialTimes, retryIntervalTime)
			time.Sleep(retryIntervalTime * time.Second)
		} else {
			log.Printf("Dial to [%s] failed. %s\n", targetAddr.String(), err.Error())
			return nil
		}
	}
}

// 监听端口
func listen(port uint32, id string) net.Listener {
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Println("Listen failed, the port may be used or closed", port)
		return nil
	}
	log.Printf("Listening at address %s by %s\n", address, id)
	return listener
}

// 受理请求
func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Accept connect failed ->", err.Error())
		return nil
	}
	//log.Println("Accept a new client ->", conn.RemoteAddr())
	return conn
}

// 定时器，每间隔一段时间执行一遍函数
func setInterval(callback func(), duration time.Duration) {
	c := time.Tick(duration)
	for {
		callback()
		<-c
	}
}
