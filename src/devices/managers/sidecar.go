/*与sidecar之间的通信*/
package managers

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	. "rc"
	"syscall"
)

var MyIp string //本机IP

/*
   ======get version 返回===========================
*/
type verData struct {
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
	BuildTime   string `json:"build_time"`
	CommitID    string `json:"commit_id"`

	//for detail=yes
	Ip   string  `json:"ip,omitempty"`
	Port int     `json:"port,omitempty"`
	Urls []UrlSt `json:"urls"`
}

type UrlSt struct {
	Url  string `json:"url"`
	Port int    `json:"port"`
}

func GetVerHandle(w http.ResponseWriter, r *http.Request) {
	//仅支持GET方法
	if r.Method != "GET" {
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	detail := r.Form.Get("detail")

	LOG_TRAC.Println("get version~")

	data := verData{
		ServiceName: GlobalConfig.ServiceName,
		Version:     GlobalConfig.Version,
		BuildTime:   GlobalConfig.BuildTime,
		CommitID:    GlobalConfig.CommitID,
	}

	if detail == "yes" {
		data.Ip = MyIp
		data.Port = GlobalConfig.ServicePort
	}

	SuccessResponse(w, data)
}

/*
   ======get status 返回===========================
*/
type statusData struct {
	Run         string     `json:"run"`
	ServiceName string     `json:"service_name"`
	FailMessage string     `json:"fail_message,omitempty"`
	Mem         *statusMem `json:"mem,omitempty"`
}

type statusMem struct {
	Used int `json:"used"`
	Free int `json:"free"`
}

var getStatus_num int

func GetStatusHandle(w http.ResponseWriter, r *http.Request) {

	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}
	if getStatus_num >= 10 {
		LOG_TRAC.Println("get status~ for", getStatus_num, "times")
		getStatus_num = 0
	}
	getStatus_num++

	status := statusData{
		Run:         "ok",
		ServiceName: GlobalConfig.ServiceName,
	}
	if err := SuccessResponse(w, status); err != nil {
		LOG_ERRO.Println(err)
	}
}

/*
   ======退出接口===========================
*/
type safeQuitSt struct {
	QuitChan   chan struct{}
	chanStatus bool //QuitChan status, true:open;false:closed
}

var q safeQuitSt

/*
   监听退出信号: SIGINT,SIGQUIT,SIGTERM
   监听日志重置信号：SIGHUP
*/
func (this *safeQuitSt) listenSignal() {

	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	hupChan := make(chan os.Signal)
	signal.Notify(hupChan, syscall.SIGHUP)

	for {
		select {
		case s := <-quitChan:
			LOG_ERRO.Println(s)
			if this.chanStatus {
				close(this.QuitChan)
				this.chanStatus = false
			}

		case <-hupChan:
			LOG_INFO.Println("got sighup~")
			LogReset()

			dbLogReset()
		}
	}

}

func StartListenSignal() <-chan struct{} {
	go q.listenSignal()

	return q.QuitChan
}

/*
	初始化创建退出chan
*/
func SafeQuitInit() {
	q.QuitChan = make(chan struct{})
	q.chanStatus = true
}

/*
	服务退出接口
*/
func HttpQuit(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("Start quit!")

	if r.Method != "DELETE" {
		ErrorResponse(w, http.StatusMethodNotAllowed, "", nil, nil)
		LOG_ERRO.Println("Method not supported: " + r.Method)
		return
	}

	r.ParseForm()
	r.Form.Get("reason")
	LOG_ERRO.Println(r.Form.Get("reason"))

	if q.chanStatus {
		close(q.QuitChan)
		q.chanStatus = false
	}

	SuccessResponse(w, nil)
}

/*
	获取本机ip
*/
func GetMyIp() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		panic(err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	LOG_ERRO.Println("No IP!")
	return ""
}
