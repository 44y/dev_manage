//向订阅者发布场所状态变更消息

package managers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gmux "github.com/gorilla/mux"
	"io/ioutil"
	"net"
	"net/http"
	. "rc"
	"strconv"
	"sync"
	"time"
)

var (
	http2pubChan = make(chan *http2pubSt)             //订阅新增或删除消息
	pubMsgChan   = make(chan *pubMsgSt, MAX_MSG_CHAN) //场所状态变更发送消息
)

//http server向publisher发送
type http2pubSt struct {
	req          *http.Request
	writer       http.ResponseWriter
	pub2httpChan chan *pub2httpSt //publisher 处理完订阅消息后回复
}

//http请求处回复
type pub2httpSt struct {
	ret  bool
	code int //错误码
	err  error
}

//订阅者信息
type subInfoSt struct {
	Users  map[string]*UserInfoSt //key:订阅者ID，value:注册详细信息
	req    *http.Request
	writer http.ResponseWriter

	connErrorChan chan string //连接断开时发送
}

type UserInfoSt struct {
	UsersInfo //数据库字段

	Wg *sync.WaitGroup `xorm:"-"   json:"-"` //各个user的wg
}

//备用结构，订阅port为string类型时
type UserInfo_bak struct {
	TransType    string   `json:"trans_type"` //tcp,udp
	Host         string   `json:"host"`       //ip address
	Port         string   `json:"port"`
	Msgtype_json []string `json:"msgtype"` //"netbars"
	RegId        string   `json:"-"`       //订阅者ID
}

//发布消息格式
type pubMsgSt struct {
	MsgType   string `json:"msgtype"` //"netbars", "aps"
	OrgCode   string `json:"security_software_orgcode"`
	OrgName   string `json:"security_software_orgname"`
	PlaceName string `json:"place_name"`
	WaCode    string `json:"netbar_wacode"`
	Action    int    `json:"action"`  //0：新增场所，1：场所信息变更，2：删除场所
	EnType    int    `json:"en_type"` //0:des-cbc加密
	DeKey     string `json:"de_key"`
	DeIv      string `json:"de_iv"`
}

/*
	从数据中加载已注册用户
*/
func (this *subInfoSt) loadFromDb() error {
	LOG_TRAC.Println("loadFromDb")

	var err error
	data := make([]UsersInfo, 0)

	err = X.Find(&data)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Println(data)
	for _, v := range data {
		user := &UserInfoSt{}
		user.Wg = &sync.WaitGroup{}
		user.UsersInfo = v
		if err = sendAllData(&v); err != nil {
			LOG_ERRO.Println(err)
			X.Delete(v)
			continue
		}

		this.Users[v.RegId] = user
		LOG_TRAC.Println("load user:", user.RegId, "from db")
	}
	return nil
}

/*
	建立一个udp连接，发送msg
*/
func sendUdp(host string, port int, msg []byte) error {
	var (
		err  error
		conn net.Conn
	)
	LOG_TRAC.Println("Start dial..")

	if conn, err = net.DialTimeout("udp", host+":"+strconv.Itoa(port), DIAL_TIMEOUT*time.Second); err != nil {
		return err
	}

	//conn.SetDeadline(time.Now().Add(time.Second * 2))

	defer conn.Close()
	LOG_TRAC.Println("connect success:", host, port)
	LOG_TRAC.Println("start send message:", string(msg))
	if _, err = conn.Write(msg); err != nil {
		return err
	}
	return nil
}

/*
	向用户发送消息
*/
func (this *UserInfoSt) newUdpConn(errorChan chan string, msg []byte) {
	this.Wg.Add(1)
	defer LOG_INFO.Println("connect closed:", this.Host, this.Port)
	defer this.Wg.Done()

	if err := sendUdp(this.Host, this.Port, msg); err != nil {
		LOG_ERRO.Println(err)
		LOG_FATAL.Println("delete user:", this.RegId)
		errorChan <- this.RegId
		return
	}
	LOG_TRAC.Println("send success!")
}

/*
	生成发送字符串
*/
func createMsg(m *pubMsgSt) ([]byte, error) {
	var (
		json_b []byte
		err    error
	)

	if json_b, err = json.Marshal(*m); err != nil {
		LOG_ERRO.Println(err)
		return nil, err
	}
	/*
		//获取消息长度
		if len_b, err = GetStrlenByte(json_b); err != nil {
			LOG_ERRO.Println(err)
			return nil, err
		}

		msg := CombineByte(len_b, json_b)
	*/
	json_b = bytes.Replace(json_b, []byte("\\u0026"), []byte("&"), -1)
	json_b = bytes.Replace(json_b, []byte("\\u003c"), []byte("<"), -1)
	json_b = bytes.Replace(json_b, []byte("\\u003e"), []byte(">"), -1)
	json_b = bytes.Replace(json_b, []byte("\\u003d"), []byte("="), -1)
	return json_b, nil
}

/*
	将变更信息组成消息发送到各个订阅者
*/
func (this *subInfoSt) publish(m *pubMsgSt) error {
	LOG_TRAC.Println("new msg:", m)
	if len(this.Users) == 0 {
		LOG_TRAC.Println("No sub user")
		return nil
	}

	msg, err := createMsg(m)
	if err != nil {
		LOG_ERRO.Println(err)
		return err
	}
	LOG_TRAC.Println(string(msg))

	var mt int
	switch m.MsgType {
	case MSGTYPE_NETBARS_STR:
		mt = MSGTYPE_NETBARS_BIT

	case MSGTYPE_APS_STR:
		mt = MSGTYPE_APS_BIT
	default:
		LOG_ERRO.Println("unsupported msg type!")
		return errors.New("unsupported msg type!")
	}

	for _, v := range this.Users {
		if v.Msgtype&mt > 0 {
			go v.newUdpConn(this.connErrorChan, msg)
			LOG_TRAC.Println("send msg", v)
		}
	}
	return nil
}

/*
	发送所有数据库数据
*/
func sendAllData(user *UsersInfo) error {
	var (
		err error
	)

	//发送场所消息
	if user.Msgtype&MSGTYPE_NETBARS_BIT > 0 {
		netbars := make([]NetbarInfo, 0)
		if err = X.Where(fmt.Sprintf("approval = %d", APPROVED)).Find(&netbars); err != nil {
			return err
		}
		for _, v := range netbars {
			org := &OrgInfo{Orgcode: v.Orgcode}
			if _, err = X.Get(org); err != nil {
				LOG_ERRO.Println(err)
				continue
			}

			var en struct {
				EnType int
				DeKey  string
				DeIv   string
			}
			if err = getEncryption(org.EncryptId, &en); err != nil {
				LOG_ERRO.Println("getEncryption wrong! encryptid:", org.EncryptId)
				continue
			}

			mst := &pubMsgSt{
				MsgType:   MSGTYPE_NETBARS_STR,
				OrgCode:   v.Orgcode,
				OrgName:   v.Orgname,
				PlaceName: v.PlaceName,
				WaCode:    v.Wacode,
				Action:    ACTION_ADD,
				EnType:    en.EnType,
				DeKey:     en.DeKey,
				DeIv:      en.DeIv,
			}
			msg, err := createMsg(mst)
			if err != nil {
				LOG_ERRO.Println(err)
				continue

			}

			if err = sendUdp(user.Host, user.Port, msg); err != nil {
				LOG_ERRO.Println(err)
				continue
			}
		}
	}

	//发送设备消息
	if user.Msgtype&MSGTYPE_APS_BIT > 0 {
		LOG_TRAC.Println("TODO:发送设备消息")
	}

	return nil
}

/*
   处理POST请求
*/
func (this *subInfoSt) httpPOST(st *pub2httpSt) {
	//解析订阅ID
	params := gmux.Vars(this.req)
	LOG_TRAC.Println("here")
	var (
		reg_id    string
		ok        bool
		err       error
		err_code  int
		user_info = &UserInfoSt{
			Wg: &sync.WaitGroup{},
		}
		user_bak = &UserInfo_bak{}
		bd_b     []byte
	)

	if reg_id, ok = params["reg_id"]; ok == false {
		err_code = http.StatusBadRequest
		err = errors.New("No register RegId!")
		goto ERR
	}

	//解析请求body

	if bd_b, err = ioutil.ReadAll(this.req.Body); err != nil {
		LOG_ERRO.Println(err)
		ErrorResponse(this.writer, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}
	if err = json.Unmarshal(bd_b, user_info); err != nil {
		LOG_ERRO.Println(err)
		//解析body失败是，尝试用备用结构解析
		if err = json.Unmarshal(bd_b, user_bak); err != nil {
			err_code = http.StatusBadRequest
			LOG_ERRO.Println(err)
			goto ERR
		} else {
			user_info.TransType = user_bak.TransType
			user_info.Host = user_bak.Host
			user_info.Port, err = strconv.Atoi(user_bak.Port)
			if err != nil {
				err_code = http.StatusBadRequest
				LOG_ERRO.Println(err)
				goto ERR
			}
			user_info.Msgtype_json = user_bak.Msgtype_json
		}
	}

	LOG_TRAC.Printf("body is %+v\n", user_info)
	//检查body参数
	if err := this.checkPostBody(user_info); err != nil {
		err_code = http.StatusBadRequest
		goto ERR
	}

	user_info.RegId = reg_id

	//ID已存在
	if value, ok := this.Users[reg_id]; ok == true {
		LOG_TRAC.Println("User already exists,update parameter")
		st.ret = true
		st.code = 0
		st.err = nil
		//向新用户发送已有数据
		if err = sendAllData(&user_info.UsersInfo); err != nil {
			LOG_ERRO.Println(err)
			err_code = http.StatusInternalServerError
			goto ERR
		}

		//更新到数据库
		value.TransType = user_info.TransType
		value.Host = user_info.Host
		value.Port = user_info.Port
		value.Msgtype = user_info.Msgtype
		X.Update(&value.UsersInfo)
		return
	}

	//向新用户发送已有数据
	if err = sendAllData(&user_info.UsersInfo); err != nil {
		LOG_ERRO.Println(err)
		err_code = http.StatusInternalServerError
		goto ERR
	}

	st.ret = true
	st.code = 0
	st.err = nil

	//检查是否有ip和端口都相同的用户，有则覆盖新的注册ID
	for k, v := range this.Users {
		if v.Host == user_info.Host && v.Port == user_info.Port {
			LOG_TRAC.Println("old user, reg id changed!", k)

			this.Users[reg_id] = user_info

			if _, err := X.Where("reg_id = ?", k).Update(&this.Users[reg_id].UsersInfo); err != nil {
				LOG_ERRO.Println(err)
				goto ERR
			}
			delete(this.Users, k)
			return
		}
	}

	this.Users[reg_id] = user_info

	LOG_TRAC.Printf("New user:%s\n", user_info.RegId)
	//插入数据库
	if _, err := X.InsertOne(&this.Users[reg_id].UsersInfo); err != nil {
		LOG_FATAL.Println(err)
		goto ERR
	}

	return

ERR:
	st.ret = false
	st.code = err_code
	st.err = err
}

/*
	POST时检查body参数是否合法
*/
func (this *subInfoSt) checkPostBody(body_st *UserInfoSt) error {
	if body_st.TransType != "udp" {
		LOG_ERRO.Println("trans_type wrong!")
		return errors.New("trans_type wrong!")
	}

	for _, m := range body_st.Msgtype_json {
		switch m {
		case MSGTYPE_NETBARS_STR:
			body_st.Msgtype |= MSGTYPE_NETBARS_BIT

		case MSGTYPE_APS_STR:
			body_st.Msgtype |= MSGTYPE_APS_BIT
		default:
			LOG_ERRO.Println("unsupported msg type, ", m)
		}
	}

	return nil
}

/*
   处理DELETE请求
*/
func (this *subInfoSt) httpDELETE(st *pub2httpSt) {
	//解析订阅ID
	params := gmux.Vars(this.req)
	var reg_id string
	var ok bool
	if reg_id, ok = params["reg_id"]; ok == false {
		st.ret = false
		st.code = http.StatusBadRequest
		st.err = errors.New("No register RegId!")
		return
	}
	LOG_TRAC.Println("New delete subscriber:", reg_id)

	//ID未订阅
	if _, ok := this.Users[reg_id]; ok == false {
		st.ret = false
		st.code = http.StatusBadRequest
		st.err = errors.New("ID haven't subscribed!")
		return
	}

	//从数据库删除
	if _, err := X.Delete(&this.Users[reg_id].UsersInfo); err != nil {
		LOG_FATAL.Println(err)
	}

	delete(this.Users, reg_id)
	LOG_TRAC.Println("Delete done:", reg_id)

	st.ret = true
	st.code = 0
	st.err = nil
}

/*
	http订阅入口,检查版本号，将消息发给publisher
*/
func SubMngHandle(w http.ResponseWriter, r *http.Request) {
	LOG_TRAC.Println("url:", r.RequestURI, "method:", r.Method)

	//检测版本号
	ret, data := CheckVersion(r)
	if !ret {
		LOG_ERRO.Println(data)
		ErrorResponse(w, http.StatusBadRequest, "", data, nil)
		return
	}

	//get all registers
	if r.Method == http.MethodGet {
		rslt := make([]UsersInfo, 0)
		if count, err := X.FindAndCount(&rslt); err != nil {
			LOG_ERRO.Println(err)
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, nil)
			return
		} else {
			QueryResponse(w, count, count, rslt)
			return
		}

	}

	//发送消息给publisher
	msg := &http2pubSt{
		req:          r,
		writer:       w,
		pub2httpChan: make(chan *pub2httpSt),
	}
	http2pubChan <- msg

	//publisher回复
	rsp := <-msg.pub2httpChan
	if rsp.ret {
		SuccessResponse(w, nil)
	} else if rsp.err != nil {
		LOG_ERRO.Println(rsp.err)
		ErrorResponse(w, rsp.code, rsp.err.Error(), nil, nil)
	} else {
		ErrorResponse(w, rsp.code, "", nil, nil)
	}

}

/*
=======publisher LOOP，select channels：
	quitChan：主程序发来的退出信号，优雅退出
	http2pubChan: http收到的订阅消息
	pubMsgChan:  状态变更，需要发给订阅者的消息
	connErrorChan: 连接发送失败，删除用户
*/
func StartPublisher(quit context.Context, wg *sync.WaitGroup) {
	LOG_TRAC.Println("Publisher start!")
	wg.Add(1)
	defer LOG_INFO.Println("Publisher done!")
	defer wg.Done()

	subMng := &subInfoSt{
		Users: make(map[string]*UserInfoSt),

		connErrorChan: make(chan string, 1),
	}

	//从数据库读取现有数据
	if err := subMng.loadFromDb(); err != nil {
		LOG_FATAL.Println(err)
		panic(err)
	}
	//for k, v := range subMng.Users {
	//	LOG_TRAC.Println(k, ":", v)
	//}

	for {
		select {
		case <-quit.Done():
			//关闭所有tcp连接
			for _, v := range subMng.Users {
				v.Wg.Wait()
			}

			//LOG_WARN.Println("All conn closed")
			return

		case httpreq := <-http2pubChan:
			//处理http请求
			rsp := new(pub2httpSt)

			subMng.req = httpreq.req
			subMng.writer = httpreq.writer

			switch httpreq.req.Method {
			case http.MethodPost:
				subMng.httpPOST(rsp)

			case http.MethodDelete:
				subMng.httpDELETE(rsp)

			default:
				rsp.ret = false
				rsp.code = http.StatusMethodNotAllowed
				rsp.err = nil
			}
			//处理完成后回复
			httpreq.pub2httpChan <- rsp

		case msg := <-pubMsgChan:
			subMng.publish(msg)

		case id := <-subMng.connErrorChan:
			//待该用户数据处理完毕后 删除注册用户
			subMng.Users[id].Wg.Wait()
			if _, err := X.Delete(&subMng.Users[id].UsersInfo); err != nil {
				LOG_FATAL.Println(err)
			}
			delete(subMng.Users, id)
		}
	}
}
