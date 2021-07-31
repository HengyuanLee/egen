package generator

import (
	"path/filepath"
	"strings"
)
//是否合法
func Islegal() bool {
	for _, f := range Xlsxfiles {
		fn := filepath.Base(f)
		if !strings.HasPrefix(fn, "~$") && strings.HasSuffix(fn, ".xlsx") {
			f := filepath.Base(f)
			f = strings.TrimSuffix(f, ".xlsx")
			xf, ok := getFile(f)
			if ok {
				_0fieldNames := make([]string, 0)
				_0typeNames := make([]string, 0)
				for i, sheet := range xf.Sheets {
					if !strings.HasPrefix(sheet.Name, "!") && getSheetType(sheet) == "object" {
						fieldCells := sheet.Rows[2].Cells
						typeCells := sheet.Rows[3].Cells
						if len(fieldCells) != len(typeCells) {
							Error("字段和字段值类型需要一一对应 : "+sheet.Name)
							return false
						}
						//第0张sheet
						if i == 0 {
							_0fieldNames = make([]string, len(fieldCells))
							_0typeNames = make([]string, len(typeCells))
							for j, cell := range fieldCells{
								_0fieldNames[j] = cell.String()
							}
							for j, cell := range typeCells{
								_0typeNames[j] = cell.String()
							}
						}else{
							//除了第0张sheet，其它sheet以第0张为准则，字段和类型要求一致
							if len(_0fieldNames) != len(fieldCells){
								Error("字段或字段值类型 配置数量 不一致的表: "+xf.Sheets[0].Name + " | "+sheet.Name)
								return false
							}
							for j, cell := range fieldCells{
								if _0fieldNames[j] != cell.String(){
									Error("字段不一致的表: "+xf.Sheets[0].Name+"."+ _0fieldNames[j] + " | "+sheet.Name+"."+cell.String())
									return false
								}
							}
							for j, cell := range typeCells{
								if _0typeNames[j] != cell.String(){
									Error("字段值类型不一致的表: "+xf.Sheets[0].Name+"."+ _0typeNames[j] + " | "+sheet.Name+"."+cell.String())
									return false
								}
							}
						}
					}
				}
			}
		} else {
			Error("打开xlsx文件失败 ：" + f)
			return false
		}
	}
	return true
}