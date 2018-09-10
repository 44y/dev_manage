//监控设备、场所状态; 定时删除过期文件
package managers

import (
	"context"
	"io/ioutil"
	"os"
	. "rc"
	"sync"
	"time"
)

const (
	STATUS_MONITOR_PERIOD = time.Second * 10

	FILE_MONITOR_PERIOD = time.Second * 10
)

func StartMonitor(quit context.Context, wg *sync.WaitGroup) {
	LOG_TRAC.Println("Monitor start!")
	wg.Add(1)
	defer wg.Done()
	defer LOG_INFO.Println("Monitor done!")

	LOG_TRAC.Println("start monitor")

	timeUp_status := time.After(STATUS_MONITOR_PERIOD)
	timeUp_files := time.After(FILE_MONITOR_PERIOD)

	for {
		select {
		case <-quit.Done():
			//LOG_TRAC.Println("got quit,bye~")
			return

		case <-timeUp_status:
			statusHandle()
			timeUp_status = time.After(STATUS_MONITOR_PERIOD)

		case <-timeUp_files:
			filesHandle()
			timeUp_files = time.After(FILE_MONITOR_PERIOD)
		}
	}

}

//monitor netbars and aps data and heartbeat time, send message if necessary
func statusHandle() {
	var (
		err      error
		ex       bool
		time_now = time.Now().Unix()
		update   bool
	)

	aps := make([]DevInfo, 0)
	if err = X.Where("approval = ?", APPROVED).Find(&aps); err != nil {
		LOG_ERRO.Println(err)
		return
	}

	for _, ap := range aps {
		update = false

		if ap.Approval != APPROVED {
			continue
		}

		//心跳超时
		if time_now-ap.LastAliveTime > GlobalConfig.ApAliveTimeout &&
			ap.AliveStatus == STATUS_NORMAL {
			LOG_INFO.Printf("ap:%s alive abnormal, last alive time is %d\n", ap.ApId, ap.LastAliveTime)
			ap.AliveStatus = STATUS_ABNORMAL

			nb := &NetbarInfo{Id: ap.NetbarIndex}
			if ex, err = X.Get(nb); err != nil {
				LOG_ERRO.Println(err)
				continue
			} else if !ex {
				LOG_ERRO.Println("netbar index not exist,", ap.NetbarIndex)
				continue
			}

			msg := &MessageCenter{
				Type:        "ap",
				Orgname:     nb.Orgname,
				Orgcode:     nb.Orgcode,
				Wacode:      nb.Wacode,
				PlaceName:   nb.PlaceName,
				ApId:        ap.ApId,
				Read:        "no",
				CreatedTime: time.Now().Unix(),
			}
			X.InsertOne(msg)
			update = true
		}

		//数据超时
		if time_now-ap.LastDataTime > GlobalConfig.ApDataTimeout &&
			ap.DataStatus == STATUS_NORMAL {
			LOG_INFO.Printf("ap:%s data abnormal, last data time is %d\n", ap.ApId, ap.LastDataTime)
			ap.DataStatus = STATUS_ABNORMAL
			update = true

		}

		//basic超时
		if time_now-ap.LastBasicTime > GlobalConfig.APBasicTimeout &&
			ap.BasicStatus == STATUS_NORMAL {
			LOG_INFO.Printf("ap:%s basic abnormal, last basic time is %d\n", ap.ApId, ap.LastBasicTime)
			ap.BasicStatus = STATUS_ABNORMAL
			update = true

		}
		if update {
			if _, err = updateData(ap.Id, &ap); err != nil {
				LOG_ERRO.Println(err)
			}
		}
	}

	netbars := make([]NetbarInfo, 0)
	if err = X.Where("approval = ?", APPROVED).Find(&netbars); err != nil {
		LOG_ERRO.Println(err)
		return
	}

	for _, netbar := range netbars {
		update = false
		if netbar.Approval != APPROVED {
			continue
		}

		//心跳超时
		if time_now-netbar.LastAliveTime > GlobalConfig.NetbarAliveTimeout &&
			netbar.AliveStatus == STATUS_NORMAL {
			LOG_INFO.Printf("netbar:%s alive abnormal, last alive time is %d\n", netbar.Wacode, netbar.LastAliveTime)
			netbar.AliveStatus = STATUS_ABNORMAL
			netbar.BusinessStatus = NETBAR_BUSINESS_CLOSE

			msg := &MessageCenter{
				Type:        "netbar",
				Orgname:     netbar.Orgname,
				Orgcode:     netbar.Orgcode,
				Wacode:      netbar.Wacode,
				PlaceName:   netbar.PlaceName,
				Read:        "no",
				CreatedTime: time.Now().Unix(),
			}
			X.InsertOne(msg)
			update = true

		}

		//basic超时
		if time_now-netbar.LastBasicTime > GlobalConfig.NetbarBasicTimeout &&
			netbar.BasicStatus == STATUS_NORMAL {
			LOG_INFO.Printf("netbar:%s basic abnormal, last basic time is %d\n", netbar.Wacode, netbar.LastBasicTime)
			netbar.BasicStatus = STATUS_ABNORMAL
			update = true

		}

		if update {
			if _, err = updateData(netbar.Id, &netbar); err != nil {
				LOG_ERRO.Println(err)
			}
		}
	}
}

//remove expired files
func filesHandle() {
	files, err := ioutil.ReadDir(GlobalConfig.FileSysPath)
	if err != nil {
		LOG_ERRO.Println(err)
		return
	}

	var whole_name string
	for _, file := range files {
		if !file.IsDir() &&
			time.Now().Unix()-file.ModTime().Unix() >= GlobalConfig.FileExpiredTime {

			whole_name = GlobalConfig.FileSysPath + "/" + file.Name()
			os.Remove(whole_name)
			LOG_TRAC.Println("delete expired file:", file.Name())

			X.Delete(&FileSys{FileName: whole_name})
		}
	}
}
