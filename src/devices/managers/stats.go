//场所、设备统计查询
package managers

import (
	"encoding/json"
	"errors"
	"fmt"
	gmux "github.com/gorilla/mux"
	"net/http"
	. "rc"
	"strconv"
)

const (
	//supported group
	GROUP_AREA_CODE    = "area_code"
	GROUP_ORG_NAME     = "org_name"
	GROUP_NETSITE_TYPE = "netsite_type"

	//group return
	NETBAR_NUM          = "number_num"
	DEV_ONLINE_NUM      = "dev_online_num"
	DEV_OFFLINE_NUM     = "dev_offline_num"
	DEV_MAINTAIN_NUM    = "dev_maintain_num"
	DEV_TOTAL_NUM       = "dev_total_num"
	DEV_DATA_ONLINE_NUM = "dev_data_online_num"
	DEV_ALIVE_RATE      = "dev_alive_rate"
	DEV_DATA_RATE       = "dev_data_rate"

	//group sql
	GROUP_SQL_FORMAT = "SELECT %s, " +
		"COUNT(*) AS %s," +
		"SUM(dev_online_num) as %s," +
		"SUM(dev_offline_num) AS %s," +
		"SUM(dev_abnormal_num) AS %s," +
		"SUM(dev_total_num) AS %s," +
		"SUM(dev_data_online_num) AS %s," +
		"ROUND(SUM(dev_online_num)/SUM(dev_total_num)*100,2) AS %s," +
		"ROUND(SUM(dev_data_online_num)/SUM(dev_total_num)*100,2) AS %s " +
		"FROM NetbarInfo " +
		"where approval = 1 "
)

type groupSt struct {
	GroupId          string `json:"group_id,omitempty"`
	NetbarNum        int    `json:"netbar_num"`
	DevOnlineNum     int    `json:"dev_online_num"`
	DevOfflineNum    int    `json:"dev_offline_num"`
	DevMaintainNum   int    `json:"dev_maintain_num"`
	DevTotalNum      int    `json:"dev_total_num"`
	DevDataOnlineNum int    `json:"dev_data_online_num"`
	DevAliveRate     string `json:"dev_alive_rate"`
	DevDataRate      string `json:"dev_data_rate"`
}

type groupTotalSt struct {
	groupSt
	NetbarDataOnlineNum   int    `json:"netbar_data_online_num"`
	NetbarDataOnlineRate  string `json:"netbar_data_online_rate"`
	NetbarAliveOnlineNum  int    `json:"netbar_alive_online_num"`
	NetbarAliveOnlineRate string `json:"netbar_alive_online_rate"`
}

type NetbarStatsRetSt struct {
	Total groupTotalSt `json:"total"`
	Data  []groupSt    `json:"data"`
}

/*
	更新场所表中已审核场所下各种状态设备的数量
*/
func updateNetbarColumns(netbars []NetbarInfo) ([]NetbarInfo, error) {
	var (
		err   error
		rsl_n = make([]NetbarInfo, 0)
	)

	for _, n := range netbars {
		n.DevTotalNum = 0
		n.DevOnlineNum = 0
		n.DevOfflineNum = 0
		n.DevDataOnlineNum = 0
		n.DevAbnormalNum = 0
		n.DataStatus = STATUS_ABNORMAL

		aps := make([]DevInfo, 0)
		err = X.Where("netbar_wacode = ?", n.Wacode).And("approval = ?", APPROVED).
			Find(&aps)
		if err != nil {
			return nil, err
		}
		for _, ap := range aps {
			n.DevTotalNum++
			if ap.AliveStatus == STATUS_ABNORMAL {
				n.DevOfflineNum++
				LOG_TRAC.Printf("发现一个离线设备")
			} else {
				n.DevOnlineNum++
				LOG_TRAC.Printf("发现一个在线设备")
			}

			if ap.DataStatus == STATUS_NORMAL {
				n.DevDataOnlineNum++
				n.DataStatus = STATUS_NORMAL
				LOG_TRAC.Printf("发现一个数据在线设备")
			}
			if ap.DeviceStatus != DEVICE_ONLINE &&
				ap.DeviceStatus != DEVICE_OFFLINE {
				n.DevAbnormalNum++
			}
		}
		/*
			if time_now-n.LastBasicTime > GlobalConfig.NetbarBasicTimeout {
				n.BasicStatus = STATUS_ABNORMAL
			} else {
				n.BasicStatus = STATUS_NORMAL
			}

			if time_now-n.LastAliveTime > GlobalConfig.NetbarAliveTimeout {
				n.AliveStatus = STATUS_ABNORMAL
			} else {
				n.AliveStatus = STATUS_NORMAL
			}
		*/

		rsl_n = append(rsl_n, n)
	}
	return rsl_n, nil

}

/*
	场所统计查询
*/
func NetbarsStatsHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	var (
		err  error
		code int
		rslt []groupSt
	)

	//解析body
	body := new(DevicesGetBody)
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		if err.Error() == "EOF" {
			LOG_INFO.Println("No get body!")
		} else {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
			return
		}
	}

	rslt, code, err = doGroup(r, &body.Scope)
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, code, err.Error(), nil, nil)
		return
	}

	count64 := int64(len(rslt))

	total := groupTotalSt{}

	for _, v := range rslt {
		total.NetbarNum += v.NetbarNum
		total.DevOnlineNum += v.DevOnlineNum
		total.DevOfflineNum += v.DevOfflineNum
		total.DevMaintainNum += v.DevMaintainNum
		total.DevTotalNum += v.DevTotalNum
		total.DevDataOnlineNum += v.DevDataOnlineNum
	}
	if total.DevTotalNum == 0 {
		total.DevAliveRate = "0.00%"
		total.DevDataRate = "0.00%"
	} else {
		total.DevAliveRate = fmt.Sprintf("%0.2f",
			(float32(total.DevOnlineNum)/float32(total.DevTotalNum))*100) + "%"
		total.DevDataRate = fmt.Sprintf("%0.2f",
			(float32(total.DevDataOnlineNum)/float32(total.DevTotalNum))*100) + "%"
	}

	//统计total中的场所信息
	netbars := make([]NetbarInfo, 0)
	if err = X.Where("approval = ?", APPROVED).Find(&netbars); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
		return
	}
	for _, v := range netbars {
		if v.DataStatus == STATUS_NORMAL {
			total.NetbarDataOnlineNum++
		}
		if v.AliveStatus == STATUS_NORMAL {
			total.NetbarAliveOnlineNum++
		}
	}
	if total.NetbarNum == 0 {
		total.NetbarDataOnlineRate = "0.00%"
		total.NetbarAliveOnlineRate = "0.00%"
	} else {
		total.NetbarDataOnlineRate = fmt.Sprintf("%0.2f",
			(float32(total.NetbarDataOnlineNum)/float32(total.NetbarNum))*100) + "%"
		total.NetbarAliveOnlineRate = fmt.Sprintf("%0.2f",
			(float32(total.NetbarAliveOnlineNum)/float32(total.NetbarNum))*100) + "%"
	}

	rt := &NetbarStatsRetSt{
		Total: total,
		Data:  rslt,
	}

	//LOG_TRAC.Println(rslt)
	QueryResponse(w, count64, count64, rt)
}

/*
	执行group查询
*/
func doGroup(r *http.Request, body *DataScope) ([]groupSt, int, error) {
	var (
		err          error
		qret         []map[string][]byte
		g            string
		areacode_str string //根据地区编码组成查询字符串
		orgcode_str  string //根据厂商编码组成查询字符串
	)

	if err = r.ParseForm(); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if body != nil {
		//生成地区编码查询字符串
		if areacode_str, err = getAreaCodeStr(nil, &body.AreaCode); err != nil {
			LOG_ERRO.Println(err)
			return nil, http.StatusBadRequest, err
		}

		//根据scope中的厂商编码进行查询限制
		for _, v := range body.Orgcodes {
			orgcode_str = orgcode_str + fmt.Sprintf(" security_software_orgcode = %s or", v)
		}
		if len(orgcode_str) > 0 {
			orgcode_str = orgcode_str[1 : len(orgcode_str)-2]
		}
		LOG_TRAC.Println("orgcode_str:", orgcode_str)
	}

	group := r.Form.Get("group")
	switch group {
	case GROUP_AREA_CODE:
		g = "area_code_3"

	case GROUP_ORG_NAME:
		g = "security_software_orgname"

	case GROUP_NETSITE_TYPE:
		g = "netsite_type"

	default:
		return nil, http.StatusBadRequest, errors.New("unsupported group :" + group)
	}

	sql1 := fmt.Sprintf(
		GROUP_SQL_FORMAT, g,
		NETBAR_NUM,
		DEV_ONLINE_NUM,
		DEV_OFFLINE_NUM,
		DEV_MAINTAIN_NUM,
		DEV_TOTAL_NUM,
		DEV_DATA_ONLINE_NUM,
		DEV_ALIVE_RATE,
		DEV_DATA_RATE)

	if len(areacode_str) > 0 {
		sql1 += "and " + areacode_str
	}
	if len(orgcode_str) > 0 {
		sql1 += "and " + orgcode_str
	}

	sql2 := fmt.Sprintf(" GROUP by %s", g)
	sql := sql1 + sql2

	//group之前更新设备数量
	netbars := make([]NetbarInfo, 0)
	if err = X.Find(&netbars); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if netbars, err = updateNetbarColumns(netbars); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for _, v := range netbars {
		if _, err = updateData(v.Id, &v); err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	if qret, err = X.Query(sql); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	rslt := make([]groupSt, 0)
	for _, v := range qret {
		gret := groupSt{
			GroupId:      string(v[g]),
			DevAliveRate: string(v[DEV_ALIVE_RATE]) + "%",
			DevDataRate:  string(v[DEV_DATA_RATE]) + "%",
		}
		if gret.DevAliveRate == "%" {
			gret.DevAliveRate = "0.00%"
		}
		if gret.DevDataRate == "%" {
			gret.DevDataRate = "0.00%"
		}

		gret.NetbarNum, _ = strconv.Atoi(string(v[NETBAR_NUM]))
		gret.DevOnlineNum, _ = strconv.Atoi(string(v[DEV_ONLINE_NUM]))
		gret.DevOfflineNum, _ = strconv.Atoi(string(v[DEV_OFFLINE_NUM]))
		gret.DevDataOnlineNum, _ = strconv.Atoi(string(v[DEV_DATA_ONLINE_NUM]))
		gret.DevMaintainNum, _ = strconv.Atoi(string(v[DEV_MAINTAIN_NUM]))
		gret.DevTotalNum, _ = strconv.Atoi(string(v[DEV_TOTAL_NUM]))

		rslt = append(rslt, gret)
	}
	return rslt, 0, nil
}

/*
	场所统计查询导出文件
*/
func NetbarsStats2FileHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	switch r.Method {
	case http.MethodPost:
		statsFilePOST(w, r)

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

func statsFilePOST(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("statsFilePOST:", r.RequestURI)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	var (
		body *FilePostBody
		err  error
		code int
		fret *fileRet
		rslt []groupSt
	)
	//解析参数
	if err = r.ParseForm(); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	//解析body
	if body, err = getFilePostBody(r.Body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	LOG_TRAC.Printf("body:%+v\n", body)

	if body != nil {
		rslt, code, err = doGroup(r, &body.Scope)
	} else {
		rslt, code, err = doGroup(r, nil)
	}
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, code, err.Error(), nil, nil)
		return
	}

	//生成文件
	if body != nil {
		fret, err = createFile(r.Form.Get("file"), rslt, body.Fields)
	} else {
		fret, err = createFile(r.Form.Get("file"), rslt, nil)
	}
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	SuccessResponse(w, fret)
}
