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
		var buf bytes.Buffer
		buf.WriteString(getTabs(0) + "{\n")
		for i, sheet := range xf.Sheets {
			if !strings.HasPrefix(sheet.Name, "!") && getSheetType(sheet) == "object" {
				//只生成主表，即生成与文件名称相同的sheet
				if i > 0 {
					buf.WriteString(",\n")
				}
				b := g.processSheet(filename, sheet.Name)
				buf.WriteString(b.String())
			}
		}
		buf.WriteString("\n" + getTabs(0) + "}")

		outfile := g.genPath + filename + ".json"
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
		if len(sheet.Rows) < 4 {
			Fatal("json: 表头小于4行" + sheet.Name)
			return buf
		}
		g.processSheetLoop(buf, filename, sheetName, level)
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

func (g *Genjson) processSheetLoop(buf *bytes.Buffer, filename string, sheetName string, level int) {
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
				ownerRows = append(ownerRows, row)
			}
		}
	}
	isFirstSheetWirte := true
	level++
	for _, row := range ownerRows {

		if isFirstSheetWirte == false {
			buf.WriteString(",\n")
			buf.WriteString(getTabs(level))
		}
		if isFirstSheetWirte {
			buf.WriteString(getTabs(level))
		}
		isFirstSheetWirte = false
		curid := row.Cells[0].String()
		tname := typeCells[0].String()
		if isBaseType(tname) || tname == "string" {
			buf.WriteString("\"" + curid + "\"" + ":")
		} else {
			Error("lua: 每行数据头的字段 id 数据类型只能为基本类型或string")
			continue
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
					_valList := strings.Split(row.Cells[index].String(), listSplit)
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
					value += newv
				}
			} else if tname == "string" {
				buf.WriteString("\"" + fieldName + "\":")
				if isList {
					value = ""
					_val := row.Cells[index].String()
					if isAlias {
						_val = getFileAliasValue(filename, _val)
					}
					value += ", "
					value += "\"" + _val + "\""
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
					buf.WriteString("\"" + fname + "\" : {")
					_keyValueStrs := strings.Split(row.Cells[index].String(), ",")
					for i := 0; i < len(_keyValueStrs); i++ {
						_kvStr := strings.Split(_keyValueStrs[i], ":")
						if len(_kvStr) != 2{
							Warn("json: map的key或者value为空，将被忽略，id:%s, 配置",row.Cells[i].String())
							continue
						}
						_key := strings.TrimSpace(_kvStr[0])
						_val := strings.TrimSpace(_kvStr[1])
						if _key == "" || _val == ""{
							Warn("json: map的key或者value为空，将被忽略，id:%s, key:%s  value:%s",row.Cells[0].String(), _key, _val)
							continue
						}
						isAlias := (getCmdValue(sheet, index, "alias") == "true")
						if isAlias {
							_key = getFileAliasValue(filename, _key)
							_val = getFileAliasValue(filename, _val)
						}
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
									Error("json: map的value不支持枚举类型 " + fname)
									continue
								} else if st == "object" {
									Error("json: map的value不支持枚举类型 object " + fname)
								}
							}
						}
						_val = strings.TrimSuffix(_val, ",")
						value += "\n" + getTabs(level+2) + "\"" + _key + "\" : " + _val + ","
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
					Error("json: 未知类型"+ subSheetName + fieldName)
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
