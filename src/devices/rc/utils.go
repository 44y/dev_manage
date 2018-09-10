package rc

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	gmux "github.com/gorilla/mux"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

const BadVersion = "版本不支持"

type CodeStatus struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
}

/*
   ======错误返回===========================
*/
type errorRsp struct {
	CodeStatus
	Msg  string    `json:"message"`
	Data errorData `json:"data,omitempty"`
}

type errorData struct {
	Reason    string   `json:"reason,omitempty"`
	Parameter ErrorMap `json:"parameters,omitempty"`
	Body      ErrorMap `json:"body,omitempty"`
}

type ErrorMap map[string]interface{}

//导出文件的字段域名
type OutFieldsSlice []map[string]string

/*
   向客户端返回错误
*/
func ErrorResponse(w http.ResponseWriter, code int, reason string, parameter, body ErrorMap) error {
	rsp := &errorRsp{
		CodeStatus: CodeStatus{
			Code: code,
		},
		Msg: http.StatusText(code),
		Data: errorData{
			Reason: reason,
		},
	}

	//错误400
	if code == http.StatusBadRequest {
		rsp.Data.Parameter = parameter
		rsp.Data.Body = body
	}

	if code >= http.StatusBadRequest &&
		code < http.StatusInternalServerError {
		rsp.Status = "error"
	}
	if code >= http.StatusInternalServerError {
		rsp.Status = "fail"
	}
	//rsp.Code = code
	//rsp.Status = "error"
	return json.NewEncoder(w).Encode(rsp)
}

/*
	======成功返回========================
*/
type succRsp struct {
	CodeStatus
	Data interface{} `json:"data,omitempty"`
}

/*
   向客户端返回成功,v为data结构
*/
func SuccessResponse(w http.ResponseWriter, data interface{}) error {
	rsp := succRsp{
		CodeStatus: CodeStatus{
			Code:   http.StatusOK,
			Status: "success",
		},
		Data: data,
	}
	//rsp.Code = http.StatusOK
	//rsp.Status = "success"
	return json.NewEncoder(w).Encode(rsp)
}

type queryRsp struct {
	CodeStatus
	Data queryData `json:"data"`
}

type queryData struct {
	Total   int64       `json:"total"`
	Count   int64       `json:"count"`
	Results interface{} `json:"results,omitempty"`
}

/*
   向客户端返回查询结果,v为data结构
*/
func QueryResponse(w http.ResponseWriter, total, count int64, data interface{}) error {
	rsp := queryRsp{
		CodeStatus: CodeStatus{
			Code:   http.StatusOK,
			Status: "success",
		},
		Data: queryData{
			Total:   total,
			Count:   count,
			Results: data,
		},
	}
	//rsp.Code = http.StatusOK
	//rsp.Status = "success"
	return json.NewEncoder(w).Encode(rsp)
}

/*
	error不为nil则打印出来，但不做其他操作
*/
func CheckErr(e error) {
	if e != nil {
		LOG_ERRO.Println(e)
	}
}

/*
   获取字符串长度，单位[]byte，最大两字节，不足补零
*/
func GetStrlenByte(str []byte) ([]byte, error) {
	strlen := len(str)
	fmt.Println(strlen)
	if strlen > 65535 {
		return nil, fmt.Errorf("strlen is too long\n")
	}

	len_16 := fmt.Sprintf("%04x", strlen)

	len_byte, err := hex.DecodeString(len_16)
	if err != nil {
		panic(err)
	}

	return len_byte, nil
}

/*
	将[]byte解析成int
*/
func DecodeByte2int(b []byte) (int, error) {
	str := hex.EncodeToString(b)

	var (
		n   int
		err error
	)

	n, err = fmt.Sscanf(str, "%04x", &n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

/*
   将多个[]byte合并为一个[]byte
*/
func CombineByte(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

/*
	判断版本号是否支持
*/
func CheckVersion(r *http.Request) (bool, ErrorMap) {
	params := gmux.Vars(r)
	//LOG_TRAC.Println(params)
	var x, y, z int

	if ver, ok := params["version"]; ok == false {
		data := make(ErrorMap)
		data["version"] = "No version in url!"
		return false, data
	} else {
		fmt.Sscanf(ver, "%d.%d.%d", &x, &y, &z)
		data := make(ErrorMap)
		data["version"] = []string{BadVersion}
		if x != GlobalConfig.VerX || y > GlobalConfig.VerY || z > GlobalConfig.VerZ {
			return false, data
		}
	}

	return true, nil
}

/*
	判断scope中包含的查询条件
*/
func ParseScope(keys []string, scope_str []string) map[string]bool {
	ret := make(map[string]bool, 0)

	if len(scope_str) == 0 {
		return ret
	}
	for _, key := range keys {
		for _, scp_v := range scope_str {
			//LOG_TRAC.Println(key, scp_v)
			if key == scp_v {
				ret[key] = true
			}
		}
	}

	return ret
}

/*
	复制src中不为空的字段到dst中，dst中其余字段不变
	src为struct，dst为struct指针
*/
func CopyStruct(dst, src interface{}) error {
	src_v := reflect.ValueOf(src)
	src_t := reflect.TypeOf(src)
	dst_v := reflect.ValueOf(dst)
	dst_t := reflect.TypeOf(dst)
	//找到指针所指的值
	dst_vValue := dst_v.Elem()
	dst_tValue := dst_t.Elem()

	//只接受struct类型
	if src_t.Kind() != reflect.Struct || dst_tValue.Kind() != reflect.Struct {
		return fmt.Errorf("Only struct type accpted!\n")
	}

	for i := 0; i < src_t.NumField(); i++ {
		field_name := src_t.Field(i).Name

		is_found := false
		//在目的结构中找原结构中的字段，字段名和类型都要匹配
		for j := 0; j < dst_tValue.NumField(); j++ {
			if field_name == dst_tValue.Field(j).Name &&
				src_t.Field(i).Type.Kind() == dst_tValue.Field(j).Type.Kind() {
				is_found = true
				break
			}
		}
		//目的结构中未找到该字段
		if !is_found {
			LOG_WARN.Printf("Field not found in dst struct, %s\n", field_name)
			continue
		}

		kd := src_t.Field(i).Type.Kind()
		switch kd {
		case reflect.Int, reflect.Int64:
			if src_v.Field(i).Int() != 0 {
				v := dst_vValue.Field(i)
				if v.CanSet() {
					v.SetInt(src_v.Field(i).Int())
				}
			}
		case reflect.Float32, reflect.Float64:
			if src_v.Field(i).Float() != 0 {
				v := dst_vValue.Field(i)
				if v.CanSet() {
					v.SetFloat(src_v.Field(i).Float())
				}
			}
		case reflect.String:
			if len(src_v.Field(i).String()) != 0 {
				v := dst_vValue.Field(i)
				if v.CanSet() {
					v.SetString(src_v.Field(i).String())
				}
			}
		default:
			LOG_INFO.Printf("Unsupported type: %v", kd)
		}
	}

	return nil
}

/*
	将查询结果写入csv文件，data为结构slice
*/
func WriteCSV(name string, data interface{}, fields OutFieldsSlice) error {
	var (
		ex  bool
		tag string
	)
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Slice {
		return errors.New("Only slice of struct accepted")
	}

	v := reflect.ValueOf(data)
	data_len := v.Len()
	if data_len == 0 {
		//LOG_ERRO.Println("data lenth is 0")
		return errors.New("data lenth is 0")
	}
	if v.Index(0).Type().Kind() != reflect.Struct {
		return errors.New("Only slice of struct accepted")
	}

	f, err := os.Create(name)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)

	//文件头
	title_str := make([]string, 0)
	if fields == nil || len(fields) == 0 {
		LOG_TRAC.Println("no fields , write all fields")
		for i := 0; i < v.Index(0).NumField(); i++ {
			fdi := v.Index(0).Type().Field(i)
			if tag, ex = fdi.Tag.Lookup("json"); ex && tag != "-" {
				title_str = append(title_str, fdi.Name)
			}
		}
		w.Write(title_str)
		//写入全部数据
		for i := 0; i < data_len; i++ {
			str := make([]string, 0)

			dv := v.Index(i)
			dt := v.Index(i).Type()
			for j := 0; j < dv.NumField(); j++ {
				fd := dt.Field(j)
				switch fd.Type.Kind() {
				case reflect.String:
					tag, ex = fd.Tag.Lookup("json")
					//LOG_TRAC.Println(ex, tag)
					if ex && tag != "-" {
						str = append(str, dv.Field(j).String())
					}
				case reflect.Int, reflect.Int64:
					tag, ex = fd.Tag.Lookup("json")
					//LOG_TRAC.Println(ex, tag)
					if ex && tag != "-" {
						str = append(str, strconv.FormatInt(dv.Field(j).Int(), 10))
					}
				case reflect.Float64, reflect.Float32:
					tag, ex = fd.Tag.Lookup("json")
					//LOG_TRAC.Println(ex, tag)
					if ex && tag != "-" {
						str = append(str, strconv.FormatFloat(dv.Field(j).Float(), 'f', 5, 64))
					}
				default:
					LOG_TRAC.Println("unsupported type:", fd.Type.Kind())
				}
			}

			LOG_TRAC.Println(str)
			w.Write(str)
		}

	} else {
		key_str := make([]string, 0)
		for _, mv := range fields {
			for k, v := range mv {
				title_str = append(title_str, v)
				key_str = append(key_str, k)
			}
		}
		w.Write(title_str)
		//写入fields要的数据
		for i := 0; i < data_len; i++ {
			str := make([]string, 0)
			var found_key bool
			for _, key := range key_str {
				found_key = false
				LOG_TRAC.Println("key:", key)
				dv := v.Index(i)
				dt := v.Index(i).Type()

			OUT_LOOP:
				for j := 0; j < dv.NumField(); j++ {
					fd := dt.Field(j)
					switch fd.Type.Kind() {
					case reflect.String:
						tag, ex = fd.Tag.Lookup("json")
						//LOG_TRAC.Println(ex, tag)
						if ex && tag == key {
							str = append(str, dv.Field(j).String())
							found_key = true
							break OUT_LOOP
						}
					case reflect.Int, reflect.Int64:
						tag, ex = fd.Tag.Lookup("json")
						//LOG_TRAC.Println(ex, tag)
						if ex && tag == key {
							str = append(str, strconv.FormatInt(dv.Field(j).Int(), 10))
							found_key = true
							break OUT_LOOP
						}
					case reflect.Float64, reflect.Float32:
						tag, ex = fd.Tag.Lookup("json")
						//LOG_TRAC.Println(ex, tag)
						if ex && tag == key {
							str = append(str, strconv.FormatFloat(dv.Field(j).Float(), 'f', 5, 64))
							found_key = true
							break OUT_LOOP
						}
					default:
						LOG_TRAC.Println("unsupported type:", fd.Type.Kind())
					}
				}
				if !found_key {
					return errors.New("unsupported field:" + key)
				}
			}
			LOG_TRAC.Println(str)
			w.Write(str)
		}
	}
	LOG_TRAC.Println("title_str = ", title_str)

	w.Flush()
	return nil
}
