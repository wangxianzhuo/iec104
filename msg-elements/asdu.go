package elements

type ASDU struct {
	DUI         DUI
	MessageBody BytesConverter
}

func (asdu ASDU) ConvertBytes() []byte {
	var dui []byte
	dui = append(dui, asdu.DUI.TypeIdentification)
	dui = append(dui, asdu.DUI.VariableStructureQualifier)
	dui = append(dui, asdu.DUI.Cause)
	if asdu.DUI.CauseExtEnable {
		dui = append(dui, asdu.DUI.CauseExt)
	}
	dui = append(dui, asdu.DUI.PublicAddressLow)
	if asdu.DUI.PublicAddressHigEnable {
		dui = append(dui, asdu.DUI.PublicAddressHig)
	}

	return append(dui, asdu.MessageBody.ConvertBytes()...)
}

// DUI 数据单元标识符
type DUI struct {
	TypeIdentification         byte // 类型标识 1-127，《DLT 634.5101-2002》 7.2.1.1
	VariableStructureQualifier byte // 可变结构限定词，《DLT 634.5101-2002》 7.2.2.1
	Cause                      byte // 传送原因，《DLT 634.5101-2002》 7.2.3.1
	CauseExt                   byte // 源发站地址
	CauseExtEnable             bool // 源发站地址s使能
	PublicAddressLow           byte // 应用服务数据单元公共地址低8位
	PublicAddressHig           byte // 应用服务数据单元公共地址高8位
	PublicAddressHigEnable     bool // 应用服务数据单元公共地址高8位使能
}

type BytesConverter interface {
	ConvertBytes() []byte
}
