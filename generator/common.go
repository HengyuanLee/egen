package generator

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"github.com/tealeg/xlsx"
)

var (
	PackageName = "TableData"
	Comment     = true
	Xlsxfiles   = make([]string, 0)
)
var (
	g_aliasMap    = map[string]string{}
	xlsxFiles     = map[string]*xlsx.File{} //保存所加载过的xlsx文件
	aliasMap      = map[string]map[string]string{}
	enumCellIndex = map[string]int{
		"Field":   0,
		"Value":   1,
		"Alias":   2,
		"Comment": 3,
	}
	basetypes = []string{
		"int",
		"uint",
		"int32",
		"int64",
		"uint32",
		"uint64",
		"float",
		"float32",
		"long",
		"bool",
	}
)

func getFullFilename(baseFilename string) string {
	for _, ff := range Xlsxfiles {
		basef := filepath.Base(strings.TrimSpace(ff))
		if baseFilename == strings.TrimSuffix(basef, ".xlsx") {
			return ff
		}
	}
	Error("找不到xlsx文件： " + baseFilename)
	return ""
}
func isBaseType(t string) bool {
	for _, v := range basetypes {
		if t == v {
			return true
		}
	}
	return false
}

func getSheetType(sheet *xlsx.Sheet) string {
	if len(sheet.Rows) == 0 || len(sheet.Rows[0].Cells) == 0 {
		Error(" 空表：无法判断sheet类型定义是Object还是eunm。")
		return ""
	}
	ts := sheet.Rows[0].Cells[0].String()
	ts = strings.TrimSpace(ts)
	if ts == "enum" || ts == "alias" {
		return ts
	} else {
		return "object"
	}
}
func loadXlsxFile(filename string) (*xlsx.File, error) {
	fullFilename := getFullFilename(filename)
	if fullFilename != "" {
		xlf, err := xlsx.OpenFile(fullFilename)
		if err != nil {
			Error("读取xlsx表格出错：" + err.Error())
			return nil, err
		} else {
			return xlf, nil
		}
	}
	return nil, errors.New("找不到文件" + filename)
}
func getFile(filename string) (*xlsx.File, bool) {
	filename = strings.TrimPrefix(filename, "!")
	xlf, ok := xlsxFiles[filename]
	if !ok {
		newXlf, err := loadXlsxFile(filename)
		if err == nil {
			xlf = newXlf
			xlsxFiles[filename] = newXlf
			return newXlf, true
		} else {
			return nil, false
		}
	} else {
		return xlf, true
	}
}
func getFileSheet(filename string, sheetName string) (*xlsx.Sheet, bool) {
	filename = strings.TrimPrefix(filename, "!")
	sheetName = strings.TrimPrefix(sheetName, "!")
	xlf, ok := xlsxFiles[filename]
	if !ok {
		newXlf, err := loadXlsxFile(filename)
		if err == nil {
			xlf = newXlf
			xlsxFiles[filename] = newXlf
		} else {
			return nil, false
		}
	}
	for _, sheet := range xlf.Sheets {
		if strings.TrimPrefix(sheet.Name, "!") == sheetName {
			return sheet, true
		}
	}
	Error("文件 " + filename + " 中找不到sheet: " + sheetName)
	return nil, false
}
func getFileAliasValue(filename string, alias string) string {
	_, ok := aliasMap[filename]
	if !ok {
		aliasMap[filename] = map[string]string{}
		xfs, fok := getFile(filename)
		if fok {
			for _, sheet := range xfs.Sheets {
				if getSheetType(sheet) == "alias" {
					for index, row := range sheet.Rows {
						if index >= 2 { //前面两行为注解
							if len(row.Cells) >= 2 {
								calias := row.Cells[0].String()
								cvalue := row.Cells[1].String()
								calias = strings.TrimSpace(calias)
								cvalue = strings.TrimSpace(cvalue)
								aliasMap[filename][calias] = cvalue
							}
						}
					}
				}
			}
		}
	}
	value, ok := aliasMap[filename][alias]
	if ok {
		return value
	}
	return alias
}

func getEnumValue(row *xlsx.Row) (string, string, string, string, bool) {
	if len(row.Cells) < 2 {
		return "", "", "", "", false
	} else {
		i0 := row.Cells[0].String()
		i1 := row.Cells[1].String()
		i2 := ""
		i3 := ""
		if len(row.Cells) >= 3 {
			i2 = row.Cells[2].String()
		}
		if len(row.Cells) >= 4 && Comment {
			i3 = row.Cells[3].String()
		}
		return i0, i1, i2, i3, true
	}
}

//给定一行，获取去对应枚举配置类型的值
func getEnumCellValue(row *xlsx.Row, filedType string) string {
	if filedType == "Comment" && !Comment {
		return ""
	}
	index, ok := enumCellIndex[filedType]
	if ok {
		return row.Cells[index].String()
	}
	return ""
}
func isStrEmpty(s string) bool {
	if strings.Trim(strings.Trim(s, " "), "\n") == "" {
		return true
	} else {
		return false
	}
}
func getTabs(level int) string {
	var sbuf bytes.Buffer
	for i := 0; i < level; i++ {
		sbuf.WriteString("\t")
	}
	return sbuf.String()
}
func getCmdValue(sheet *xlsx.Sheet, cellIndex int, cmd string) string {
	cmds := getCellCmds(sheet, cellIndex)
	return cmds[cmd]
}

//获取第三行自定义命令参数
func getCellCmds(sheet *xlsx.Sheet, cellIndex int) map[string]string {
	cmap := make(map[string]string)
	if len(sheet.Rows[1].Cells) <= cellIndex {
		return cmap
	}
	cmdStrs := sheet.Rows[1].Cells[cellIndex].String()
	cmdss := strings.Split(cmdStrs, " ")

	for _, cmds := range cmdss {

		kv := strings.Split(cmds, ":")
		if len(kv) != 2 {
			continue
		}
		k := strings.Trim(kv[0], " ")
		v := strings.Trim(kv[1], " ")
		cmap[k] = v
	}
	return cmap
}

//获取注释
func getComment(sheet *xlsx.Sheet, cellIndex int) string {
	if Comment == false {
		return ""
	}
	name := sheet.Rows[0].Cells[cellIndex].String()
	return name
}
func toUp(str string) string {
	if str == "id" {
		return "ID"
	}
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] -= 32
	}
	return string(strArry)
}
func toLow(str string) string {
	if str == "id" {
		return "ID"
	}
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] += 32
	}
	return string(strArry)
}
func isListField(name string) bool {
	rightIndex := strings.Index(name, ",")
	exist_c := rightIndex != -1
	return exist_c == false && strings.HasPrefix(name, "<") && strings.HasSuffix(name, ">")
}
func isMapField(name string) bool {
	rightIndex := strings.Index(name, ",")
	exist_c := rightIndex != -1
	return exist_c && strings.HasPrefix(name, "<") && strings.HasSuffix(name, ">")
}