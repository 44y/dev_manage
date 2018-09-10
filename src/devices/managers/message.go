//消息中心
package managers

import (
	"encoding/json"
	"github.com/go-xorm/xorm"
	"net/http"
	. "rc"
	"strconv"
	"strings"
)

type MsgCtSt struct {
	HttpSt
}

type MsgCtBody struct {
	Read string `json:"read"`
	Idxs string `json:"idxs"`
}

//处理GET请求
func (this *MsgCtSt) httpGET() {
	LOG_TRAC.Println("httpGet")

	var (
		err          error
		read         string
		count, total int64
		ret          = make([]MessageCenter, 0)
		ret_t        = make([]MessageCenter, 0)
	)
	//解析参数
	if err = this.req.ParseForm(); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	offset, _ := strconv.Atoi(this.req.Form.Get("offset"))
	limit, _ := strconv.Atoi(this.req.Form.Get("limit"))
	if limit == 0 {
		limit = LIMIT_DEFAULT
	}

	read = this.req.Form.Get("read")

	var ss, ss_t *xorm.Session
	ss = X.Select("SQL_CALC_FOUND_ROWS *")
	ss_t = X.Select("SQL_CALC_FOUND_ROWS *")

	if read != "" {
		if read != "yes" && read != "no" {
			LOG_ERRO.Println("read is wrong:", read)
			ErrorResponse(this.writer, http.StatusBadRequest,
				"read is wrong:"+read,
				ErrorMap{"read": []string{BadRead}}, nil)
			return
		}
		ss = ss.Where("has_read like ?", read)
		ss_t = ss_t.Where("has_read like ?", read)
	}
	if err = ss.Limit(limit, offset).Find(&ret); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	count = int64(len(ret))
	if count == 0 {
		total = 0
	} else {
		if err = ss_t.Find(&ret_t); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
		total = int64(len(ret_t))
	}

	QueryResponse(this.writer, total, count, ret)
}

//处理DELETE请求
func (this *MsgCtSt) httpDELETE() {
	LOG_TRAC.Println("httpDELETE")

	var (
		err  error
		body = &MsgCtBody{}
	)

	if err = json.NewDecoder(this.req.Body).Decode(body); err != nil {
		if err.Error() == "EOF" {
			LOG_INFO.Println("No get body!")
		} else {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
	}

	if len(body.Idxs) == 0 {
		LOG_INFO.Println("delete all messages!")
		if _, err = X.NotIn("id", -1).Delete(new(MessageCenter)); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}

	} else {
		ids := strings.Split(body.Idxs, ",")
		ss := X.Where("id = ?", ids[0])
		for i := 1; i < len(ids); i++ {
			ss = ss.Or("id = ?", ids[i])
		}
		if _, err = ss.Delete(new(MessageCenter)); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
	}

	SuccessResponse(this.writer, nil)
}

//处理PATCH请求
func (this *MsgCtSt) httpPATCH() {
	LOG_TRAC.Println("httpPATCH")

	var (
		err  error
		body = &MsgCtBody{}
	)

	if err = json.NewDecoder(this.req.Body).Decode(body); err != nil {
		if err.Error() == "EOF" {
			LOG_INFO.Println("No get body!")
		} else {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
	}

	if body.Read != "yes" && body.Read != "no" {
		LOG_ERRO.Println("unsupported read values:", body.Read)
		ErrorResponse(this.writer, http.StatusBadRequest,
			"unsupported read values:"+body.Read, nil,
			ErrorMap{"read": []string{BadRead}})
		return
	}

	if len(body.Idxs) == 0 {
		LOG_INFO.Println("update all messages!")
		if _, err = X.Table(new(MessageCenter).TableName()).
			NotIn("id", -1).Update(UpdateMap{"has_read": body.Read}); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}

	} else {
		ids := strings.Split(body.Idxs, ",")
		ss := X.Table(new(MessageCenter).TableName()).Where("id = ?", ids[0])
		for i := 1; i < len(ids); i++ {
			ss = ss.Or("id = ?", ids[i])
		}
		if _, err = ss.Update(UpdateMap{"has_read": body.Read}); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(this.writer, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		}
	}
	SuccessResponse(this.writer, nil)
}

func MessageHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("MessageHandle:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	msg := &MsgCtSt{
		HttpSt: HttpSt{
			req:    r,
			writer: w,
		},
	}

	switch r.Method {
	case http.MethodGet:
		msg.httpGET()

	case http.MethodPatch:
		msg.httpPATCH()

	case http.MethodDelete:
		msg.httpDELETE()

	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}
