#### 一款跨平台灵活的导Excel表工具:
###### 基于go语言编写，支持Windows，OSX，linux
* ###### 下载：
* ###### [工具下载地址](https://github.com/HengyuanLee/egen/releases/tag/v1.0.0)
* ###### [项目源码地址](https://github.com/HengyuanLee/egen)
## egen
#### 特色功能：
* **别名机制**：数值填写支持别名，如填表的int类型内容可以是别名中文，生成代码时会自定替换成真正的int值。方便直观阅读，避免数据量大时阅读混乱。
* **嵌入外表**：填表时如果引用到别的表定义的数据，类型一栏中只需“Excel文件名.表名”，然后在值栏中填写对应的外表id即可。程序不必读表时各种乱跳获取id去查表。
* **支持枚举定义**：不用再手动写各种注释0、1、2等各个代表什么，并且某个值一改容易出乱子。
*  **C#和GO支持第三方库的定义**：如希望能够在表中定义角色位置时，C#中期望的类型是结构体UnityEngine.Vector3，那么可以在自定义行设置，避免在程序开发时还得写一遍赋值代码。
###### 以上的功能是针对开发中可能遇到的问题，提高开发效率和表格配置内容的健壮性。
---------------------

**使用说明**：
* **仅支持.xlsx格式的excel表**
* **支持生成json、lua、go、c# 代码。**
* **无需安装环境，下载即用（用法与protobuf的生成类似）**
* **生成方法：**
**执行gen.bat或者gen.sh即可。**
如，生成json：
`./egen -json_out=./json ./Character.xlsx`

* **egen命令使用说明：**
**输入任意不存在的命令"egen -xxx"即可显示命令说明列表：**


  **version**
	查看当前版本
  **-color_log**
	log输出，是否打开彩色log，仅支持shell终端 (default true)
  **-comment**
	是否生成代码注释 (default true)
	  **-increment string**
   false：生成全部xlxs，true:增量生成xlsx (default "false")
  **-cs_out string**
	生成c#代码路径 output golang code (.cs) (default "./gen/cs")
  **-go_out string**
	生成go代码路径 output golang code (.go) (default "./gen/go")
  **-json_out string**
	生成json代码路径 output json format (.json) (default "./gen/json")
  **-lua_out string**
	生成lua代码路径 output lua code (.lua) (default "./gen/lua")
  **-package string**
	设置生成代码的包名/命名空间 (default "ConfigData")

----------------------------------
#### Excel配置格式：
#### 支持多种数据类型
* **int** **uint** **int32** **int64** **uint32** **uint64** **float** **float32** **long** **bool** **list** **map** **oject**


#### excel配置格式
* **excel表头定义4行。**
+ ##### 第一行：字段的文字描述。
+ ##### 第二行：自定义规则行，
   + 填写格式：key1:value1 key2:value2 (中间有一个空格)
	 **自定义命令有：**
	* **alias**：值为bool类型，定义此列所填的值是否使用别名，使用别名时此列的值会去找	@Alias定义的值来替换本身。
	* **split**：值为string类型，定义此列值为map或list时分割元素使用的分割符。
	* **csType**：值为string类型，定义生成C#类时这个类名用所填的值来强行替换。
	* **goType**：同理上。
	有多个命令时用空格隔开。
+ ##### 第三行：字段名称，应填合法的字段名。
+ ##### 第四行：数据类型，即基本类型，string，自定义object，枚举。
------
#### 填表格式：
###### 如要生成Char.xlsx文件，并引用Global.xlsx文件：
###### 主表定义：
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191207204113940.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3UwMTI3NDA5OTI=,size_16,color_FFFFFF,t_70)
###### 表对象Pos定义：
![在这里插入图片描述](https://img-blog.csdnimg.cn/2019112219572116.png)
###### 表枚举CareerType定义：
值从第3行开始有效。
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191122195859194.png)
###### 表别名alias定义：
值从第3行开始有效。
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191123194054629.png)
###### 执行：
其中，最后的./xlsx为excel文件的根目录，将目录下所有文件生成，也可以只生成指定文件./xlsx/Char.xlsx
```bash
egen -color_log=true -lua_out=./gen/lua -json_out=./gen/json -go_out=./gen/go -cs_out=./gen/cs ./xlsx
```
控制台输出：
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191122201353388.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3UwMTI3NDA5OTI=,size_16,color_FFFFFF,t_70)
#### 生成结果(局部代码)：
json

```json
		{
		"1":{
			"id":1,
			"careerType":1,
			"name":"盖利特",
			"iconId":"10012",
			"attr" : {
				"7" : 44,
				"3" : 5,
				"1" : 34
			},
			"piceIds":[1,2,3, 3,9,7,4],
			"worldPos":{
				"id":1,
				"x":5,
				"y":4,
				"z":9
			},
			"localPos":{
				"x":2,
				"y":54,
				"z":0
			}
		}
	}
```
lua

```c
	{
		[1]={
			id = 1,    --唯一ID
			careerType = 1,    --职业
			name = "盖利特",    --名字
			iconId = "10012",    --图标ID
			attr = {
				[7] = 44,   --战斗/物理
				[3] = 5,   --战斗/物理
				[1] = 34   --战斗/法术
			},    --战斗/物理
			piceIds = {1,2,3, 3,9,7,4},    --合成碎片
			worldPos = {
				id = 1,    --唯一ID
				x = 5,
				y = 4,
				z = 9
			},    --全图位置
			localPos = {
				x = 2,    --x坐标
				y = 54,    --y坐标
				z = 0
			}    --初始位置
		}
	}
```

C#
```csharp
	public class Char{
		public uint id;	 //唯一ID
		public CareerType careerType;	 //职业
		public string name;	 //名字
		public string iconId;	 //图标ID
		public Dictionary<Int32,Int32> attr;	 //战斗/物理
		public int[] piceIds;	 //合成碎片
		public UnityEngine.Vector3 worldPos;	 //全图位置
		public Pos localPos;	 //初始位置
	}
```
go

```go
type Char struct{
	ID	uint `json:"id"` //唯一ID
	CareerType	int `json:"careerType"` //职业
	Name	string `json:"name"` //名字
	IconId	string `json:"iconId"` //图标ID
	Attr	map[int32]int32 `json:"attr"` //战斗/物理
	PiceIds	[]int `json:"piceIds"` //合成碎片
	WorldPos	Vector3 `json:"worldPos"` //全图位置
	LocalPos	Pos `json:"localPos"` //初始位置
}
```
#### C#和GO都是采用读取json方式，
#### C#使用第三方库Newtonsoft来读取。
在Unity中使用，
C#读取json方式：
```csharp
using System;
using UnityEngine;
using ConfigData;
using System.Collections.Generic;

public class GameDataJsonParse : MonoBehaviour
{
    [ContextMenu("测试")]
    private void Start()
    {
        Dictionary<UInt32, Char> chars = DeJson<Dictionary<UInt32, Char>>("Char");
        foreach (var c in chars)
        {
            Debug.Log("key:"+c.Key+"   valu.name : " + c.Value.name);
        }
    }
    public static T DeJson<T>(string jsonFile)
    {
        string text = Resources.Load<TextAsset>(jsonFile).text;
        return Newtonsoft.Json.JsonConvert.DeserializeObject<T>(text);
    }
}
```
go读取json方式：

```go
	var chars[int]Character
	data, _ := ioutil.ReadFile("./Char.json")
	err := json.Unmarshal(data, &chars)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(chars[0].Name)
	}
```

