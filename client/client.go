package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/wangxianzhuo/iec104"

	"github.com/sirupsen/logrus"
)

type Client struct {
	conn     net.Conn
	DataChan chan []byte
	ctrChan  chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
	Log      *logrus.Entry
}

func NewClient(address string, dataChan chan []byte, logger *logrus.Entry) (Client, context.CancelFunc, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return Client{}, nil, fmt.Errorf("创建TCP连接异常: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return Client{
		conn:     conn,
		DataChan: dataChan,
		ctrChan:  make(chan []byte),
		ctx:      ctx,
		cancel:   cancel,
		Log:      logger,
	}, cancel, nil
}

func (c Client) Close() {
	c.cancel()
	c.conn.Close()
	c.Log.Info("IEC104客户端停止")
}

func (c Client) Start(testInterval time.Duration) {
	c.Log.Info("IEC104客户端通讯启动")
	go c.control()
	c.init()
	go c.connectionTest(testInterval)
	c.totalCall()
	c.receive()
}

func (c Client) connectionTest(testInterval time.Duration) {
	c.Log.Infof("IEC104连接测试启动，每%v执行一次连接测试", testInterval*time.Second)
	ticker := time.NewTicker(testInterval)
	for {
		select {
		case <-ticker.C:
			err := c.test()
			if err != nil {
				c.Log.Errorf("连接测试失败: %v", err)
			}
			err = c.reconnect()
			if err != nil {
				c.Log.Errorf("重连失败: %v", err)
				c.Log.Errorf("结束通讯")
				c.cancel()
				return
			}
		case <-c.ctx.Done():
			c.Log.Info("IEC104连接测试停止")
		}
	}

}
func (c Client) receive() {
	c.Log.Info("数据接收线程启动")
	for {
		select {
		case data := <-c.DataChan:
			c.Log.Debugf("收到原始数据: [%X]", data)
		case <-c.ctx.Done():
			c.Log.Info("数据接收线程停止")
		}
	}
}

func (c Client) control() {
	c.Log.Info("控制线程启动")
	for {
		select {
		case ctr := <-c.ctrChan:
			c.Log.Debugf("收到控制帧数据: [%X]", ctr)
		case <-c.ctx.Done():
			c.Log.Info("控制线程停止")
		}
	}
}

func (c Client) init() {
	c.stop()
	c.start()
}

func (c Client) stop() {

}

func (c Client) start() error {
	uFrame := iec104.UFrame{
		STARTDT_ACT: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		return fmt.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		return fmt.Errorf("启动帧创建异常: %v", err)
	}

	_, err = c.conn.Write(apdu.ConvertBytes())
	if err != nil {
		return fmt.Errorf("启动帧发送异常: %v", err)
	}

	return nil
}

func (c Client) test() error {
	uFrame := iec104.UFrame{
		TESTFR_ACT: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		return fmt.Errorf("测试帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		return fmt.Errorf("测试帧创建异常: %v", err)
	}

	_, err = c.conn.Write(apdu.ConvertBytes())
	if err != nil {
		return fmt.Errorf("测试帧发送异常: %v", err)
	}

	return nil
}

func (c Client) totalCall() {

}

func (c Client) reconnect() error {
	c.stop()
	c.start()
	c.totalCall()
	return nil
}
