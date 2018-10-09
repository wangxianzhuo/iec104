package iec104

import (
	"fmt"

	elements "github.com/wangxianzhuo/iec104/msg-elements"
)

const (
	ApciLen = 6
)

type APDU struct {
	APCI     APCI
	ASDU     elements.ASDU
	Len      int
	ASDULen  int
	ctrType  byte
	ctrFrame interface{}
}

func NewAPDU(apci APCI, asdu *elements.ASDU) (APDU, error) {
	var asduLen int
	if asdu == nil {
		asduLen = 0
	} else {
		asduLen = len(asdu.ConvertBytes())
	}
	t, f, err := ParseCtr(apci)
	if err != nil {
		return APDU{}, fmt.Errorf("解析控制域异常: %v", err)
	}
	return APDU{
		APCI:     apci,
		ASDU:     *asdu,
		Len:      ApciLen + asduLen,
		ASDULen:  asduLen,
		ctrType:  t,
		ctrFrame: f,
	}, nil
}

func ParseAPDU(input []byte) (APDU, error) {
	if input == nil || len(input) < 6 {
		return APDU{}, fmt.Errorf("APDU报文[%X]非法", input)
	}
	start := input[0]
	if start != 0x68 {
		return APDU{}, fmt.Errorf("APDU报文[%X]不是完整报文，找不到启动字符68H", input)
	}

	var apci APCI
	apci.Start = input[0]
	apci.ApduLen = int(input[1])
	apci.Ctr1 = input[2]
	apci.Ctr2 = input[3]
	apci.Ctr3 = input[4]
	apci.Ctr4 = input[5]

	fType, ctrFrame, err := ParseCtr(apci)
	if err != nil {
		return APDU{}, fmt.Errorf("APDU报文[%X]解析控制域异常: %v", input, err)
	}

	var asdu elements.ASDU
	var asduLen int

	if len(input[6:len(input)]) < 1 {
		asdu = elements.ASDU{}
		asduLen = 0
	} else {
		asdu, err = elements.ParseASDU(input[6:len(input)])
		if err != nil {
			return APDU{}, fmt.Errorf("APDU报文[%X]解析ASDU域[%X]异常: %v", input, input[6:len(input)], err)
		}
		asduLen = len(input[6:len(input)])
	}

	return APDU{
		APCI:     apci,
		ASDU:     asdu,
		Len:      apci.ApduLen,
		ASDULen:  asduLen,
		ctrType:  fType,
		ctrFrame: ctrFrame,
	}, nil
}

func (apdu APDU) ConvertBytes() []byte {
	var asdu []byte
	if apdu.ASDULen < 1 {
		asdu = []byte{}
	} else {
		asdu = apdu.ASDU.ConvertBytes()
	}
	return append(apdu.APCI.ConvertBytes(),
		asdu...,
	)
}
