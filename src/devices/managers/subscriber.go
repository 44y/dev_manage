//向import订阅事件消息，接受事件

package managers

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	. "rc"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type udpServerSt struct {
	addr string
	wg   *sync.WaitGroup
}

const MAX_ERR_TIMES = 3

type MsgType struct {
	MT string `json:"msgtype"`
}

//场所心跳 bjson
type SiteStatusMsg struct {
	Data SiteStatusSt `json:"data"`
}
type SiteStatusSt struct {
	Orgcode string `json:"fc"`
	Wacode  string `json:"sc"`
	Status  int    `json:"ss"`
}

//设备心跳 bjson
type DeviceStatusMsg struct {
	Data DeviceStatusSt `json:"data"`
}
type DeviceStatusSt struct {
	Orgcode string `json:"fc"`
	Wacode  string `json:"sc"`
	ApId    string `json:"dc"`
	Status  int    `json:"ds"`
}

//厂商基础信息
type WA_BASIC_ORG_MSG struct {
	Data WA_BASIC_ORG `json:"data"`
}
type WA_BASIC_ORG struct {
	Orgname       string `json:"G020014"`
	Orgcode       string `json:"G020013"`
	Address       string `json:"G020037"`
	Contactor     string `json:"E020017"`
	ContactorTel  string `json:"E020020"`
	ContactorMail string `json:"E020023"`
}

//场所基础信息
type WA_BASIC_NETBAR_MSG struct {
	Data WA_BASIC_NETBAR `json:"data"`
}
type WA_BASIC_NETBAR struct {
	Wacode          string  `json:"G020004"`
	PlaceName       string  `json:"F040002"`
	SiteAddress     string  `json:"G020017"`
	Longitude       float64 `json:"-"`
	Longitude_ori   string  `json:"F010016"`
	Latitude        float64 `json:"-"`
	Latitude_ori    string  `json:"F010017"`
	NetsiteType     string  `json:"F040011"`
	BusinessNature  string  `json:"E010007"`
	LPName          string  `json:"E020001"`
	LPCfType        string  `json:"E020003"`
	LPCfID          string  `json:"E020004"`
	RelationAccount string  `json:"B070003"`
	StartTime       string  `json:"I070009"`
	EndTime         string  `json:"I070010"`
	AccessType      string  `json:"G020010"`
	OperatorNet     string  `json:"B020001"`
	AccessIP        string  `json:"I070016"`
	Orgcode         string  `json:"G020013"`

	orgIndex  int64  //安全厂商的数据库索引
	orgName   string //安全厂商名称
	orgcode_2 string //安全厂商定义的厂商代码
}

//设备基础信息
type WA_BASIC_DEVICE_MSG struct {
	Data WA_BASIC_DEVICE `json:"data"`
}
type WA_BASIC_DEVICE struct {
	Wacode             string  `json:"G020004"`
	ApId               string  `json:"I070011"`
	ApMac              string  `json:"F030011"`
	ApName             string  `json:"I070012"`
	ApAddress          string  `json:"I070013"`
	Type               int     `json:"-"`
	Type_ori           string  `json:"I070014"`
	Orgcode            string  `json:"G020013"`
	Longitude          float64 `json:"-"`
	LongitudeCol_ori   string  `json:"F010018"` //采集设备经度
	LongitudeAP_ori    string  `json:"F010001"` //AP设备经度
	Latitude           float64 `json:"-"`
	LatitudeCol_ori    string  `json:"F010019"` //AP设备纬度
	LatitudeAP_ori     string  `json:"F010002"` //采集设备纬度
	Radius             int     `json:"-"`
	Radius_ori         string  `json:"I070004"`
	DataPeriod         int     `json:"-"`
	DataPeriod_ori     string  `json:"I070015"`
	Floor              string  `json:"F040012"`
	Station            string  `json:"F040013"`
	LineInfo           string  `json:"C030006"`
	VehicleInfo        string  `json:"C030007"`
	CompartmentNO      string  `json:"C030008"`
	CarCode            string  `json:"C030002"`
	CaptureTime        int     `json:"-"`
	Time_ori           string  `json:"H010014"`
	UploadInterval     int     `json:"-"`
	UploadInterval_ori string  `json:"I070015"`

	placeName   string //场所名称
	netbarIndex int64  //场所的数据库索引
}

//事件触发信息
type WA_SOURCE_MSG struct {
	Data WA_SOURCE `json:"data"`
}
type WA_SOURCE struct {
	ApId string `json:"I070011"`
}

const (
	//场所营业状态
	NETBAR_BUSINESS_OPEN = iota
	NETBAR_BUSINESS_CLOSE
)
const (
	//设备维护状态
	DEVICE_ONLINE = iota
	DEVICE_OFFLINE
	DEVICE_MAINTAIN //维护
	DEVICE_REMOVED  //已拆除
	DEVICE_DELETED  //已删除
	DEVICE_NO_PING  //网络不通
)
const (
	//心跳、数据状态
	STATUS_NORMAL = iota
	STATUS_ABNORMAL
)
const (
	//采集设备类型
	DEVICE_TYPE        = iota
	DEVICE_TYPE_FIXED      //固定采集设备
	DEVICE_TYPE_MOBILE     //移动车载采集设备
	DEVICE_TYPE_SINGLE     //单兵采集设备
	DEVICE_TYPE_OTHER  = 9 //其他设备类型
)

/*
	场所心跳
	将场所心跳更新至数据库
*/
func siteStatusHandle(data []byte) error {
	var (
		ex  bool
		err error
	)
	//LOG_TRAC.Println("data:", string(data))
	msg := &SiteStatusMsg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("site status:%+v\n", msg.Data)
	//检查数据格式
	if len(msg.Data.Orgcode) != ORGCODE_LEN ||
		len(msg.Data.Wacode) != WACODE_LEN ||
		msg.Data.Status < NETBAR_BUSINESS_OPEN ||
		msg.Data.Status > NETBAR_BUSINESS_CLOSE {
		LOG_ERRO.Println("sitestatus data format wrong!")
		//TODO:记录错误日志
		return errors.New("sitestatus data format wrong!")
	}
	//orgcode是否存在
	org := &OrgInfo{Orgcode: msg.Data.Orgcode}
	if ex, err = X.Get(org); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_ERRO.Println("orgcode not exist!", msg.Data.Orgcode)
		//TODO:记录错误日志
		return errors.New("orgcode not exist!")
	}

	netbar := &NetbarInfo{}
	netbar.Wacode = msg.Data.Wacode

	if ex, err = X.Get(netbar); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if ex { //场所已存在
		//wacode与orgcode是否对的上
		if netbar.Orgcode != msg.Data.Orgcode {
			LOG_ERRO.Println("orgcode wrong!", msg.Data.Orgcode)
			//TODO:记录错误日志
			return errors.New("orgcode wrong!")
		}
		//if netbar.Approval == APPROVED {
		netbar.LastAliveTime = time.Now().Unix()
		netbar.AliveStatus = STATUS_NORMAL
		netbar.BusinessStatus = msg.Data.Status

		if _, err = updateData(netbar.Id, netbar); err != nil {
			LOG_ERRO.Println(err)
			return err
		}

		LOG_TRAC.Println("update netbar success", netbar.Wacode)
		return nil
		//}
	} else {
		LOG_ERRO.Println("wacode not exists!")
	}
	return nil
}

/*
	设备心跳
*/
func deviceStatusHandle(data []byte) error {
	var (
		ex  bool
		err error
	)
	//LOG_TRAC.Println("data:", string(data))

	msg := &DeviceStatusMsg{}

	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("device status:%+v\n", msg.Data)
	//检查数据格式
	if len(msg.Data.Orgcode) != ORGCODE_LEN ||
		len(msg.Data.Wacode) != WACODE_LEN ||
		len(msg.Data.ApId) != APID_LEN ||
		msg.Data.Status < DEVICE_ONLINE ||
		msg.Data.Status > DEVICE_NO_PING {
		LOG_ERRO.Println("devstatus data format wrong!")
		//TODO:记录错误日志
		return errors.New("devstatus data format wrong!")
	}
	//orgcode 是否存在
	org := &OrgInfo{Orgcode: msg.Data.Orgcode}
	if ex, err = X.Get(org); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_ERRO.Println("orgcode not exist,", msg.Data.Orgcode)
		return errors.New("orgcode not exist" + msg.Data.Orgcode)
	}
	//wacode是否存在
	netbar := &NetbarInfo{Wacode: msg.Data.Wacode}
	if ex, err = X.Get(netbar); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_ERRO.Println("Wacode not exist,", msg.Data.Wacode)
		return errors.New("Wacode not exist" + msg.Data.Wacode)
	} else if netbar.Orgcode != org.Orgcode {
		LOG_ERRO.Println("Wacode wrong,", msg.Data.Wacode)
		return errors.New("Wacode wrong" + msg.Data.Wacode)
	}
	//apid是否存在
	ap := &DevInfo{ApId: msg.Data.ApId}
	if ex, err = X.Get(ap); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if ex { //设备存在，更新状态
		if ap.Wacode != netbar.Wacode {
			LOG_ERRO.Println("Apid wrong,", msg.Data.ApId)
			return errors.New("Apid wrong" + msg.Data.ApId)
		}

		ap.LastAliveTime = time.Now().Unix()

		ap.AliveStatus = STATUS_NORMAL
		ap.DeviceStatus = msg.Data.Status
		ap.PlaceName = netbar.PlaceName
		ap.NetbarIndex = netbar.Id

		if _, err = updateData(ap.Id, ap); err != nil {
			LOG_ERRO.Println(err)
			return err
		}
		LOG_TRAC.Println("update device success", ap.ApId)
		return nil

	} else {
		LOG_INFO.Println("device not exist:", msg.Data.ApId)
	}

	return nil
}

/*
	检查厂商基础信息字段
*/
func checkOrgBaisc(data *WA_BASIC_ORG) error {
	if len(data.Orgcode) != ORGCODE_LEN {
		return errors.New("org code wrong!")
	}

	if len(data.Orgname) == 0 {
		return errors.New("org name wrong!")
	}

	if len(data.Address) == 0 {
		return errors.New("org address wrong!")
	}

	if len(data.Contactor) == 0 {
		return errors.New("Contactor wrong!")
	}

	if len(data.ContactorTel) == 0 {
		return errors.New("Contactor Tel wrong!")
	}

	if len(data.ContactorMail) == 0 {
		return errors.New("Contactor mail wrong!")
	}
	return nil
}

/*
	厂商基础信息
*/
func orgBasicHandle(data []byte) error {
	var (
		ex  bool
		err error
	)

	msg := &WA_BASIC_ORG_MSG{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("org basic:%+v\n", msg.Data)
	m_data := &msg.Data
	if err = checkOrgBaisc(m_data); err != nil {
		LOG_ERRO.Println(err)
		return err
	}

	org := &OrgInfo{Orgcode: m_data.Orgcode}

	if ex, err = X.Get(org); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if ex {
		org.Orgname = m_data.Orgname
		org.Address = m_data.Address
		org.Contactor = m_data.Contactor
		org.ContactorTel = m_data.ContactorTel
		org.ContactorMail = m_data.ContactorMail
		org.BasicStatus = STATUS_NORMAL
		org.LastBaiscTime = time.Now().Unix()

		if _, err = updateData(org.Id, org); err != nil {
			LOG_ERRO.Println(err)
			return err
		}

		LOG_TRAC.Println("update org success", org.Orgcode)
		return nil
	} else {
		LOG_ERRO.Println("org not exists,", m_data.Orgcode)
		return errors.New("org not exists," + m_data.Orgcode)
	}
}

/*
	检查场所基础信息字段
*/
func checkNetbarBaisc(data *WA_BASIC_NETBAR) error {
	if len(data.Wacode) != WACODE_LEN {
		return errors.New("netbar_wacode wrong!")
	}

	if len(data.PlaceName) == 0 {
		return errors.New("place_name empty!")
	}

	if len(data.SiteAddress) == 0 {
		return errors.New("site_address wrong!")
	}

	var err error
	if data.Longitude, err = strconv.ParseFloat(data.Longitude_ori, 64); err != nil {
		return err
	}
	if data.Latitude, err = strconv.ParseFloat(data.Latitude_ori, 64); err != nil {
		return err
	}

	var ex bool
	nett := &NetsiteType{Code: data.NetsiteType}
	if ex, err = X.Exist(nett); err != nil {
		return err
	} else if !ex {
		return errors.New("netsite type wrong!")
	}

	busit := &BusinessNature{Code: data.BusinessNature}
	if ex, err = X.Exist(busit); err != nil {
		return err
	} else if !ex {
		return errors.New("business nature wrong!")
	}

	//经营性上网服务场所，第7、第8，取值为“10”
	if data.BusinessNature == OPERATING &&
		data.NetsiteType != "0" {
		return errors.New("netsite type or business nature wrong!")
	}

	org := &OrgInfo{Orgcode: data.Orgcode}
	if ex, err = X.Get(org); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_ERRO.Println("orgcode not exist!")
		return errors.New("orgcode not exist!")
	}

	//检查场所编码各部分组成是否合法
	if string(data.Wacode[6]) != data.BusinessNature {
		return errors.New("bussiness nature mismatches wacode!")
	}
	if string(data.Wacode[7]) != data.NetsiteType {
		return errors.New("netsite type mismatches wacode!")
	}
	if string(data.Wacode[8:10]) != org.Code {
		return errors.New("orgcode mismatches wacode!")
	}

	data.orgIndex = org.Id
	data.orgcode_2 = org.Code
	data.orgName = org.Orgname

	return nil
}

/*
	场所基础信息
*/
func netbarBasicHandle(data []byte) error {
	var (
		ex  bool
		err error
	)

	msg := &WA_BASIC_NETBAR_MSG{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("netbar basic:%+v\n", msg.Data)

	m_data := &msg.Data
	if err = checkNetbarBaisc(m_data); err != nil {
		LOG_ERRO.Println(err)
		return err
	}

	netbar := &NetbarInfo{Wacode: m_data.Wacode}
	if ex, err = X.Get(netbar); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if ex { //更新场所基础信息
		if netbar.Orgcode != m_data.Orgcode {
			LOG_ERRO.Println("Wacode wrong,", m_data.Wacode)
			return errors.New("Wacode wrong" + m_data.Wacode)
		}
		if netbar.Approval != APPROVED {
			LOG_INFO.Println("netbar not approved", m_data.Wacode)
		}
		netbar.PlaceName = m_data.PlaceName
		netbar.SiteAddress = m_data.SiteAddress
		netbar.Longitude = m_data.Longitude
		netbar.Latitude = m_data.Latitude
		netbar.NetsiteType = m_data.NetsiteType
		netbar.BusinessNature = m_data.BusinessNature
		netbar.LPName = m_data.LPName
		netbar.LPCfType = m_data.LPCfType
		netbar.LPCfID = m_data.LPCfID
		netbar.RelationAccount = m_data.RelationAccount
		netbar.StartTime = m_data.StartTime
		netbar.EndTime = m_data.EndTime
		netbar.BasicStatus = STATUS_NORMAL
		netbar.LastBasicTime = time.Now().Unix()
		netbar.Orgname = m_data.orgName
		netbar.Orgcode_2 = m_data.orgcode_2
		netbar.OrgIndex = m_data.orgIndex

		//拆分场所编码组成部分
		netbar.AreaCode1 = m_data.Wacode[0:2] + "0000"
		netbar.AreaCode2 = m_data.Wacode[0:4] + "00"
		netbar.AreaCode3 = m_data.Wacode[0:6]
		netbar.NetbarSerialNO = m_data.Wacode[10:]

		if len(m_data.AccessType) > 0 {
			netbar.AccessType = m_data.AccessType
		}
		if len(m_data.OperatorNet) > 0 {
			netbar.OperatorNet = m_data.OperatorNet
		}
		if len(m_data.AccessIP) > 0 {
			netbar.AccessIP = m_data.AccessIP
		}

		if _, err = updateData(netbar.Id, netbar); err != nil {
			LOG_ERRO.Println(err)
			return err
		}

		LOG_TRAC.Println("update netbar success", netbar.Wacode)
		return nil
	} else { //场所不存在，新增未审核场所
		new_netbar := &NetbarInfo{
			Wacode:          m_data.Wacode,
			PlaceName:       m_data.PlaceName,
			SiteAddress:     m_data.SiteAddress,
			Longitude:       m_data.Longitude,
			Latitude:        m_data.Latitude,
			BusinessNature:  m_data.BusinessNature,
			NetsiteType:     m_data.NetsiteType,
			LPName:          m_data.LPName,
			LPCfID:          m_data.LPCfID,
			LPCfType:        m_data.LPCfType,
			RelationAccount: m_data.RelationAccount,
			StartTime:       m_data.StartTime,
			EndTime:         m_data.EndTime,
			BasicStatus:     STATUS_NORMAL,
			AliveStatus:     STATUS_NORMAL,
			LastAliveTime:   time.Now().Unix(),
			LastBasicTime:   time.Now().Unix(),
			Approval:        NOTAPPROVED,
			Orgcode:         m_data.Orgcode,
			Orgname:         m_data.orgName,
			Orgcode_2:       m_data.orgcode_2,
			OrgIndex:        m_data.orgIndex,

			//拆分场所编码组成部分
			AreaCode1:      m_data.Wacode[0:2] + "0000",
			AreaCode2:      m_data.Wacode[0:4] + "00",
			AreaCode3:      m_data.Wacode[0:6],
			NetbarSerialNO: m_data.Wacode[10:],
		}

		if _, err = X.InsertOne(new_netbar); err != nil {
			LOG_ERRO.Println(err)
			return err
		}
		return nil
	}
}

/*
	检查设备基础信息字段
*/
func checkDeviceBaisc(data *WA_BASIC_DEVICE, msg_type string) error {
	var (
		err error
		ex  bool
	)

	netbar := &NetbarInfo{Wacode: data.Wacode}
	if ex, err = X.Get(netbar); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_INFO.Println("netbar not exist! :", data.Wacode)
		//return errors.New("wacode not exist!")
	}
	data.placeName = netbar.PlaceName
	data.netbarIndex = netbar.Id

	if len(data.ApId) != APID_LEN {
		return errors.New("collection_equipmentid wrong!")
	}

	if msg_type == "WA_BASIC_FJ_0003_1" {
		if data.Longitude, err = strconv.ParseFloat(data.LongitudeAP_ori, 64); err != nil {
			return err
		} else if data.Longitude == 0 {
			return errors.New("longitude wrong!")
		}

		if data.Latitude, err = strconv.ParseFloat(data.LatitudeAP_ori, 64); err != nil {
			return err
		} else if data.Latitude == 0 {
			return errors.New("latitude wrong!")
		}
	}

	if msg_type == "WA_BASIC_FJ_1001" {
		if len(data.Time_ori) != TIMESTAMP_LEN {
			return errors.New("time wrong!")
		}
		if data.CaptureTime, err = strconv.Atoi(data.Time_ori); err != nil {
			return errors.New("time wrong!")
		}

		if data.Longitude, err = strconv.ParseFloat(data.LongitudeCol_ori, 64); err != nil {
			return err
		} else if data.Longitude == 0 {
			return errors.New("longitude wrong!")
		}

		if data.Latitude, err = strconv.ParseFloat(data.LongitudeCol_ori, 64); err != nil {
			return err
		} else if data.Latitude == 0 {
			return errors.New("latitude wrong!")
		}
	}

	if msg_type == "WA_BASIC_FJ_1002" {
		if len(data.ApName) == 0 {
			return errors.New("collection_equipment_name wrong")
		}

		if len(data.ApAddress) == 0 {
			return errors.New("collection_equipment_address wrong")
		}

		if len(data.Type_ori) != 1 {
			return errors.New("collection_equipment_type wrong")
		}

		if data.Longitude, err = strconv.ParseFloat(data.LongitudeCol_ori, 64); err != nil {
			return err
		} else if data.Longitude == 0 {
			return errors.New("longitude wrong!")
		}

		if data.Latitude, err = strconv.ParseFloat(data.LongitudeCol_ori, 64); err != nil {
			return err
		} else if data.Latitude == 0 {
			return errors.New("latitude wrong!")
		}

		if data.Type, err = strconv.Atoi(data.Type_ori); err != nil {
			return errors.New("collection_equipment_type wrong")
		}
		if data.Type != DEVICE_TYPE_FIXED &&
			data.Type != DEVICE_TYPE_MOBILE &&
			data.Type != DEVICE_TYPE_SINGLE &&
			data.Type != DEVICE_TYPE_OTHER {
			return errors.New("collection_equipment_type wrong")
		}

		if len(data.UploadInterval_ori) == 0 {
			return errors.New("upload_time_interval wrong")
		}
		if data.UploadInterval, err = strconv.Atoi(data.UploadInterval_ori); err != nil {
			return errors.New("upload_time_interval wrong")
		}

		if len(data.Radius_ori) == 0 {
			return errors.New("cellection_radius wrong")
		}
		if data.Radius, err = strconv.Atoi(data.Radius_ori); err != nil {
			return errors.New("collection_radius wrong")
		}
	}

	if msg_type == "WA_BASIC_FJ_0003_1" {
		data.Type = DEVICE_TYPE_FIXED
		match, err := regexp.MatchString(REGEXP_MAC2, data.ApMac)
		if err != nil {
			return err
		}
		if !match {
			return errors.New("ap_mac wrong")
		}
	}

	if msg_type == "WA_BASIC_FJ_0003_2" {
		data.Type = DEVICE_TYPE_MOBILE
		match, err := regexp.MatchString(REGEXP_MAC2, data.ApMac)
		if err != nil {
			return err
		}
		if !match {
			return errors.New("ap_mac wrong")
		}
	}

	//data.ApMac = transMacFormat2(data.ApMac)

	return nil
}

/*
	设备基础信息
*/
func deviceBasicHandle(data []byte, msg_type string) error {
	var (
		ex  bool
		err error
	)

	msg := &WA_BASIC_DEVICE_MSG{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("device basic:%+v\n", msg.Data)

	m_data := &msg.Data
	if err = checkDeviceBaisc(m_data, msg_type); err != nil {
		LOG_ERRO.Println(err)
		return err
	}

	ap := &DevInfo{ApId: m_data.ApId}
	if ex, err = X.Get(ap); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if ex { //更新设备基础信息
		if ap.Wacode != m_data.Wacode {
			LOG_ERRO.Println("wacode wrong")
			return errors.New("wacode wrong")
		}
		if ap.Approval != APPROVED {
			LOG_INFO.Println("ap not approved!", ap.ApId)
		}

		if m_data.Longitude != 0 {
			ap.Latitude = m_data.Longitude
		}
		if m_data.Latitude != 0 {
			ap.Latitude = msg.Data.Latitude
		}
		if m_data.CaptureTime != 0 {
			ap.CaptureTime = m_data.CaptureTime
		}
		if len(m_data.ApName) != 0 {
			ap.ApName = m_data.ApName
		}
		if len(m_data.ApAddress) != 0 {
			ap.ApName = m_data.ApAddress
		}
		if m_data.Type != 0 {
			ap.Type = m_data.Type
		}
		if m_data.UploadInterval != 0 {
			ap.UploadInterval = m_data.UploadInterval
		}
		if m_data.Radius != 0 {
			ap.Radius = m_data.Radius
		}
		if len(m_data.CarCode) != 0 {
			ap.CarCode = m_data.CarCode
		}
		if len(m_data.LineInfo) != 0 {
			ap.LineInfo = m_data.LineInfo
		}
		if len(m_data.VehicleInfo) != 0 {
			ap.VehicleInfo = m_data.VehicleInfo
		}
		if len(m_data.CompartmentNO) != 0 {
			ap.CompartmentNO = m_data.CompartmentNO
		}
		if len(m_data.Floor) != 0 {
			ap.Floor = m_data.Floor
		}
		if len(m_data.Station) != 0 {
			ap.Station = m_data.Station
		}
		if len(m_data.ApMac) != 0 {
			ap.ApMac = m_data.ApMac
		}

		ap.PlaceName = m_data.placeName
		ap.NetbarIndex = m_data.netbarIndex

		ap.BasicStatus = STATUS_NORMAL
		ap.LastBasicTime = time.Now().Unix()

		ap.AreaCode1 = m_data.Wacode[0:2] + "0000"
		ap.AreaCode2 = m_data.Wacode[0:4] + "00"
		ap.AreaCode3 = m_data.Wacode[0:6]

		if _, err = updateData(ap.Id, ap); err != nil {
			LOG_ERRO.Println(err)
			return err
		}
		LOG_TRAC.Println("update ap success", ap.ApId)
		return nil
	} else { //设备不存在，新增未审核设备
		new_data := &DevInfo{
			ApId:           m_data.ApId,
			ApMac:          m_data.ApMac,
			Wacode:         m_data.Wacode,
			ApName:         m_data.ApName,
			ApAddress:      m_data.ApAddress,
			Type:           m_data.Type,
			AreaCode1:      m_data.Wacode[0:2] + "0000",
			AreaCode2:      m_data.Wacode[0:4] + "00",
			AreaCode3:      m_data.Wacode[0:6],
			Radius:         m_data.Radius,
			Longitude:      m_data.Longitude,
			Latitude:       m_data.Latitude,
			Floor:          m_data.Floor,
			Station:        m_data.Station,
			LineInfo:       m_data.LineInfo,
			VehicleInfo:    m_data.VehicleInfo,
			CompartmentNO:  m_data.CompartmentNO,
			CarCode:        m_data.CarCode,
			Approval:       NOTAPPROVED,
			UploadInterval: m_data.UploadInterval,
			CaptureTime:    m_data.CaptureTime,
			BasicStatus:    STATUS_NORMAL,
			AliveStatus:    STATUS_ABNORMAL,
			DataStatus:     STATUS_ABNORMAL,
			LastBasicTime:  time.Now().Unix(),
			PlaceName:      m_data.placeName,
			NetbarIndex:    m_data.netbarIndex,
		}

		if _, err = X.InsertOne(new_data); err != nil {
			LOG_ERRO.Println(err)
			return err
		}
		return nil
	}

	return nil
}

/*
	审计数据,更新设备数据时间
*/
func sourceEventHandle(data []byte) error {
	var (
		ex  bool
		err error
	)

	msg := &WA_SOURCE_MSG{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Printf("source:%+v\n", msg.Data)

	ap := &DevInfo{ApId: msg.Data.ApId}

	if ex, err = X.Get(ap); err != nil {
		LOG_ERRO.Println(err)
		return err
	} else if !ex {
		LOG_ERRO.Println("ap:", msg.Data.ApId, "not exists!")
		return errors.New("ap: " + msg.Data.ApId + " not exists!")
	} else if ap.Approval != APPROVED {
		LOG_INFO.Println("ap:", msg.Data.ApId, "not approval!")
		return errors.New("ap: " + msg.Data.ApId + " not approval!")
	}

	/*
		//设备数据状态变更，需要更新场所内设备数量
		if ap.DataStatus == STATUS_ABNORMAL {
			netbar := &NetbarInfo{Wacode: ap.Wacode}
			if ex, err = X.Get(netbar); err != nil {
				return err
			} else if !ex {
				return errors.New("netbar:" + ap.Wacode + " not exists!")
			}
			netbar.DevDataOnlineNum++
			updateData(netbar.Id, netbar)
		}
	*/
	ap.DataStatus = STATUS_NORMAL
	ap.LastDataTime = time.Now().Unix()

	if _, err = updateData(ap.Id, ap); err != nil {
		LOG_ERRO.Println(err)
		return err
	}

	return nil
}

func (this *udpServerSt) msgHandle(buf []byte) {
	this.wg.Add(1)
	defer this.wg.Done()
	var (
		//lenth int
		err error
	)
	/*
		if lenth, err = DecodeByte2int(buf[0:2]); err != nil {
			LOG_ERRO.Println(err)
			return
		}

		data := buf[2:]
		LOG_TRAC.Println(lenth, "len()", len(data))
		if lenth != len(data) {
			LOG_ERRO.Println("data lenth wrong!")
			return
		}
	*/
	//	LOG_TRAC.Println("buf:", string(buf))
	data := buf
	mt := &MsgType{}
	err = json.Unmarshal(data, mt)
	if err != nil {
		LOG_ERRO.Println(err)
		return
	}

	LOG_TRAC.Println("msg type:", mt.MT)

	switch mt.MT {
	case "SITESTATUS":
		if err = siteStatusHandle(data); err != nil {
			LOG_ERRO.Println(err)
		}

	case "DEVICESTATUS":
		if err = deviceStatusHandle(data); err != nil {
			LOG_ERRO.Println(err)
		}

	case "WA_BASIC_FJ_1004", "WA_BASIC_FJ_0002":
		if err = orgBasicHandle(data); err != nil {
			LOG_ERRO.Println(err)
		}

	case "WA_BASIC_FJ_1003", "WA_BASIC_FJ_0001":
		if err = netbarBasicHandle(data); err != nil {
			LOG_ERRO.Println(err)
		}

	case "WA_BASIC_FJ_1001", "WA_BASIC_FJ_1002", "WA_BASIC_FJ_0003_1", "WA_BASIC_FJ_0003_2":
		if err = deviceBasicHandle(data, mt.MT); err != nil {
			LOG_ERRO.Println(err)
		}

	case "WA_SOURCE_FJ_1001", "WA_SOURCE_FJ_1002", "WA_SOURCE_FJ_0001",
		"WA_SOURCE_FJ_0002", "WA_SOURCE_0005", "WA_SOURCE_IK_0001":
		if err = sourceEventHandle(data); err != nil {
			LOG_ERRO.Println(err)
		}
	default:
		LOG_ERRO.Println("msgtype error:", mt.MT)
	}
}

const MAX_READ_BUF = 65535

//conn处理
func (this *udpServerSt) connHandle(conn *net.UDPConn) {
	this.wg.Add(1)
	defer this.wg.Done()
	defer conn.Close()

	for {
		LOG_TRAC.Println("keep reading..")
		buf := make([]byte, MAX_READ_BUF)
		n, _, err := conn.ReadFromUDP(buf)

		if err != nil {
			LOG_FATAL.Println(err)
			return
		}
		//LOG_DBG.Println("read from", addr, "msg:", string(buf[:n]), "lenth:", n)

		go this.msgHandle(buf[:n])
	}
}

//开启udp服务
func (this *udpServerSt) serve(q <-chan struct{}) {
	this.wg.Add(1)
	defer this.wg.Done()

	var (
		conn *net.UDPConn
		err  error
	)

	udp_addr, err := net.ResolveUDPAddr("udp", this.addr)
	if err != nil {
		panic(err)
	}

	//开始监听
	LOG_INFO.Println("Start listen..")
	conn, err = net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(conn)
	}

	//处理消息
	go this.connHandle(conn)

	//退出信号
	<-q
	conn.Close()

}

type subBodySt struct {
	TransType string   `json:"trans_type"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	MsgType   []string `json:"msgtype"`
}

var SubType = []string{
	//终端特征采集数据
	"WA_SOURCE_FJ_1001",
	"WA_SOURCE_FJ_1002",
	"WA_BASIC_FJ_1001",
	"WA_BASIC_FJ_1002",
	"WA_BASIC_FJ_1003",
	"WA_BASIC_FJ_1004",

	//审计数据
	"WA_SOURCE_FJ_0001",
	"WA_SOURCE_FJ_0002",
	"WA_SOURCE_IK_0001",
	"WA_SOURCE_0005",
	"WA_BASIC_FJ_0001",
	"WA_BASIC_FJ_0002",
	"WA_BASIC_FJ_0003_1",
	"WA_BASIC_FJ_0003_2",

	//心跳 bjson
	"SITESTATUS",
	"DEVICESTATUS",
}

const SUB_PERIOD = 2

/*
	调用rest API向import订阅消息
*/
func subToImport(quitChan <-chan struct{}) bool {
	var pub DependentService

	for _, pub = range GlobalConfig.Dependent {
		if pub.Use == "subscriber" {
			break
		}
	}

	url := "http://" + GlobalConfig.ApiGw + "/" + pub.Name +
		"/" + pub.Version + "/" + pub.Rc + "/" +
		GlobalConfig.ServiceName

	body := &subBodySt{
		TransType: "udp",
		Host:      MyIp,
		Port:      GlobalConfig.UDPClientPort,
		MsgType:   SubType,
	}

	LOG_TRAC.Println("url:", url)
	LOG_TRAC.Printf("%+v\n", body)
	body_b, err := json.Marshal(body)
	if err != nil {
		LOG_ERRO.Println(err)
		return false
	}

	time_up := time.After(time.Duration(SUB_PERIOD) * time.Second)

	req_body := ioutil.NopCloser(strings.NewReader(string(body_b)))
	for {
		resp, err := http.Post(url, "application/json; encoding=utf-8", req_body)
		if err != nil {
			LOG_ERRO.Println(err)
		} else if resp.StatusCode == http.StatusOK {
			LOG_TRAC.Println("sub to import success!")
			return true
		}

		select {
		case <-time_up:
			time_up = time.After(time.Duration(SUB_PERIOD) * time.Second)
			continue

		case <-quitChan:
			return false
		}
	}

}

//订阅
func StartSubscriber(quit context.Context, wg *sync.WaitGroup) {
	LOG_TRAC.Println("Subscriber start!")
	wg.Add(1)
	defer LOG_INFO.Println("Subscriber done!")
	defer wg.Done()

	srv := &udpServerSt{
		addr: MyIp + ":" + strconv.Itoa(GlobalConfig.UDPClientPort),
		wg:   &sync.WaitGroup{},
	}

	srvQuit := make(chan struct{})
	go srv.serve(srvQuit)

	//向import订阅
	if !subToImport(quit.Done()) {
		close(srvQuit)
		srv.wg.Wait()
		return
	}

	//退出时关闭tcp server
	<-quit.Done()
	//LOG_WARN.Println("Got quit,close tcp server")
	close(srvQuit)
	srv.wg.Wait()
	//LOG_WARN.Println("udp server closed")
}

/*
	basic和bjson文件错误记录日志
*/
func ErrorLog() {
	//TODO:
}
