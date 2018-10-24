package elements

// cot cause of translation 传送原因，《DLT 634.5101-2002》 7.2.3.1
const (
	COT_ACTIVE   = 3
	COT_INIT     = 4  // 初始化
	COT_REQ      = 5  // 请求或者被请求
	COT_ACT      = 6  // 激活
	COT_ACTCON   = 7  // 激活确认
	COT_DEACT    = 8  // 停止激活
	COT_DEACTCON = 9  // 停止激活确认
	COT_ACTTERM  = 10 // 激活终止
	COT_INTRGEN  = 20 // 相应站召唤
)
