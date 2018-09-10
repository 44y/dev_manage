//厂商管理
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
	"strconv"
	"strings"
)

type orgsMngSt struct {
	HttpSt
	file_type string
	file_name string //full name
}

/*
	处理http GET请求
*/
func (this *orgsMngSt) httpGET() {
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

	//查询
	results, code, err := this.find(&body.Scope)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, code, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Println("results:", results)
	QueryResponse(this.writer, results.total, results.count, results.results)

}

/*
	根据条件查询厂商
*/
func (this *orgsMngSt) find(body *DataScope) (*findResults, int, error) {

	//解析查询参数
	LOG_TRAC.Println(this.req.Form)
	keyword := this.req.Form.Get("keyword")
	offset, _ := strconv.Atoi(this.req.Form.Get("offset"))
	limit, _ := strconv.Atoi(this.req.Form.Get("limit"))
	if limit == 0 {
		limit = LIMIT_DEFAULT
	}
	scope := strings.Split(this.req.Form.Get("scope"), " ")
	LOG_TRAC.Println("limit:", limit, "offset:", offset, "scope:", scope)

	r := make([]OrgInfo, 0)
	rslt := &r //查询结果
	t := make([]OrgInfo, 0)
	rslt_tt := &t //查询结果

	var (
		total        int64
		err          error
		areacode_str string //根据地区编码组成查询字符串
		orgcode_str  string //根据厂商编码组成的查询字符串
	)

	if body != nil {
		//生成地区编码查询字符串
		if areacode_str, err = getAreaCodeStr(nil, &body.AreaCode); err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusBadRequest, err
		}
		LOG_TRAC.Println(areacode_str)

		for _, v := range body.Orgcodes {
			orgcode_str += fmt.Sprintf(" security_software_orgcode = %s or", v)
		}
		if len(orgcode_str) > 0 {
			orgcode_str = orgcode_str[1 : len(orgcode_str)-2]
		}
		LOG_TRAC.Println("orgcode_str:", orgcode_str)
	}

	var ss, ss_t *xorm.Session
	ss = X.Select("SQL_CALC_FOUND_ROWS *")
	ss_t = X.Select("SQL_CALC_FOUND_ROWS *")

	//orgcode查询条件
	if len(orgcode_str) > 0 {
		ss = ss.And(orgcode_str)
		ss_t = ss_t.And(orgcode_str)
	}

	//查询关键词
	if len(keyword) != 0 {
		//解析查询范围
		scope_keys := []string{"orgcode", "orgname"}
		scope_map := ParseScope(scope_keys, scope)
		LOG_TRAC.Println(scope_map)

		orgcode_str := "security_software_orgcode like " + " '%" + keyword + "%' "
		orgname_str := "security_software_orgname like " + " '%" + keyword + "%'"
		//查询厂商编码和名字
		if (scope_map["orgcode"] && scope_map["orgname"]) ||
			(!scope_map["orgcode"] && !scope_map["orgname"]) {

			ss = ss.And(orgcode_str + " or " + orgname_str)
			ss_t = ss_t.And(orgcode_str + " or " + orgname_str)

		} else if scope_map["orgcode"] { //查询厂商编码

			ss = ss.And(orgcode_str)
			ss_t = ss_t.And(orgcode_str)

		} else { //查询厂商名字

			ss = ss.And(orgname_str)
			ss_t = ss_t.And(orgname_str)
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

	ret := &findResults{
		count:   count,
		total:   total,
		results: reflect.ValueOf(rslt).Elem().Interface(),
	}

	return ret, 0, nil
}

/*
	执行sql查询语句，支持表：
	- OrgInfo;
	- netbarInfoApproved
	- NetbarInfoNotApproved
	- NetbarInfoDeleted
*/
func doQuery(sql, sql_total string, table_type interface{}) (*findResults, error) {
	//根据表类型创建slice
	t := reflect.TypeOf(table_type)
	sliceT := reflect.SliceOf(t)

	var sql_ret, total_results interface{}
	sql_ret = reflect.New(sliceT).Interface() //返回slice指针
	total_results = reflect.New(sliceT).Interface()
	//if err := X.SQL(sql).Find(sql_ret); err != nil {
	var count, total int
	if err := X.SQL(sql).Find(sql_ret); err != nil {
		LOG_ERRO.Println(err)
		return nil, err
	} else {
		//计算返回slice的数量
		ret_val := reflect.ValueOf(sql_ret).Elem()
		count = ret_val.Len()

		err := X.SQL(sql_total).Find(total_results)
		if err != nil {
			LOG_ERRO.Println(err)
			return nil, err
		} else {
			total_val := reflect.ValueOf(total_results).Elem()
			total = total_val.Len()
			LOG_TRAC.Println("count is ", count, "total is", total)

		}
	}
	ret := &findResults{
		count:   int64(count),
		total:   int64(total),
		results: reflect.ValueOf(sql_ret).Elem().Interface(),
	}
	return ret, nil
}

/*
	处理http POST请求
*/
func (this *orgsMngSt) httpPOST() {
	var (
		err   error
		exist bool
		emap  ErrorMap
	)

	//解析body
	body := new(OrgInfo)
	if err := json.NewDecoder(this.req.Body).Decode(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	LOG_TRAC.Printf("body is %+v\n", body)
	//检查body参数
	if emap, err = this.checkPostBody(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, emap)
		return
	}

	//查询记录是否已存在
	if exist, err = X.Exist(&OrgInfo{Orgcode: body.Orgcode}); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	} else if exist { //记录已存在
		LOG_ERRO.Println("Record alread exists!")
		ErrorResponse(this.writer, http.StatusConflict, "Record alread exists!", nil, nil)
		return
	}

	//插入记录
	if _, err = X.InsertOne(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	X.Get(body)
	LOG_TRAC.Println("body id=", body.Id)

	SuccessResponse(this.writer, *body)
}

/*
	POST时检查body参数是否合法
*/
func (this *orgsMngSt) checkPostBody(body_st *OrgInfo) (ErrorMap, error) {
	var (
		ex  bool
		err error
	)
	//orgcode
	if len(body_st.Orgcode) != ORGCODE_LEN {
		return ErrorMap{"security_software_orgcode": []string{LenWrong + strconv.Itoa(ORGCODE_LEN)}},
			errors.New("security_software_orgcode wrong!")
	}
	if ex, err = X.Exist(&OrgInfo{Orgcode: body_st.Orgcode}); err != nil {
		return nil, err
	}
	if ex {
		return ErrorMap{"security_software_orgcode": []string{RecordAlreadyExist}},
			errors.New("security_software_orgcode wrong!")
	}

	//orgname
	if len(body_st.Orgname) == 0 {
		return ErrorMap{"security_software_orgname": []string{EmptNotAllowed}},
			errors.New("security_software_orgname wrong!")
	}
	if ex, err = X.Exist(&OrgInfo{Orgname: body_st.Orgname}); err != nil {
		return nil, err
	}
	if ex {
		return ErrorMap{"security_software_orgname": []string{RecordAlreadyExist}},
			errors.New("security_software_orgname wrong!")
	}

	//address
	if len(body_st.Address) == 0 {
		return ErrorMap{"security_software_address": []string{EmptNotAllowed}},
			errors.New("security_software_address wrong!")
	}

	//code
	if len(body_st.Code) != CODE_LEN {
		return ErrorMap{"security_software_code": []string{EmptNotAllowed}},
			errors.New("security_software_code wrong!")
	}
	if ex, err = X.Exist(&OrgInfo{Code: body_st.Code}); err != nil {
		return nil, err
	}
	if ex {
		return ErrorMap{"security_software_code": []string{RecordAlreadyExist}},
			errors.New("security_software_code wrong!")
	}

	if len(body_st.Contactor) == 0 {
		return ErrorMap{"contactor": []string{EmptNotAllowed}},
			errors.New("contactor wrong!")
	}

	if len(body_st.ContactorTel) == 0 {
		return ErrorMap{"contactor_tel": []string{EmptNotAllowed}},
			errors.New("contactor_tel wrong!")
	}

	if len(body_st.ContactorMail) == 0 {
		return ErrorMap{"contactor_mail": []string{EmptNotAllowed}},
			errors.New("contactor_mail wrong!")
	}

	if body_st.EncryptId == 0 {
		body_st.EncryptId = 1
	}

	return nil, nil
}

/*
	处理http DELETE请求
	请求参数force为 "1"表示强制删除下挂所有场所
*/
func (this *orgsMngSt) httpDELETE() {
	var (
		index, count int64
		err          error
		emap         ErrorMap
	)

	//检查url中index是否合法
	if index, emap, err = CheckIndex(this.req, new(OrgInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), emap, nil)
		return
	}

	LOG_TRAC.Println("index is ", index)

	netbars := make([]NetbarInfo, 0)
	count, err = X.Where(fmt.Sprintf("org_index = %d", index)).FindAndCount(&netbars)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Println("netbars count=", count)
	if count > 0 { //有下挂场所
		LOG_TRAC.Println(netbars)
		//循环删除所有场所和场所下所有设备
		//TODO: 软删除
		for _, n_v := range netbars {
			devs := make([]DevInfo, 0)
			dev_count, err := X.Where(fmt.Sprintf("netbar_index = %d", n_v.Id)).FindAndCount(&devs)
			if err != nil {
				LOG_ERRO.Println(err)
				ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
			if dev_count > 0 {
				LOG_TRAC.Println("dev_count=", dev_count)
				LOG_TRAC.Println(devs)
				for _, d_v := range devs {
					if err = doDELETE(d_v.Id, new(DevInfo).TableName(), new(DevInfoDeleted)); err != nil {
						LOG_ERRO.Println(err)
						ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
						return
					}
				}
			}
			if err = doDELETE(n_v.Id, new(NetbarInfo).TableName(), new(NetbarInfoDeleted)); err != nil {
				LOG_ERRO.Println(err)
				ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
			send2Publisher(ACTION_DELETE, nil, n_v)
		}
	}

	if _, err = X.ID(index).Delete(new(OrgInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	SuccessResponse(this.writer, nil)
	return
}

/*
	处理http PATCH请求
*/
func (this *orgsMngSt) httpPATCH() {
	//检查url中index是否合法
	var (
		index, count int64
		err          error
		bd_b         []byte
		emap         ErrorMap
	)

	if index, emap, err = CheckIndex(this.req, new(OrgInfo)); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), emap, nil)
		return
	}

	LOG_TRAC.Println("index is ", index)

	//解析body
	body := new(OrgInfo)
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

	LOG_TRAC.Printf("body :%+v\n", body)

	old := &OrgInfo{Id: index}
	if _, err = X.Get(old); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	*body = *old
	json.Unmarshal(bd_b, body)

	//更新数据
	data, err := updateData(index, body)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	//orgcode或orgname改变，更新下挂场所;orgcode改变时 更新下挂设备apid
	var netbar_update, ap_update bool
	var orgcode, orgname string
	if len(body.Orgcode) != 0 && body.Orgcode != old.Orgcode {
		netbar_update = true
		ap_update = true
		orgcode = body.Orgcode
	} else {
		orgcode = old.Orgcode
	}

	if len(body.Orgname) != 0 && body.Orgname != old.Orgname {
		netbar_update = true
		orgname = body.Orgname
	} else {
		orgname = old.Orgname
	}
	if netbar_update {
		netbar := make([]NetbarInfo, 0)
		if count, err = X.Where(fmt.Sprintf("org_index = %d", old.Id)).FindAndCount(&netbar); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		} else if count > 0 {
			for _, v := range netbar {
				old_data := v
				v.Orgname = orgname
				v.Orgcode = orgcode
				if _, err = X.Update(&v); err != nil {
					ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
					return
				}
				LOG_TRAC.Println("update netbar:", v.Wacode, "success")
				send2Publisher(ACTION_MODIFY, v, old_data)

				if ap_update {
					aps := make([]DevInfo, 0)
					if count, err = X.Where("netbar_index = ?", v.Id).FindAndCount(&aps); err != nil {
						ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
						return
					} else if count > 0 {
						for _, ap := range aps {
							ap.ApId = orgcode + ap.ApId[ORGCODE_LEN:]
							X.Update(&ap)
						}
					}
				}
			}
		} else {
			LOG_TRAC.Println("no netbar needs to update!")
		}
	}

	SuccessResponse(this.writer, reflect.ValueOf(data).Elem().Interface())
	return
}

/*
	比较两个结构体，找出dst中不为0，src中为0的字段.（返回包含xorm tag的字段）
	src,dst都为结构指针
*/
func structCmp(dst, src interface{}) (UpdateMap, error) {
	dst_t := reflect.TypeOf(dst).Elem()
	src_t := reflect.TypeOf(src).Elem()

	if dst_t.Kind() != reflect.Struct || src_t.Kind() != reflect.Struct {
		return nil, errors.New("Only pointer to struct accpted!")
	}
	dst_v := reflect.ValueOf(dst).Elem()
	src_v := reflect.ValueOf(src).Elem()

	retMap := make(UpdateMap)

	for i := 0; i < src_t.NumField(); i++ {
		field_name := src_t.Field(i).Name
		//解析字段的数据库名称
		tag_str, ok := src_t.Field(i).Tag.Lookup("xorm")
		tag_sl := strings.Split(tag_str, "'")
		//LOG_TRAC.Println(tag_str)
		if !ok || len(tag_sl) < 2 {
			continue
		}
		tag_name := tag_sl[1]

		//在目的结构中找原结构中的字段，字段名和类型都要匹配
		for j := 0; j < dst_v.NumField(); j++ {
			if field_name == dst_t.Field(j).Name &&
				src_t.Field(i).Type.Kind() == dst_t.Field(j).Type.Kind() {

				dfv := dst_v.Field(j)
				sfv := src_v.Field(i)
				kd := src_t.Field(i).Type.Kind()
				switch kd {
				case reflect.Int, reflect.Int64:
					if dfv.Int() != 0 && sfv.Int() == 0 {
						retMap[tag_name] = 0
					}
				case reflect.Float32, reflect.Float64:
					if dfv.Float() != 0 && sfv.Float() == 0 {
						retMap[tag_name] = 0
					}
				case reflect.String:
					if dfv.String() != "" && sfv.String() == "" {
						retMap[tag_name] = ""
					}
				}
				break
			}
		}
	}
	return retMap, nil

}

/*
	更新数据到数据库，
	data 为结构指针
	支持表：
	- OrgInfo;
	- netbarInfoApproved
	- NetbarInfoNotApproved
	- NetbarInfoDeleted
*/
func updateData(index int64, data interface{}) (interface{}, error) {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return nil, errors.New("Only pointer to struct accepted")
	}
	//根据表类型创建slice
	t := reflect.TypeOf(data).Elem()
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Only pointer to struct accepted")
	}
	//sliceT := reflect.SliceOf(t)
	//old := reflect.New(sliceT).Interface() //返回slice指针
	old := reflect.New(t).Interface()

	var err error

	//根据index查找数据库
	ex, err := X.ID(index).Get(old)
	if err != nil {
		return nil, err
	} else if !ex {
		return nil, fmt.Errorf("invalid records, index :%d,%v\n", index, old)
	}

	upd_map, err := structCmp(old, data)
	if err != nil {
		LOG_ERRO.Println(err)
		return nil, err
	}
	LOG_TRAC.Println("upd_map:", upd_map)

	//更新数据
	if _, err = X.ID(index).Update(data); err != nil {
		LOG_ERRO.Println(err)
		return nil, err
	}

	if len(upd_map) > 0 {
		//更新0值字段
		if _, err = X.Table(reflect.New(t).Interface()).ID(index).Update(upd_map); err != nil {
			LOG_ERRO.Println(err)
			return nil, err
		}
	}
	send2Publisher(ACTION_MODIFY, data, old)

	new_data := reflect.New(t).Interface()
	X.ID(index).Get(new_data)

	return new_data, nil
}

/*
	从url中解析index，返回int64
*/
func ParseIndexFromUrl(req *http.Request) (int64, error) {
	params := gmux.Vars(req)

	var (
		v  string
		ok bool
	)
	if v, ok = params["idx"]; !ok {
		return -1, errors.New("No index in url")
	}

	id, err := strconv.Atoi(v)
	if err != nil {
		return -1, errors.New("Invalid index")
	}

	return int64(id), nil
}

/*
	检查一个url中的index是否存在于数据库中，table_type为数据库表实例
	返回：
	-存在：index值和nil
	-不存在：-1和错误
*/
func CheckIndex(req *http.Request, table_type interface{}) (int64, ErrorMap, error) {
	var (
		id    int64
		err   error
		exist bool
	)

	id, err = ParseIndexFromUrl(req)
	if err != nil {
		LOG_ERRO.Println(err)
		return -1, ErrorMap{"idx": FormatWrong}, err
	}

	LOG_TRAC.Println("id:", id)
	//查询数据库
	if exist, err = X.ID(id).Exist(table_type); err != nil {
		LOG_ERRO.Println(err)
		return -1, nil, err
	} else if exist {
		return id, nil, nil
	} else {
		return 0, ErrorMap{"idx": RecordNotExist}, fmt.Errorf("index not in database\n")
	}

}

/*
	厂商管理处理handle
*/
func OrgsMngHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI, "method:", r.Method, r.URL.Path)

	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	mng := &orgsMngSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
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
		LOG_ERRO.Println("Method not supported: "+r.Method, "")
	}
}

//===================文件管理================
/*
	新建文件
*/
func (this *orgsMngSt) filePOST() {
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
func OrgsFileHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI, "method:", r.Method)

	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	//解析参数
	if err := r.ParseForm(); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	mng := &orgsMngSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
	}
	switch r.Method {
	case http.MethodPost:
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
