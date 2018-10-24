package elements

import (
	"encoding/binary"
	"math"
)

const (
	M_ME_NC_1_SQ_1_MSG_LEN = 5
	M_ME_NC_1_SQ_0_MSG_LEN = 8
)

// MessageElement_13_SQ_1 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=1的信息元素
type MessageElement_13_SQ_1 struct {
	Address uint32
	Cores   []MessageElementCore_13
}

func (e MessageElement_13_SQ_1) ConvertBytes() []byte {
	var cores []byte
	for _, c := range e.Cores {
		cores = append(cores, c.ConvertBytes()...)
	}
	return append([]byte{byte(e.Address & 0xFF), byte(e.Address >> 8 & 0xFF), byte(e.Address >> 16 & 0xFF)}, cores...)
}

// MessageElement_13_SQ_0_Ele 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=0的信息元素
type MessageElement_13_SQ_0_Ele struct {
	Address uint32
	Core    MessageElementCore_13
}

func (e MessageElement_13_SQ_0_Ele) ConvertBytes() []byte {
	return append([]byte{byte(e.Address & 0xFF), byte(e.Address)}, e.Core.ConvertBytes()...)
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
	// fmt.Printf("解析asdu: % X\n", asdu)
	sq := dui.VariableStructureQualifier >> 7
	number := dui.VariableStructureQualifier & 0x7F

	switch sq {
	case 0:
		size := len(asdu[6:len(asdu)]) / int(number)
		msgBody := asdu[6:len(asdu)]
		var elements MessageElement_13_SQ_0
		for i := 0; i < int(number)*size; i += size {
			// fmt.Printf("msgBody: % X\n", msgBody)
			address := binary.LittleEndian.Uint32(append([]byte{msgBody[i], msgBody[i+1], msgBody[i+2]}, 0x00))
			value := math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i+3 : i+7]))
			// fmt.Printf("raw: %X, value: %v\n", msgBody[i+3:i+7], math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i+3:i+7])))
			qds := ParseQDS(msgBody[i+7])
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
		address := binary.LittleEndian.Uint32(append([]byte{asdu[6], asdu[7], asdu[8]}, 0x00))
		msgBody := asdu[9:len(asdu)]
		size := len(asdu[9:len(asdu)]) / int(number)
		var elements MessageElement_13_SQ_1
		elements.Address = address
		for i := 0; i < int(number)*size; i += size {
			value := math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i : i+4]))
			qds := ParseQDS(msgBody[i+4])
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
