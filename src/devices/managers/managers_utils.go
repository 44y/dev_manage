//managers 包中的常量、类型等定义及通用的函数
package managers

import (
	"net/http"
	. "rc"
)

//==================常量定义==================

//字段长度限制
const (
	ORGCODE_LEN  = 9  //安全厂商代码长度
	CODE_LEN     = 2  //厂商编码长度
	WACODE_LEN   = 14 //场所编码长度
	AREACODE_LEN = 6  //地区编码长度
	APID_LEN     = 21 //设备编号长度
	MAC_LEN      = 17 //MAC长度
	//LONGI_LEN     = 9  //经纬度长度
	TIMESTAMP_LEN = 10 //绝对秒数长度
)

//错误信息定义
const (
	BadLoniLati     = "经纬度为浮点数，东经大于西经，北纬大于南纬"
	BadStartEndTime = "起始结束时间为绝对秒数，结束时间大于起始时间"
	BadAreaCode     = "字段应为6位数字"
	BadNetsiteType  = "经营性上网服务场所固定为0"
	BadRead         = "read为yes或no"

	LenWrong           = "字段长度应为："
	TooLong            = "字段超长"
	TypeWrong          = "类型错误"
	FormatWrong        = "格式错误"
	EmptNotAllowed     = "字段不能为空"
	RecordAlreadyExist = "记录已存在"
	RecordNotExist     = "记录不存在"
	NetbarNotApproved  = "场所未审核"
)

//正则表达式
const (
	REGEXP_MAC1     = "([a-f0-9]{2}:){5}[a-f0-9]{2}" //aa:bb:cc
	REGEXP_MAC2     = "([A-F0-9]{2}-){5}[A-F0-9]{2}" //AA-BB-CC
	REGEXP_AREACODE = `\d{6}`
)

//查询limit默认值100
const (
	LIMIT_DEFAULT = 100
)

//审核状态
const (
	NOTAPPROVED = iota
	APPROVED
	DELETED
)

//场所下的设备状态
const (
	NETBARSTATUS_DEV = iota
	NETBARSTATUS_DEV_ONLINE
	NETBARSTATUS_DEV_ABNORMAL
	NETBARSTATUS_DEV_OFFLINE
	NETBARSTATUS_DEV_EMPTY
)

//场所经营性质
const (
	OPERATING       = "1" //经营性上网服务场所
	NON_OPERATING   = "2" //非经营性上网服务场所
	WIFI_COLLECTING = "3" //WiFi无线采集前端场所
)

const (
	MAX_MSG_CHAN = 5 //接受消息chan最大长度
	DIAL_TIMEOUT = 2 //dial超时时间
)

//场所变更动作
const (
	ACTION_ADD = iota
	ACTION_MODIFY
	ACTION_DELETE
)

//订阅消息类型
const (
	MSGTYPE_NETBARS_BIT = 1 << iota
	MSGTYPE_APS_BIT

	MSGTYPE_NETBARS_STR = "netbars"
	MSGTYPE_APS_STR     = "aps"
)

//==================结构定义==================
//通用查询结果
type findResults struct {
	count, total int64
	//results      []OrgInfo
	results interface{} //slice
}

//全局查询数据范围，用户权限控制
type DataScope struct {
	Orgcodes []string `json:"security_software_orgcode"`
	AreaCode `json:"area_code"`
}

//区域编码结构
type AreaCode struct {
	Level1 []string `json:"1"`
	Level2 []string `json:"2"`
	Level3 []string `json:"3"`
}

//场所过滤条件
type NetbarFilter struct {
	Area_code       AreaCode `json:"area_code"`
	BussinessNature string   `json:"business_nature"`
	NetsiteType     string   `json:"netsite_type"`
	Orgname         string   `json:"security_software_orgname"`
	Orgcode         string   `json:"security_software_orgcode"`
	Status          int      `json:"status"`
}

//查询body结构，场所厂商设备通用
type DevicesGetBody struct {
	Scope  DataScope    `json:"data_scope"`
	Filter NetbarFilter `json:"filter"`
}

//查询导出文件POST body结构，场所厂商设备通用
type FilePostBody struct {
	DevicesGetBody
	Fields OutFieldsSlice `json:"output_fields"`
}

//http request require
type HttpSt struct {
	req    *http.Request
	writer http.ResponseWriter
}

//==================类型定义==================
//更新时用于指定需要更新为0的字段
type UpdateMap map[string]interface{}

//==================函数==================
