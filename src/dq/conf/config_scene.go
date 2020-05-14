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
	SceneFileDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadSceneFileData() {
	_, SceneFileDatas = utils.ReadXlsxData("bin/conf/scenes.xlsx", (*SceneFileData)(nil))

}
func GetSceneFileData(typeid int32) *SceneFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (SceneFileDatas[int(typeid)])
	if re == nil {
		log.Info("not find Scenefile:%d", typeid)
		return nil
	}
	return (SceneFileDatas[int(typeid)]).(*SceneFileData)
}
func GetAllScene() map[interface{}]interface{} {
	return SceneFileDatas
}

//单位配置文件数据
type SceneFileData struct {
	//配置文件数据
	TypeID          int32  //类型ID 唯一ID
	ScenePath       string //场景路径
	CreateUnit      string //创建单位
	UnitExperience  int32  //击杀单位获得经验
	UnitGold        int32  //击杀单位获得金币
	UnitDiamond     int32  //击杀单位获得的砖石
	StartX          float32
	StartY          float32
	EndX            float32
	EndY            float32
	IsOpen          int32  //1表示开放 2表示关闭
	InitOpen        int32  //游戏初始化的时候是否开启 1表示开放 2表示关闭
	SceneBuff       string //场景BUFF 进入场景的单位都会添加此BUFF
	ChangeEquipAble int32  //是否可以更换装备 1表示可以 其他表示否
	DeathHuicheng   int32  //死亡后是否回到和平世界 1表示是 其他表示否
	ForceAttackMode int32  //是否强制设置玩家的攻击模式 0表是否  其他表示设置为的攻击模式 (1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)
	//StartX	StartY	EndX	EndY
	CreateBossRule int32 //刷新boss额外条件 0表示无 1表示场景中的全部普通单位死亡后再刷新

	NoPlayerCloseTime int32 //当场景运行超过此时间 且 没有玩家在场景中了就关闭此场景 -1表示永久不关闭
	CloseTime         int32 //当场景运行超过此时间就关闭此场景 -1表示永久不关闭 (竞技场)

	HuiChengMode int32 //0表示回到和平世界 1表示回到本地图随机位置
	PiPeiAble    int32 //是否可以匹配 0否 1可以

	//特殊情况处理
	Exception      int32  //0表示没有特殊情况 1:工会战 2夺宝奇兵 3竞技场
	ExceptionParam string //特殊情况处理参数:1(根据排名获得的公会经验) 2(需要保留宝箱的时间,宝箱道具ID)

}
