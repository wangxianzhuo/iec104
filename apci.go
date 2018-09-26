package iec

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

type IFrame struct {
	Send int16
	Recv int16
}

type SFrame struct {
	Recv int16
}

type UFrame struct {
	TESTFR_CON byte //U帧，测试确认
	TESTFR_ACT byte //U帧，测试激活

	STOPDT_CON byte //U帧，停止确认
	STOPDT_ACT byte //U帧，停止激活

	STARTDT_CON byte //U帧，启动确认
	STARTDT_ACT byte //U帧，启动激活
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
	testfrCon := apci.Ctr1 & 0x80
	testfrAct := apci.Ctr1 & 0x40

	stopdtCon := apci.Ctr1 & 0x20
	stopdtAct := apci.Ctr1 & 0x10

	startdtCon := apci.Ctr1 & 0x08
	startdtAct := apci.Ctr1 & 0x04

	return 3, UFrame{
		TESTFR_ACT:  testfrAct,
		TESTFR_CON:  testfrCon,
		STOPDT_ACT:  stopdtAct,
		STOPDT_CON:  stopdtCon,
		STARTDT_ACT: startdtAct,
		STARTDT_CON: startdtCon,
	}
}
