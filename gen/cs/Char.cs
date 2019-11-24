//Generated by sxgen, version=1.0. By DESKTOP-E03E9F2\leme at time 2019-11-24 09:39:14
using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
namespace ConfigData{
	public enum CareerType {
		Warrior = 1,    //战士  
		Mage = 2,    //法师  
		GG = 3,    //上的  
	}
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
	public class Pos{
		public float x;	 //x坐标
		public float y;	 //y坐标
		public float z;	
	}
	public class Tree{
		public Int64 id;	 //唯一ID
		public string name;	
		public int flowers;	
		public int leaf;	 //叶子数量
	}
}