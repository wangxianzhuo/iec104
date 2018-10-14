package elements

import (
	"fmt"
)

// ParseASDU 解析asdu
func ParseASDU(asdu []byte) (ASDU, error) {
	if asdu == nil || len(asdu) < 4 {
		return ASDU{}, fmt.Errorf("asdu[%X]非法", asdu)
	}
	dui, err := parseDUI(asdu)
	if err != nil {
		return ASDU{}, fmt.Errorf("解析asdu[%X]的DUI异常: %v", asdu, err)
	}

	var messageBody BytesConverter
	switch dui.TypeIdentification {
	case M_ME_NC_1:
		var err error
		messageBody, err = parseM_ME_NC_1(asdu, dui)
		if err != nil {
			return ASDU{}, fmt.Errorf("解析asdu[%X]的messageBody异常: %v", asdu, err)
		}
	case C_IC_NA_1:
		messageBody = parseC_IC_NA_1(asdu)
	default:
		return ASDU{}, fmt.Errorf("未知类型标识[%v]", dui.TypeIdentification)
	}

	return ASDU{
		DUI:         dui,
		MessageBody: messageBody,
	}, nil
}

func parseDUI(asdu []byte) (DUI, error) {
	var dui DUI
	var err error

	dui.TypeIdentification, err = parseDUITypeIdentification(asdu[0])
	if err != nil {
		return DUI{}, fmt.Errorf("解析asdu的dui的TypeIdentification异常:%v", err)
	}

	switch dui.TypeIdentification {
	case M_ME_NC_1:
		dui.VariableStructureQualifier = asdu[1]
		dui.Cause = asdu[2]
		dui.PublicAddressLow = asdu[3]
	}
	return dui, nil
}

func parseDUITypeIdentification(t byte) (byte, error) {
	switch t {
	case M_ME_NC_1:
		return t, nil
	case C_IC_NA_1:
		return t, nil
	default:
		return 0x00, fmt.Errorf("未知类型标识[%X]", t)
	}
}
