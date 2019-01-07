package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/wangxianzhuo/iec104"
)

func main() {

}

var (
	connectDeadline time.Duration
)

// Server IEC104从站
type Server struct {
	conn     net.Conn
	dataChan chan iec104.APDU
	ctrChan  chan iec104.APDU
	outChan  chan map[string]float32
	ctx      context.Context
	cancel   context.CancelFunc
	Log      *logrus.Entry
	mux      *sync.Mutex
	vs       int
	vr       int
	ack      int
}

func (s Server) uFrameResp() {
	defer s.cancel()
	s.Log.Info("U帧响应线程启动")
	for {
		select {
		case <-s.ctx.Done():
			s.Log.Info("U帧接收线程停止")
		case apdu := <-s.ctrChan:
			s.Log.Debugf("接收U帧[%v]", apdu)
			uFrame := apdu.CtrFrame.(iec104.UFrame)
			if uFrame.STARTDT_ACT {
				uFrame.STARTDT_ACT = false
				uFrame.STARTDT_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := s.conn.Write(resp.ConvertBytes())
				if err != nil {
					s.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				s.Log.Debugf("响应U帧[%v]", resp)
			} else if uFrame.STOPDT_ACT {
				uFrame.STOPDT_ACT = false
				uFrame.STOPDT_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := s.conn.Write(resp.ConvertBytes())
				if err != nil {
					s.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				s.Log.Debugf("响应U帧[%v]", resp)
			} else if uFrame.TESTFR_ACT {
				uFrame.TESTFR_ACT = false
				uFrame.TESTFR_CON = true
				apci, _ := iec104.NewAPCI(iec104.ApciLen, uFrame)
				resp, _ := iec104.NewAPDU(apci, nil)
				_, err := s.conn.Write(resp.ConvertBytes())
				if err != nil {
					s.Log.Errorf("响应U帧[%v]异常: %v", apdu, err)
					continue
				}
				s.Log.Debugf("响应U帧[%v]", resp)
			} else {
				s.Log.Debugf("U帧无需响应")
			}
		}
	}
}

func (s Server) send() {

}
func (s Server) receive() {
	defer s.cancel()
	s.Log.Info("数据接收线程启动")
	for {
		select {
		case resp := <-s.dataChan:
			switch f := resp.CtrFrame.(type) {
			case iec104.IFrame:
			// 	// 处理I帧
			// 	if resp.ASDU.DUI.TypeIdentification != elements.C_IC_NA_1 || resp.ASDU.DUI.Cause != elements.COT_ACTCON {
			// 		// 非总召唤确认
			// 		// 这里可能会有异常
			// 		c.Log.Debugf("准备解析APDU[%v]浮点数", resp)
			// 		data, err := handleData(resp)
			// 		if err != nil {
			// 			c.Log.Errorf("解析处理I帧异常: %v", err)
			// 			continue LOOP
			// 		}

			// 		c.outChan <- data
			// 		c.Log.Debugf("获得数据: %v", data)
			// 	}

			// 	// 响应S帧
			// 	sFrame := iec104.SFrame{
			// 		Recv: f.Send + 1,
			// 	}
			// 	apci, _ := iec104.NewAPCI(iec104.ApciLen, sFrame)
			// 	resp, _ := iec104.NewAPDU(apci, nil)
			// 	_, err := c.conn.Write(resp.ConvertBytes())
			// 	if err != nil {
			// 		c.Log.Errorf("响应S帧[%v]异常: %v", resp, err)
			// 		continue LOOP
			// 	}
			// 	c.Log.Debugf("响应S帧[%X]", resp.ConvertBytes())
			case iec104.SFrame:
				s.Log.Infof("接收S帧: [%X]", resp.ConvertBytes())
			default:
				s.Log.Warnf("未知类型APDU[%X]", resp.ConvertBytes())
			}
		case <-s.ctx.Done():
			s.Log.Info("数据接收线程停止")
		}
	}
}
func (s Server) totalCallResp() {
	b := []byte{0x68, 0x3A, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x06, 0x03, 0x00, 0x01, 0x00, 0x09, 0x40, 0x00, 0x23, 0xDB, 0x11, 0x40, 0x00, 0x06, 0x40, 0x00, 0x75, 0x3C, 0x58, 0x3F, 0x00, 0x01, 0x40, 0x00, 0x94, 0x17, 0xC3, 0x40, 0x00, 0x0E, 0x40, 0x00, 0xCD, 0x4C, 0x7A, 0x42, 0x00, 0x02, 0x40, 0x00, 0x9D, 0x68, 0x27, 0x3C, 0x00, 0x04, 0x40, 0x00, 0xFF, 0x04, 0x08, 0x40, 0x00}
	fmt.Println(b)
}
func (s Server) read() {
	defer s.cancel()
	s.Log.Info("socket读线程启动")
	for {
		select {
		case <-s.ctx.Done():
			s.Log.Info("socket读线程停止")
		default:
		}
		s.mux.Lock()
		buf := make([]byte, 1024)
		n, err := s.conn.Read(buf)
		if err != nil {
			s.Log.Errorf("socket读操作异常: %v", err)
			s.mux.Unlock()
			return
		}

		s.conn.SetDeadline(time.Now().Add(connectDeadline))
		s.Log.Debugf("下一次超时时间为: %v", time.Now().Add(connectDeadline).Format(time.RFC3339))

		s.Log.Debugf("收到原始数据: [% X]", buf[:n])
		apdu, err := iec104.ParseAPDU(buf[:n])
		if err != nil {
			s.Log.Warnf("解析APDU异常: %v", err)
			s.mux.Unlock()
			continue
		}

		switch apdu.CtrFrame.(type) {
		case iec104.IFrame:
			s.dataChan <- apdu
		case iec104.SFrame:
			s.dataChan <- apdu
		case iec104.UFrame:
			s.ctrChan <- apdu
		}
		s.mux.Unlock()
	}
}
