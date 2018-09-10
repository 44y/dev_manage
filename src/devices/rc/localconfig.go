/*配置管理*/

package rc

import (
	"encoding/json"
	"fmt"
	"github.com/go-xorm/xorm"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	//localConfig = "localcfg.json"
	localConfig = "/usr/local/bin/localcfg.json"

	second = 1
	minute = 60 * second
	hour   = 60 * minute

	AP_ALIVE_TIMEOUT_DEFAULT     = 3 * 5 * minute //3个5分钟
	NETBAR_ALIVE_TIMEOUT_DEFAULT = 3 * 24 * hour  //3个24小时
	DATA_TIMEOUT_DEFAULT         = hour           //1小时
	BASIC_TIMEOUT_DEFAULT        = 3 * hour       //3个1小时

	//period to remove expired files
	FILE_EXPIRED_DEFAULT = 1 * hour
)

type confSt struct {
	ServiceName string `json:"service_name"`
	ServicePort int    `json:"service_port"`
	Version     string `json:"version"`

	//dir to save output csv files
	FileSysPath string `json:"file_system_path"`

	//files host, like :http://dockers.ikuai8.com:18087
	FileHost string `json:"-"`

	//devices log file
	DevicesLogFile string `json:"devices_log_file"`

	//ftp server dir
	FtpServerDir string `json:"ftp_server_dir"`

	//xorm database log file
	XormLogFile string `json:"xorm_log_file"`

	ApiGw         string `json:"api_gw"`
	UDPClientPort int    `json:"udp_client_port"`

	BuildTime string `json:"-"`
	CommitID  string `json:"-"`

	//dependent services config
	Dependent []DependentService `json:"dependent"`

	//devices config ,can be set by restful API
	DevicesConfig `json:"-"`

	//database config
	DbConfig DbConfigSt `json:"database"`

	VerX int `json:"-"`
	VerY int `json:"-"`
	VerZ int `json:"-"`

	req    *http.Request       `json:"-"`
	writer http.ResponseWriter `json:"-"`
}

//依赖服务
type DependentService struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Rc      string `json:"rc"`
	Use     string `json:"use"`
}

//subcriber配置
type SubConfigSt struct {
	ApiGwAddr string `json:"api_gw"`   //api网关地址
	Port      int    `json:"sub_port"` //监听
	PubName   string `json:"pub_name"` //import服务名称
	PubVer    string `json:"pub_ver"`  //import版本
}

//数据库配置
type DbConfigSt struct {
	User     string `json:"db_user"`
	Password string `json:"db_password"`
	Ip       string `json:"db_ip"`
	Port     int    `json:"db_port"`
	Name     string `json:"db_name"`
}

var (
	GlobalConfig confSt
)

/*
	load from database
*/
func LoadConfigFromDB(x *xorm.Engine) {
	var (
		err error
	)

	if _, err = x.Get(&GlobalConfig.DevicesConfig); err != nil {
		LOG_ERRO.Println(err)
		GlobalConfig.NetbarAliveTimeout = NETBAR_ALIVE_TIMEOUT_DEFAULT
		GlobalConfig.NetbarBasicTimeout = BASIC_TIMEOUT_DEFAULT
		GlobalConfig.ApAliveTimeout = AP_ALIVE_TIMEOUT_DEFAULT
		GlobalConfig.ApDataTimeout = DATA_TIMEOUT_DEFAULT
		GlobalConfig.APBasicTimeout = BASIC_TIMEOUT_DEFAULT
		GlobalConfig.FileExpiredTime = FILE_EXPIRED_DEFAULT
	}

	GlobalConfig.Println()
}

/*
	get env
*/
func getEnv() {
	GlobalConfig.FileHost = os.Getenv("FILES_HOST")
}

/*
   初始化config
*/
func RcInit(buildtime, commit_id string) {
	fmt.Println("PKG:rc init!")
	var (
		err error
	)
	if err = GlobalConfig.loadLocalConf(localConfig); err != nil {
		panic(err)
	} else {
		GlobalConfig.BuildTime = buildtime
		GlobalConfig.CommitID = commit_id
		getEnv()
		GlobalConfig.Println()
	}

}

/*
   加载本地配置
*/
func (this *confSt) loadLocalConf(filename string) error {

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(buf, this); err != nil {
		return err
	}

	fmt.Sscanf(this.Version, "%d.%d.%d", &this.VerX, &this.VerY, &this.VerZ)

	return nil
}

/*
   打印config
*/
func (this *confSt) Println() {
	fmt.Printf("%+v\n", this)
}
