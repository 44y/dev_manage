package managers

import (
	"encoding/json"
	"net/http"
	. "rc"
)

/*
   通过http GET 获取config
*/
func configHttpGET(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("get config")
	SuccessResponse(w, GlobalConfig.DevicesConfig)
}

/*
   通过http PATCH 修改config
*/
func configHttpPATCH(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("config patch")
	var (
		err error
	)

	body := new(DevicesConfig)
	if err = json.NewDecoder(r.Body).Decode(body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body:%+v\n", body)

	if err = CopyStruct(&GlobalConfig.DevicesConfig, *body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	LOG_TRAC.Printf("GlobalConfig:%+v\n", GlobalConfig.DevicesConfig)
	//X.ID(1).Update(&GlobalConfig.Timeout)
	updateData(1, &GlobalConfig.DevicesConfig)

	LoadConfigFromDB(X)

	SuccessResponse(w, GlobalConfig.DevicesConfig)
}

func ConfigHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("ConfigHandle")

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	switch r.Method {
	case "GET":
		configHttpGET(w, r)
	case "PATCH":
		configHttpPATCH(w, r)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
	}
}
