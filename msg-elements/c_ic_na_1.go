package elements

// QOI:
// 	20 站召唤（全局）
// 	21 第1组召唤
// 	22 第2组召唤
// 	23 第3组召唤
// 	24 第4组召唤
// 	25 第5组召唤
// 	26 第6组召唤
// 	27 第7组召唤
// 	28 第8组召唤
// 	29 第9组召唤
// 	30 第10组召唤
// 	31 第11组召唤
// 	32 第12组召唤
// 	33 第13组召唤
// 	34 第14组召唤
// 	35 第15组召唤
// 	36 第16组召唤

const (
	QOI_GLOBAL_CALL = 20
	QIO_GROUP_1     = 21
)

type MessageElement_100 struct {
	Address byte // 信息对象地址
	QOI     byte // 召唤限定词，《DLT 634.5101-2002》 7.2.6.22
}

func (e MessageElement_100) ConvertBytes() []byte {
	return []byte{
		e.Address,
		e.QOI,
	}
}

func parseC_IC_NA_1(asdu []byte) MessageElement_100 {
	return MessageElement_100{
		Address: asdu[4],
		QOI:     asdu[5],
	}
}

func NewASDUC_IC_NA_1(cause, publicAddress, qoi byte) ASDU {
	return ASDU{
		DUI: DUI{
			TypeIdentification:         C_IC_NA_1,
			VariableStructureQualifier: 0x01,
			Cause:                  cause,
			PublicAddressLow:       publicAddress,
			PublicAddressHigEnable: false,
		},
		MessageBody: MessageElement_100{
			Address: 0,
			QOI:     qoi,
		},
	}
}
