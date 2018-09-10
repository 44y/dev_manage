package main

import (
	"context"
	. "managers"
	. "rc"
	. "restful"
	"sync"
	"time"
)

//build var
var (
	BuildVersion string
	BuildTime    string
	BuildName    string
	CommitID     string
)

/*
	初始化入口，调用各个包的初始化
*/
func init() {
	//初始化本地配置文件
	RcInit(BuildTime, CommitID)

	//初始化log
	LogInit()

	//初始化managers
	MngInit()
}

/*
	反初始化
*/
func unInit() {
	//LOG_TRAC.Println("unInit")

	MngUnInit()
}

func main() {
	LOG_TRAC.Println("main start!")

	defer LOG_INFO.Println("main done!")
	defer unInit()

	//init context
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx3, cancel3 := context.WithCancel(context.Background())

	//init waitgroups
	var wg_httpsrv, wg_pub, wg_sub, wg_moni sync.WaitGroup

	//start http server
	go StartHttpSrv(ctx2, &wg_httpsrv)

	//start publisher
	go StartPublisher(ctx3, &wg_pub)

	//start subscriber
	go StartSubscriber(ctx2, &wg_sub)

	//start monitor
	go StartMonitor(ctx1, &wg_moni)

	//开始监听信号
	quitChan := StartListenSignal()

	//主程序等待退出信号
	<-quitChan

	cancel1()
	//等待monitor退出
	wg_moni.Wait()

	cancel2()
	//等待http退出
	wg_httpsrv.Wait()

	//等待sub退出
	wg_sub.Wait()

	cancel3()
	//等待pub退出
	wg_pub.Wait()

}
