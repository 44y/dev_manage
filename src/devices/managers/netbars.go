//场所管理
package managers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-xorm/xorm"
	gmux "github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	. "rc"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NetbarsMngSt struct {
	HttpSt
	approval  int
	file_type string
	file_name string //full name
}

/*
	处理http GET请求
*/
func (this *NetbarsMngSt) httpGET() {
	LOG_TRAC.Println("httpGet")

	//解析body
	getbody := new(DevicesGetBody)
	if err := json.NewDecoder(this.req.Body).Decode(getbody); err != nil {
		if err.Error() == "EOF" {
			LOG_INFO.Println("No get body!")
		} else {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
	}
	LOG_TRAC.Printf("getbody:%+v\n", getbody)

	//查询
	results, code, err := this.find(getbody)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, code, err.Error(), nil, nil)
		return
	}
	QueryResponse(this.writer, results.total, results.count, results.results)
}

/*
	检查地区编码长度是否符合
*/
func checkAreaCodeLen(code *AreaCode) bool {
	if code == nil {
		return true
	}

	for _, v := range code.Level1 {
		if len(v) != AREACODE_LEN {
			return false
		}
	}

	for _, v := range code.Level2 {
		if len(v) != AREACODE_LEN {
			return false
		}
	}
	for _, v := range code.Level3 {
		if len(v) != AREACODE_LEN {
			return false
		}
	}

	return true
}

/*
	将code2中的编码在code1中查询，生成查询字符串。没有地区编码返回 "",nil
	code1: 查询范围
	code2：权限范围
*/
func getAreaCodeStr(code1, code2 *AreaCode) (string, error) {
	LOG_TRAC.Println("code1 :", code1, "code2:", code2)

	if !checkAreaCodeLen(code1) || !checkAreaCodeLen(code2) {
		return "", errors.New("code format wrong!")
	}

	var (
		code1_len, code2_len int
		ret_str              string
		str_format           = " area_code_3 like '%s' or"
		v1, v2               string
		found                bool
		code_tmp             *AreaCode
	)

	if code1 != nil {
		code1_len = len(code1.Level1) + len(code1.Level2) + len(code1.Level3)
	}
	if code2 != nil {
		code2_len = len(code2.Level1) + len(code2.Level2) + len(code2.Level3)
	}

	//code1 code2都为空，返回""
	if code1_len == 0 && code2_len == 0 {
		LOG_INFO.Println("No area code!")
		return "", nil
	} else if code1_len*code2_len == 0 {
		//没有权限范围，仅根据查询范围查询
		if code1_len == 0 {
			code_tmp = code2
		}
		//没有查询范围，仅根据权限范围查询
		if code2_len == 0 {
			code_tmp = code1
		}

		for _, v1 = range code_tmp.Level1 {
			ret_str = ret_str + fmt.Sprintf(str_format, v1[0:AREACODE_LEN-4]+"%")
		}

		for _, v1 = range code_tmp.Level2 {
			ret_str = ret_str + fmt.Sprintf(str_format, v1[0:AREACODE_LEN-2]+"%")
		}

		for _, v1 = range code_tmp.Level3 {
			ret_str = ret_str + fmt.Sprintf(str_format, v1)
		}

		//去除开头空格和末尾"or"
		ret_str = ret_str[1 : len(ret_str)-2]
		LOG_TRAC.Println("ret_str :", ret_str)
		return ret_str, nil
	}

	//两个范围都有时，取交集
	for _, v1 = range code1.Level1 {
		found = false
		//一级地区编码完全相等
		for _, v2 = range code2.Level1 {
			if v1 == v2 {
				//去除末尾0
				ret_str = ret_str + fmt.Sprintf(str_format, v1[0:AREACODE_LEN-4]+"%")
				found = true
				break
			}
		}

		//二级编码属于一级范围
		if !found {
			for _, v2 = range code2.Level2 {
				if v1[0:2] == v2[0:2] {
					ret_str = ret_str + fmt.Sprintf(str_format, v2[0:AREACODE_LEN-2]+"%")
					found = true
				}
			}
		}

		//三级编码属于一级范围
		if !found {
			for _, v2 = range code2.Level3 {
				if v1[0:2] == v2[0:2] {
					ret_str = ret_str + fmt.Sprintf(str_format, v2)
				}
			}
		}
	}

	for _, v1 = range code1.Level2 {
		found = false
		//二级编码属于一级范围
		for _, v2 = range code2.Level1 {
			if v1[0:2] == v2[0:2] {
				//去除末尾0
				ret_str = ret_str + fmt.Sprintf(str_format, v1[0:AREACODE_LEN-2]+"%")
				found = true
				break
			}
		}

		//二级地区编码完全相等
		if !found {
			for _, v2 = range code2.Level2 {
				if v1 == v2 {
					//去除末尾0
					ret_str = ret_str + fmt.Sprintf(str_format, v1[0:AREACODE_LEN-2]+"%")
					found = true
					break
				}
			}
		}

		//三级编码属于二级范围
		if !found {
			for _, v2 = range code2.Level3 {
				if v1[0:4] == v2[0:4] {
					ret_str = ret_str + fmt.Sprintf(str_format, v2)
				}
			}
		}
	}

	for _, v1 = range code1.Level3 {
		found = false
		//三级编码属于一级范围
		for _, v2 = range code2.Level1 {
			if v1[0:2] == v2[0:2] {
				ret_str = ret_str + fmt.Sprintf(str_format, v1)
				found = true
				break
			}
		}

		//三级编码属于二级范围
		if !found {
			for _, v2 = range code2.Level2 {
				if v1[0:4] == v2[0:4] {
					ret_str = ret_str + fmt.Sprintf(str_format, v1)
					found = true
					break
				}
			}
		}

		//三级编码完全相等
		if !found {
			for _, v2 = range code2.Level3 {
				if v1 == v2 {
					ret_str = ret_str + fmt.Sprintf(str_format, v1)
					break
				}
			}
		}
	}

	//去除开头空格和末尾"or"
	ret_str = ret_str[1 : len(ret_str)-2]
	LOG_TRAC.Println("ret_str :", ret_str)
	return ret_str, nil
}

/*
	根据过滤条件返回查询字符串
*/
func getNetbarFilterStr(filter *NetbarFilter) string {
	var ret_str string

	if len(filter.BussinessNature) > 0 {
		ret_str = ret_str + fmt.Sprintf(" business_nature like '%s' and", filter.BussinessNature)
	}

	if len(filter.NetsiteType) > 0 {
		ret_str = ret_str + fmt.Sprintf(" netsite_type like '%s' and", filter.NetsiteType)
	}

	//根据厂商名字查询厂商编码
	if len(filter.Orgname) > 0 {
		org := &OrgInfo{Orgname: filter.Orgname}
		if _, err := X.Get(org); err != nil {
			LOG_ERRO.Println(err)
			return ""
		}
		ret_str = ret_str + fmt.Sprintf(" security_software_orgcode like '%s' and", org.Orgcode)
	}
	//根据厂商编码查询
	if len(filter.Orgcode) > 0 {
		ret_str = ret_str + fmt.Sprintf(" security_software_orgcode like '%s' and", filter.Orgcode)
	}

	switch filter.Status {
	case NETBARSTATUS_DEV_ONLINE:
		ret_str = ret_str + fmt.Sprintf(" dev_online_num > 0 and")
	case NETBARSTATUS_DEV_OFFLINE:
		ret_str = ret_str + fmt.Sprintf(" dev_offline_num > 0 and")
	case NETBARSTATUS_DEV_ABNORMAL:
		ret_str = ret_str + fmt.Sprintf(" dev_abnormal_num > 0 and")
	case NETBARSTATUS_DEV_EMPTY:
		ret_str = ret_str + fmt.Sprintf(" dev_total_num = 0 and")
	}

	//去除开头空格和末尾"and"
	if len(ret_str) != 0 {
		ret_str = ret_str[1 : len(ret_str)-3]
	}

	return ret_str
}

/*
	根据条件查询场所
*/
func (this *NetbarsMngSt) find(getbody *DevicesGetBody) (*findResults, int, error) {
	LOG_TRAC.Println(this.req.Form)

	var (
		total         int64
		index         int64
		err           error
		ex            bool
		rslt, rslt_tt interface{} //slice类型指针
		//tbl          interface{} //表结构实例指针
		isDeleted    bool   //true表示查询deleted表
		areacode_str string //根据地区编码组成查询字符串
		filter_str   string //根据过滤条件生成查询字符串
		orgcode_str  string //根据厂商编码组成查询字符串
	)

	//检查url中index是否合法，合法则为精确查询index
	index, err = ParseIndexFromUrl(this.req)
	if err == nil && index != 0 {
		netbar := &NetbarInfo{Id: index}
		ex, err = X.Get(netbar)
		if err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusInternalServerError, err
		}
		if ex {
			sl := make([]NetbarInfo, 0)
			sl = append(sl, *netbar)
			return &findResults{count: 1, total: 1, results: sl}, 0, nil
		}

		del_netbar := &NetbarInfoDeleted{OriId: index}
		ex, err = X.Get(del_netbar)
		if err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusInternalServerError, err
		}
		if ex {
			sl := make([]NetbarInfoDeleted, 0)
			sl = append(sl, *del_netbar)
			return &findResults{count: 1, total: 1, results: sl}, 0, nil
		}
		return nil, http.StatusBadRequest, errors.New(RecordNotExist)
	}

	//解析查询参数
	keyword := this.req.Form.Get("keyword")

	offset, _ := strconv.Atoi(this.req.Form.Get("offset"))
	limit, _ := strconv.Atoi(this.req.Form.Get("limit"))
	if limit == 0 {
		limit = LIMIT_DEFAULT
	}
	scope := strings.Split(this.req.Form.Get("scope"), " ")

	LOG_TRAC.Println("limit:", limit, "offset:", offset, "scope:", scope)

	if getbody != nil {
		//生成地区编码查询字符串
		if areacode_str, err = getAreaCodeStr(&getbody.Filter.Area_code, &getbody.Scope.AreaCode); err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusBadRequest, err
		}

		//根据scope中的厂商编码进行查询限制
		for _, v := range getbody.Scope.Orgcodes {
			orgcode_str = orgcode_str + fmt.Sprintf(" security_software_orgcode = %s or", v)
		}
		if len(orgcode_str) > 0 {
			orgcode_str = orgcode_str[1 : len(orgcode_str)-2]
		}
		LOG_TRAC.Println("orgcode_str:", orgcode_str)

		//生成过滤条件查询字符串
		filter_str = getNetbarFilterStr(&getbody.Filter)
		LOG_TRAC.Println(filter_str)
	}

	//根据approval指定查询的表
	switch this.approval {
	case APPROVED, NOTAPPROVED: //查询NetbarInfo
		r := make([]NetbarInfo, 0)
		rslt = &r
		t := make([]NetbarInfo, 0)
		rslt_tt = &t
		//tbl = new(NetbarInfo)
		isDeleted = false
	case DELETED: //NetbarInfoDeleted
		r := make([]NetbarInfoDeleted, 0)
		rslt = &r
		t := make([]NetbarInfoDeleted, 0)
		rslt_tt = &t
		//tbl = new(NetbarInfoDeleted)
		isDeleted = true
	default:
		LOG_ERRO.Println("Unknown approval value:", this.approval)
		return nil, http.StatusBadRequest, errors.New("Unknown approval value")
	}

	approval_str := "approval = " + strconv.Itoa(this.approval)

	//根据表和地区编码区分查询条件
	var ss, ss_t *xorm.Session

	if isDeleted {
		ss = X.Select("SQL_CALC_FOUND_ROWS *")
		ss_t = X.Select("SQL_CALC_FOUND_ROWS *")
	} else {
		ss = X.Where(approval_str).Select("SQL_CALC_FOUND_ROWS *")
		ss_t = X.Where(approval_str).Select("SQL_CALC_FOUND_ROWS *")
	}

	//areacode查询条件
	if len(areacode_str) > 0 {
		ss = ss.And(areacode_str)
		ss_t = ss_t.And(areacode_str)
	}

	//filter查询条件
	if len(filter_str) > 0 {
		ss = ss.And(filter_str)
		ss_t = ss_t.And(filter_str)
	}

	//orgcode查询条件
	if len(orgcode_str) > 0 {
		ss = ss.And(orgcode_str)
		ss_t = ss_t.And(orgcode_str)
	}

	//查询关键字
	if len(keyword) != 0 {
		scope_keys := []string{"wacode", "place_name"} //支持的查询关键字
		scope_map := ParseScope(scope_keys, scope)
		LOG_TRAC.Println(scope_map)

		wacode_str := "netbar_wacode like " + " '%" + keyword + "%' "
		name_str := "place_name like " + " '%" + keyword + "%'"

		//查询场所编码和名字
		if (scope_map["wacode"] && scope_map["place_name"]) ||
			(!scope_map["wacode"] && !scope_map["place_name"]) {

			ss = ss.And(wacode_str + " or " + name_str)
			ss_t = ss_t.And(wacode_str + " or " + name_str)

		} else if scope_map["wacode"] { //查询场所编码

			ss = ss.And(wacode_str)
			ss_t = ss_t.And(wacode_str)

		} else { //查询场所名字

			ss = ss.And(name_str)
			ss_t = ss_t.And(name_str)

		}
	}

	err = ss.Limit(limit, offset).Find(rslt)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	count := int64(reflect.ValueOf(rslt).Elem().Len())
	if count == 0 {
		total = 0
	} else {
		if err = ss_t.Find(rslt_tt); err != nil {
			return nil, http.StatusInternalServerError, err
		}
		total = int64(reflect.ValueOf(rslt_tt).Elem().Len())
	}
	rslt_v := reflect.ValueOf(rslt).Elem().Interface()
	LOG_TRAC.Println(count, total)

	//更新场所status和下挂设备状态
	if v, ok := rslt_v.([]NetbarInfo); ok {
		if rslt_v, err = updateNetbarColumns(v); err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusInternalServerError, err
		}
	}

	LOG_TRAC.Println(rslt_v)
	ret := &findResults{
		count:   count,
		total:   total,
		results: rslt_v,
	}

	return ret, 0, nil
}

/*
	处理http POST请求
*/
func (this *NetbarsMngSt) httpPOST() {
	var (
		err error
	)
	//解析body
	body := new(NetbarInfo)
	if err = json.NewDecoder(this.req.Body).Decode(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body is %+v\n", body)
	//检查body参数
	var emap ErrorMap
	if emap, err = this.checkPostBody(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
		return
	}
	body.DataStatus = STATUS_ABNORMAL
	body.AliveStatus = STATUS_ABNORMAL
	body.BasicStatus = STATUS_ABNORMAL
	body.BusinessStatus = NETBAR_BUSINESS_CLOSE

	if _, err = X.InsertOne(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	X.Get(body)

	send2Publisher(ACTION_ADD, body, nil)
	SuccessResponse(this.writer, *body)
	/*
			bean := make([]NetbarInfo, 0)

			//插入数据库
			if ret_data, err = doPOST(body, &bean); err != nil {
				LOG_ERRO.Println(err)
				ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(),nil,nil)
				return
			}

		SuccessResponse(this.writer, reflect.ValueOf(ret_data).Elem().Interface())
	*/
}

/*
	支持直接添加审核/未审核的场所/设备，故post请求时可能执行更新表操作
	根据是否存在数据进行插入或更新数据
	data: 要插入的数据
	bean: 表结构slice指针，供查询用
	code: wacode或ap RegId
*/
func doPOST(data, bean interface{}) (interface{}, error) {
	var (
		id, count int64
		err       error
		where_str string
	)

	//类型断言，判断出表名.检查记录是否已存在
	if v, ok := data.(*DevInfo); ok {
		where_str = "ap_id = " + v.ApId
	} else if v, ok := data.(*NetbarInfo); ok {
		where_str = "netbar_wacode = " + v.Wacode
	}
	if count, err = X.Where(where_str).FindAndCount(bean); err != nil {
		return nil, err
	} else if count == 1 { //记录已存在，更新

		//取出记录id
		old := reflect.ValueOf(bean).Elem().Index(0)
		id = old.FieldByName("Id").Int()
		if ret_data, err := updateData(id, data); err != nil {
			return nil, err
		} else {
			return ret_data, nil
		}

	} else if count == 0 { //记录不存在，插入
		if id, err = X.InsertOne(data); err != nil {
			return nil, err
		}
		//赋值新的id
		id_field := reflect.ValueOf(data).Elem().FieldByName("Id")
		if id_field.CanSet() {
			id_field.SetInt(id)
		} else {
			//publish
			send2Publisher(ACTION_ADD, data, nil)

			return nil, errors.New("Id can not set!")
		}
		return data, nil
	} else { //同一个wacode不止一条记录
		panic("Too many records for 1 wacode," + where_str + ":" + strconv.Itoa(int(count)))
	}
}

/*
	根据id获取加密信息
	en为结构体指针需包含字段：
	EnType  int
	DeKey   string
	DeIv    string
*/
func getEncryption(id int64, en interface{}) error {
	LOG_TRAC.Println("RegId: ", id)

	en_data := new(EncryptionInfo)

	if ex, err := X.ID(id).Get(en_data); err != nil {
		return err
	} else if !ex {
		return errors.New("No encryption")
	}

	v := reflect.ValueOf(en).Elem()

	if v.FieldByName("EnType").CanSet() {
		v.FieldByName("EnType").SetInt(int64(en_data.Type))
	}

	if v.FieldByName("DeKey").CanSet() {
		v.FieldByName("DeKey").SetString(en_data.DeKey)
	}

	if v.FieldByName("DeIv").CanSet() {
		v.FieldByName("DeIv").SetString(en_data.DeIv)
	}

	return nil
}

/*
	比较场所信息是否变更
	ret:
	1 信息变更
	2 厂商编码或场所编码变更
	0 没有变更
*/
func ifNetbarInfoChanged(x, y *NetbarInfo) int {

	if x.Orgcode != y.Orgcode || x.Wacode != y.Wacode {
		return 2
	}

	if x.PlaceName != y.PlaceName {
		return 1
	}

	return 0
}

/*
	场所状态变更时，构建消息，交由publisher发布
	action: 变更动作
	new_data,old_data: 变更前后数据指针，根据需要决定是否发送
*/
func send2Publisher(action int, new_data, old_data interface{}) {
	var (
		v_new, v_old *NetbarInfo
		ok, ex       bool
		err          error
	)

	switch action {
	case ACTION_ADD: //只发送已审核场所
		if v_new, ok = new_data.(*NetbarInfo); ok && v_new.Approval == APPROVED {
			msg := &pubMsgSt{
				MsgType:   MSGTYPE_NETBARS_STR,
				OrgCode:   v_new.Orgcode,
				OrgName:   v_new.Orgname,
				PlaceName: v_new.PlaceName,
				WaCode:    v_new.Wacode,
				Action:    action,
			}
			//根据厂商编码取出加密信息
			org := &OrgInfo{Orgcode: v_new.Orgcode}
			if ex, err = X.Get(org); err != nil {
				LOG_ERRO.Println(err)
				return
			} else if ex {
				if err = getEncryption(org.EncryptId, msg); err != nil {
					LOG_ERRO.Println("getEncryption wrong! encryptid:", org.EncryptId)
					return
				}
				pubMsgChan <- msg
			}

			//update netbars code file in ftp server
			WriteApporvalNetbars2File()

		} else {
			LOG_TRAC.Println("do nothing")
			return
		}
	case ACTION_MODIFY: //发送所关心参数变化的场所，只发送审核场所
		if v_new, ok = new_data.(*NetbarInfo); !ok {
			LOG_TRAC.Println("do nothing")
			return
		}
		if v_old, ok = old_data.(*NetbarInfo); !ok {
			LOG_TRAC.Println("do nothing")
			return
		}
		if v_old.Approval == NOTAPPROVED && v_new.Approval == NOTAPPROVED {
			LOG_TRAC.Println("do nothing")
			return

		}
		var en_new struct {
			EnType int
			DeKey  string
			DeIv   string
		}
		msg := &pubMsgSt{
			MsgType: MSGTYPE_NETBARS_STR,
		}
		//已审核场所信息变更
		if v_old.Approval == APPROVED && v_new.Approval == APPROVED {
			//根据厂商编码取出加密信息
			org_old := &OrgInfo{Orgcode: v_old.Orgcode}
			if ex, err = X.Get(org_old); err != nil {
				LOG_ERRO.Println(err)
				return
			}

			org_new := &OrgInfo{Orgcode: v_new.Orgcode}
			if ex, err = X.Get(org_new); err != nil {
				LOG_ERRO.Println(err)
				return
			}
			if err = getEncryption(org_new.EncryptId, &en_new); err != nil {
				LOG_ERRO.Println("getEncryption wrong! encryptid:", org_new.EncryptId)
				return
			}
			msg.OrgCode = v_new.Orgcode
			msg.OrgName = v_new.Orgname
			msg.PlaceName = v_new.PlaceName
			msg.WaCode = v_new.Wacode
			msg.Action = ACTION_MODIFY
			msg.EnType = en_new.EnType
			msg.DeKey = en_new.DeKey
			msg.DeIv = en_new.DeIv
			ischanged := ifNetbarInfoChanged(v_old, v_new)

			if ischanged == 2 {
				//场所编码 厂商编码变更  发送删除和新增
				msg1 := &pubMsgSt{
					MsgType:   MSGTYPE_NETBARS_STR,
					OrgCode:   v_old.Orgcode,
					OrgName:   v_old.Orgname,
					PlaceName: v_old.PlaceName,
					WaCode:    v_old.Wacode,
					Action:    ACTION_DELETE,
				}
				pubMsgChan <- msg1
				msg.Action = ACTION_ADD
			} else if ischanged == 0 {
				LOG_TRAC.Println("do nothing")
				return
			}
		}
		//审核变未审核
		if v_old.Approval == APPROVED && v_new.Approval == NOTAPPROVED {
			msg.OrgCode = v_old.Orgcode
			msg.PlaceName = v_old.PlaceName
			msg.WaCode = v_old.Wacode
			msg.Action = ACTION_DELETE

			//update netbars code file in ftp server
			WriteApporvalNetbars2File()
		}

		//未审核变审核
		if v_old.Approval == NOTAPPROVED && v_new.Approval == APPROVED {
			//根据厂商编码取出加密信息
			org := &OrgInfo{Orgcode: v_new.Orgcode}
			if ex, err = X.Get(org); err != nil {
				LOG_ERRO.Println(err)
				return
			}

			if err = getEncryption(org.EncryptId, &en_new); err != nil {
				LOG_ERRO.Println("getEncryption wrong! encryptid:", org.EncryptId)
			}
			msg.OrgCode = v_new.Orgcode
			msg.PlaceName = v_new.PlaceName
			msg.WaCode = v_new.Wacode
			msg.Action = ACTION_ADD
			msg.EnType = en_new.EnType
			msg.DeKey = en_new.DeKey
			msg.DeIv = en_new.DeIv

			//update netbars code file in ftp server
			WriteApporvalNetbars2File()
		}
		pubMsgChan <- msg

	case ACTION_DELETE: //只发送审核场所
		if v_deleted, ok := old_data.(*NetbarInfoDeleted); !ok || v_deleted.Approval != APPROVED {
			LOG_TRAC.Println("do nothing")
			return
		} else {
			msg := &pubMsgSt{
				MsgType:   MSGTYPE_NETBARS_STR,
				OrgCode:   v_deleted.Orgcode,
				OrgName:   v_deleted.Orgname,
				PlaceName: v_deleted.PlaceName,
				WaCode:    v_deleted.Wacode,
				Action:    action,
			}
			pubMsgChan <- msg
			//update netbars code file in ftp server
			WriteApporvalNetbars2File()
		}
	default:
		LOG_TRAC.Println("do nothing")
		return
	}
}

/*
	创建一个小于max的随机数
*/
var RandSeed int64 = 0

func createRandNum(max int) int {
	rand.Seed(RandSeed + time.Now().Unix())
	RandSeed++
	return rand.Intn(max)
}

/*
	检查场所编码相关字段是否合法，返回合法的场所编码
*/
func checkWacode(area, busi, netsite, orgcode2, serial *string) (string, ErrorMap, error) {
	var (
		ex, match bool
		err       error
	)

	if match, err = regexp.MatchString(REGEXP_AREACODE, *area); err != nil || !match {
		return "", ErrorMap{"area_code_3": []string{BadAreaCode}}, errors.New("area code wrong!")
	}

	//check business_type
	if len(*busi) == 0 {
		return "", ErrorMap{"business_nature": []string{BadNetsiteType}}, errors.New("business nature wrong!")
	} else {
		bust := &BusinessNature{Code: *busi}
		if ex, err = X.Exist(bust); err != nil {
			LOG_ERRO.Println(err)
			return "", nil, err
		} else if !ex {
			return "", ErrorMap{"business_nature": []string{TypeWrong}}, errors.New("BusinessNature wrong!")
		}
	}

	//check netsite_type
	if len(*netsite) == 0 {
		return "", ErrorMap{"netsite_type": []string{EmptNotAllowed}}, errors.New("netsite type wrong!")
	} else {
		if *busi == OPERATING {
			if *netsite != "0" {
				return "", ErrorMap{"netsite_type": []string{TypeWrong}}, errors.New("netsite type wrong!")
			}
		} else {
			nett := &NetsiteType{Code: *netsite}
			if ex, err = X.Get(nett); err != nil {
				LOG_ERRO.Println(err)
				return "", nil, err
			} else if !ex {
				return "", ErrorMap{"netsite_type": []string{TypeWrong}}, errors.New("NetsiteType wrong!")
			}
		}
	}

	if len(*orgcode2) == 0 {
		return "", nil, errors.New("orgcode2 wrong!")
	}

	//check wacode
	var wacode string
	if len(*serial) == 0 { //随机生成场所序列号
		for {
			*serial = fmt.Sprintf("%04d", createRandNum(9999))
			wacode = *area + *busi + *netsite + *orgcode2 + *serial

			LOG_TRAC.Println("wacode :", wacode)

			netbar := &NetbarInfo{Wacode: wacode}
			if ex, err = X.Exist(netbar); err != nil {
				return "", nil, err
			} else if !ex {
				LOG_TRAC.Println("create random wacode success")
				break
			}
		}
	} else {
		wacode = *area + *busi + *netsite + *orgcode2 + *serial

		LOG_TRAC.Println("wacode :", wacode)

		netbar := &NetbarInfo{Wacode: wacode}
		if ex, err = X.Exist(netbar); err != nil {
			return "", nil, err
		} else if ex {
			return "", ErrorMap{"netbar_serialNO": RecordAlreadyExist}, errors.New("Wacode exists!")
		}
	}

	return wacode, nil, nil
}

/*
	POST时检查body参数是否合法
*/
func (this *NetbarsMngSt) checkPostBody(body_st *NetbarInfo) (ErrorMap, error) {
	var (
		err error
		ex  bool
	)

	if len(body_st.Orgcode) != ORGCODE_LEN {
		return ErrorMap{"security_software_orgcode": []string{LenWrong + strconv.Itoa(ORGCODE_LEN)}}, errors.New("org code wrong!")
	} else {
		org := &OrgInfo{Orgcode: body_st.Orgcode}
		if ex, err = X.Get(org); err != nil {
			LOG_ERRO.Println(err)
			return nil, err
		} else if !ex {
			return ErrorMap{"security_software_orgcode": []string{RecordNotExist}}, errors.New("Orgcode wrong!")
		}
		body_st.Orgcode_2 = org.Code
		body_st.OrgIndex = org.Id
		body_st.Orgname = org.Orgname
	}

	if len(body_st.AreaCode3) != AREACODE_LEN {
		return ErrorMap{"area_code_3": []string{BadAreaCode}}, errors.New("area code wrong!")
	}

	body_st.AreaCode1 = body_st.AreaCode3[0:2] + "0000"
	body_st.AreaCode2 = body_st.AreaCode3[0:4] + "00"

	if len(body_st.PlaceName) == 0 {
		return ErrorMap{"place_name": []string{EmptNotAllowed}}, errors.New("place name wrong!")
	}

	if len(body_st.SiteAddress) == 0 {
		return ErrorMap{"site_address": []string{EmptNotAllowed}}, errors.New("site address wrong!")
	}

	//check wacode
	var emap ErrorMap
	if body_st.Wacode, emap, err = checkWacode(
		&body_st.AreaCode3,
		&body_st.BusinessNature,
		&body_st.NetsiteType,
		&body_st.Orgcode_2,
		&body_st.NetbarSerialNO); err != nil {
		LOG_ERRO.Println(emap)
		return emap, err
	}

	if body_st.Approval != APPROVED && body_st.Approval != NOTAPPROVED {
		return ErrorMap{"approval": []string{TypeWrong}}, errors.New("Approval wrong!")
	}

	//审核时间为空时，默认为当前时间
	if body_st.Approval == APPROVED && body_st.ApprovalTime == 0 {
		LOG_INFO.Println("No approval time, use time now")
		body_st.ApprovalTime = time.Now().Unix()
	}

	return nil, nil
}

/*
	处理http DELETE请求
*/
func (this *NetbarsMngSt) httpDELETE() {
	var (
		index, count int64
		err          error
		//emap         ErrorMap
		ex bool
	)

	//检查url中index是否合法
	index, err = ParseIndexFromUrl(this.req)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(),
			ErrorMap{"idx": FormatWrong}, nil)
		return
	}

	//查询index是否存在于已删除表中
	if ex, err = X.Exist(&NetbarInfoDeleted{OriId: index}); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(),
			nil, nil)
		return
	}
	//删除已删除表中数据
	if ex {
		if _, err = X.Delete(&NetbarInfoDeleted{OriId: index}); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}

		SuccessResponse(this.writer, nil)
		return
	}

	//查询index是否存在于场所表
	if ex, err = X.ID(index).Exist(new(NetbarInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(),
			nil, nil)
		return
	}
	if !ex {
		LOG_ERRO.Println("index not exist!")
		ErrorResponse(this.writer, http.StatusBadRequest, "index not exist!",
			ErrorMap{"idx": RecordNotExist}, nil)
		return
	}

	//软删除场所
	devs := make([]DevInfo, 0)
	count, err = X.Where(fmt.Sprintf("netbar_index = %d", index)).FindAndCount(&devs)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Println("devices count=", count)
	if count > 0 { //下挂有设备
		LOG_TRAC.Println(devs)
		//循环删除场所下所有设备
		//TODO: 软删除
		for _, v := range devs {
			if err = doDELETE(v.Id, new(DevInfo).TableName(), new(DevInfoDeleted)); err != nil {
				LOG_ERRO.Println(err)
				ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
		}
	}

	if err = doDELETE(index, new(NetbarInfo).TableName(), new(NetbarInfoDeleted)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	SuccessResponse(this.writer, nil)

}

/*
	执行删除，场所/设备删除时将记录插入到deleted表
	table_name: 原表名称
	table_deleted: deleted表实例， 结构指针
	返回：插入删除表中的OriIndex和error
*/
func doDELETE(index int64, table_name string, table_deleted interface{}) error {
	var (
		err error
		ex  bool
	)

	//取出该记录
	if ex, err = X.Table(table_name).ID(index).Get(table_deleted); err != nil {
		return err
	} else if !ex {
		return errors.New("index not in database")
	}

	LOG_TRAC.Printf("%+v\n", table_deleted)

	//从NetbarInfo中删除并插入netbarInfoDeleted表中
	if _, err = X.Table(table_name).ID(index).Delete(table_deleted); err != nil {
		return err
	}

	//publish
	send2Publisher(ACTION_DELETE, nil, table_deleted)

	//设置删除时间
	time_field := reflect.ValueOf(table_deleted).Elem().FieldByName("DeletedTime")
	if time_field.CanSet() {
		time_field.SetInt(time.Now().Unix())
	} else {
		return errors.New("DeletedTime can not set!")
	}

	//设置approval
	approval_field := reflect.ValueOf(table_deleted).Elem().FieldByName("Approval")
	if approval_field.CanSet() {
		approval_field.SetInt(DELETED)
	} else {
		return errors.New("Approval can not set!")
	}

	//设置原表index
	index_field := reflect.ValueOf(table_deleted).Elem().FieldByName("OriId")
	if index_field.CanSet() {
		index_field.SetInt(index)
	} else {
		return errors.New("OriId can not set!")
	}

	if _, err = X.InsertOne(table_deleted); err != nil {
		return err
	}

	return nil
}

/*
	PATCH时检查body参数是否合法
*/
func (this *NetbarsMngSt) checkPatchBody(body_st *NetbarInfo) (ErrorMap, error) {
	if body_st.Approval != APPROVED &&
		body_st.Approval != NOTAPPROVED {
		return ErrorMap{"approval": TypeWrong}, errors.New("approval wrong")
	}

	return nil, nil
}

/*
	处理http PATCH请求
*/
func (this *NetbarsMngSt) httpPATCH() {
	var (
		count int64
		err   error
		index int64
		bd_b  []byte
		emap  ErrorMap
	)
	//解析body
	body := new(NetbarInfo)
	if bd_b, err = ioutil.ReadAll(this.req.Body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	if err = json.Unmarshal(bd_b, body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body is %+v\n", body)

	//检查请求body参数
	if emap, err = this.checkPatchBody(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
		return
	}

	//检查index是否合法
	if index, emap, err = CheckIndex(this.req, new(NetbarInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), emap, nil)
		return
	}
	LOG_TRAC.Println("index is ", index)

	//检查场所编码相关参数
	old := &NetbarInfo{Id: index}
	if _, err = X.Get(old); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	*body = *old
	json.Unmarshal(bd_b, body)

	var wacode_changed bool
	wa := new(NetbarInfo)
	if len(body.AreaCode3) != 0 && old.AreaCode3 != body.AreaCode3 {
		LOG_TRAC.Println("AreaCode3 changed")
		wacode_changed = true
		wa.AreaCode3 = body.AreaCode3
		body.AreaCode1 = body.AreaCode3[0:2] + "0000"
		body.AreaCode2 = body.AreaCode3[0:4] + "00"
	} else {
		wa.AreaCode3 = old.AreaCode3
	}

	if len(body.NetsiteType) != 0 && old.NetsiteType != body.NetsiteType {
		wacode_changed = true
		wa.NetsiteType = body.NetsiteType
	} else {
		wa.NetsiteType = old.NetsiteType
	}

	if len(body.BusinessNature) != 0 && old.BusinessNature != body.BusinessNature {
		wacode_changed = true
		wa.BusinessNature = body.BusinessNature
	} else {
		wa.BusinessNature = old.BusinessNature
	}

	LOG_TRAC.Printf("old body :%+v\n", old)
	if len(body.Orgcode) != 0 && old.Orgcode != body.Orgcode {
		wacode_changed = true
		LOG_TRAC.Println("orgcode changed!")
		org := &OrgInfo{}
		org.Orgcode = body.Orgcode
		if _, err = X.Get(org); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		body.Orgcode_2 = org.Code
		body.OrgIndex = org.Id
		wa.Orgcode_2 = org.Code
	} else {
		wa.Orgcode_2 = old.Orgcode_2
		body.Orgcode_2 = old.Orgcode_2
		body.OrgIndex = old.OrgIndex
	}

	if len(body.NetbarSerialNO) != 0 && old.NetbarSerialNO != body.NetbarSerialNO {
		wacode_changed = true
		wa.NetbarSerialNO = body.NetbarSerialNO
	} else {
		wa.NetbarSerialNO = old.NetbarSerialNO
	}

	if wacode_changed {
		//LOG_TRAC.Println("wacode changed")
		var emap ErrorMap
		if body.Wacode, emap, err = checkWacode(
			&wa.AreaCode3,
			&wa.BusinessNature,
			&wa.NetsiteType,
			&wa.Orgcode_2,
			&wa.NetbarSerialNO); err != nil {
			LOG_ERRO.Println(err, emap)
			ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
			return
		}
	} else {
		LOG_TRAC.Println("wacode not changed")
	}

	//审核时增加默认审核时间
	if old.Approval == NOTAPPROVED &&
		body.Approval == APPROVED &&
		body.ApprovalTime == 0 {
		body.ApprovalTime = time.Now().Unix()
	}
	//已审核设备变未审核，清空已审核时间
	if body.Approval == NOTAPPROVED &&
		old.Approval == APPROVED {
		body.ApprovalTime = 0
	}

	//执行更新场所
	ret_data, err := updateData(index, body)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	//场所编码或场所名称变化，更新下挂设备
	var need_update bool
	var wacode, placename string
	if len(body.Wacode) != 0 && body.Wacode != old.Wacode {
		need_update = true
		wacode = body.Wacode
	} else {
		wacode = old.Wacode
	}
	if len(body.PlaceName) != 0 && body.PlaceName != old.PlaceName {
		need_update = true
		placename = body.PlaceName
	} else {
		placename = old.PlaceName
	}
	if need_update {
		dev := make([]DevInfo, 0)
		if count, err = X.Where(fmt.Sprintf("netbar_index = %d", old.Id)).FindAndCount(&dev); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		} else if count > 0 {
			for _, v := range dev {
				v.Wacode = wacode
				v.AreaCode1 = wacode[0:2] + "0000"
				v.AreaCode2 = wacode[0:3] + "00"
				v.AreaCode3 = wacode[0:6]
				v.PlaceName = placename
				if _, err = updateData(v.Id, &v); err != nil {
					LOG_ERRO.Println(err)
					ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
					return
				}
				LOG_TRAC.Println("update ap:", v.ApId, "success")
			}
		} else {
			LOG_TRAC.Println("no ap needs to update!")
		}
	}

	SuccessResponse(this.writer, reflect.ValueOf(ret_data).Elem().Interface())
}

/*
	解析url参数中的approval，默认设为0
*/
func parseApproval(r *http.Request) (int, ErrorMap, error) {
	var (
		aprl int
		err  error
	)
	//解析参数
	if err = r.ParseForm(); err != nil {
		return -1, nil, err
	}

	if aprl_str := r.Form.Get("approval"); len(aprl_str) == 0 {
		LOG_INFO.Println("No approval in url")
		aprl = APPROVED
	} else if aprl, err = strconv.Atoi(aprl_str); err != nil {
		return -1, ErrorMap{"approval": []string{TypeWrong}}, errors.New("wrong approval format: " + aprl_str)
	} else if aprl != APPROVED && aprl != NOTAPPROVED && aprl != DELETED {
		return -1, ErrorMap{"approval": []string{TypeWrong}}, errors.New("wrong approval format: " + aprl_str)
	}

	return aprl, nil, nil
}

func NetbarsMngHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("NetbarsMngHandle")
	LOG_TRAC.Println("url:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	//获取approval，默认值设为"0"
	aprl, data, err := parseApproval(r)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), data, nil)
		return
	}

	mng := &NetbarsMngSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
		approval: aprl,
	}

	switch r.Method {
	case http.MethodGet:
		mng.httpGET()
	case http.MethodPost:
		mng.httpPOST()
	case http.MethodDelete:
		mng.httpDELETE()
	case http.MethodPatch:
		mng.httpPATCH()
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}

/*
	根据http请求中的file_name和file_type生成文件全名(带路径)
	文件名规则：时间戳-三位随机数.cvs
*/
func createFileName(file_type string) (string, string, error) {
	var (
		err             error
		ex              bool
		full_name, name string
	)

	if file_type != "csv" {
		return "", "", errors.New("unsupported file type :" + file_type)
	}

	for {
		name = fmt.Sprintf("%d-%03d", time.Now().Unix(), createRandNum(999)) +
			"." + file_type
		full_name = GlobalConfig.FileSysPath + "/" + name

		fs := &FileSys{FileName: full_name}
		if ex, err = X.Exist(fs); err != nil {
			return "", "", err
		} else if !ex {
			LOG_TRAC.Println("create random file name,", full_name)
			break
		}
	}

	return full_name, name, nil
}

//===================文件管理================
type fileRet struct {
	FileIdx     int64  `json:"file_idx"`
	DownloadUrl string `json:"download_url"`
	CreateTime  int64  `json:"create_time"`
	Duration    int64  `json:"duration"`
}

/*
	新建文件
*/
func (this *NetbarsMngSt) filePOST() {
	this.file_type = this.req.Form.Get("file")

	var (
		err     error
		code    int
		results *findResults
		fret    *fileRet
		body    *FilePostBody
	)
	//解析body
	if body, err = getFilePostBody(this.req.Body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body:%+v\n", body)

	//查询
	if body != nil {
		results, _, err = this.find(&body.DevicesGetBody)
	} else {
		results, _, err = this.find(nil)
	}
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, code, err.Error(), nil, nil)
		return
	}

	//生成文件
	if body != nil {
		fret, err = createFile(this.file_type, results.results, body.Fields)
	} else {
		fret, err = createFile(this.file_type, results.results, nil)
	}
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	SuccessResponse(this.writer, fret)
}

/*
	decode file POST body
*/
func getFilePostBody(data io.ReadCloser) (*FilePostBody, error) {

	var (
		err  error
		body = new(FilePostBody)
		bd_b []byte
	)

	/*
		if err := json.NewDecoder(data).Decode(body); err != nil {
			if err.Error() == "EOF" {
				LOG_INFO.Println("No body!")
			} else {
				return body, err
			}
		}*/
	if bd_b, err = ioutil.ReadAll(data); err != nil {
		return nil, err
	}

	if len(bd_b) == 0 {
		LOG_INFO.Println("No Body")
		return nil, nil
	}

	LOG_TRAC.Println(string(bd_b))

	if err = json.Unmarshal(bd_b, body); err != nil {
		return nil, err
	}

	return body, nil
}

/*
	data为结构slice
*/
func createFile(file_type string, data interface{}, fields OutFieldsSlice) (*fileRet, error) {
	var (
		file_name, name string
		err             error
	)

	//生成文件名
	file_name, name, err = createFileName(file_type)
	if err != nil {
		return nil, err
	}

	//写文件
	if err = WriteCSV(file_name, data, fields); err != nil {
		os.Remove(file_name)
		return nil, err
	}
	fs := &FileSys{FileName: file_name, FileType: file_type}
	_, err = X.InsertOne(fs)
	if err != nil {
		os.Remove(file_name)
		return nil, err
	}
	X.Get(fs)
	download_url := GlobalConfig.FileHost + "/" +
		GlobalConfig.ServiceName + "/" +
		GlobalConfig.Version +
		"/files/" + name
	return &fileRet{FileIdx: fs.Id,
		DownloadUrl: download_url,
		CreateTime:  time.Now().Unix(),
		Duration:    GlobalConfig.FileExpiredTime,
	}, nil
}

/*
	根据文件索引，删除文件和数据库中数据
*/
func deleteFile(index string) error {
	var (
		err error
		id  int64
		ex  bool
	)

	id, err = strconv.ParseInt(index, 10, 64)
	if err != nil {
		return err
	}

	fs := &FileSys{Id: id}
	if ex, err = X.Get(fs); err != nil {
		return err
	} else if !ex {
		return errors.New("file not exist!")
	}

	os.Remove(fs.FileName)
	X.Delete(fs)
	return nil
}

/*
	文件管理handle
*/
func NetbarsFileHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("NetbarsFileHandle")
	LOG_TRAC.Println("url:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	mng := &NetbarsMngSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
	}
	switch r.Method {
	case http.MethodPost:
		//获取approval，默认值设为已审核
		aprl, data, err := parseApproval(r)
		if err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), data, nil)
			return
		}

		mng.approval = aprl
		mng.filePOST()

	case http.MethodDelete:
		var (
			err error
			ok  bool
			v   string
		)

		params := gmux.Vars(r)
		if v, ok = params["file_idx"]; !ok {
			LOG_ERRO.Println("No file_idx in url")
			ErrorResponse(w, http.StatusBadRequest, "No file_idx in url", nil, nil)
			return
		}

		if err = deleteFile(v); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
		SuccessResponse(w, nil)

	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}
