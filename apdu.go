package iec

type APDU struct {
	APCI APCI
	ASDU ASDU
	Len  int
}
