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
	GuildPostFileDatas     = make(map[interface{}]interface{})
	GuildPinLevelFileDatas = make(map[interface{}]interface{})
	GuildLevelFileDatas    = make(map[interface{}]interface{})
	GuildMaxPinLevel       = int32(10)
	GuildMaxLevel          = int32(10)
)

//场景配置文件
func LoadGuildFileData() {
	_, GuildPostFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "Post", (*GuildPostFileData)(nil))
	_, GuildPinLevelFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "PinLevel", (*GuildPinLevelFileData)(nil))
	_, GuildLevelFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "Guild", (*GuildLevelFileData)(nil))

}
func GetGuildPostFileData(level int32) *GuildPostFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (GuildPostFileDatas[int(level)])
	if re == nil {
		log.Info("not find GuildPostFileDatas:%d", level)
		return nil
	}
	return (GuildPostFileDatas[int(level)]).(*GuildPostFileData)
}

//单位配置文件数据
type GuildPostFileData struct {
	//配置文件数据
	Post                        int32  //职位
	Name                        string //名字
	NoticeWriteAble             int32  //修改公告的权利 0表示无 1表示有
	DeletePlayerWriteAble       int32  //踢人的权利 0表示无 1表示有
	ResponseJoinPlayerWriteAble int32  ////回复玩家申请加入公会的权利 0表示无 1表示有
	DismissWriteAble            int32  //解散公会的权利 0表示无 1表示有
	ExitWriteAble               int32  //退出公会的权限 0表示没有 1表示有
}

func GetGuildPinLevelFileData(pinlevel int32) *GuildPinLevelFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (GuildPinLevelFileDatas[int(pinlevel)])
	if re == nil {
		log.Info("not find GuildPinLevelFileDatas:%d", pinlevel)
		return nil
	}
	return (GuildPinLevelFileDatas[int(pinlevel)]).(*GuildPinLevelFileData)
}

//单位配置文件数据
type GuildPinLevelFileData struct {
	//配置文件数据
	PinLevel  int32   //职位
	Name      string  //名字
	Receive   float32 //分成比列Receive
	UpgradeEx int32   //升级所需要的经验
}

func GetGuildLevelFileData(level int32) *GuildLevelFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (GuildLevelFileDatas[int(level)])
	if re == nil {
		log.Info("not find GuildLevelFileDatas:%d", level)
		return nil
	}
	return (GuildLevelFileDatas[int(level)]).(*GuildLevelFileData)
}

//单位配置文件数据
type GuildLevelFileData struct {
	//配置文件数据
	GuildLevel int32 //职位
	UpgradeEx  int32 //升级所需要的经验
	MaxCount   int32 //最大成员数量
}
