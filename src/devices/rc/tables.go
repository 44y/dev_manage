//定义数据库表项
package rc

import (
	"time"
)

//安全厂商信息
type OrgInfo struct {
	Id            int64     `json:"idx"`
	Orgname       string    `xorm:"varchar(70) notnull unique 'security_software_orgname'"    json:"security_software_orgname"`
	Orgcode       string    `xorm:"char(9) notnull unique 'security_software_orgcode'"        json:"security_software_orgcode"`
	Address       string    `xorm:"varchar(255) notnull unique 'security_software_address'"   json:"security_software_address"`
	Code          string    `xorm:"char(2) notnull unique 'security_software_code'"           json:"security_software_code"`
	Contactor     string    `xorm:"varchar(128) notnull 'contactor'"                          json:"contactor"`
	ContactorTel  string    `xorm:"varchar(32) notnull 'contactor_tel'"                       json:"contactor_tel"`
	ContactorMail string    `xorm:"varchar(32) notnull 'contactor_mail'"                      json:"contactor_mail"`
	BasicStatus   int       `xorm:"tinyint(1) not null default(1) 'basic_status'"              json:"basic_status"`
	LastBaiscTime int64     `xorm:"bigint(14) default(null) 'last_basic_time'"                  json:"last_basic_time"`
	EncryptId     int64     `xorm:"int(11) default(1)"                                        json:"encrypt_id"`
	CreatedAt     time.Time `xorm:"created"                                                   json:"-"` //json忽略该字段
	UpdatedAt     time.Time `xorm:"updated"                                                   json:"-"`
	//Version       int       `xorm:"version"`
}

func (*OrgInfo) TableName() string {
	return "OrgInfo"
}

func (this *OrgInfo) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "orgcode:", this.Orgcode, "Id:", this.Id)
}

func (this *OrgInfo) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "orgcode:", this.Orgcode, "Id:", this.Id)
}

func (this *OrgInfo) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "orgcode:", this.Orgcode, "Id:", this.Id)
	time.Now()
}

//场所信息
type NetbarInfo struct {
	Id             int64   `json:"idx"`
	OriId          int64   `xorm:"bigint(20) 'ori_id'"                                json:"-"`
	Orgname        string  `xorm:"varchar(255) notnull 'security_software_orgname'"    json:"security_software_orgname"`
	Orgcode        string  `xorm:"char(9) notnull 'security_software_orgcode'"        json:"security_software_orgcode"`
	OrgIndex       int64   `xorm:"bigint notnull 'org_index'"                         json:"-"` //厂商的数据库索引
	Orgcode_2      string  `xorm:"char(2) notnull 'security_software_code'"  json:"-"`          //安全厂商自定义的编码，用于组成场所编码
	Wacode         string  `xorm:"char(14) notnull 'netbar_wacode'"                   json:"netbar_wacode"`
	NetbarSerialNO string  `xorm:"char(14) notnull 'netbar_serialNO'"                 json:"netbar_serialNO"`
	PlaceName      string  `xorm:"varchar(255) not null 'place_name'"                 json:"place_name"`
	AreaCode1      string  `xorm:"char(6) not null 'area_code_1'"                       json:"area_code_1"`
	AreaCode2      string  `xorm:"char(6) not null 'area_code_2'"                       json:"area_code_2"`
	AreaCode3      string  `xorm:"char(6) not null 'area_code_3'"                       json:"area_code_3"`
	SiteAddress    string  `xorm:"varchar(255) not null 'site_address'"               json:"site_address"`
	Longitude      float64 `xorm:"decimal(10,7) default(null) 'longitude'"                 json:"longitude"`
	Latitude       float64 `xorm:"decimal(10,7) default(null) 'latitude'"                  json:"latitude"`
	NetsiteType    string  `xorm:"varchar(1) not null 'netsite_type'"                 json:"netsite_type"`
	BusinessNature string  `xorm:"char(1) not null 'business_nature'"                 json:"business_nature"`
	//LP for LawPrincipal 法人
	LPName          string `xorm:"varchar(64) default(null) 'law_principal_name'"             json:"law_principal_name"`
	LPCfType        string `xorm:"varchar(3) default(null) 'law_principal_certificate_type'"  json:"law_principal_certificate_type"`
	LPCfID          string `xorm:"varchar(128) default(null) 'law_principal_certificate_id'"  json:"law_principal_certificate_id"`
	RelationAccount string `xorm:"varchar(128) default(null) 'relationship_account'"          json:"relationship_account"`
	StartTime       string `xorm:"char(5) default(null) 'start_time'"                         json:"start_time"`
	EndTime         string `xorm:"char(5) default(null) 'end_time'"                           json:"end_time"`
	AccessType      string `xorm:"char(2) default(null) 'access_type'"                        json:"access_type"`
	OperatorNet     string `xorm:"varchar(2) default(null) 'operator_net'"                    json:"operator_net"`
	AccessIP        string `xorm:"varchar(64) default(null) 'access_ip'"                      json:"access_ip"`
	Approval        int    `xorm:"tinyint(1) not null default(0) 'approval'"                  json:"approval"`
	ApprovalTime    int64  `xorm:"bigint(14) default(null) 'approval_time'"                   json:"approval_time"`
	//EncryptionId    int64  `xorm:"tinyint(1) default(1) 'encryption_id'"                      json:"encryption_id"`
	Comments string `xorm:"varchar(255) default(null) 'comments'"    json:"comments"`
	//删除时间
	DeletedTime      int64 `xorm:"bigint(14) default(null) 'deleted_time'"                     json:"deleted_time"`
	DataStatus       int   `xorm:"tinyint(1) not null default(1) 'data_status'"               json:"data_status"`
	AliveStatus      int   `xorm:"tinyint(1) not null default(1) 'alive_status'"              json:"alive_status"`
	LastAliveTime    int64 `xorm:"bigint(14) default(null) 'last_alive_time'"                 json:"last_alive_time"`
	LastDataTime     int64 `xorm:"bigint(14) default(null) 'last_data_time'"                  json:"last_data_time"`
	BasicStatus      int   `xorm:"tinyint(1) not null default(1) 'basic_status'"               json:"basic_status"`
	LastBasicTime    int64 `xorm:"bigint(14) default(null) 'last_basic_time'"                  json:"last_basic_time"`
	DevTotalNum      int   `xorm:"int(5) default(null) 'dev_total_num'"                        json:"dev_total_num"`
	DevOnlineNum     int   `xorm:"int(5) default(null) 'dev_online_num'"                        json:"dev_online_num"`
	DevOfflineNum    int   `xorm:"int(5) default(null) 'dev_offline_num'"                        json:"dev_offline_num"`
	DevAbnormalNum   int   `xorm:"int(5) default(null) 'dev_abnormal_num'"                        json:"dev_abnormal_num"`
	DevDataOnlineNum int   `xorm:"int(5) default(null) 'dev_data_online_num'"                        json:"dev_data_online_num"`
	//ServiceStatus  int       `xorm:"tinyint(1) default(0) 'service_status'"        json:"service_status"`
	BusinessStatus int       `xorm:"tinyint(1) default(0) 'business_status'"        json:"business_status"`
	CreatedAt      time.Time `xorm:"created"                                                    json:"-"` //json忽略该字段
	UpdatedAt      time.Time `xorm:"updated"                                                    json:"-"`
}

func (*NetbarInfo) TableName() string {
	return "NetbarInfo"
}

func (this *NetbarInfo) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
}

func (this *NetbarInfo) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
}

func (this *NetbarInfo) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
	time.Now()
}

/*
	场所信息 ————已删除
	从审核和未审核表中删除的数据存在这里，记录审核和删除时间，字段内容可重复不覆盖
*/
type NetbarInfoDeleted struct {
	Id             int64   `json:"-"`
	OriId          int64   `xorm:"bigint(20) notnull 'ori_id'"                             json:"idx"` //在未删除表中的index
	Orgname        string  `xorm:"varchar(70) notnull 'security_software_orgname'"    json:"security_software_orgname"`
	Orgcode        string  `xorm:"char(9) notnull 'security_software_orgcode'"        json:"security_software_orgcode"`
	OrgIndex       int64   `xorm:"bigint notnull 'org_index'"                         json:"-"` //厂商的数据库索引
	Orgcode_2      string  `xorm:"char(2) notnull 'security_software_code'"  json:"-"`
	Wacode         string  `xorm:"char(14) notnull 'netbar_wacode'"                   json:"netbar_wacode"`
	NetbarSerialNO string  `xorm:"char(14) notnull 'netbar_serialNO'"                 json:"netbar_serialNO"`
	PlaceName      string  `xorm:"varchar(255) not null 'place_name'"                 json:"place_name"`
	AreaCode1      string  `xorm:"char(6) not null 'area_code_1'"                       json:"area_code_1"`
	AreaCode2      string  `xorm:"char(6) not null 'area_code_2'"                       json:"area_code_2"`
	AreaCode3      string  `xorm:"char(6) not null 'area_code_3'"                       json:"area_code_3"`
	SiteAddress    string  `xorm:"varchar(255) not null 'site_address'"               json:"site_address"`
	Longitude      float64 `xorm:"decimal(10,7) default(null) 'longitude'"                 json:"longitude"`
	Latitude       float64 `xorm:"decimal(10,7) default(null) 'latitude'"                  json:"latitude"`
	NetsiteType    string  `xorm:"varchar(1) not null 'netsite_type'"                 json:"netsite_type"`
	BusinessNature string  `xorm:"char(1) not null 'business_nature'"                 json:"business_nature"`
	//LP for LawPrincipal 法人
	LPName          string `xorm:"varchar(64) default(null) 'law_principal_name'"             json:"law_principal_name"`
	LPCfType        string `xorm:"varchar(3) default(null) 'law_principal_certificate_type'"  json:"law_principal_certificate_type"`
	LPCfID          string `xorm:"varchar(128) default(null) 'law_principal_certificate_id'"  json:"law_principal_certificate_id"`
	RelationAccount string `xorm:"varchar(128) default(null) 'relationship_account'"          json:"relationship_account"`
	StartTime       string `xorm:"char(5) default(null) 'start_time'"                         json:"start_time"`
	EndTime         string `xorm:"char(5) default(null) 'end_time'"                           json:"end_time"`
	AccessType      string `xorm:"char(2) default(null) 'access_type'"                        json:"access_type"`
	OperatorNet     string `xorm:"varchar(2) default(null) 'operator_net'"                    json:"operator_net"`
	AccessIP        string `xorm:"varchar(64) default(null) 'access_ip'"                      json:"access_ip"`
	Approval        int    `xorm:"tinyint(1) not null default(0) 'approval'"                  json:"approval"`
	ApprovalTime    int64  `xorm:"bigint(14) default(null) 'approval_time'"                   json:"approval_time"`
	//EncryptionId    int64  `xorm:"tinyint(1) default(1) 'encryption_id'"                      json:"encryption_id"`
	Comments string `xorm:"varchar(255) default(null) 'comments'"    json:"comments"`
	//删除时间
	DeletedTime      int64 `xorm:"bigint(14) default(null) 'deleted_time'"                     json:"deleted_time"`
	DataStatus       int   `xorm:"tinyint(1) not null default(1) 'data_status'"                json:"data_status"`
	AliveStatus      int   `xorm:"tinyint(1) not null default(1) 'alive_status'"               json:"alive_status"`
	LastAliveTime    int64 `xorm:"bigint(14) default(null) 'last_alive_time'"                 json:"last_alive_time"`
	LastDataTime     int64 `xorm:"bigint(14) default(null) 'last_data_time'"                  json:"last_data_time"`
	BasicStatus      int   `xorm:"tinyint(1) not null default(1) 'basic_status'"               json:"basic_status"`
	LastBaiscTime    int64 `xorm:"bigint(14) default(null) 'last_basic_time'"                  json:"last_basic_time"`
	DevTotalNum      int   `xorm:"int(5) default(null) 'dev_total_num'"                        json:"dev_total_num"`
	DevOnlineNum     int   `xorm:"int(5) default(null) 'dev_online_num'"                        json:"dev_online_num"`
	DevOfflineNum    int   `xorm:"int(5) default(null) 'dev_offline_num'"                        json:"dev_offline_num"`
	DevAbnormalNum   int   `xorm:"int(5) default(null) 'dev_abnormal_num'"                        json:"dev_abnormal_num"`
	DevDataOnlineNum int   `xorm:"int(5) default(null) 'dev_data_online_num'"                        json:"dev_data_online_num"`
	//ServiceStatus  int       `xorm:"tinyint(1) default(0) 'service_status'"        json:"service_status"`
	BusinessStatus int       `xorm:"tinyint(1) default(0) 'business_status'"        json:"business_status"`
	CreatedAt      time.Time `xorm:"created"                                                    json:"-"` //json忽略该字段
	UpdatedAt      time.Time `xorm:"updated"                                                    json:"-"`
}

func (*NetbarInfoDeleted) TableName() string {
	return "NetbarInfo_Deleted"
}

func (this *NetbarInfoDeleted) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
}

func (this *NetbarInfoDeleted) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
}

func (this *NetbarInfoDeleted) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "wacode:", this.Wacode, "Id:", this.Id)
}

/*
//场所信息 ————未审核
type NetbarInfoNotApproved struct {
	NetbarInfo
}

func (*NetbarInfoNotApproved) TableName() string {
	return "NetbarInfo_NotApproved"
}

func (this *NetbarInfoNotApproved) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}

func (this *NetbarInfoNotApproved) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}

func (this *NetbarInfoNotApproved) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}

//场所信息 ————已审核
type netbarInfoApproved struct {
	NetbarInfo
}

func (*netbarInfoApproved) TableName() string {
	return "NetbarInfo_Approved"
}

func (this *netbarInfoApproved) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}

func (this *netbarInfoApproved) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}

func (this *netbarInfoApproved) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "wacode:", this.Wacode, "RegId:", this.Id)
}


*/

//设备信息
type DevInfo struct {
	Id          int64  `json:"idx"`
	OriId       int64  `xorm:"bigint(20) 'ori_id'"                                     json:"-"`
	Wacode      string `xorm:"char(14) notnull 'netbar_wacode'"                        json:"netbar_wacode"`
	PlaceName   string `xorm:"varchar(255) not null 'place_name'"                 json:"place_name"`
	NetbarIndex int64  `xorm:"bigint notnull 'netbar_index'"                              json:"-"` //场所的数据库索引
	ApId        string `xorm:"char(21) notnull unique 'ap_id'"                         json:"ap_id"`
	ApMac       string `xorm:"char(17) notnull unique 'ap_mac'"                        json:"ap_mac"`
	ApName      string `xorm:"char(128) default(null) 'ap_name'"                       json:"ap_name"`
	ApAddress   string `xorm:"char(255) default(null)"                                 json:"ap_address"`
	Type        int    `xorm:"tinyint(1) notnull default(1) 'type'"                    json:"type"`
	Status      int    `xorm:"tinyint(1) notnull default(0) 'status'"                  json:"status"`
	OfflineTime int64  `xorm:"bigint(14) default(null) 'last_offline_time'"            json:"last_offline_time"`
	AreaCode1   string `xorm:"char(6) not null 'area_code_1'"                       json:"area_code_1"`
	AreaCode2   string `xorm:"char(6) not null 'area_code_2'"                       json:"area_code_2"`
	AreaCode3   string `xorm:"char(6) not null 'area_code_3'"                       json:"area_code_3"`
	Radius      int    `xorm:"int(4) default(0) 'collection_radius'"                   json:"collection_radius"`
	//DataPeriod  int    `xorm:"int(6) default(0) 'data_period'"                         json:"data_period"`
	//ProbePeriod int    `xorm:"int(60) default(0) 'probe_period'"                       json:"probe_period"`

	Longitude          float64 `xorm:"decimal(10,7) default(null) 'longitude'"                 json:"longitude"`
	Latitude           float64 `xorm:"decimal(10,7) default(null) 'latitude'"                  json:"latitude"`
	Floor              string  `xorm:"varchar(16) default(null) 'floor'"                       json:"floor"`
	Station            string  `xorm:"varchar(128) default(null) 'subway_station'"             json:"subway_station"`
	LineInfo           string  `xorm:"varchar(255) default(null) 'subway_line_info'"           json:"subway_line_info"`
	VehicleInfo        string  `xorm:"varchar(255) default(null) 'subway_vehicle_info'"        json:"subway_vehicle_info"`
	CompartmentNO      string  `xorm:"varchar(255) default(null) 'subway_compartment_number'"  json:"subway_compartment_number"`
	CarCode            string  `xorm:"varchar(64) default(null) 'car_code'"                    json:"car_code"`
	Approval           int     `xorm:"tinyint(1) not null default(0) 'approval'"               json:"approval"`
	ApprovalTime       int64   `xorm:"bigint(14) default(null) 'approval_time'"                json:"approval_time"`
	Comments           string  `xorm:"varchar(255) default(null) 'comments'"                   json:"comments"`
	UploadInterval     int     `xorm:"int(6) default(0) 'upload_interval'"                     json:"upload_interval"`
	CollectionInterval int     `xorm:"int(6) default(0) 'collection_interval'"                 json:"collection_interval"`
	CaptureTime        int     `xorm:"bigint(14) default(0) 'capture_time'"                    json:"capture_time"`
	//删除时间
	DeletedTime   int64     `xorm:"bigint(14) default(null) 'deleted_time'"                 json:"-"`
	DataStatus    int       `xorm:"tinyint(1) not null default(1) 'data_status'"            json:"data_status"`
	AliveStatus   int       `xorm:"tinyint(1) not null default(1) 'alive_status'"           json:"alive_status"`
	LastAliveTime int64     `xorm:"bigint(14) default(null) 'last_alive_time'"              json:"last_alive_time"`
	LastDataTime  int64     `xorm:"bigint(14) default(null) 'last_data_time'"               json:"last_data_time"`
	BasicStatus   int       `xorm:"tinyint(1) not null default(1) 'basic_status'"           json:"basic_status"`
	LastBasicTime int64     `xorm:"bigint(14) default(null) 'last_basic_time'"              json:"last_basic_time"`
	DeviceStatus  int       `xorm:"tinyint(1) default(null) 'device_status'"                json:"device_status"`
	CreatedAt     time.Time `xorm:"created"                                                 json:"-"` //json忽略该字段
	UpdatedAt     time.Time `xorm:"updated"                                                 json:"-"`
}

func (*DevInfo) TableName() string {
	return "DevInfo"
}

func (this *DevInfo) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
}

func (this *DevInfo) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
}

func (this *DevInfo) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
	time.Now()
}

//设备信息 ————已删除
type DevInfoDeleted struct {
	Id          int64  `json:"-"`
	OriId       int64  `xorm:"bigint(20) notnull 'ori_id'"                             json:"idx"` //在未删除表中的index
	Wacode      string `xorm:"char(14) notnull 'netbar_wacode'"                        json:"netbar_wacode"`
	PlaceName   string `xorm:"varchar(255) not null 'place_name'"                      json:"place_name"`
	NetbarIndex int64  `xorm:"bigint notnull 'netbar_index'"                           json:"-"` //场所的数据库索引
	ApId        string `xorm:"char(21) notnull  'ap_id'"                               json:"ap_id"`
	ApMac       string `xorm:"char(17) notnull  'ap_mac'"                              json:"ap_mac"`
	ApName      string `xorm:"char(128) default(null) 'ap_name'"                       json:"ap_name"`
	ApAddress   string `xorm:"char(255) default(null)"                                 json:"ap_address"`
	Type        int    `xorm:"tinyint(1) notnull default(1) 'type'"                    json:"type"`
	Status      int    `xorm:"tinyint(1) notnull default(0) 'status'"                  json:"status"`
	OfflineTime int64  `xorm:"bigint(14) default(null) 'last_offline_time'"            json:"last_offline_time"`
	AreaCode1   string `xorm:"char(6) not null 'area_code_1'"                       json:"area_code_1"`
	AreaCode2   string `xorm:"char(6) not null 'area_code_2'"                       json:"area_code_2"`
	AreaCode3   string `xorm:"char(6) not null 'area_code_3'"                       json:"area_code_3"`
	Radius      int    `xorm:"int(4) default(0) 'collection_radius'"                   json:"collection_radius"`
	//DataPeriod  int    `xorm:"int(6) default(0) 'data_period'"                         json:"data_period"`
	//ProbePeriod int    `xorm:"int(60) default(0) 'probe_period'"                       json:"probe_period"`

	Longitude          float64 `xorm:"decimal(10,7) default(null) 'longitude'"                 json:"longitude"`
	Latitude           float64 `xorm:"decimal(10,7) default(null) 'latitude'"                  json:"latitude"`
	Floor              string  `xorm:"varchar(16) default(null) 'floor'"                       json:"floor"`
	Station            string  `xorm:"varchar(128) default(null) 'subway_station'"             json:"subway_station"`
	LineInfo           string  `xorm:"varchar(255) default(null) 'subway_line_info'"           json:"subway_line_info"`
	VehicleInfo        string  `xorm:"varchar(255) default(null) 'subway_vehicle_info'"        json:"subway_vehicle_info"`
	CompartmentNO      string  `xorm:"varchar(255) default(null) 'subway_compartment_number'"  json:"subway_compartment_number"`
	CarCode            string  `xorm:"varchar(64) default(null) 'car_code'"                    json:"car_code"`
	Approval           int     `xorm:"tinyint(1) not null default(0) 'approval'"               json:"approval"`
	ApprovalTime       int64   `xorm:"bigint(14) default(null) 'approval_time'"                json:"approval_time"`
	Comments           string  `xorm:"varchar(255) default(null) 'comments'"                   json:"comments"`
	UploadInterval     int     `xorm:"int(6) default(0) 'upload_interval'"                     json:"upload_interval"`
	CollectionInterval int     `xorm:"int(6) default(0) 'collection_interval'"                 json:"collection_interval"`
	CaptureTime        int     `xorm:"bigint(14) default(0) 'capture_time'"                    json:"capture_time"`
	//删除时间
	DeletedTime   int64     `xorm:"bigint(14) default(null) 'deleted_time'"                 json:"-"`
	DataStatus    int       `xorm:"tinyint(1) not null default(1) 'data_status'"            json:"data_status"`
	AliveStatus   int       `xorm:"tinyint(1) not null default(1) 'alive_status'"            json:"alive_status"`
	LastAliveTime int64     `xorm:"bigint(14) default(null) 'last_alive_time'"              json:"last_alive_time"`
	LastDataTime  int64     `xorm:"bigint(14) default(null) 'last_data_time'"               json:"last_data_time"`
	BasicStatus   int       `xorm:"tinyint(1) not null default(1) 'basic_status'"            json:"basic_status"`
	LastBaiscTime int64     `xorm:"bigint(14) default(null) 'last_basic_time'"                  json:"last_basic_time"`
	DeviceStatus  int       `xorm:"tinyint(1) default(null) 'device_status'"                json:"device_status"`
	CreatedAt     time.Time `xorm:"created"                                                 json:"-"` //json忽略该字段
	UpdatedAt     time.Time `xorm:"updated"                                                 json:"-"`
}

func (*DevInfoDeleted) TableName() string {
	return "DevInfo_Deleted"
}

func (this *DevInfoDeleted) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
}

func (this *DevInfoDeleted) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
}

func (this *DevInfoDeleted) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "apid:", this.ApId, "Id:", this.Id)
}

/*
//设备信息 ————未审核
type devInfoNotApproved struct {
	DevInfo
}

func (*devInfoNotApproved) TableName() string {
	return "DevInfo_NotApproved"
}

func (this *devInfoNotApproved) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

func (this *devInfoNotApproved) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

func (this *devInfoNotApproved) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

//设备信息 ————已审核
type devInfoApproved struct {
	DevInfo
}

func (*devInfoApproved) TableName() string {
	return "DevInfo_Approved"
}

func (this *devInfoApproved) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

func (this *devInfoApproved) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

func (this *devInfoApproved) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "apid:", this.ApId, "RegId:", this.Id)
}

*/
//加密信息表
type EncryptionInfo struct {
	Id          int64     `json:"idx"`
	Name        string    `xorm:"varchar(255) notnull 'name'"                   json:"name"`
	Description string    `xorm:"varchar(255) default(null) 'description'"      json:"description"`
	Type        int       `xorm:"tinyint(2) default(0) 'type'"                  json:"type"`
	EnKey       string    `xorm:"varchar(255) default(null) 'en_key'"           json:"en_key"`
	EnIv        string    `xorm:"varchar(255) default(null) 'en_iv'"            json:"en_iv"`
	DeKey       string    `xorm:"varchar(255) default(null) 'de_key'"           json:"de_key"`
	DeIv        string    `xorm:"varchar(255) default(null) 'de_iv'"            json:"de_iv"`
	CreatedAt   time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt   time.Time `xorm:"updated"                                       json:"-"`
}

func (*EncryptionInfo) TableName() string {
	return "EncryptionInfo"
}

func (this *EncryptionInfo) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *EncryptionInfo) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *EncryptionInfo) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

//注册用户信息表
type UsersInfo struct {
	Id           int64     `json:"idx"`
	TransType    string    `xorm:"char(3) notnull 'trans_type'"         json:"trans_type"` //tcp,udp
	Host         string    `xorm:"char(15) notnull 'host'"              json:"host"`       //ip address
	Port         int       `xorm:"int(5) notnull 'port'"                json:"port"`
	Msgtype      int       `xorm:"bigint(14) notnull 'msg_type'"       json:"-"` //"netbars"
	Msgtype_json []string  `xorm:"-"                                   json:"msgtype"`
	RegId        string    `xorm:"varchar(32) notnull unique 'reg_id'"  json:"regid"`      //订阅者ID
	CreatedAt    time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt    time.Time `xorm:"updated"                                       json:"-"`
}

func (*UsersInfo) TableName() string {
	return "UsersInfo"
}

func (this *UsersInfo) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "reg_id:", this.RegId, "Id:", this.Id)
}

func (this *UsersInfo) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "reg_id:", this.RegId, "Id:", this.Id)
}

func (this *UsersInfo) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "reg_id:", this.RegId, "Id:", this.Id)
}

//config表
type DevicesConfig struct {
	Id                 int64     `json:"idx"`
	NetbarAliveTimeout int64     `xorm:"bigint(10) default(0) 'netbar_alive_timeout'"    json:"netbar_alive_timeout"`
	NetbarBasicTimeout int64     `xorm:"bigint(10) default(0) 'netbar_basic_timeout'"    json:"netbar_basic_timeout"`
	ApAliveTimeout     int64     `xorm:"bigint(10) default(0) 'ap_alive_timeout'"        json:"ap_alive_timeout"`
	ApDataTimeout      int64     `xorm:"bigint(10) default(0) 'ap_data_timeout'"         json:"ap_data_timeout"`
	APBasicTimeout     int64     `xorm:"bigint(10) default(0) 'ap_basic_timeout'"        json:"ap_basic_timeout"`
	FileExpiredTime    int64     `xorm:"bigint(10) default(0) 'file_expired_time'"        json:"file_expired_time"`
	CreatedAt          time.Time `xorm:"created"                                         json:"-"` //json忽略该字段
	UpdatedAt          time.Time `xorm:"updated"                                         json:"-"`
}

func (*DevicesConfig) TableName() string {
	return "DevicesConfig"
}

func (this *DevicesConfig) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "Id:", this.Id)
}

func (this *DevicesConfig) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "Id:", this.Id)
}

func (this *DevicesConfig) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "Id:", this.Id)
}

//csv文件表
type FileSys struct {
	Id        int64     `json:"idx"`
	FileName  string    `xorm:"varchar(128) notnull unique 'file_name'"       json:"file_name"`
	FileType  string    `xorm:"varchar(5) notnull 'file_type'"                json:"file_type"`
	CreatedAt time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                       json:"-"`
}

func (*FileSys) TableName() string {
	return "FileSys"
}

func (this *FileSys) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "FileName:", this.FileName, "Id:", this.Id)
}

func (this *FileSys) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "FileName:", this.FileName, "Id:", this.Id)
}

func (this *FileSys) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "FileName:", this.FileName, "Id:", this.Id)
}

//场所经营性质表
type BusinessNature struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"             json:"name"`
	Code      string    `xorm:"char(1) notnull 'code'"              json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"  json:"comments"`
	CreatedAt time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                       json:"-"`
}

func (*BusinessNature) TableName() string {
	return "BussinessNature"
}

func (this *BusinessNature) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *BusinessNature) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *BusinessNature) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//场所服务类型表
type NetsiteType struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"             json:"name"`
	Code      string    `xorm:"char(1) notnull 'code'"              json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"  json:"comments"`
	Nature    string    `xorm:"varchar(64) notnull 'nature'"     json:"nature"`
	CreatedAt time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                       json:"-"`
}

func (*NetsiteType) TableName() string {
	return "NetsiteType"
}

func (this *NetsiteType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *NetsiteType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *NetsiteType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//常用证件代码表
type CertificateType struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"                    json:"name"`
	Code      string    `xorm:"char(3) notnull 'code'"                        json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"         json:"comments"`
	CreatedAt time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                       json:"-"`
}

func (*CertificateType) TableName() string {
	return "CertificateType"
}

func (this *CertificateType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *CertificateType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *CertificateType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//接入服务商表
type OperatorNet struct {
	Id          int64     `json:"idx"`
	Name        string    `xorm:"varchar(128) notnull 'name'"             json:"name"`
	ServiceName string    `xorm:"varchar(32) notnull 'service_name'"     json:"service_name"`
	Code        string    `xorm:"char(2) notnull 'code'"              json:"code"`
	Comments    string    `xorm:"varchar(255) default(null) 'comments'"  json:"comments"`
	CreatedAt   time.Time `xorm:"created"                                       json:"-"` //json忽略该字段
	UpdatedAt   time.Time `xorm:"updated"                                       json:"-"`
}

func (*OperatorNet) TableName() string {
	return "OperatorNet"
}

func (this *OperatorNet) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *OperatorNet) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *OperatorNet) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//接入方式表
type AccessType struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"                json:"name"`
	Code      string    `xorm:"char(2) notnull 'code'"                    json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"     json:"comments"`
	CreatedAt time.Time `xorm:"created"                                   json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                   json:"-"`
}

func (*AccessType) TableName() string {
	return "AccessType"
}

func (this *AccessType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *AccessType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *AccessType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//采集设备类型表
type ApType struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"                json:"name"`
	Code      string    `xorm:"char(2) notnull 'code'"                    json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"     json:"comments"`
	CreatedAt time.Time `xorm:"created"                                   json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                   json:"-"`
}

func (*ApType) TableName() string {
	return "ApType"
}

func (this *ApType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *ApType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *ApType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//认证类型表
type AuthType struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"                json:"name"`
	Code      string    `xorm:"char(10) notnull 'code'"                    json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"     json:"comments"`
	CreatedAt time.Time `xorm:"created"                                   json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                   json:"-"`
}

func (*AuthType) TableName() string {
	return "AuthType"
}

func (this *AuthType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *AuthType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *AuthType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//虚拟身份类型表
type ImType struct {
	Id        int64     `json:"-"`
	Name      string    `xorm:"varchar(128) notnull 'name'"               json:"name"`
	Code      string    `xorm:"char(10) notnull 'code'"                   json:"code"`
	Type      string    `xorm:"varchar(32) notnull 'type'"                json:"type"`
	TypeName  string    `xorm:"varchar(64) notnull 'type_name'"           json:"type_name"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"     json:"comments"`
	CreatedAt time.Time `xorm:"created"                                   json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                   json:"-"`
}

func (*ImType) TableName() string {
	return "ImType"
}

func (this *ImType) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *ImType) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *ImType) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//应用服务类型代码表
type NetworkApp struct {
	Id        int64     `json:"idx"`
	Name      string    `xorm:"varchar(128) notnull 'name'"                json:"name"`
	Code      string    `xorm:"char(5) notnull 'code'"                    json:"code"`
	Comments  string    `xorm:"varchar(255) default(null) 'comments'"     json:"comments"`
	CreatedAt time.Time `xorm:"created"                                   json:"-"` //json忽略该字段
	UpdatedAt time.Time `xorm:"updated"                                   json:"-"`
}

func (*NetworkApp) TableName() string {
	return "NetworkApp"
}

func (this *NetworkApp) AfterInsert() {
	LOG_TRAC.Println("Insert in", this.TableName(), "name:", this.Name, "Id:", this.Id)
}

func (this *NetworkApp) AfterUpdate() {
	LOG_TRAC.Println("Update in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

func (this *NetworkApp) AfterDelete() {
	LOG_TRAC.Println("Delete in", this.TableName(), "id:", this.Name, "Id:", this.Id)
}

//消息中心事件表
type MessageCenter struct {
	Id          int64  `json:"idx"`
	Type        string `xorm:"varchar(8) default(null) 'type'"                             json:"type"`
	Orgname     string `xorm:"varchar(255) default(null) 'security_software_orgname'"      json:"security_software_orgname"`
	Orgcode     string `xorm:"char(9) default(null) 'security_software_orgcode'"           json:"security_software_orgcode"`
	Wacode      string `xorm:"char(14) default(null) 'netbar_wacode'"                      json:"netbar_wacode"`
	PlaceName   string `xorm:"varchar(255) default(null) 'place_name'"                     json:"place_name"`
	ApId        string `xorm:"char(21)  default(null) 'ap_id'"                             json:"ap_id"`
	Read        string `xorm:"varchar(3) 'has_read'"                                           json:"read"`
	CreatedTime int64  `json:"created_time"`
}

func (*MessageCenter) TableName() string {
	return "MessageCenter"
}

func (this *MessageCenter) AfterInsert() {
	LOG_TRAC.Println("Insert in", "Id:", this.Id)
}

func (this *MessageCenter) AfterUpdate() {
	LOG_TRAC.Println("Update in", "Id:", this.Id)
}

func (this *MessageCenter) AfterDelete() {
	LOG_TRAC.Println("Delete in", "Id:", this.Id)
}
