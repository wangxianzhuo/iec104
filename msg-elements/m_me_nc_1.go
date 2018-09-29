package elements

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	M_ME_NC_1_SQ_1_MSG_LEN = 5
	M_ME_NC_1_SQ_0_MSG_LEN = 6
)

// MessageElement_13_SQ_1 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=1的信息元素
type MessageElement_13_SQ_1 struct {
	Address byte
	Cores   []MessageElementCore_13
}

func (e MessageElement_13_SQ_1) ConvertBytes() []byte {
	var cores []byte
	for _, c := range e.Cores {
		cores = append(cores, c.ConvertBytes()...)
	}
	return append([]byte{e.Address}, cores...)
}

// MessageElement_13_SQ_0_Ele 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=0的信息元素
type MessageElement_13_SQ_0_Ele struct {
	Address byte
	Core    MessageElementCore_13
}

func (e MessageElement_13_SQ_0_Ele) ConvertBytes() []byte {
	return append([]byte{e.Address}, e.Core.ConvertBytes()...)
}

type MessageElement_13_SQ_0 []MessageElement_13_SQ_0_Ele

func (e MessageElement_13_SQ_0) ConvertBytes() []byte {
	var result []byte
	for _, ele := range e {
		result = append(result, ele.ConvertBytes()...)
	}
	return result
}

type MessageElementCore_13 struct {
	Value float32 // 浮点值
	QDS   QDS
}

func (c MessageElementCore_13) ConvertBytes() []byte {
	v := math.Float32bits(c.Value)
	return append([]byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}, c.QDS.ConvertBytes()...)
}

// QDS 品质描述词，《DLT 634.5101-2002》 7.2.6.3
type QDS struct {
	OV bool // false(0) = 未溢出 | true(1) = 溢出
	BL bool // false(0) = 未被锁闭 | true(1) = 被锁闭
	SB bool // false(0) = 未被取代 | true(1) = 被取代
	NT bool // false(0) = 非当前值 | true(1) = 当前值
	IV bool // false(0) = 无效 | true(1) = 有效
}

func (qds QDS) ConvertBytes() []byte {
	var result byte = 0x00
	if qds.OV {
		result += 0x01
	}
	if qds.BL {
		result += 0x10
	}
	if qds.SB {
		result += 0x20
	}
	if qds.NT {
		result += 0x40
	}
	if qds.IV {
		result += 0x80
	}
	return []byte{
		result,
	}
}

func parseM_ME_NC_1(asdu []byte, dui DUI) (BytesConverter, error) {
	sq := dui.VariableStructureQualifier >> 7
	number := dui.VariableStructureQualifier & 0x7F

	switch sq {
	case 0:
		msgBody := asdu[4:len(asdu)]
		if len(msgBody)/M_ME_NC_1_SQ_0_MSG_LEN != int(number) {
			return nil, fmt.Errorf("asdu的信息对象[%x]数量[%v]与dui中的数量[%v]不匹配", msgBody, len(msgBody)/M_ME_NC_1_SQ_0_MSG_LEN, number)
		}
		var elements MessageElement_13_SQ_0
		for i := 0; i < int(number)*M_ME_NC_1_SQ_0_MSG_LEN; i += M_ME_NC_1_SQ_0_MSG_LEN {
			address := msgBody[i]
			value := math.Float32frombits(binary.BigEndian.Uint32(msgBody[i+1 : i+5]))
			qds := ParseQDS(msgBody[i+5])
			element := MessageElement_13_SQ_0_Ele{
				Address: address,
				Core: MessageElementCore_13{
					Value: value,
					QDS:   qds,
				},
			}
			elements = append(elements, element)
		}
		return elements, nil
	default:
		address := asdu[4]
		msgBody := asdu[5:len(asdu)]
		if len(msgBody)/M_ME_NC_1_SQ_1_MSG_LEN != int(number) {
			return nil, fmt.Errorf("asdu的信息对象[%x]数量[%v]与dui中的数量[%v]不匹配", msgBody, len(msgBody)/M_ME_NC_1_SQ_1_MSG_LEN, number)
		}
		var elements MessageElement_13_SQ_1
		elements.Address = address
		for i := 0; i < int(number)*M_ME_NC_1_SQ_1_MSG_LEN; i += M_ME_NC_1_SQ_1_MSG_LEN {
			value := math.Float32frombits(binary.BigEndian.Uint32(msgBody[i+1 : i+5]))
			qds := ParseQDS(msgBody[i+5])
			core := MessageElementCore_13{
				Value: value,
				QDS:   qds,
			}
			elements.Cores = append(elements.Cores, core)
		}
		return elements, nil
	}
}

// ParseQDS 解析QDS
func ParseQDS(qds byte) QDS {
	return QDS{
		OV: qds&0x01 == 1,
		BL: qds&0x10 == 1,
		SB: qds&0x20 == 1,
		NT: qds&0x40 == 1,
		IV: qds&0x80 == 1,
	}
}
