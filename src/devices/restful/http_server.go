package restful

import (
	"context"
	gmux "github.com/gorilla/mux"
	mng "managers"
	"net/http"
	. "rc"
	"strconv"
	"sync"
)

/*
	注册服务路由，挂载handle
*/
func registerMux(home_url string, mux *gmux.Router) {

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			w.Write([]byte("欢迎进入隐藏关卡！"))
			return
		}
		w.Write([]byte("欢迎来到主页！但是这里什么都没有.."))
	})
	sub_r := mux.PathPrefix(home_url).Subrouter()
	//厂商管理
	sub_r.HandleFunc("/orgs", mng.OrgsMngHandle).Methods(http.MethodGet, http.MethodPost)
	sub_r.HandleFunc("/orgs/file", mng.OrgsFileHandle).Methods(http.MethodPost)
	sub_r.HandleFunc("/orgs/file/{file_idx}", mng.OrgsFileHandle).Methods(http.MethodDelete)
	sub_r.HandleFunc("/orgs/{idx}", mng.OrgsMngHandle).Methods(http.MethodPatch, http.MethodDelete)

	//场所管理
	sub_r.HandleFunc("/netbars", mng.NetbarsMngHandle).Methods(http.MethodGet, http.MethodPost)
	sub_r.HandleFunc("/netbars/file", mng.NetbarsFileHandle).Methods(http.MethodPost)
	sub_r.HandleFunc("/netbars/file/{file_idx}", mng.NetbarsFileHandle).Methods(http.MethodDelete)
	sub_r.HandleFunc("/netbars/{idx}", mng.NetbarsMngHandle).Methods(http.MethodPatch, http.MethodDelete, http.MethodGet)

	//设备管理
	sub_r.HandleFunc("/aps", mng.APsMngHandle).Methods(http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete)
	sub_r.HandleFunc("/aps/file", mng.ApsFileHandle).Methods(http.MethodPost)
	sub_r.HandleFunc("/aps/file/{file_idx}", mng.ApsFileHandle).Methods(http.MethodDelete)
	sub_r.HandleFunc("/aps/{idx}", mng.APsMngHandle).Methods(http.MethodPatch, http.MethodDelete, http.MethodGet)

	//人员热力图
	sub_r.HandleFunc("/stats/hotstatus", mng.HotStatusHandle).Methods(http.MethodGet)

	//网安编码管理
	sub_r.HandleFunc("/sys_config/{code_type}", mng.WaCodeHandle).Methods(http.MethodGet, http.MethodPost,
		http.MethodPatch, http.MethodDelete)

	//配置查询、变更
	sub_r.HandleFunc("/config", mng.ConfigHandle).Methods(http.MethodGet, http.MethodPatch)

	//订阅消息
	sub_r.HandleFunc("/sub/{reg_id}", mng.SubMngHandle).Methods(http.MethodPost, http.MethodDelete)
	sub_r.HandleFunc("/sub", mng.SubMngHandle).Methods(http.MethodGet)

	//sidecar调用
	mux.HandleFunc("/version", mng.GetVerHandle).Methods(http.MethodGet)
	sub_r.HandleFunc("/service", mng.HttpQuit).Methods(http.MethodDelete)
	sub_r.HandleFunc("/status", mng.GetStatusHandle).Methods(http.MethodGet)

	//统计查询
	sub_r.HandleFunc("/stats/netbars", mng.NetbarsStatsHandle).Methods(http.MethodGet)
	sub_r.HandleFunc("/stats/netbars/file", mng.NetbarsStats2FileHandle).Methods(http.MethodPost)
	sub_r.HandleFunc("/stats/netbars/file/{file_idx}", mng.NetbarsStats2FileHandle).Methods(http.MethodDelete)

	//消息中心
	sub_r.HandleFunc("/message", mng.MessageHandle).Methods(http.MethodGet, http.MethodDelete, http.MethodPatch)

	//加密信息
	sub_r.HandleFunc("/encryption/default", mng.EncrytionHandle).Methods(http.MethodGet)

	//file system test
	fs_prefix := "/" + GlobalConfig.ServiceName + "/" + GlobalConfig.Version + "/files/"
	mux.PathPrefix(fs_prefix).Handler(http.StripPrefix(fs_prefix, http.FileServer(http.Dir(GlobalConfig.FileSysPath))))
}

/*
	启动http服务
*/
func StartHttpSrv(quit context.Context, wg *sync.WaitGroup) {
	LOG_TRAC.Println("Http server start!")
	wg.Add(1)
	defer LOG_INFO.Println("Http server done!")
	defer wg.Done()

	home_url := "/" + GlobalConfig.ServiceName + "/{version}"
	//LOG_TRAC.Println("homepage:", home_url)

	mux := gmux.NewRouter()
	registerMux(home_url, mux)

	LOG_TRAC.Println("http server start!")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(GlobalConfig.ServicePort),
		Handler: mux,
	}

	h_wg := &sync.WaitGroup{}
	go func() {
		h_wg.Add(1)
		defer h_wg.Done()
		server.ListenAndServe()
	}()

	//等待退出信号关闭服务
	<-quit.Done()
	//server.Close()
	shutdownSrv(server)

	h_wg.Wait()
}

/*
	优雅的退出http服务
*/
func shutdownSrv(srv *http.Server) {
	if err := srv.Shutdown(context.Background()); err != nil {
		LOG_ERRO.Println(err)
	}
}
