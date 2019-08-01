package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const LogMaxFileSize = 1024 * 1024 * 32

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	fileLogger *log.Logger
	baseFile   *os.File
	count      int32
}

var OnDebugEvent func(logLevel int, errStr string)
var newFileState = int32(0) //新建文件
var ServerName = ""
var LogLevel = "debug"
var LogPath = ""
var LogFlag = 3

func New(strPrev string, strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%v%d%02d%02d_%02d_%02d_%02d.log",
			strPrev,
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}
		baseLogger = log.New(file, "", flag)
		baseFile = file
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.fileLogger = baseLogger
	logger.baseLogger = log.New(os.Stdout, "", flag)
	logger.baseFile = baseFile

	return logger, nil
}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}
	logger.fileLogger = nil
	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	str := fmt.Sprintf(format, a...)
	if OnDebugEvent != nil {
		OnDebugEvent(level, str)
	}

	logger.baseLogger.Output(3, str)
	if logger.fileLogger != nil {
		v := atomic.AddInt32(&logger.count, int32(len(str)+1))
		if v > LogMaxFileSize {
			v1 := atomic.AddInt32(&newFileState, 1)
			if v1 == 1 {
				atomic.StoreInt32(&logger.count, 0)
				New1(ServerName, LogLevel, LogPath, LogFlag)
				atomic.AddInt32(&newFileState, -1)
			} else {
				atomic.AddInt32(&newFileState, -1)
			}
		}
		logger.fileLogger.Output(3, str)
	}

	//+shang yige da
	if level == fatalLevel {
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

var gLogger, _ = New(ServerName, "debug", "", log.LstdFlags)

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Release(format string, a ...interface{}) {
	gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	format = LightRed(format)
	gLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

func Recover(r interface{}) {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, false)
	Error("%v: %s", r, string(buf[:l]))
}

func ReceiveMsg(format string, a ...interface{}) {
	format = Blue(format)
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func SendMsg(format string, a ...interface{}) {
	format = Cyan(format)
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Warn(format string, a ...interface{}) {
	format = Brown(format)
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}
func GM(format string, a ...interface{}) {
	format = Purple(format)
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}
func FightValueChange(format string, a ...interface{}) {
	format = Green(format)
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Close() {
	gLogger.Close()
}
func New1(strPrev string, strLevel string, pathname string, flag int) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return
	}
	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()
		filename := fmt.Sprintf("%v%d%02d%02d_%02d_%02d_%02d_%d.log",
			strPrev,
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second(),
			now.Nanosecond(),
		)
		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return
		}
		baseLogger = log.New(file, "", flag)
		baseFile = file
		// new
		logger := new(Logger)
		logger.level = level
		logger.fileLogger = baseLogger
		logger.baseLogger = log.New(os.Stdout, "", flag)
		logger.baseFile = baseFile

		del := gLogger
		gLogger = logger
		go func() {
			time.Sleep(5 * time.Second)
			del.Close()
		}()
	}
}

//绿色字体，modifier里，第一个控制闪烁，第二个控制下划线
func Green(str string, modifier ...interface{}) string {
	return cliColorRender(str, 32, 0, modifier...)
}

//淡绿
func LightGreen(str string, modifier ...interface{}) string {
	return cliColorRender(str, 32, 1, modifier...)
}

//青色/蓝绿色
func Cyan(str string, modifier ...interface{}) string {
	return cliColorRender(str, 36, 0, modifier...)
}

//淡青色
func LightCyan(str string, modifier ...interface{}) string {
	return cliColorRender(str, 36, 1, modifier...)
}

//红字体
func Red(str string, modifier ...interface{}) string {
	return cliColorRender(str, 31, 0, modifier...)
}

//淡红色
func LightRed(str string, modifier ...interface{}) string {
	return cliColorRender(str, 31, 1, modifier...)
}

//黄色字体
func Yellow(str string, modifier ...interface{}) string {
	return cliColorRender(str, 33, 0, modifier...)
}

//黑色
func Black(str string, modifier ...interface{}) string {
	return cliColorRender(str, 30, 0, modifier...)
}

//深灰色
func DarkGray(str string, modifier ...interface{}) string {
	return cliColorRender(str, 30, 1, modifier...)
}

//浅灰色
func LightGray(str string, modifier ...interface{}) string {
	return cliColorRender(str, 37, 0, modifier...)
}

//白色
func White(str string, modifier ...interface{}) string {
	return cliColorRender(str, 37, 1, modifier...)
}

//蓝色
func Blue(str string, modifier ...interface{}) string {
	return cliColorRender(str, 34, 0, modifier...)
}

//淡蓝
func LightBlue(str string, modifier ...interface{}) string {
	return cliColorRender(str, 34, 1, modifier...)
}

//紫色
func Purple(str string, modifier ...interface{}) string {
	return cliColorRender(str, 35, 0, modifier...)
}

//淡紫色
func LightPurple(str string, modifier ...interface{}) string {
	return cliColorRender(str, 35, 1, modifier...)
}

//棕色
func Brown(str string, modifier ...interface{}) string {
	return cliColorRender(str, 33, 0, modifier...)
}

func cliColorRender(str string, color int, weight int, extraArgs ...interface{}) string {
	var mo []string

	if weight > 0 {
		mo = append(mo, fmt.Sprintf("%d", weight))
	}
	if len(mo) <= 0 {
		mo = append(mo, "0")
	}
	return "\033[" + strings.Join(mo, ";") + ";" + strconv.Itoa(color) + "m" + str + "\033[0m"
}
