package generator

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

var HasError = false
var ColorLog = true

const (
	debugLevel int = iota
	infoLevel
	warningLevel
	errorLevel
	fatalLevel
	unknownLevel
)
const (
	EMPTY_STR = ""
)
const (
	color_balck = uint8(iota + 30)
	color_red
	color_green
	color_yellow
	color_blue
	color_magenta //洋红
	color_cyan    //洋蓝色
	color_white
)

var (
	levelNameMap = map[int]string{
		debugLevel:   "DEBUG",
		infoLevel:    "INFO",
		warningLevel: "WARN",
		errorLevel:   "ERROR",
		fatalLevel:   "FATAL",
		unknownLevel: "UNKNOWN",
	}
	levelColorMap = map[int]uint8{
		debugLevel:   color_green,
		infoLevel:    color_cyan,
		warningLevel: color_yellow,
		errorLevel:   color_red,
		fatalLevel:   color_magenta,
		unknownLevel: color_magenta,
	}
)
var (
	printLogger             *log.Logger
	fileLogger              *log.Logger
	lastCreateFileTimeStamp int64
	lastCreateFile          string
)

//利用正则表达式压缩字符串，去除空格或制表符
func trim(str string) string {
	if str == EMPTY_STR {
		return EMPTY_STR
	}
	return strings.Trim(str, " ")
}

func init() {
	printLogger = log.New(os.Stdout, "", log.LstdFlags)
}
func output(level int, format string, a ...interface{}) {
	var buff bytes.Buffer
	tag := levelNameMap[level]
	buff.WriteString(" [")
	buff.WriteString(tag)
	buff.WriteString("] ")
	buff.WriteString(format)
	format = buff.String()
	logstr := fmt.Sprintf(format, a...)
	newLogstr := logstr
	newLogstr = setColor(newLogstr, levelColorMap[level])
	printLogger.Output(3, newLogstr)

}

func Debug(format string, a ...interface{}) {
	output(debugLevel, format, a...)
}
func Info(format string, a ...interface{}) {
	output(infoLevel, format, a...)
}
func Warn(format string, a ...interface{}) {
	output(warningLevel, format, a...)
}
func Error(format string, a ...interface{}) {
	HasError = true
	output(errorLevel, format, a...)
	output(errorLevel, "生成失败！")
	var i string
	fmt.Println("按Enter键退出...")
	fmt.Scanln(&i)
	os.Exit(1)

}
func Fatal(format string, a ...interface{}) {
	output(fatalLevel, format, a...)
}
func setColor(s string, color uint8) string {
	if !ColorLog {
		return s
	}
	showPattern := 0
	strColor := color
	backColor := 40
	fs := "\x1b[%d;%d;%dm%s\x1b[0m"
	return fmt.Sprintf(fs, showPattern, strColor, backColor, s)
}
