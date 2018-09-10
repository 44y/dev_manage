package rc

import (
	"fmt"
	"log"
	"os"
)

/*
func New(out io.writer, prefix string, flag int) *Logger
该函数一共有三个参数：
（1）输出位置out，是一个io.Writer对象，该对象可以是一个文件也可以是实现了该接口的对象。通常我们可以用这个来指定日志输出到哪个文件。
（2）prefix 我们在前面已经看到，就是在日志内容前面的东西。我们可以将其置为 "[Info]" 、 "[Warning]"等来帮助区分日志级别。
（3） flags 是一个选项，显示日志开头的东西，可选的值有：
Ldate         = 1 << iota     // 形如 2009/01/23 的日期
Ltime                         // 形如 01:23:23   的时间
Lmicroseconds                 // 形如 01:23:23.123123   的时间
Llongfile                     // 全路径文件名和行号: /a/b/c/d.go:23
Lshortfile                    // 文件名和行号: d.go:23
LstdFlags     = Ldate | Ltime // 日期和时间
*/

var LOG_DBG, LOG_INFO, LOG_TRAC, LOG_WARN, LOG_ERRO, LOG_FATAL *log.Logger

var outputFile *os.File

const (
	NORMAL = iota
	DEBUG
)

/*
	初始化日志
*/
func LogInit() {
	fmt.Println("PKG:ik_log init!")
	/*
		var out, out_file *os.File
		out_file, _ = os.Create(Running_log)
		switch lvl {
		case NORMAL:
			out = nil
		case DEBUG:
			out = os.Stdout
			out_file = os.Stdout
		default:
			out = nil
		}
	*/
	var err error
	outputFile, err = os.OpenFile(GlobalConfig.DevicesLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	LOG_DBG = log.New(outputFile, "[DBG]", log.LstdFlags|log.Lshortfile)
	LOG_INFO = log.New(outputFile, "[INFO]", log.LstdFlags|log.Lshortfile)
	LOG_TRAC = log.New(outputFile, "[TRAC]", log.LstdFlags|log.Lshortfile)
	LOG_WARN = log.New(outputFile, "[WARNING]", log.LstdFlags|log.Lshortfile)
	LOG_ERRO = log.New(outputFile, "[ERROR]", log.LstdFlags|log.Lshortfile)
	LOG_FATAL = log.New(outputFile, "[FATAL]", log.LstdFlags|log.Lshortfile)
}

/*
	重置日志
*/
func LogReset() {
	var (
		err      error
		tmp_file *os.File
	)

	tmp_file, err = os.OpenFile(GlobalConfig.DevicesLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	LOG_DBG.SetOutput(tmp_file)
	LOG_INFO.SetOutput(tmp_file)
	LOG_TRAC.SetOutput(tmp_file)
	LOG_WARN.SetOutput(tmp_file)
	LOG_ERRO.SetOutput(tmp_file)
	LOG_FATAL.SetOutput(tmp_file)

	outputFile.Close()
	outputFile = tmp_file

	LOG_INFO.Println("reset log file success!")
}
