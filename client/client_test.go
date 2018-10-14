package client

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/wangxianzhuo/iec104/msg-elements"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/wangxianzhuo/iec104"
)

func Test_start(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	uFrame := iec104.UFrame{
		STARTDT_CON: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	err = c.start()
	if err != nil {
		t.Fatal(err)
	}
}
func Test_test(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	uFrame := iec104.UFrame{
		TESTFR_CON: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	err = c.test()
	if err != nil {
		t.Fatal(err)
	}

	err = c.test()
	if err != nil {
		t.Fatal(err)
	}
}
func Test_init(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	uFrame := iec104.UFrame{
		STOPDT_CON: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	uFrame2 := iec104.UFrame{
		STARTDT_CON: true,
	}
	apci2, err := iec104.NewAPCI(iec104.ApciLen, uFrame2)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu2, err := iec104.NewAPDU(apci2, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	go echoServer2(":2404", [][]byte{apdu.ConvertBytes(), apdu2.ConvertBytes()}, ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	err = c.init()
	if err != nil {
		t.Fatal(err)
	}
}
func Test_stop(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	uFrame := iec104.UFrame{
		STOPDT_CON: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	err = c.start()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_connectionTest(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	uFrame := iec104.UFrame{
		TESTFR_CON: true,
	}
	apci, err := iec104.NewAPCI(iec104.ApciLen, uFrame)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}

	apdu, err := iec104.NewAPDU(apci, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	c.connectionTest(5 * time.Second)
}
func Test_totalCall(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	iFrame := iec104.IFrame{
		Send: 1,
		Recv: 0,
	}
	apci, _ := iec104.NewAPCI(14, iFrame)
	asdu := elements.NewASDUC_IC_NA_1(elements.COT_ACTCON, 0x01, byte(elements.QOI_GLOBAL_CALL))
	apdu, err := iec104.NewAPDU(apci, &asdu)
	if err != nil {
		t.Fatal(err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)

	time.Sleep(1 * time.Second)

	c := startClient(address)
	err = c.totalCall()
	if err != nil {
		t.Fatal(err)
	}
}
func Test_uFrameResp(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	address := "127.0.0.1:2404"

	iFrame := iec104.IFrame{
		Send: 1,
		Recv: 0,
	}
	apci, _ := iec104.NewAPCI(14, iFrame)
	asdu := elements.NewASDUC_IC_NA_1(elements.COT_ACTCON, 0x01, byte(elements.QOI_GLOBAL_CALL))
	apdu, err := iec104.NewAPDU(apci, &asdu)
	if err != nil {
		t.Fatal(err)
	}

	go echoServer(":2404", apdu.ConvertBytes(), ctx)
	time.Sleep(1 * time.Second)

	uFrame1 := iec104.UFrame{
		STARTDT_ACT: true,
	}
	apci1, err := iec104.NewAPCI(iec104.ApciLen, uFrame1)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}
	apdu1, err := iec104.NewAPDU(apci1, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	uFrame2 := iec104.UFrame{
		STOPDT_ACT: true,
	}
	apci2, err := iec104.NewAPCI(iec104.ApciLen, uFrame2)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}
	apdu2, err := iec104.NewAPDU(apci2, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	uFrame3 := iec104.UFrame{
		TESTFR_ACT: true,
	}
	apci3, err := iec104.NewAPCI(iec104.ApciLen, uFrame3)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}
	apdu3, err := iec104.NewAPDU(apci3, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	uFrame4 := iec104.UFrame{
		STARTDT_CON: true,
	}
	apci4, err := iec104.NewAPCI(iec104.ApciLen, uFrame4)
	if err != nil {
		t.Errorf("启动帧控制域创建异常: %v", err)
	}
	apdu4, err := iec104.NewAPDU(apci4, nil)
	if err != nil {
		t.Errorf("启动帧创建异常: %v", err)
	}

	c := startClient(address)
	go c.uFrameResp()
	c.ctrChan <- apdu1
	c.ctrChan <- apdu2
	c.ctrChan <- apdu3
	c.ctrChan <- apdu4

}

func echoServer(address string, resp []byte, ctx context.Context) error {
	defer log.Printf("echo server stop")
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Printf("echo server start")

	conn, err := l.Accept()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}

		log.Printf("READ: [%X]", buf[:n])
		_, err = conn.Write(resp)
		if err != nil {
			return err
		}
		log.Printf("SEND: [%X]", resp)
	}
}

func echoServer2(address string, resp [][]byte, ctx context.Context) error {
	defer log.Printf("echo server stop")
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Printf("echo server start")

	conn, err := l.Accept()
	if err != nil {
		return err
	}

	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}

		log.Printf("READ: [%X]", buf[:n])
		_, err = conn.Write(resp[i])
		if err != nil {
			return err
		}
		log.Printf("SEND: [%X]", resp[i])
		i++
	}
}

func startClient(address string) Client {
	outChan := make(chan map[string]float32)
	logrus.SetLevel(logrus.DebugLevel)
	c, _, err := New(address, outChan, logrus.WithField("client", "iec104"))
	if err != nil {
		panic(err)
	}

	return c
}
