//Generated by sxgen, version=1.0. By DESKTOP-E03E9F2\leme at time 2019-11-24 09:18:27
package ConfigData
const(
	CareerType_Warrior = 1    //战士  
	CareerType_Mage = 2    //法师  
	CareerType_GG = 3    //上的  
)
type Char struct{
	ID	uint `json:"id"` //唯一ID
	CareerType	int `json:"careerType"` //职业
	Name	string `json:"name"` //名字
	IconId	string `json:"iconId"` //图标ID
	Attr	map[int32]int32 `json:"attr"` //战斗/物理
	PiceIds	[]int `json:"piceIds"` //合成碎片
	WorldPos	!Vector3 `json:"worldPos"` //全图位置
	LocalPos	Pos `json:"localPos"` //初始位置
}
type Pos struct{
	X	float `json:"x"` //x坐标
	Y	float `json:"y"` //y坐标
	Z	float `json:"z"`
}
type Tree struct{
	ID	int64 `json:"id"` //唯一ID
	Name	string `json:"name"`
	Flowers	int `json:"flowers"`
	Leaf	int `json:"leaf"` //叶子数量
}
