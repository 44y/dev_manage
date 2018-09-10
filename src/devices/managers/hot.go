//人员热力图处理

package managers

import (
	"encoding/json"
	"fmt"
	"net/http"
	. "rc"
	"strconv"
	"time"
)

//人员热力图返回结构
type HotRetSt struct {
	UserTotalNum int               `json:"user_total_num"`
	Data         []NetbarUserNumSt `json:"data"`
}

//场所经纬度、人数结构
type NetbarUserNumSt struct {
	Wacode    string  `json:"netbar_wacode"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	User_num  int     `json:"user_number"`
}

//统计查询返回body
type userNumBody struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   userNumData `json:"data"`
}
type userNumData struct {
	Count int `json:"count"`
}

var hotConfig *DependentService

func HotStatusHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("HotStatusHandle")

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	var (
		long_w, long_e, lati_n, lati_s float64
		start, end                     int64
		err                            error
		count                          int64
		url                            string
		resp                           *http.Response
	)
	//解析参数
	if err = r.ParseForm(); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	//解析经纬度
	long_w, err = strconv.ParseFloat(r.Form.Get("lw"), 64)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"lw": []string{BadLoniLati}}, nil)
		return
	}

	long_e, err = strconv.ParseFloat(r.Form.Get("le"), 64)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"le": []string{BadLoniLati}}, nil)
		return
	}

	lati_n, err = strconv.ParseFloat(r.Form.Get("ln"), 64)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"ln": []string{BadLoniLati}}, nil)
		return
	}

	lati_s, err = strconv.ParseFloat(r.Form.Get("ls"), 64)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"ls": []string{BadLoniLati}}, nil)
		return
	}

	if r.Form.Get("start") == "" {
		start = 0
	} else {
		start, err = strconv.ParseInt(r.Form.Get("start"), 10, 64)
		if err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"start": []string{BadStartEndTime}}, nil)
			return
		}
	}

	if r.Form.Get("end") == "" {
		end = 0
	} else {
		end, err = strconv.ParseInt(r.Form.Get("end"), 10, 64)
		if err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"end": []string{BadStartEndTime}}, nil)
			return
		}
	}

	if start > end {
		LOG_ERRO.Println(BadStartEndTime)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), ErrorMap{"start": []string{BadStartEndTime}}, nil)
		return
	}

	LOG_TRAC.Println("long_w:", long_w, "long_e:", long_e, "lati_n:", lati_n, "lati_s:", lati_s)
	//东经大于西经，北纬大于南纬
	if long_w >= long_e || lati_s >= lati_n {
		LOG_ERRO.Println("parameter error!")
		ErrorResponse(w, http.StatusBadRequest, "parameter error!", ErrorMap{"lw": []string{BadLoniLati}}, nil)
		return
	}

	str_format := " longitude >= %f and longitude <= %f and latitude >= %f and latitude <= %f "
	query_str := fmt.Sprintf(str_format, long_w, long_e, lati_s, lati_n)

	LOG_TRAC.Println("query_str:", query_str)

	netbars := make([]NetbarInfo, 0)

	//根据条件查找
	count, err = X.Where(query_str).And("approval = ?", APPROVED).FindAndCount(&netbars)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Println(count)
	if count == 0 {
		QueryResponse(w, 0, 0, nil)
		return
	}

	if hotConfig == nil {
		for _, v := range GlobalConfig.Dependent {
			if v.Use == "hot" {
				hotConfig = &v
				break
			}
		}
	}
	url = "http://" + GlobalConfig.ApiGw + "/" + hotConfig.Name +
		"/" + hotConfig.Version + "/" + hotConfig.Rc

	ret_data := &HotRetSt{}
	ret_data.Data = make([]NetbarUserNumSt, count)
	for i, v := range netbars {
		now := time.Now()
		if start == 0 {
			start = now.Unix() - int64(now.Second()+now.Minute()*60+now.Hour()*60*60)
		}
		if end == 0 {
			end = now.Unix()
		}

		n_url := url + fmt.Sprintf("?F_netbar_code=%s", v.Wacode) +
			fmt.Sprintf("&start=%d&end=%d", start, end)
		LOG_TRAC.Println("url:", n_url)
		ret_data.Data[i].Wacode = v.Wacode
		ret_data.Data[i].Latitude = v.Latitude
		ret_data.Data[i].Longitude = v.Longitude

		req, err := http.NewRequest(http.MethodGet, n_url, nil)
		if err != nil {
			LOG_ERRO.Println(err)
		}
		//req.ParseForm()
		//req.Form.Add("F_netbar_code", v.Wacode)

		client := &http.Client{}
		client.Timeout = 5 * time.Second

		if resp, err = client.Do(req); err != nil {
			LOG_ERRO.Println(err)
			continue
		}
		LOG_TRAC.Printf("%+v\n", resp)
		if resp.StatusCode != http.StatusOK {
			LOG_ERRO.Println("bad resp, code :", resp.StatusCode)
			continue
		}
		body := new(userNumBody)
		if err = json.NewDecoder(resp.Body).Decode(body); err != nil {
			LOG_ERRO.Println(err)
			continue
		}
		//LOG_TRAC.Printf("body:%+v\n", body)
		if body.Code != http.StatusOK {
			LOG_ERRO.Println("bad resp, code :", body.Code, "status:", body.Status)
			ErrorResponse(w, http.StatusInternalServerError,
				hotConfig.Name+" response fail", nil, nil)
			return
		}

		ret_data.Data[i].User_num = body.Data.Count
		ret_data.UserTotalNum += body.Data.Count
	}

	//LOG_TRAC.Println(datas)
	QueryResponse(w, count, count, ret_data)
}
