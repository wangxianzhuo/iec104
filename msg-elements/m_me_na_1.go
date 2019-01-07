package elements

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// MessageElement_9_SQ_1 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=1的信息元素
type MessageElement_9_SQ_1 struct {
	Address uint32
	Cores   []MessageElementCore_9
}

func (e MessageElement_9_SQ_1) ConvertBytes() []byte {
	var cores []byte
	for _, c := range e.Cores {
		cores = append(cores, c.ConvertBytes()...)
	}
	return append([]byte{byte(e.Address & 0xFF), byte(e.Address >> 8 & 0xFF), byte(e.Address >> 16 & 0xFF)}, cores...)
}

// MessageElement_9_SQ_0_Ele 测量值，短浮点数，《DLT 634.5101-2002》 7.3.1.13 13:M_ME_NC_1，SQ=0的信息元素
type MessageElement_9_SQ_0_Ele struct {
	Address uint32
	Core    MessageElementCore_9
}

func (e MessageElement_9_SQ_0_Ele) ConvertBytes() []byte {
	return append([]byte{byte(e.Address & 0xFF), byte(e.Address)}, e.Core.ConvertBytes()...)
}

type MessageElement_9_SQ_0 []MessageElement_9_SQ_0_Ele

func (e MessageElement_9_SQ_0) ConvertBytes() []byte {
	var result []byte
	for _, ele := range e {
		result = append(result, ele.ConvertBytes()...)
	}
	return result
}

type MessageElementCore_9 struct {
	Value int16 // 规一化值
	QDS   QDS
}

func (c MessageElementCore_9) ConvertBytes() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, int16(c.Value))
	v := bytesBuffer.Bytes()
	return append([]byte{
		v[0],
		v[1],
	}, c.QDS.ConvertBytes()...)
}

func parseM_ME_NA_1(asdu []byte, dui DUI) (BytesConverter, error) {
	// fmt.Printf("解析asdu: % X\n", asdu)
	sq := dui.VariableStructureQualifier >> 7
	number := dui.VariableStructureQualifier & 0x7F

	switch sq {
	case 0:
		size := len(asdu[6:len(asdu)]) / int(number)
		msgBody := asdu[6:len(asdu)]
		var elements MessageElement_9_SQ_0
		for i := 0; i < int(number)*size; i += size {
			// fmt.Printf("msgBody: % X\n", msgBody)
			address := binary.LittleEndian.Uint32(append([]byte{msgBody[i], msgBody[i+1], msgBody[i+2]}, 0x00))
			// value := math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i+3 : i+7]))
			value, _ := getValueWithComplementUseLittleEndian(msgBody[i+3 : i+5])
			// fmt.Printf("raw: %X, value: %v\n", msgBody[i+3:i+7], math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i+3:i+7])))
			qds := ParseQDS(msgBody[i+5])
			element := MessageElement_9_SQ_0_Ele{
				Address: address,
				Core: MessageElementCore_9{
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
		var elements MessageElement_9_SQ_1
		elements.Address = address
		for i := 0; i < int(number)*size; i += size {
			// value := math.Float32frombits(binary.LittleEndian.Uint32(msgBody[i : i+4]))
			value, _ := getValueWithComplementUseLittleEndian(msgBody[i : i+2])
			qds := ParseQDS(msgBody[i+2])
			core := MessageElementCore_9{
				Value: value,
				QDS:   qds,
			}
			elements.Cores = append(elements.Cores, core)
		}
		return elements, nil
	}
}

func getValueWithComplementUseLittleEndian(int16Bytes []byte) (int16, error) {
	if len(int16Bytes) < 2 {
		return 0, fmt.Errorf("only parse 2 bytes value")
	}

	if (int16Bytes[1] & 0x80) == 0x80 {
		v := binary.LittleEndian.Uint16(int16Bytes)
		t := (^v + 1)
		return int16(t) * -1, nil
	} else {
		return int16(binary.LittleEndian.Uint16(int16Bytes)), nil
	}
}
