package iec104

import (
	"fmt"
)

type APCI struct {
	Start   byte
	ApduLen int
	Ctr1    byte
	Ctr2    byte
	Ctr3    byte
	Ctr4    byte
}

func (apci APCI) ConvertBytes() []byte {
	return []byte{
		apci.Start,
		byte(apci.ApduLen),
		apci.Ctr1,
		apci.Ctr2,
		apci.Ctr3,
		apci.Ctr4,
	}
}

type IFrame struct {
	Send int16
	Recv int16
}

type SFrame struct {
	Recv int16
}

type UFrame struct {
	TESTFR_CON bool //U帧，测试确认
	TESTFR_ACT bool //U帧，测试激活

	STOPDT_CON bool //U帧，停止确认
	STOPDT_ACT bool //U帧，停止激活

	STARTDT_CON bool //U帧，启动确认
	STARTDT_ACT bool //U帧，启动激活
}

func ParseCtr(apci APCI) (byte, interface{}, error) {
	frameType := apci.Ctr1 & 0x03
	switch frameType {
	case 0:
		t, f := parseIFrame(apci)
		return t, f, nil
	case 1:
		t, f := parseSFrame(apci)
		return t, f, nil
	case 3:
		t, f := parseUFrame(apci)
		return t, f, nil
	default:
		return 0xFF, nil, fmt.Errorf("未知APCI帧类型: [%X]", frameType)
	}
}

func parseIFrame(apci APCI) (byte, IFrame) {
	send := int16(apci.Ctr1)>>1 + int16(apci.Ctr2)<<7
	recv := int16(apci.Ctr3)>>1 + int16(apci.Ctr4)<<7

	return 0, IFrame{
		Send: send,
		Recv: recv,
	}
}

func parseSFrame(apci APCI) (byte, SFrame) {
	recv := int16(apci.Ctr3)>>1 + int16(apci.Ctr4)<<7

	return 1, SFrame{
		Recv: recv,
	}
}

func parseUFrame(apci APCI) (byte, UFrame) {
	testfrCon := apci.Ctr1&0x80 != 0
	testfrAct := apci.Ctr1&0x40 != 0

	stopdtCon := apci.Ctr1&0x20 != 0
	stopdtAct := apci.Ctr1&0x10 != 0

	startdtCon := apci.Ctr1&0x08 != 0
	startdtAct := apci.Ctr1&0x04 != 0

	return 3, UFrame{
		TESTFR_ACT:  testfrAct,
		TESTFR_CON:  testfrCon,
		STOPDT_ACT:  stopdtAct,
		STOPDT_CON:  stopdtCon,
		STARTDT_ACT: startdtAct,
		STARTDT_CON: startdtCon,
	}
}

func NewAPCI(len int, ctr interface{}) (APCI, error) {
	var apci APCI
	apci.Start = 0x68
	apci.ApduLen = len

	switch frame := ctr.(type) {
	case IFrame:
		apci.Ctr1 = byte(frame.Send) << 1
		apci.Ctr2 = byte(frame.Send >> 8)
		apci.Ctr3 = byte(frame.Recv) << 1
		apci.Ctr4 = byte(frame.Recv >> 8)
	case SFrame:
		apci.Ctr1 = 1
		apci.Ctr2 = 0
		apci.Ctr3 = byte(frame.Recv) << 1
		apci.Ctr4 = byte(frame.Recv >> 8)
	case UFrame:
		var ctr1 byte
		if frame.STARTDT_ACT {
			ctr1 += 0x04
		}
		if frame.STARTDT_CON {
			ctr1 += 0x08
		}
		if frame.STOPDT_ACT {
			ctr1 += 0x10
		}
		if frame.STOPDT_CON {
			ctr1 += 0x20
		}
		if frame.TESTFR_ACT {
			ctr1 += 0x40
		}
		if frame.TESTFR_CON {
			ctr1 += 0x80
		}
		apci.Ctr1 = ctr1 + 3
		apci.Ctr2 = 0
		apci.Ctr3 = 0
		apci.Ctr4 = 0
	default:
		return APCI{}, fmt.Errorf("未知控制域类型[%T]", ctr)
	}

	return apci, nil
}
