//设备管理
package managers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-xorm/xorm"
	gmux "github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	. "rc"
	"reflect"
	//"regexp"
	"strconv"
	"strings"
	"time"
)

type APsMngSt struct {
	HttpSt
	approval  int
	file_type string
	file_name string //full name
}

/*
	处理http GET请求或 生成查询文件
	rc- 0:查询， 1:生成查询文件
*/
func (this *APsMngSt) httpGET() {
	LOG_TRAC.Println("httpGet")
	//解析body
	body := new(DevicesGetBody)
	if err := json.NewDecoder(this.req.Body).Decode(body); err != nil {
		if err.Error() == "EOF" {
			LOG_INFO.Println("No get body!")
		} else {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
	}
	LOG_TRAC.Printf("body:%+v\n", body)

	//查询
	results, code, err := this.find(&body.Scope)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, code, err.Error(), nil, nil)
		return
	}
	QueryResponse(this.writer, results.total, results.count, results.results)
}

/*
	根据条件查询设备
*/
func (this *APsMngSt) find(body *DataScope) (*findResults, int, error) {
	LOG_TRAC.Println(this.req.Form)
	var (
		total, index  int64
		ex            bool
		err           error
		rslt, rslt_tt interface{} //slice类型指针
		//tbl          interface{} //表结构实例指针
		isDeleted    bool   //true表示查询deleted表
		areacode_str string //根据地区编码组成查询字符串
		wacode_str   string //根据厂商编码组成查询字符串
		//wacode_str1  string //根据url中wacode参数指定的场所编码组成查询字符串
	)
	//检查url中index是否合法，合法则为精确查询index
	index, err = ParseIndexFromUrl(this.req)
	if err == nil && index != 0 {
		ap := &DevInfo{Id: index}
		ex, err = X.Get(ap)
		if err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusInternalServerError, err
		}
		if ex {
			sl := make([]DevInfo, 0)
			sl = append(sl, *ap)
			return &findResults{count: 1, total: 1, results: sl}, 0, nil
		}

		del_ap := &DevInfoDeleted{OriId: index}
		ex, err = X.Get(del_ap)
		if err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusInternalServerError, err
		}
		if ex {
			sl := make([]DevInfoDeleted, 0)
			sl = append(sl, *del_ap)
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

	wacode := this.req.Form.Get("wacode")

	if body != nil {
		//生成地区编码查询字符串
		if areacode_str, err = getAreaCodeStr(nil, &body.AreaCode); err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusBadRequest, err
		}
		LOG_TRAC.Println(areacode_str)

		//根据scope中的厂商编码查询场所编码进行查询限制
		for _, v := range body.Orgcodes {
			netbars := make([]NetbarInfo, 0)
			if err = X.Where("security_software_orgcode = ?", v).Find(&netbars); err != nil {
				LOG_ERRO.Println(err)
				return nil, http.StatusBadRequest, err
			}
			for _, n := range netbars {
				wacode_str += fmt.Sprintf(" netbar_wacode like '%s' or", n.Wacode)
			}

		}
		if len(wacode_str) > 0 {
			wacode_str = wacode_str[1 : len(wacode_str)-2]
		}
		LOG_TRAC.Println("wacode_str:", wacode_str)
	}

	//根据approval指定查询条件
	switch this.approval {
	case APPROVED, NOTAPPROVED: //查询NetbarInfo
		r := make([]DevInfo, 0)
		rslt = &r
		t := make([]DevInfo, 0)
		rslt_tt = &t
		//tbl = new(DevInfo)
		isDeleted = false
	case DELETED: //NetbarInfoDeleted
		r := make([]DevInfoDeleted, 0)
		rslt = &r
		t := make([]DevInfoDeleted, 0)
		rslt_tt = &t
		//tbl = new(DevInfoDeleted)
		isDeleted = true
	default:
		LOG_ERRO.Printf("Unknown approval value:", this.approval)
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

	//wacode查询条件1
	if len(wacode_str) > 0 {
		ss = ss.And(wacode_str)
		ss_t = ss_t.And(wacode_str)
	}

	//wacode查询条件2
	if wacode != "" {
		ss = ss.And(" netbar_wacode = ?", wacode)
		ss_t = ss_t.And(" netbar_wacode = ?", wacode)
	}

	//查询关键字
	if len(keyword) != 0 {
		scope_keys := []string{"ap_id", "ap_mac"} //支持的查询关键字
		scope_map := ParseScope(scope_keys, scope)
		LOG_TRAC.Println(scope_map)

		apid_str := "ap_id like " + " '%" + keyword + "%' "
		apmac_str := "ap_mac like " + " '%" + keyword + "%'"

		//查询ap id和ap mac
		if (scope_map["ap_id"] && scope_map["ap_mac"]) ||
			(!scope_map["ap_id"] && !scope_map["ap_mac"]) {

			ss = ss.And(apid_str + " or " + apmac_str)
			ss_t = ss_t.And(apid_str + " or " + apmac_str)

		} else if scope_map["ap_id"] { //查询ap RegId

			ss = ss.And(apid_str)
			ss_t = ss_t.And(apid_str)

		} else if scope_map["ap_mac"] { //查询ap mac

			ss = ss.And(apmac_str)
			ss_t = ss_t.And(apmac_str)
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
		err = ss_t.Find(rslt_tt)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		total = int64(reflect.ValueOf(rslt_tt).Elem().Len())
	}

	rslt_v := reflect.ValueOf(rslt).Elem().Interface()

	/*
		//更新设备状态
		time_now := time.Now().Unix()
		if v, ok := rslt_v.([]DevInfo); ok {
			rsl_n := make([]DevInfo, 0)
			for _, n := range v {
				if time_now-n.LastDataTime > GlobalConfig.ApDataTimeout {
					n.DataStatus = STATUS_ABNORMAL
				} else {
					n.DataStatus = STATUS_NORMAL
				}

				if time_now-n.LastAliveTime > GlobalConfig.ApAliveTimeout {
					n.AliveStatus = STATUS_ABNORMAL
				} else {
					n.AliveStatus = STATUS_NORMAL
				}

				if time_now-n.LastBasicTime > GlobalConfig.APBasicTimeout {
					n.BasicStatus = STATUS_ABNORMAL
				} else {
					n.BasicStatus = STATUS_NORMAL
				}

				rsl_n = append(rsl_n, n)
			}
			rslt_v = rsl_n
		}
	*/
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
func (this *APsMngSt) httpPOST() {
	var (
		err    error
		netbar *NetbarInfo
	)
	//解析body
	body := new(DevInfo)
	if err = json.NewDecoder(this.req.Body).Decode(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body is %+v\n", body)
	//检查body参数
	var emap ErrorMap
	if netbar, emap, err = this.checkPostBody(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
		return
	}
	body.DataStatus = STATUS_ABNORMAL
	body.AliveStatus = STATUS_ABNORMAL
	body.DeviceStatus = DEVICE_OFFLINE
	body.BasicStatus = STATUS_ABNORMAL

	if _, err = X.InsertOne(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	X.Get(body)

	//更新场所表
	netbar.DevTotalNum++
	netbar.DevOfflineNum++
	if _, err = updateData(netbar.Id, netbar); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	send2Publisher(ACTION_ADD, body, nil)
	SuccessResponse(this.writer, *body)

	/*
		bean := make([]DevInfo, 0)
		//插入设备表
		if ret_data, err = doPOST(body, &bean); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}

		SuccessResponse(this.writer, reflect.ValueOf(ret_data).Elem().Interface())
	*/
}

/*
	aa:bb:cc --> AABBCC
*/
func transMacFormat1(ori_mac string) string {
	return strings.ToUpper(ori_mac[0:2] +
		ori_mac[3:5] +
		ori_mac[6:8] +
		ori_mac[9:11] +
		ori_mac[12:14] +
		ori_mac[15:17])
}

/*
	AA-BB-CC --> aa:bb:cc
*/
func transMacFormat2(ori_mac string) string {
	return strings.ToLower(ori_mac[0:2] + ":" +
		ori_mac[3:5] + ":" +
		ori_mac[6:8] + ":" +
		ori_mac[9:11] + ":" +
		ori_mac[12:14] + ":" +
		ori_mac[15:17])
}

/*
	aa:bb:cc --> AA-BB-CC
*/
func transMacFormat3(ori_mac string) string {
	return strings.ToUpper(ori_mac[0:2] + "-" +
		ori_mac[3:5] + "-" +
		ori_mac[6:8] + "-" +
		ori_mac[9:11] + "-" +
		ori_mac[12:14] + "-" +
		ori_mac[15:17])
}

/*
	POST时检查body参数是否合法
	返回所属场所信息和错误
*/
func (this *APsMngSt) checkPostBody(body_st *DevInfo) (*NetbarInfo, ErrorMap, error) {
	if len(body_st.Wacode) != WACODE_LEN {
		return nil, ErrorMap{"netbar_wacode": []string{LenWrong + strconv.Itoa(WACODE_LEN)}}, errors.New("NETBAR wacode wrong!")
	}

	//所属场所是否存在
	netbar := &NetbarInfo{Wacode: body_st.Wacode}
	var (
		err error
		ex  bool
	)
	if ex, err = X.Get(netbar); err != nil {
		return nil, nil, err
	} else if !ex {
		return nil, ErrorMap{"netbar_wacode": []string{RecordNotExist}}, errors.New("netbar not exists: " + body_st.Wacode)
	}

	body_st.PlaceName = netbar.PlaceName
	body_st.NetbarIndex = netbar.Id

	body_st.AreaCode1 = netbar.Wacode[0:2] + "0000"
	body_st.AreaCode2 = netbar.Wacode[0:4] + "00"
	body_st.AreaCode3 = netbar.Wacode[0:6]

	if len(body_st.ApMac) != MAC_LEN {
		return nil, ErrorMap{"ap_mac": []string{LenWrong + strconv.Itoa(MAC_LEN)}}, errors.New("ap mac wrong!")
	}
	body_st.ApMac = transMacFormat3(body_st.ApMac)

	body_st.ApId = netbar.Orgcode + transMacFormat1(body_st.ApMac)
	LOG_TRAC.Println("new ap id is ", body_st.ApId)

	aptype := &ApType{Code: strconv.Itoa(body_st.Type)}
	if ex, err = X.Exist(aptype); err != nil {
		return nil, ErrorMap{"type": []string{TypeWrong}}, err
	} else if !ex {
		return nil, ErrorMap{"type": []string{TypeWrong}}, errors.New("type not exists")
	}

	if body_st.Approval != APPROVED && body_st.Approval != NOTAPPROVED {
		return nil, ErrorMap{"approval": []string{TypeWrong}}, errors.New("Approval wrong!")
	}

	// 设备已审核时，所属场所必须为已审核；
	// 审核时间为空，则填上当前时间
	if body_st.Approval == APPROVED {
		if netbar.Approval != APPROVED {
			return nil, ErrorMap{"netbar_wacode": NetbarNotApproved},
				errors.New("can't approve device if its netbar haven't approved")
		}
		if body_st.ApprovalTime == 0 {
			LOG_INFO.Println("No approval time, use time now")
			body_st.ApprovalTime = time.Now().Unix()
		}
	}

	//经纬度为空时，使用场所经纬度
	if body_st.Latitude == 0 && body_st.Longitude == 0 {
		body_st.Longitude = netbar.Longitude
		body_st.Latitude = netbar.Latitude
	}

	return netbar, nil, nil
}

/*
	处理http DELETE请求
*/
func (this *APsMngSt) httpDELETE() {
	var (
		err   error
		ex    bool
		index int64
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
	if ex, err = X.Exist(&DevInfoDeleted{OriId: index}); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(),
			nil, nil)
		return
	}
	if ex {
		X.Delete(&DevInfoDeleted{OriId: index})
		SuccessResponse(this.writer, nil)
		return
	}

	//查询index是否存在于设备表
	if ex, err = X.ID(index).Exist(new(DevInfo)); err != nil {
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

	//软删除设备

	err = doDELETE(index, new(DevInfo).TableName(), new(DevInfoDeleted))
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	SuccessResponse(this.writer, nil)

}

/*
	PATCH时检查body参数是否合法
*/
func (this *APsMngSt) checkPatchBody(body_st *DevInfo) (ErrorMap, error) {
	var (
		ex  bool
		err error
	)

	if len(body_st.ApMac) > 0 {
		if len(body_st.ApMac) != MAC_LEN {
			return ErrorMap{"ap_mac": []string{LenWrong + strconv.Itoa(MAC_LEN)}}, errors.New("ap mac wrong!")
		}
	}

	if len(body_st.Wacode) > 0 && len(body_st.Wacode) != WACODE_LEN {
		return ErrorMap{"netbar_wacode": FormatWrong}, errors.New("netbar wacode wrong!")
	}

	if body_st.Type != 0 {
		aptype := &ApType{Code: strconv.Itoa(body_st.Type)}
		if ex, err = X.Exist(aptype); err != nil {
			return ErrorMap{"type": []string{TypeWrong}}, err
		} else if !ex {
			return ErrorMap{"type": []string{TypeWrong}}, errors.New("type not exists")
		}
	}

	if body_st.Approval != APPROVED &&
		body_st.Approval != NOTAPPROVED {
		return ErrorMap{"approval": TypeWrong}, errors.New("approval wrong")
	}

	return nil, nil
}

/*
	处理http PATCH请求
*/
func (this *APsMngSt) httpPATCH() {
	var (
		err   error
		index int64
		bd_b  []byte
		emap  ErrorMap
	)

	//解析body
	body := new(DevInfo)
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

	//检查index是否合法
	if index, emap, err = CheckIndex(this.req, new(DevInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), emap, nil)
		return
	}
	LOG_TRAC.Println("index is ", index)

	//检查请求body参数
	if emap, err = this.checkPatchBody(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
		return
	}

	old := &DevInfo{Id: index}
	if _, err = X.Get(old); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	//LOG_TRAC.Printf("old:%+v\n", old)
	*body = *old
	json.Unmarshal(bd_b, body)
	//LOG_TRAC.Printf("body:%+v\n", body)
	//wacode或ap mac改变，更新ap id等
	var need_update, ex bool
	var apmac, orgcode string

	netbar := &NetbarInfo{Wacode: body.Wacode}
	if ex, err = X.Get(netbar); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	} else if !ex {
		LOG_ERRO.Println("wacode not exists:", body.Wacode)
		ErrorResponse(this.writer, http.StatusBadRequest, "wacode not exists:"+body.Wacode, nil,
			ErrorMap{"netbar_wacode": RecordNotExist})
		return
	}

	if len(body.Wacode) != 0 && body.Wacode != old.Wacode {
		need_update = true
		body.PlaceName = netbar.PlaceName
		body.NetbarIndex = netbar.Id
		body.AreaCode1 = body.Wacode[0:2] + "0000"
		body.AreaCode2 = body.Wacode[0:4] + "00"
		body.AreaCode3 = body.Wacode[0:6]
		orgcode = netbar.Orgcode
	} else {
		netbar := &NetbarInfo{Id: old.NetbarIndex}
		if _, err = X.Get(netbar); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		orgcode = netbar.Orgcode
	}
	if len(body.ApMac) != 0 && body.ApMac != old.ApMac {
		body.ApMac = transMacFormat3(body.ApMac)
		LOG_TRAC.Println("apmac:", body.ApMac)
		need_update = true
		apmac = body.ApMac
	} else {
		apmac = old.ApMac
	}
	if need_update {
		body.ApId = orgcode + transMacFormat1(apmac)
		LOG_TRAC.Println("new ap id is ", body.ApId)
	}

	// 审核设备时，所属场所必须为已审核；
	// 审核时间为空，则填上当前时间
	if body.Approval == APPROVED &&
		old.Approval == NOTAPPROVED {
		if netbar.Approval != APPROVED {
			LOG_ERRO.Println("can't approve device if its netbar haven't approved")
			ErrorResponse(this.writer, http.StatusBadRequest, "AP所属场所未审核！", nil,
				ErrorMap{"netbar_wacode": NetbarNotApproved})
			return
		}

		if body.ApprovalTime == 0 {
			body.ApprovalTime = time.Now().Unix()
		}
	}
	//已审核设备变未审核，清空已审核时间
	if body.Approval == NOTAPPROVED &&
		old.Approval == APPROVED {
		body.ApprovalTime = 0
	}

	//执行更新
	if ret_data, err := updateData(index, body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	} else {
		SuccessResponse(this.writer, reflect.ValueOf(ret_data).Elem().Interface())
		return
	}

	SuccessResponse(this.writer, http.StatusOK)

}

func APsMngHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("APsMngHandle")
	LOG_TRAC.Printf(r.RequestURI)

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

	mng := &APsMngSt{
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

//===================文件管理================
/*
	新建文件
*/
func (this *APsMngSt) filePOST() {
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
		results, code, err = this.find(&body.Scope)
	} else {
		results, code, err = this.find(nil)
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
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	SuccessResponse(this.writer, fret)
}

/*
	文件管理handle
*/
func ApsFileHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI, "method:", r.Method)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	mng := &APsMngSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
	}
	switch r.Method {
	case http.MethodPost:
		//获取approval，默认值设为0
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
