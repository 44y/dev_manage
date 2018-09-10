package managers

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"io/ioutil"
	"os"
	"os/exec"
	. "rc"
	"strconv"
	"strings"
	"time"
)

var (
	X *xorm.Engine
	//db_config *DbConfigSt

	//log file
	logFile *os.File
)

type newTable interface {
	TableName() string
}

/*
   managers包初始化，加载数据库中已有数据
*/
func MngInit() {
	LOG_TRAC.Println("PKG:managers init!")

	//获取本机Ip
	MyIp = GetMyIp()
	LOG_TRAC.Println("my ip is", MyIp)

	//初始化数据库
	if err := connectDatabase(); err != nil {
		panic(err)
	}

	LoadConfigFromDB(X)

	syncFileSys()

	WriteApporvalNetbars2File()

	//初始化safe quit
	SafeQuitInit()

}

func MngUnInit() {
	if err := X.Close(); err != nil {
		LOG_ERRO.Println(err)
	} else {
		LOG_INFO.Println("Database done!")
	}

}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

/*
	对数据进行DES-CBC加密，然后进行base64
*/
func EncryptAndBase64(data, key, iv string) (string, error) {
	blk, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	data = string(PKCS5Padding([]byte(data), blk.BlockSize()))
	blk_mode := cipher.NewCBCEncrypter(blk, []byte(iv))

	en_data := make([]byte, len(data))

	blk_mode.CryptBlocks(en_data, []byte(data))

	//base64后的字符串要在末尾加换行符，否则openssl解析不出来
	en_str := base64.StdEncoding.EncodeToString(en_data) + "\n"
	return en_str, nil
}

/*
	将已审核场所的场所编码写入ftp目录下的文件里
*/
var (
	netbar_wacodes_filename         = "netbar_wacodes"
	netbar_wacodes_update_time_file = "netbar_wacodes.time"
	netbar_wacodes_filename_tmp     = "netbar_wacodes_tmp"
)

func WriteApporvalNetbars2File() {
	var (
		err       error
		netbars   = make([]NetbarInfo, 0)
		buf       string
		tmp_file  *os.File
		file_name = GlobalConfig.FtpServerDir + "/" + netbar_wacodes_filename
		tmp_name  = GlobalConfig.FtpServerDir + "/" + netbar_wacodes_filename_tmp
		time_name = GlobalConfig.FtpServerDir + "/" + netbar_wacodes_update_time_file
	)
	if err = X.Where("approval = ?", APPROVED).Find(&netbars); err != nil {
		LOG_ERRO.Println(err)
		return
	}

	for _, v := range netbars {
		buf += v.Wacode + "\n"
	}

	if l := len(buf); l > 0 {
		buf = buf[:l-1]
	}
	LOG_TRAC.Println("tmp_file:", tmp_name)
	tmp_file, err = os.OpenFile(tmp_name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		LOG_ERRO.Println(err)
		return
	}

	//加密
	encryption := &EncryptionInfo{}
	X.ID(1).Get(encryption)
	en_buf, err := EncryptAndBase64(buf, encryption.EnKey, encryption.EnIv)
	if err != nil {
		LOG_ERRO.Println(err)
		tmp_file.Close()
		return
	}
	tmp_file.WriteString(en_buf)
	tmp_file.Close()

	err = exec.Command("mv", tmp_name, file_name).Run()
	if err != nil {
		LOG_ERRO.Println(err)
	}

	tmp_file, err = os.OpenFile(tmp_name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		LOG_ERRO.Println(err)
		return
	}

	tmp_file.WriteString(strconv.Itoa(int(time.Now().Unix())))
	tmp_file.Close()
	err = exec.Command("mv", tmp_name, time_name).Run()
	if err != nil {
		LOG_ERRO.Println(err)
	}
}

/*
	加载文件系统中的文件，与数据库同步
*/
func syncFileSys() {
	var (
		err   error
		found bool
		fsys  = make([]FileSys, 0)
	)
	files, err := ioutil.ReadDir(GlobalConfig.FileSysPath)
	if err != nil {
		panic(err)
	}

	if err = X.Find(&fsys); err != nil {
		panic(err)
	}

	//将本地不在数据库中的文件加入数据库
	for _, file := range files {
		found = false
		if !file.IsDir() {
			whole_name := GlobalConfig.FileSysPath + "/" + file.Name()
			LOG_TRAC.Println("file name:", file.Name())

			for _, fs := range fsys {
				if fs.FileName == whole_name {
					found = true
					break
				}
			}

			if !found {
				LOG_TRAC.Println("file", file.Name(), "not in database")
				ds := strings.Split(file.Name(), ".")
				if ds[len(ds)-1] != "csv" {
					LOG_ERRO.Println("delete unknown type file:", file.Name())
					os.Remove(whole_name)
				} else {
					n_fs := &FileSys{FileName: whole_name, FileType: ds[1]}
					X.InsertOne(n_fs)
				}
			}
		}
	}

	//将数据库中不存在的文件记录删除
	for _, fs := range fsys {
		found = false
		for _, file := range files {
			whole_name := GlobalConfig.FileSysPath + "/" + file.Name()
			if fs.FileName == whole_name {
				found = true
				break
			}
		}

		if !found {
			LOG_TRAC.Println("file", fs.FileName, "not in file system")
			LOG_TRAC.Println(X.Delete(&fs))
		}
	}
}

/*
	重置数据库日志
*/
func dbLogReset() {
	var (
		err      error
		tmp_file *os.File
	)

	if tmp_file, err = os.OpenFile(GlobalConfig.XormLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		panic(err)
	} else {
		X.SetLogger(xorm.NewSimpleLogger(tmp_file))
	}

	logFile.Close()
	logFile = tmp_file

	LOG_INFO.Println("database log reset success!")
}

/*
   连接数据库
*/
func connectDatabase() error {
	db_cfg := &GlobalConfig.DbConfig

	dataSourceName := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8", db_cfg.User, db_cfg.Password, db_cfg.Ip, db_cfg.Port, db_cfg.Name)

	var err error
	//新建引擎
	X, err = xorm.NewEngine("mysql", dataSourceName)

	for err != nil {
		LOG_ERRO.Println(err)
		time.Sleep(2 * time.Second)
		X, err = xorm.NewEngine("mysql", dataSourceName)
	}
	LOG_TRAC.Println("new engine success")

	//ping 测试
	err = X.Ping()
	for err != nil {
		LOG_ERRO.Println(err)
		time.Sleep(2 * time.Second)
		err = X.Ping()
	}

	LOG_TRAC.Println("ping success")

	//开启日志
	X.ShowSQL(true)
	if logFile, err = os.OpenFile(GlobalConfig.XormLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		panic(err)
	} else {
		X.SetLogger(xorm.NewSimpleLogger(logFile))
	}

	//设置连接池大小
	X.SetMaxIdleConns(5)
	//设置最大连接数
	X.SetMaxOpenConns(6)

	//名称映射规则主要负责结构体名称到表名和结构体field到表字段的名称映射
	// * SnakeMapper 支持struct为驼峰式命名，表结构为下划线命名之间的转换
	X.SetTableMapper(core.SnakeMapper{})

	//同步数据库表
	syncTables()
	LOG_TRAC.Println("同步数据表成功")

	return nil
}

/*
   同步数据表项
*/
func syncTables() {
	tables, err := X.DBMetas()
	if err != nil {
		panic(err)
	}

	LOG_TRAC.Println("table num is:", len(tables))

	//for _, v := range tables {
	//	LOG_TRAC.Printf("%+v\n", *v)
	//}

	//创建厂商表
	if err = createIfNotExists(new(OrgInfo)); err != nil {
		panic(err)
	}

	//创建场所表
	if err = createIfNotExists(new(NetbarInfo)); err != nil {
		panic(err)
	}

	//创建已删除场所表
	if err = createIfNotExists(new(NetbarInfoDeleted)); err != nil {
		panic(err)
	}

	//创建设备表
	if err = createIfNotExists(new(DevInfo)); err != nil {
		panic(err)
	}

	//创建已删除设备表
	if err = createIfNotExists(new(DevInfoDeleted)); err != nil {
		panic(err)
	}

	var empty bool

	//创建加密信息表，添加默认加密信息
	if err = createIfNotExists(new(EncryptionInfo)); err != nil {
		panic(err)
	}
	if empty, err = X.IsTableEmpty(new(EncryptionInfo)); err != nil {
		panic(err)
	} else if empty {
		X.InsertOne(&EncryptionInfo{
			Id:          1,
			Name:        ENCRYPTION_DEFAULT_NAME,
			Description: ENCRYPTION_DEFAULT_DESCRPITON,
			Type:        ENCRYPTION_TYPE_DES_CBS,
			EnKey:       ENCRYPTION_DEFAULT_KEY,
			EnIv:        ENCRYPTION_DEFAULT_IV,
			DeKey:       ENCRYPTION_DEFAULT_KEY,
			DeIv:        ENCRYPTION_DEFAULT_IV,
		})
	}

	//创建用户信息表
	if err = createIfNotExists(new(UsersInfo)); err != nil {
		panic(err)
	}

	//创建config表
	if err = createIfNotExists(new(DevicesConfig)); err != nil {
		panic(err)
	}
	if empty, err = X.IsTableEmpty(new(DevicesConfig)); err != nil {
		panic(err)
	} else if empty {
		X.InsertOne(&DevicesConfig{
			NetbarAliveTimeout: NETBAR_ALIVE_TIMEOUT_DEFAULT,
			NetbarBasicTimeout: BASIC_TIMEOUT_DEFAULT,
			ApAliveTimeout:     AP_ALIVE_TIMEOUT_DEFAULT,
			ApDataTimeout:      DATA_TIMEOUT_DEFAULT,
			APBasicTimeout:     BASIC_TIMEOUT_DEFAULT,
			FileExpiredTime:    FILE_EXPIRED_DEFAULT,
		})
	}

}

/*
	若表不存在则创建，存在则同步
*/
func createIfNotExists(table newTable) error {
	exist, err := X.IsTableExist(table)
	if err != nil {
		return err
	}

	if !exist {
		return X.Charset("utf8").CreateTable(table)
	}

	//同步
	if err = X.Sync2(table); err != nil {
		LOG_ERRO.Println(err)
	}

	LOG_TRAC.Println("table", table.TableName(), "exists")
	return nil
}
