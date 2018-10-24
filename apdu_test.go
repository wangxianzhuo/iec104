package iec104

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_parse(t *testing.T) {
	ins, _ := hex.DecodeString("6832000000000D050300010005400026365F3C00094000C1CA114000064000075E8D3F000240009D68273C000440008D92134000")
	apdu, err := ParseAPDU(ins)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(apdu)
}

// func Test_ParseAPDUUFrame(t *testing.T) {
// 	ins, err := hex.DecodeString("680407000000")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	apdu, err := ParseAPDU(ins)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	uFrame, ok := apdu.ctrFrame.(UFrame)
// 	if !ok {
// 		t.Fatalf("apdu不是U格式")
// 	}
// 	if !uFrame.STARTDT_ACT {
// 		t.Fatalf("apdu的U格式的控制域中的startdt不是处于act状态")
// 	}
// 	t.Logf("apdu: %v", apdu)
// 	t.Logf("apdu.apci: %v", apdu.APCI)
// 	t.Logf("apdu.asdu: %v", apdu.ASDU)
// 	t.Logf("apdu.len: %v", apdu.Len)
// 	t.Logf("apdu.ctrType: %T", apdu.ctrFrame)
// 	t.Logf("apdu.ctrFrame: %v", uFrame)
// 	t.Logf("apdu bytes: [%X]", apdu.ConvertBytes())
// 	if bytes.Compare(ins, apdu.ConvertBytes()) != 0 {
// 		t.Fatalf("apdu[%X]的字节形式[%X]异常", ins, apdu.ConvertBytes())
// 	}
// }

// func Test_NewAPCI(t *testing.T) {
// 	iFrame := IFrame{
// 		Recv: int16(3),
// 		Send: int16(8),
// 	}

// 	sFrame := SFrame{
// 		Recv: int16(10),
// 	}

// 	uFrame := UFrame{
// 		STARTDT_ACT: true,
// 	}

// 	i, err := NewAPCI(10, iFrame)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if bytes.Compare(i.ConvertBytes(), []byte{0x68, 0x0A, 0x10, 0x00, 0x06, 0x00}) != 0 {
// 		t.Fatalf("创建i格式的apci[%X]异常", i.ConvertBytes())
// 	}

// 	s, err := NewAPCI(10, sFrame)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if bytes.Compare(s.ConvertBytes(), []byte{0x68, 0x0A, 0x01, 0x00, 0x14, 0x00}) != 0 {
// 		t.Fatalf("创建s格式的apci[%X]异常", s.ConvertBytes())
// 	}

// 	u, err := NewAPCI(10, uFrame)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if bytes.Compare(u.ConvertBytes(), []byte{0x68, 0x0A, 0x07, 0x00, 0x00, 0x00}) != 0 {
// 		t.Fatalf("创建u格式的apci[%X]异常", u.ConvertBytes())
// 	}
// }
