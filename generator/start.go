package generator

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const Version = "1.0"

// 标准参数
var (

	// 文件类型导出
	flagGenAll      = flag.String("increment", "false", "false：生成全部xlxs，true:增量生成xlsx")
	flagPackageName = flag.String("package", "ConfigData", "设置生成代码的包名/命名空间")
	flagLuaOut      = flag.String("lua_out", "./gen/lua", "生成lua代码路径 output lua code (*.lua)")
	flagJsonOut     = flag.String("json_out", "./gen/json", "生成json代码路径 output json format (*.json)")
	flagGoOut       = flag.String("go_out", "./gen/go", "生成go代码路径 output golang code (*.go)")
	flagCsOut       = flag.String("cs_out", "./gen/cs", "生成c#代码路径 output golang code (*.cs)")
	flagComment     = flag.Bool("comment", true, "是否生成代码注释")
	flagColorLog    = flag.Bool("color_log", true, "log输出，是否打开彩色log，仅支持shell终端")
)

func Start() {

	flag.Parse()
	args := flag.Args()

	if len(args) == 1 && args[0] == "version" {
		fmt.Println("当前版本：" + Version)
		return
	}

	colorLog := *flagColorLog
	ColorLog = colorLog

	_, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		Info("", err)
	}

	xlsxfiles := make([]string, 0)

	g_xlsxfiles := make([]string, 0)
	//确定要生成的xlsx文件
	isValid := func(f string) bool {
		f = path.Base(f)
		if !strings.HasPrefix(f, "!") && !strings.HasPrefix(f, "~$") && strings.HasSuffix(f, ".xlsx") {
			return true
		}
		return false
	}
	for _, xf := range args {
		xf = strings.TrimSpace(xf)
		if xf == "" {
			continue
		}
		if e, _ := PathExists(xf); e {

			if IsDir(xf) {
				xfs, err := WalkDir(xf, ".xlsx")
				if err == nil {
					for _, f := range xfs {
						if isValid(f) {
							xlsxfiles = append(xlsxfiles, f)
						}
					}
				}
			}
			if IsFile(xf) {
				if isValid(xf) {
					xlsxfiles = append(xlsxfiles, xf)
				}
			}
		} else {
			Warn("不存在路径或者文件： " + xf)
		}

	}
	for _, f := range xlsxfiles {
		fname := filepath.Base(strings.TrimSpace(f))
		if strings.HasPrefix(fname, "_") {
			g_xlsxfiles = append(g_xlsxfiles, f)
		}
	}
	if len(xlsxfiles) == 0 {
		Error("指定路径或文件找到数量为0")
		return
	}

	increment := (*flagGenAll == "true")
	xlsxfiles = getGenFiles(xlsxfiles, increment)

	Info("------------------------------------------------------------------------")
	Info("sxgen Version : " + Version)
	if increment == true {
		Info("生成模式： 增量生成。")
	} else {
		Info("生成模式： 全量生成。")
	}
	for _, f := range xlsxfiles {
		Info("确定要生成的文件： " + f)
	}
	len := len(xlsxfiles)
	Info("将要生成的文件数量： " + strconv.Itoa(len))
	if len == 0 {
		return
	}

	luaout := *flagLuaOut
	jsonout := *flagJsonOut
	goout := *flagGoOut
	csout := *flagCsOut
	packageName := *flagPackageName
	comment := *flagComment

	Comment = comment
	PackageName = packageName
	Xlsxfiles = xlsxfiles
	if luaout != "" {
		Lua().Gen(luaout)
	}
	if jsonout != "" {
		Json().Gen(jsonout)
	}
	if goout != "" {
		Go().Gen(goout)
	}
	if csout != "" {
		Cs().Gen(csout)
	}
	success()
}
func success() {
	Debug("生成成功！")
	fmt.Println("按Enter键退出...")
	var i string
	fmt.Scanln(&i)
	os.Exit(0)
}

//根据全/曾量生成，过滤生成的文件
func getGenFiles(files []string, increment bool) []string {
	if len(files) == 0 {
		return files
	}
	result := make([]string, 0)
	curMd5Map := make(map[string]string, 0)
	cacheMd5Map := make(map[string]string, 0)
	for _, xfile := range files {
		md5Code, err := Md5File(xfile)
		if err == nil {
			//Info("md5Code: "+md5Code)
			curMd5Map[xfile] = md5Code
		} else {
			fmt.Println(err.Error())
		}
	}

	gobfile := "./cache"
	exists, err := PathExists(gobfile)

	if err != nil {
		fmt.Println(err.Error())
		return result
	}
	if exists {
		if !increment {
			//全量更新，全部写入
			result = files
			cacheMd5Map = curMd5Map
		} else {
			//增量更新，检测改动
			file, err := os.Open(gobfile)
			defer file.Close()
			//md5Text := string(md5Data)
			if err != nil {
				fmt.Println(err.Error())
				return result
			} else {
				dec := gob.NewDecoder(file)
				dec.Decode(&cacheMd5Map)
				for nfile, nmd5 := range curMd5Map {
					omd5, ok := cacheMd5Map[nfile]
					if !ok || omd5 != nmd5 {
						//之前不存在的文件，或者有改动，则写入
						cacheMd5Map[nfile] = nmd5
						result = append(result, nfile)
					} else {
						//无改动文件
					}
				}
			}
		}
	} else {
		//不存在文件
		result = files
		cacheMd5Map = curMd5Map
		file, err := os.Create(gobfile)
		defer file.Close()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			Debug("创建文件： " + gobfile)
		}
	}
	file, err := os.OpenFile(gobfile, os.O_WRONLY|os.O_TRUNC, 0600)
	defer file.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	enc := gob.NewEncoder(file)
	err2 := enc.Encode(curMd5Map)
	if err2 != nil {
		fmt.Println(err2.Error())
	}
	return result
}
