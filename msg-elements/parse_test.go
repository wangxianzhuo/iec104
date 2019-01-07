package elements

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_Parse(t *testing.T) {
	// ins := "685A000000000D0A030001000840006F1203BA0009400023DB1140000C4000010D6E42000640005A82813F000140007B33C340000E400000407A4200034000BAC982400002400026365F3C00044000B7B20140000D400083CD463E00"
	// ParseASDU()
	ins, _ := hex.DecodeString("0D0A030001000840006F1203BA0009400023DB1140000C4000010D6E42000640005A82813F000140007B33C340000E400000407A4200034000BAC982400002400026365F3C00044000B7B20140000D400083CD463E00")
	asdu, err := ParseASDU(ins)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(asdu)
	fmt.Println(len(asdu.MessageBody.(MessageElement_13_SQ_0)))

	ins2, _ := hex.DecodeString("090114000100010000f9ff00")
	asdu2, err := ParseASDU(ins2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(asdu2)
	fmt.Println(len(asdu2.MessageBody.(MessageElement_9_SQ_0)))

	ins3, _ := hex.DecodeString("098A01000100014000000000000000BE4E000000000000004C0D003D0000000000F20200000000")
	asdu3, err := ParseASDU(ins3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(asdu3)
	// fmt.Println(len(asdu3.MessageBody.(MessageElement_9_SQ_0)))
}
