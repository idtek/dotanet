// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import (
	"dq/log"
	"dq/utils"
)

var (
	ItemFileDatas = make(map[interface{}]interface{})

	//key 为 typeid
	ItemDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadItemFileData() {
	_, ItemFileDatas = utils.ReadXlsxData("bin/conf/item.xlsx", (*ItemFileData)(nil))

	InitItemDatas()

}

//初始化具体技能数据
func InitItemDatas() {
	for _, v := range ItemFileDatas {
		ssd := ItemData{}
		ssd.ItemBaseData = v.(*ItemFileData).ItemBaseData
		ItemDatas[v.(*ItemFileData).TypeID] = &ssd
		log.Info("item %d:%s:%s:%s", v.(*ItemFileData).TypeID, v.(*ItemFileData).Buffs, v.(*ItemFileData).Halos, v.(*ItemFileData).Skills)
	}

	//	for _, v := range ItemFileDatas {
	//		ssd := make([]HaloData, 0)
	//		v.(*HaloFileData).Trans2HaloData(&ssd)

	//		for k, v1 := range ssd {
	//			test := v1
	//			HaloDatas[string(v1.TypeID)+"_"+string(k+1)] = &test
	//		}
	//	}

	//log.Info("----------1---------")

	//log.Info("-:%v", HaloDatas)
	//	for i := 1; i < 5; i++ {
	//		t := GetHaloData(1000, int32(i))
	//		if t != nil {
	//			log.Info("halo %d:%v", i, *t)
	//		}

	//	}

	//log.Info("----------2---------")
}

//:10000表示金币 10001表示砖石  其他表示道具ID
//是否是背包道具
func IsBagItem(typeid int32) bool {
	if typeid != 10000 && typeid != 10001 {
		return true
	}
	return false
}

//获取技能数据 通过技能ID和等级
func GetItemData(typeid int32) *ItemData {
	//log.Info("find unitfile:%d", typeid)

	re := (ItemDatas[typeid])
	if re == nil {
		log.Info("not find skill unitfile:%d", typeid)
		return nil
	}
	return (re).(*ItemData)
}

//打开宝箱 typeid level
func OpenItemBox(item *ItemData) (int32, int32) {
	if item == nil {
		return -1, -1
	}
	if item.Exception != 1 {
		return -1, -1
	}

	params := utils.GetStringFromString3(item.ExceptionParam, ";")
	if len(params) <= 0 {
		return -1, -1
	}
	typeids := make([]int32, len(params))        //道具ID
	gailvquanzhong := make([]int32, len(params)) //概率权重
	//allquanzhong := 0
	for k, v := range params {
		one := utils.GetInt32FromString3(v, ":")
		if len(one) < 2 {
			continue
		}
		typeids[k] = one[0]
		gailvquanzhong[k] = one[1]
		//allquanzhong += one[1]
	}
	gettypeidindex := utils.CheckRandomInt32Arr(gailvquanzhong)
	if gettypeidindex > 0 {
		return typeids[gettypeidindex], 1
	}
	return -1, -1
}

//技能基本数据
type ItemBaseData struct {
	TypeID    int32  //类型ID
	ItemName  string //名字
	Buffs     string //buff
	Halos     string //halo
	Skills    string //skill
	MaxLevel  int32  //最高等级
	PriceType int32  //回收价格类型
	Price     int32  //回收价格

	SaleAble  int32 //是否可以回收 1:是 2:否
	EquipAble int32 //是否可以装备到身上 1:是 2:否
	UseAble   int32 //是否可以使用 1:是 2:否
	//特殊情况处理
	Exception      int32  //0表示没有特殊情况 1:宝箱
	ExceptionParam string //特殊情况处理参数 特殊情况为1的时候:(宝箱能开出的装备和概率)
}

//技能数据 (会根据等级变化的数据)
type ItemData struct {
	ItemBaseData
}

//单位配置文件数据
type ItemFileData struct {
	//配置文件数据
	ItemBaseData
}
