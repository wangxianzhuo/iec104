package client

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/wangxianzhuo/iec104/msg-elements"

	"github.com/wangxianzhuo/iec104"

	"github.com/sirupsen/logrus"
)

var (
	connectDeadline time.Duration
)

// SetConnectDeadLine 修改默认连接超时时间
func SetConnectDeadLine(d time.Duration) {
	connectDeadline = d
}

// Client IEC104客户端
type Client struct {
	conn     net.Conn
	dataChan chan iec104.APDU
	ctrChan  chan iec104.APDU
	outChan  chan map[string]float32
	ctx      context.Context
	cancel   context.CancelFunc
	Log      *logrus.Entry
	mux      *sync.Mutex
}

// New ...
func New(address string, outChan chan map[string]float32, logger *logrus.Entry) (Client, context.CancelFunc, error) {
	if logger == nil {
		panic("logrus.Entry is nil")
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return Client{}, nil, fmt.Errorf("创建TCP连接异常: %v", err)
	}

	connectDeadline = 5 * time.Minute

	ctx, cancel := context.WithCancel(context.Background())
	return Client{
		conn:     conn,
		dataChan: make(chan iec104.APDU),
		ctrChan:  make(chan iec104.APDU),
		outChan:  outChan,
		ctx:      ctx,
		cancel:   cancel,
		Log:      logger,
		mux:      new(sync.Mutex),
	}, cancel, nil
}

// Close ...
func (c Client) Close() {
	c.cancel()
	c.conn.Close()
	c.Log.Info("IEC104客户端停止")
}

// Start 启动
func (c Client) Start(testInterval time.Duration) {
	c.Log.Info("IEC104客户端通讯启动")
	go c.read()
	c.init()
	go c.connectionTest(testInterval)
	go c.uFrameResp()
	err := c.totalCall()
	if err != nil {
		c.Log.Panic(err)
	}
	c.receive()
}

func (c Client) connectionTest(testInterval time.Duration) {
	defer c.cancel()
	c.Log.Infof("IEC104连接测试启动，每%v执行一次连接测试", testInterval)
	ticker := time.NewTicker(testInterval)
LOOP:
	for {
		select {
		case <-ticker.C:
			err := c.test()
			if err != nil {
				c.Log.Errorf("连接测试失败: %v", err)
				c.Log.Errorf("尝试重新连接")
				for i := 0; i < 5; i++ {
					err := c.reconnect()
					if err != nil {
						c.Log.Errorf("第%d次重连失败: %v", i, err)
					} else {
						continue LOOP
					}
				}
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
	defer c.cancel()
	c.Log.Info("数据接收线程启动")
LOOP:
	for {
		select {
		case resp := <-c.dataChan:
			c.Log.Debugf("收到原始数据: [%X]", resp.ConvertBytes())
			switch f := resp.CtrFrame.(type) {
			case iec104.IFrame:
				// 处理I帧
				data, err := handleData(resp)
				if err != nil {
					c.Log.Errorf("解析处理I帧异常: %v", err)
					continue LOOP
				}

				c.outChan <- data
				c.Log.Debugf("获得数据: %v", data)
				// 响应S帧
				sFrame := iec104.SFrame{
					Recv: f.Send + 1,
				}
				apci, _ := iec104.NewAPCI(iec104.ApciLen, sFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err = c.conn.Write(resp.ConvertBytes())
				if err != nil {
					c.Log.Errorf("响应S帧[%v]异常: %v", resp, err)
					continue LOOP
				}
				c.Log.Debugf("响应S帧[%X]", resp.ConvertBytes())
			case iec104.SFrame:
				c.Log.Infof("接收S帧: [%X]", resp.ConvertBytes())
			default:
				c.Log.Warnf("未知类型APDU[%X]", resp.ConvertBytes())
			}
		case <-c.ctx.Done():
			c.Log.Info("数据接收线程停止")
		}
	}
}

// 《DL/T 634.5104-2009》 5.3 采用启/停的传输控制 图18
func (c Client) init() error {
	c.Log.Info("IEC104客户端通讯初始化")
	err := c.stop()
	if err != nil {
		return err
	}
	err = c.start()
	if err != nil {
		return err
	}
	return nil
}

// 《DL/T 634.5104-2009》 5.3 采用启/停的传输控制
func (c Client) stop() error {
	uFrame := iec104.UFrame{
		STOPDT_ACT: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		return fmt.Errorf("停止帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		return fmt.Errorf("停止帧创建异常: %v", err)
	}

	resp, err := c.writeUFrame(apdu)
	if err != nil {
		return fmt.Errorf("停止帧发送异常: %v", err)
	}
	if !resp.CtrFrame.(iec104.UFrame).STOPDT_CON {
		return fmt.Errorf("停止帧响应[%v]的停止确认没有置位", resp)
	}

	return nil
}

// 《DL/T 634.5104-2009》 5.3 采用启/停的传输控制
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

	resp, err := c.writeUFrame(apdu)
	if err != nil {
		return fmt.Errorf("启动帧发送异常: %v", err)
	}
	if !resp.CtrFrame.(iec104.UFrame).STARTDT_CON {
		return fmt.Errorf("启动帧响应[%v]的启动确认没有置位", resp)
	}

	return nil
}

// 《DL/T 634.5104-2009》 5.2 测试过程
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

	resp, err := c.writeUFrame(apdu)
	if err != nil {
		return fmt.Errorf("测试帧发送异常: %v", err)
	}
	if !resp.CtrFrame.(iec104.UFrame).TESTFR_CON {
		return fmt.Errorf("测试帧响应[%v]的测试确认没有置位", resp)
	}

	return nil
}

// 《DL/T 634.5104-2009》 7.5 总召唤
func (c Client) totalCall() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	iFrame := iec104.IFrame{
		Send: 0,
		Recv: 0,
	}
	apci, _ := iec104.NewAPCI(14, iFrame)
	asdu := elements.NewASDUC_IC_NA_1(elements.COT_ACT, 0x01, byte(elements.QOI_GLOBAL_CALL))
	apdu, _ := iec104.NewAPDU(apci, &asdu)

	_, err := c.conn.Write(apdu.ConvertBytes())
	if err != nil {
		return fmt.Errorf("总召唤发送异常: %v", err)
	}

	for {

		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			return fmt.Errorf("socket读操作异常: %v", err)
		}

		c.conn.SetDeadline(time.Now().Add(connectDeadline))
		c.Log.Debugf("下一次超时时间为: %v", time.Now().Add(connectDeadline).Format(time.RFC3339))

		resp, err := iec104.ParseAPDU(buf[:n])
		if err != nil {
			return fmt.Errorf("解析APDU异常: %v", err)
		}

		switch f := resp.CtrFrame.(type) {
		case iec104.IFrame:
			if resp.ASDU.DUI.TypeIdentification == elements.C_IC_NA_1 && resp.ASDU.DUI.Cause == elements.COT_ACTCON {
				// 这里可能会有异常
				sFrame := iec104.SFrame{
					Recv: f.Send + 1,
				}
				apci, _ := iec104.NewAPCI(iec104.ApciLen, sFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err = c.conn.Write(resp.ConvertBytes())
				if err != nil {
					return fmt.Errorf("响应S帧[%v]异常: %v", resp, err)
				}
				c.Log.Debugf("响应S帧[%X]", resp.ConvertBytes())
				return nil
			} else {
				c.dataChan <- resp
				return fmt.Errorf("总召唤激活确认异常")
			}
		case iec104.SFrame:
			c.dataChan <- resp
		case iec104.UFrame:
			c.ctrChan <- resp
		}
	}
}

func (c Client) uFrameResp() {
	defer c.cancel()
	c.Log.Info("U帧响应线程启动")
	for {
		select {
		case <-c.ctx.Done():
			c.Log.Info("U帧接收线程停止")
		case apdu := <-c.ctrChan:
			c.Log.Debugf("接收U帧[%v]", apdu)
			uFrame := apdu.CtrFrame.(iec104.UFrame)
			if uFrame.STARTDT_ACT {
				uFrame.STARTDT_ACT = false
				uFrame.STARTDT_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := c.conn.Write(resp.ConvertBytes())
				if err != nil {
					c.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				c.Log.Debugf("响应U帧[%v]", resp)
			} else if uFrame.STOPDT_ACT {
				uFrame.STOPDT_ACT = false
				uFrame.STOPDT_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := c.conn.Write(resp.ConvertBytes())
				if err != nil {
					c.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				c.Log.Debugf("响应U帧[%v]", resp)
			} else if uFrame.TESTFR_ACT {
				uFrame.TESTFR_ACT = false
				uFrame.TESTFR_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := c.conn.Write(resp.ConvertBytes())
				if err != nil {
					c.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				c.Log.Debugf("响应U帧[%v]", resp)
			} else {
				c.Log.Debugf("U帧无需响应")
			}
		}
	}
}

func (c Client) reconnect() error {
	err := c.stop()
	if err != nil {
		return err
	}
	err = c.start()
	if err != nil {
		return err
	}
	err = c.totalCall()
	if err != nil {
		return err
	}
	return nil
}

func (c Client) read() {
	defer c.cancel()
	c.Log.Info("socket读线程启动")
	for {
		select {
		case <-c.ctx.Done():
			c.Log.Info("socket读线程停止")
		default:
		}
		c.mux.Lock()
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			c.Log.Errorf("socket读操作异常: %v", err)
			return
		}

		c.conn.SetDeadline(time.Now().Add(connectDeadline))
		c.Log.Debugf("下一次超时时间为: %v", time.Now().Add(connectDeadline).Format(time.RFC3339))

		apdu, err := iec104.ParseAPDU(buf[:n])
		if err != nil {
			c.Log.Warnf("解析APDU异常: %v", err)
			continue
		}

		switch apdu.CtrFrame.(type) {
		case iec104.IFrame:
			c.dataChan <- apdu
		case iec104.SFrame:
			c.dataChan <- apdu
		case iec104.UFrame:
			c.ctrChan <- apdu
		}
		c.mux.Unlock()
	}
}

func (c Client) writeUFrame(apdu iec104.APDU) (iec104.APDU, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	_, err := c.conn.Write(apdu.ConvertBytes())
	if err != nil {
		return iec104.APDU{}, err
	}

	c.Log.Debugf("发送: [%X]", apdu.ConvertBytes())
	for {

		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			return iec104.APDU{}, fmt.Errorf("socket读操作异常: %v", err)
		}

		c.conn.SetDeadline(time.Now().Add(connectDeadline))
		c.Log.Debugf("下一次超时时间为: %v", time.Now().Add(connectDeadline).Format(time.RFC3339))

		apdu, err := iec104.ParseAPDU(buf[:n])
		if err != nil {
			return iec104.APDU{}, fmt.Errorf("解析APDU异常: %v", err)
		}

		switch f := apdu.CtrFrame.(type) {
		case iec104.IFrame:
			c.dataChan <- apdu
		case iec104.SFrame:
			c.dataChan <- apdu
		case iec104.UFrame:
			if f.STARTDT_ACT || f.STOPDT_ACT || f.TESTFR_ACT {
				c.ctrChan <- apdu
			}
			return apdu, nil
		}
	}
}

func handleData(apdu iec104.APDU) (map[string]float32, error) {
	switch apdu.ASDU.DUI.TypeIdentification {
	case elements.M_ME_NC_1:
		var values map[string]float32
		switch mb := apdu.ASDU.MessageBody.(type) {
		case elements.MessageElement_13_SQ_1:
			address := int(mb.Address)
			for _, e := range mb.Cores {
				//TODO:考虑QDS
				values[fmt.Sprint(address)] = e.Value
				address++
			}
			return values, nil
		case elements.MessageElement_13_SQ_0:
			for _, e := range mb {
				//TODO:考虑QDS
				values[fmt.Sprint(int(e.Address))] = e.Core.Value
			}
			return values, nil
		default:
			return nil, fmt.Errorf("未知信息元素类型[%T]", mb)
		}
	default:
		return nil, fmt.Errorf("未支持ASDU类型: %v", apdu.ASDU.DUI.TypeIdentification)
	}
}
