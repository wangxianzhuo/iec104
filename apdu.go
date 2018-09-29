package iec

import elements "github.com/wangxianzhuo/iec104/msg-elements"

type APDU struct {
	APCI APCI
	ASDU elements.ASDU
	Len  int
}
