//网安编码管理
package managers

import (
	"encoding/json"
	"errors"
	gmux "github.com/gorilla/mux"
	"net/http"
	. "rc"
	"reflect"
	"strconv"
)

const (
	IMTYPE_TYPE_IM        = "im"        //聊天工具
	IMTYPE_TYPE_GAME      = "game"      //游戏
	IMTYPE_TYPE_MAIL      = "mail"      //邮箱
	IMTYPE_TYPE_FORUM     = "forum"     //论坛
	IMTYPE_TYPE_ECOMMERCE = "ecommerce" //电子商务
	IMTYPE_TYPE_FRIEND    = "friend"    //交友
	IMTYPE_TYPE_VIDEO     = "video"     //视频
	IMTYPE_TYPE_MUSIC     = "music"     //音乐
	IMTYPE_TYPE_TRAVEL    = "travel"    //旅行
	IMTYPE_TYPE_NEWS      = "news"      //新闻
	IMTYPE_TYPE_LIFE      = "life"      //生活服务
	IMTYPE_TYPE_TRANS     = "trans"     //传输

	IMTYPE_NAME_IM        = "聊天工具" //聊天工具
	IMTYPE_NAME_GAME      = "游戏"   //游戏
	IMTYPE_NAME_MAIL      = "邮箱"   //邮箱
	IMTYPE_NAME_FORUM     = "论坛"   //论坛
	IMTYPE_NAME_ECOMMERCE = "电子商务" //电子商务
	IMTYPE_NAME_FRIEND    = "交友"   //交友
	IMTYPE_NAME_VIDEO     = "视频"   //视频
	IMTYPE_NAME_MUSIC     = "音乐"   //音乐
	IMTYPE_NAME_TRAVEL    = "旅行"   //旅行
	IMTYPE_NAME_NEWS      = "新闻"   //新闻
	IMTYPE_NAME_LIFE      = "生活服务" //生活服务
	IMTYPE_NAME_TRANS     = "传输"   //传输

	IMTYPE_NAME_LEN_MAX = 128 //name最大长度
	IMTYPE_CODE_LEN_MAX = 10  //code最大长度
)

/*
	检查各种类型的code POST body是否合法，包括name和code
*/
func checkCodePostBody(body interface{}) (ErrorMap, error) {
	var (
		ex  bool
		err error
	)

	t := reflect.TypeOf(body)
	v := reflect.ValueOf(body)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		//LOG_ERRO.Println("body is neither struct nor pointer to struct")
		return nil, errors.New("body is neither struct nor pointer to struct")
	}

	//检查name是否为空
	name := v.FieldByName("Name").String()
	if len(name) == 0 {
		return ErrorMap{"name": []string{EmptNotAllowed}}, errors.New("name is NULL")
	}
	if len(name) > IMTYPE_NAME_LEN_MAX {
		return ErrorMap{"name": []string{TooLong}}, errors.New("name is too long")
	}

	//检查name是否已存在
	tmp := reflect.New(t)
	tmp.Elem().FieldByName("Name").SetString(name)

	LOG_TRAC.Println(tmp)
	if ex, err = X.Exist(tmp.Interface()); err != nil {
		return nil, err
	}
	if ex {
		return ErrorMap{"name": []string{RecordAlreadyExist}}, errors.New("name alread exists:" + name)
	}

	//检查code是否为空
	code := v.FieldByName("Code").String()
	if len(code) == 0 {
		return ErrorMap{"code": []string{EmptNotAllowed}}, errors.New("code is NULL")
	}

	if val, ok := v.Interface().(ImType); ok {
		//检查code
		if len(val.Code) > IMTYPE_CODE_LEN_MAX {
			return ErrorMap{"code": []string{TooLong}}, errors.New("code is too long")
		}

		tmp = reflect.New(t)
		tmp.Elem().FieldByName("Code").SetString(val.Code)
		if ex, err = X.Exist(tmp.Interface()); err != nil {
			return nil, err
		}
		if ex {
			return ErrorMap{"code": []string{RecordAlreadyExist}}, errors.New("code alread exists:" + val.Code)
		}

		//检查type
		tmp = reflect.New(t)
		if ex, err = X.Where("type = ?", val.Type).Exist(tmp.Interface()); err != nil {
			return nil, err
		}
		if !ex {
			return ErrorMap{"type": []string{RecordNotExist}}, errors.New("type not exists")
		}

		//检查type name
		tmp = reflect.New(t)
		if ex, err = X.Where("type_name = ?", val.TypeName).Exist(tmp.Interface()); err != nil {
			return nil, err
		}
		if !ex {
			return ErrorMap{"type_name": []string{RecordNotExist}}, errors.New("type_name not exists")
		}

	}
	//TODO:增加检查body函数，根据类型检查code

	return nil, nil
}

/*
	bean为各表结构的指针
*/
func wacodeMng(w http.ResponseWriter, r *http.Request, bean interface{}) {
	var (
		err              error
		ex               bool
		count, id, total int64
		emap             ErrorMap
	)
	//创建表类型结构的slice
	t := reflect.TypeOf(bean).Elem()
	data_slice := reflect.New(reflect.SliceOf(t)).Interface()
	data := reflect.New(t).Interface()
	body := reflect.New(t).Interface()

	//解析参数
	code := r.Form.Get("code")
	LOG_TRAC.Println(r.Form.Get("limit"))
	limit, _ := strconv.Atoi(r.Form.Get("limit"))
	offset, _ := strconv.Atoi(r.Form.Get("offset"))
	if limit == 0 {
		limit = LIMIT_DEFAULT
	}

	switch r.Method {
	case http.MethodGet:
		//查询全部
		if code == "" {

			if err = X.Select("SQL_CALC_FOUND_ROWS *").
				Limit(limit, offset).
				Find(data_slice); err != nil {

				LOG_ERRO.Println(err)
				ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
			count = int64(reflect.ValueOf(data_slice).Elem().Len())
			if count == 0 {
				total = 0
			} else {
				var qret []map[string][]byte
				qret, err = X.Query("select FOUND_ROWS()")
				if err != nil {
					LOG_ERRO.Println(err)
					ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
					return
				}
				LOG_TRAC.Println("qret:", qret)
				total, _ = strconv.ParseInt(string(qret[0]["FOUND_ROWS()"]), 10, 64)
			}
			QueryResponse(w, total, count, reflect.ValueOf(data_slice).Elem().Interface())
		} else { //查询指定编码
			if err = X.Where("code = ?", code).Find(data_slice); err != nil {

				LOG_ERRO.Println(err)
				ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
			count = int64(reflect.ValueOf(data_slice).Elem().Len())
			total = count
			QueryResponse(w, total, count, reflect.ValueOf(data_slice).Elem().Interface())
		}

	case http.MethodPost: //新增
		if err = json.NewDecoder(r.Body).Decode(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}

		//检查body参数
		if emap, err = checkCodePostBody(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, emap)
			return
		}

		name := reflect.ValueOf(body).Elem().FieldByName("Name").String()
		if len(name) == 0 {
			LOG_ERRO.Println("name is NULL!")
			ErrorResponse(w, http.StatusBadRequest, "name is NULL!", nil, nil)
			return
		}

		if ex, err = X.Exist(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		} else if ex {
			LOG_ERRO.Println("record exists!")
			ErrorResponse(w, http.StatusConflict, "record exists!", nil, nil)
			return
		}

		if _, err = X.InsertOne(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		X.Get(body)
		SuccessResponse(w, reflect.ValueOf(body).Elem().Interface())

	case http.MethodDelete:
		if code == "" {
			LOG_ERRO.Println("code is NULL!")
			ErrorResponse(w, http.StatusBadRequest, "code is NULL!", ErrorMap{"code": []string{EmptNotAllowed}}, nil)
			return
		}

		code_field := reflect.ValueOf(data).Elem().FieldByName("Code")
		if code_field.CanSet() {
			code_field.SetString(code)
		}
		if _, err = X.Delete(data); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		SuccessResponse(w, nil)

	case http.MethodPatch:
		if code == "" {
			LOG_ERRO.Println("code is NULL!")
			ErrorResponse(w, http.StatusBadRequest, "code is NULL!", nil, nil)
			return
		}

		code_field := reflect.ValueOf(data).Elem().FieldByName("Code")
		if code_field.CanSet() {
			code_field.SetString(code)
		}
		if ex, err = X.Get(data); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		} else if !ex {
			LOG_ERRO.Println("record not exist,", code)
			ErrorResponse(w, http.StatusBadRequest, "record not exist,"+code, nil, nil)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}

		if err = CopyStruct(body, reflect.ValueOf(data).Elem().Interface()); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}

		if id, err = X.Update(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		id_field := reflect.ValueOf(body).Elem().FieldByName("Id")
		if id_field.CanSet() {
			id_field.SetInt(id)
		}
		SuccessResponse(w, reflect.ValueOf(body).Elem().Interface())

	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}

//查询im type返回的结构
type ImTypeCategory struct {
	Name     string   `json:"name"`
	Key      string   `json:"key"`
	Children []ImType `json:"children"`
}

/*
	查询特定type的im type
*/
func getImTypeByType(t string) (*ImTypeCategory, error) {
	ret := ImTypeCategory{Key: t, Children: make([]ImType, 0)}
	err := X.Where("type = ?", t).Find(&ret.Children)
	if err != nil {
		return nil, err
	}
	if len(ret.Children) > 0 {
		ret.Name = ret.Children[0].TypeName

	}
	return &ret, nil
}

/*
	im type管理
*/
func imTypeMng(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		count int64
		emap  ErrorMap
		imtc  *ImTypeCategory
	)

	//解析参数
	code := r.Form.Get("code")
	//limit := r.Form.Get("limit")
	//offset := r.Form.Get("offset")

	switch r.Method {
	case http.MethodGet:
		rslt := make([]*ImTypeCategory, 0)
		//var total int64
		//查询全部
		if code == "" {
			imtype_total := make([]ImType, 0)
			if count, err = X.FindAndCount(&imtype_total); err != nil {
				LOG_ERRO.Println(err)
				ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
				return
			}
			if count == 0 {
				LOG_TRAC.Println("not records for code:", code)
				QueryResponse(w, 0, 0, nil)
				return
			}

			//分类
			var (
				i   int
				cat *ImTypeCategory
				has bool
			)
			for _, v := range imtype_total {
				has = false
				for i, cat = range rslt {
					if v.Type == cat.Key {
						has = true
						break
					}
				}
				if has {
					rslt[i].Children = append(rslt[i].Children, v)
				} else {
					imtc = &ImTypeCategory{
						Children: make([]ImType, 0),
						Name:     v.TypeName,
						Key:      v.Type,
					}
					imtc.Children = append(imtc.Children, v)
					LOG_TRAC.Println("new type:", v.Type)
					rslt = append(rslt, imtc)
				}
			}

			QueryResponse(w, count, count, rslt)
			return
		}
		//查询特定code
		imtc = &ImTypeCategory{Children: make([]ImType, 0)}
		if count, err = X.Where("code = ?", code).FindAndCount(&imtc.Children); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		if count == 0 {
			LOG_TRAC.Println("not records for code:", code)
			QueryResponse(w, 0, 0, nil)
			return
		}

		imtc.Name = imtc.Children[0].TypeName
		imtc.Key = imtc.Children[0].Type
		rslt = append(rslt, imtc)

		QueryResponse(w, count, count, rslt)
		return

	case http.MethodPost:
		body := &ImType{}
		if err = json.NewDecoder(r.Body).Decode(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
		if emap, err = checkCodePostBody(body); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, emap)
			return
		}

		_, err = X.InsertOne(body)
		if err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		_, err = X.Get(body)
		if err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		SuccessResponse(w, *body)

	case http.MethodPatch:
		SuccessResponse(w, nil)
		return
	case http.MethodDelete:
		if code == "" {
			LOG_ERRO.Println("code is NULL!")
			ErrorResponse(w, http.StatusBadRequest, "code is NULL!", ErrorMap{"code": []string{EmptNotAllowed}}, nil)
			return
		}

	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}

/*
	网安编码管理
*/
func WaCodeHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("WaCodeHandle")
	LOG_TRAC.Println("url:", r.RequestURI)

	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	//检查url中的子资源ID是否合法
	params := gmux.Vars(r)
	if v, ok := params["code_type"]; !ok {
		LOG_ERRO.Println("No code_type in url")
		ErrorResponse(w, http.StatusBadRequest, "No code_type in url", nil, nil)
		return
	} else {
		if err := r.ParseForm(); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, "", nil, nil)
			return
		}

		switch v {
		case "business_nature":
			wacodeMng(w, r, &BusinessNature{})

		case "netsite_type":
			wacodeMng(w, r, &NetsiteType{})

		case "certificate_type":
			wacodeMng(w, r, &CertificateType{})

		case "operator_net":
			wacodeMng(w, r, &OperatorNet{})

		case "access_type":
			wacodeMng(w, r, &AccessType{})

		case "ap_type":
			wacodeMng(w, r, &ApType{})

		case "auth_type":
			wacodeMng(w, r, &AuthType{})

		case "network_app":
			wacodeMng(w, r, &NetworkApp{})

		case "im_type":
			imTypeMng(w, r)

		default:
			LOG_ERRO.Println("unsupported code type :", v)
			ErrorResponse(w, http.StatusBadRequest, "unsupported code type :"+v, nil, nil)
		}

	}
}
