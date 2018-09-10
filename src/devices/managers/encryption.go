//加密信息
package managers

import (
	"net/http"
	. "rc"
)

func EncrytionHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("EncrytionHandle:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
		return
	}

	var (
		ex  bool
		err error
		en  = &EncryptionInfo{}
	)

	if ex, err = X.ID(1).Get(en); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}

	if !ex {
		LOG_ERRO.Println("encryption table wrong!")
		ErrorResponse(w, http.StatusInternalServerError, "encryption table wrong!", nil, nil)
		return
	}

	QueryResponse(w, 1, 1, en)

}
