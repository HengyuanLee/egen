package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx"
)

var (
	genjson = newGenjson()
)

func Json() *Genjson {
	return genjson
}

type Genjson struct {
	genPath string
}

func newGenjson() *Genjson {
	g := &Genjson{}
	return g
}
func (g *Genjson) Gen(outPath string) {
	g.genPath = outPath + "/"

	for _, f := range Xlsxfiles {
		fn := filepath.Base(f)
		if !strings.HasPrefix(fn, "~$") && strings.HasSuffix(fn, ".xlsx") {
			g.genLocal(f)
		} else {
			Warn("json: 排除非xlsx文件 ： " + f)
		}
	}
}

func (g *Genjson) genLocal(file string) {
	Info("json: 正在读取Excel文件  ： " + file)

	filename := filepath.Base(file)
	filename = strings.TrimSuffix(filename, ".xlsx")

	//枚举另外生成
	g.genEnumFile(filename)

	xf, ok := getFile(filename)
	if ok {
		for _, sheet := range xf.Sheets {
			if !strings.HasPrefix(sheet.Name, "!") && getSheetType(sheet) == "object" {
				var buf bytes.Buffer
				//只生成主表，即生成与文件名称相同的sheet
				b := g.processSheet(filename, sheet.Name)

				buf.WriteString(b.String())
				outfile := g.genPath + sheet.Name + ".json"
				Info("json: 正在生成文件 ：" + outfile)
				Info("------------------------------------------------------------------------")
				_, err := os.Stat(outfile)
				if err == nil {
					os.Remove(outfile)
				}
				os.MkdirAll(filepath.Dir(outfile), 0755)
				ioutil.WriteFile(outfile, buf.Bytes(), 0666)
			}
		}
	}
}
func (g *Genjson) genEnumFile(fname string) {

	xf, ok := getFile(fname)
	if ok {
		var enumBuf bytes.Buffer
		esixt := false //是否存在枚举表
		for _, sheet := range xf.Sheets {
			if !strings.HasPrefix(sheet.Name, "!") && getSheetType(sheet) == "enum" {
				if !esixt {
					esixt = true
					enumBuf.WriteString("{\n")
				} else {
					//如果已经存在了值，后面要添加，那么要加“,”和换行再拼接。
					enumBuf.WriteString(",\n")
				}
				g.parseEnum(sheet, &enumBuf)
			}
		}
		if esixt {
			enumBuf.WriteString("\n}")
			outfile := g.genPath + fname + "_Enum.json"
			Info("json: 正在生成文件 ：" + outfile)
			Info("------------------------------------------------------------------------")
			_, euerr := os.Stat(outfile)
			if euerr == nil {
				os.Remove(outfile)
			}
			os.MkdirAll(filepath.Dir(outfile), 0755)
			ioutil.WriteFile(outfile, enumBuf.Bytes(), 0666)
		}
	}
}
func (g *Genjson) processSheet(filename string, sheetName string) *bytes.Buffer {
	Info("json: 正在读取子表 sheet : " + "/" + sheetName)
	level := 0
	buf := bytes.NewBufferString("")
	sheet, ok := getFileSheet(filename, sheetName)
	if ok {
		buf.WriteString(getTabs(level) + "{\n")
		if len(sheet.Rows) < 4 {
			Fatal("json: 表头小于4行" + sheet.Name)
			return buf
		}
		isList := true
		isRoot := true
		g.processSheetLoop(buf, filename, sheetName, make([]string, 0), level, isRoot, isList)
		buf.WriteString("\n" + getTabs(level) + "}")
	}
	return buf
}

func (g *Genjson) parseEnum(sheet *xlsx.Sheet, buf *bytes.Buffer) {
	if len(sheet.Rows) < 2 {
		Error("json: 枚举表表头小于2行 : " + sheet.Name)
		return
	}
	buf.WriteString("")
	nameList := make(map[string]bool, 0)
	valueList := make(map[string]bool, 0)

	buf.WriteString("\t\"" + sheet.Name + "\":{\n")
	for index, row := range sheet.Rows {
		if index > 1 {
			buf.WriteString("\t\t")
			name := row.Cells[0].String()
			value := row.Cells[1].String()
			if name == "" || value == "" {
				Error("json: 错误的枚举配置，字段或值配置为空：" + sheet.Name + "/" + name)
				return
			}
			_, ok := valueList[value]
			if ok {
				Error("json: 重复的枚举值：" + sheet.Name + ":" + value)
				return
			}
			_, ok = nameList[name]
			if ok {
				Error("json: 重复的枚举字段名：" + sheet.Name + ":" + name)
				return
			}
			valueList[value] = true
			nameList[name] = true
			buf.WriteString("\"" + name + "\":" + value)
			if index < len(sheet.Rows)-1 {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		}
	}
	buf.WriteString("\t}")
}

func (g *Genjson) processSheetLoop(buf *bytes.Buffer, filename string, sheetName string, pIds []string, level int, pIsRoot bool, pIsList bool) {
	sheet, ok := getFileSheet(filename, sheetName)
	if !ok {
		return
	}
	fieldCells := sheet.Rows[2].Cells
	typeCells := sheet.Rows[3].Cells
	fieldName := fieldCells[0].String()
	if strings.TrimPrefix(fieldName, "!") != "id" {
		Error("json: 错误！要求表格第一列字段必须为 id 。")
		return
	}

	ownerRows := make([]*xlsx.Row, 0)
	for index, row := range sheet.Rows {
		if index >= 4 {
			if len(row.Cells) > 0 {
				mainId := row.Cells[0].String()
				idName := fieldCells[0].String()
				if strings.TrimPrefix(fieldName, "!") != "id" {
					Error("json: " + sheet.Name + ":要求第一个主键为id, 实际为：" + idName)
					return
				}
				//如果没有填写id则忽略
				if mainId == "" {
					continue
				}
				if pIsRoot {
					ownerRows = append(ownerRows, row)
				} else {
					//id刷选，只打印id匹配成功的
					contain := func(id string) bool {
						for _, v := range pIds {
							if id == v {
								return true
							}
						}
						return false
					}
					//id刷选，只打印id匹配成功的
					if contain(mainId) {
						ownerRows = append(ownerRows, row)
					}
				}
			}
		}
	}
	isFirstSheetWirte := true
	level++
	if !pIsList && len(ownerRows) > 1 {
		pidstr := func() string {
			result := ""
			for _, v := range pIds {
				result += v + ","
			}
			return result
		}
		Error("json: 矛盾，为非数组类型，但找到超过1个实例对象:" + sheet.Name + "  ids:" + pidstr())
		return
	}
	for _, row := range ownerRows {

		if isFirstSheetWirte == false {
			buf.WriteString(",\n")
			buf.WriteString(getTabs(level))
		}
		if isFirstSheetWirte && pIsRoot {
			buf.WriteString(getTabs(level))
		}
		isFirstSheetWirte = false
		if pIsRoot {
			curid := row.Cells[0].String()
			tname := typeCells[0].String()
			if isBaseType(tname) || tname == "string" {
				buf.WriteString("\"" + curid + "\"" + ":")
			} else {
				Error("lua: 每行数据头的字段 id 数据类型只能为基本类型或string")
				continue
			}
		}
		buf.WriteString("{")

		//记录map和list写出的cell，后面不再写出
		writedCellMap := make(map[int]bool, 0)
		for index, cell := range row.Cells {

			writed, ok := writedCellMap[index]
			if ok && writed {
				continue
			}
			writedCellMap[index] = true

			fieldName := fieldCells[index].String()
			//排除属性名为空的字段
			if fieldName == "" {
				continue
			}
			if strings.HasPrefix(fieldName, "!") {
				continue
			}

			//排除属性名为空的字段
			tname := typeCells[index].String()
			fname := fieldCells[index].String()
			value := cell.String()
			//writeTypeStr := typeCells[index].String()
			if fname == "" {
				Warn("忽略空字段：" + tname)
				continue
			}
			if strings.HasPrefix(fname, "!") {
				Warn("忽略字段：" + strings.TrimPrefix(fname, "!"))
				continue
			}

			isAlias := (getCmdValue(sheet, index, "alias") == "true")
			listSplit := getCmdValue(sheet, index, "split")
			if listSplit == "" {
				listSplit = ","
			}

			buf.WriteString("\n")
			if isAlias {
				value = getFileAliasValue(filename, value)
			}
			buf.WriteString(getTabs(level + 1))

			isMap := false
			//对象是否数组类型
			//对象是否数组类型
			isList := isListField(tname)
			if isList {
				tname = strings.TrimPrefix(tname, "<")
				tname = strings.TrimSuffix(tname, ">")
			}else{
				isMap = isMapField(tname)
			}
			if isBaseType(tname) {
				buf.WriteString("\"" + fieldName + "\":")
				if isList {
					value = ""
					for i := index; i < len(row.Cells); i++ {
						_cell := row.Cells[i]
						_name := fieldCells[i].String()
						if fieldName == _name {
							writedCellMap[i] = true
							_valList := strings.Split(_cell.String(), listSplit)
							newv := ""
							for _, v := range _valList {
								if v == "" {
									continue
								}
								if isAlias {
									newv += getFileAliasValue(filename, v)
								} else {
									newv += v
								}
								newv += ","
							}
							newv = strings.TrimSuffix(newv, ",")
							if i != index {
								value += ", "
							}
							value += newv
						}
					}
				}
			} else if tname == "string" {
				buf.WriteString("\"" + fieldName + "\":")
				if isList {
					value = ""
					for i := index; i < len(row.Cells); i++ {
						_cell := row.Cells[i]
						_name := fieldCells[i].String()
						_val := _cell.String()
						if isAlias {
							_val = getFileAliasValue(filename, _val)
						}
						if fieldName == _name {
							writedCellMap[i] = true
							if i != index {
								value += ", "
							}
							value += "\"" + _val + "\""
						}
					}
				} else {
					var sbf bytes.Buffer
					sbf.WriteString("\"")
					sbf.WriteString(value)
					sbf.WriteString("\"")
					value = sbf.String()
				}
			} else if isMap {
				rightIndex := strings.Index(tname, ",")
				if strings.HasPrefix(tname, "<") && strings.HasSuffix(tname, ">") && rightIndex != 1 && rightIndex != -1 {
					kvstr := strings.TrimPrefix(tname, "<")
					kvstr = strings.TrimSuffix(kvstr, ">")
					kvs := strings.Split(kvstr, ",")
					tk := kvs[0]
					tv := kvs[1]
					if !isBaseType(tk) && tk != "string" {
						Error("json: Dictionary只支持key都是string或基本数据类型 id:%s, key:%s ", row.Cells[0].String(),  tname)
					}
					value = ""
					var curfname string
					//从新遍历出所有filedname 为当前map字段名的值对
					for i := index; i < len(row.Cells); i++ {
						_fname := fieldCells[i].String()
						if i == index {
							curfname = _fname
							buf.WriteString("\"" + curfname + "\" : {")
						}
						if _fname == "" || _fname != curfname{
							continue
						}
						if i+1 >= len(row.Cells) {
							Warn("json: map找不到值定义，将被忽略，id:%s, key:%s",row.Cells[0].String(), row.Cells[i].String())
							continue
						}
						_key := strings.TrimSpace(row.Cells[i].String())
						_val := strings.TrimSpace(row.Cells[i+1].String())
						if _key == "" || _val == ""{
							Warn("json: map的key或者value为空，将被忽略，id:%s, key:%s  value:%s",row.Cells[0].String(), _key, _val)
							continue
						}
						isAlias := (getCmdValue(sheet, i, "alias") == "true")
						if isAlias {
							_key = getFileAliasValue(filename, _key)
							_val = getFileAliasValue(filename, _val)
						}
						if curfname != "" && curfname == _fname {
							writedCellMap[i] = true
							writedCellMap[i+1] = true
							if _val == "" {
								_val = g.getEmptyVal(tv)
								continue
							}
							if isBaseType(tv) {
							} else if tv == "string" {
								_val = "\"" + _val + "\""
							} else {
								var subFilename string
								var subSheetName string
								//被点分开说明是外部表
								ss := strings.Split(tv, ".")
								if len(ss) == 2 {
									subFilename = ss[0]
									subSheetName = ss[1]
								} else {
									subFilename = filename
									subSheetName = tv
								}
								subsheet, ok := getFileSheet(subFilename, subSheetName)
								if ok {
									st := getSheetType(subsheet)
									if st == "enum" {
										Error("json: map的value不支持枚举类型 " + curfname)
										continue
									} else if st == "object" {
										isRoot := false
										isList := true
										subBuf := bytes.NewBufferString("")
										//类定义的根据所填的id去找
										ids := strings.Split(_val, listSplit)
										if isAlias {
											for index, id := range ids {
												ids[index] = getFileAliasValue(filename, id)
											}
										}
										g.processSheetLoop(subBuf, subFilename, subSheetName, ids, level+1, isRoot, isList)
										_val = subBuf.String()
									}
								}
							}
							_val = strings.TrimSuffix(_val, ",")
							value += "\n" + getTabs(level+2) + "\"" + _key + "\" : " + _val + ","
						}
					}
					value = strings.TrimSuffix(value, ",")
					value += "\n" + getTabs(level+1) + "}"
				}
			} else {//剩下是object和枚举
				if tname == "" {
					continue
				}
				subFilename := filename
				subSheetName := tname
				//被点分开说明是外部表
				ss := strings.Split(tname, ".")
				if len(ss) == 2 {
					subFilename = ss[0]
					subSheetName = ss[1]
				}
				subsheet, ok := getFileSheet(subFilename, subSheetName)
				ctype := ""
				if !ok {
					Error("json: 找不到sheet：" + tname)
					continue
				} else {
					ctype = getSheetType(subsheet)
				}

				switch ctype {
				case "object":
					buf.WriteString("\"" + fieldName + "\":")
					isRoot := false
					subBuf := bytes.NewBufferString("")
					vals := row.Cells[index].String()
					//类定义的根据所填的id去找
					ids := strings.Split(vals, listSplit)
					if isAlias {
						for index, id := range ids {
							ids[index] = getFileAliasValue(filename, id)
						}
					}
					g.processSheetLoop(subBuf, subFilename, subSheetName, ids, level, isRoot, isList)
					value = subBuf.String()
					tname = ctype
				case "enum":

					buf.WriteString("\"" + fieldName + "\":")
					find := false
					for _, row := range subsheet.Rows {
						_, _val, alias, _, _ := getEnumValue(row)
						//命中此行
						if alias == value || _val == value {
							value = _val
							find = true
							break
						}
					}
					tname = ctype
					if !find {
						Error("json: 不存在的枚举值 ： " + subsheet.Name + " : " + value)
					}
					if value == "" {
						value = subsheet.Rows[2].Cells[1].String()
						Warn("json: 字段 " + fieldName + " 枚举值为空，默认为 ：" + subsheet.Name + "/" + value)
					}
				default:
					Error("json: 找不到Sheet： " + tname)

				}
			}

			if isStrEmpty(value) {
				if isList {
					value = "[]"
				} else {
					value = g.getEmptyVal(tname)
				}
			} else {
				if isList {
					value = "[" + value + "]"
				}
			}
			buf.WriteString(value)
			//一个cell结束了
			if index < len(row.Cells)-1 {
				//如果后面还有没写出的cell，那么加“,”
				for i := index + 1; i < len(row.Cells); i++ {
					tstr := fieldCells[i].String()
					if tstr == "" || strings.HasPrefix(tstr, "!") {
						continue
					}
					writed, ok := writedCellMap[i]
					if !ok || !writed {
						buf.WriteString(",")
						break
					}
				}
			}
		}
		buf.WriteString("\n")
		buf.WriteString(getTabs(level) + "}")
	}
}

func (g *Genjson) getEmptyVal(t string) string {
	result := ""
	if t == "string" {
		result = "\"\""
	} else if isBaseType(t) {
		result = "0"
	} else if t == "object" {
		result = "null"
	} else if t == "enum" {
		result = "0"
	}
	return result
}
