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
	"time"
)

var (
	GuildPostFileDatas     = make(map[interface{}]interface{})
	GuildPinLevelFileDatas = make(map[interface{}]interface{})
	GuildLevelFileDatas    = make(map[interface{}]interface{})
	GuildMaxPinLevel       = int32(10)
	GuildMaxLevel          = int32(10)
	GuildMapFileDatas      = make(map[interface{}]interface{})
)

//场景配置文件
func LoadGuildFileData() {
	_, GuildPostFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "Post", (*GuildPostFileData)(nil))
	_, GuildPinLevelFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "PinLevel", (*GuildPinLevelFileData)(nil))
	_, GuildLevelFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "Guild", (*GuildLevelFileData)(nil))
	_, GuildMapFileDatas = utils.ReadXlsxOneSheetData("bin/conf/guild.xlsx", "Map", (*GuildMapFileData)(nil))
	format := "15:04:05"
	for k, v := range GuildMapFileDatas {
		GuildMapFileDatas[k].(*GuildMapFileData).StartTime, _ = time.Parse(format, v.(*GuildMapFileData).OpenStartTime)
		GuildMapFileDatas[k].(*GuildMapFileData).EndTime, _ = time.Parse(format, v.(*GuildMapFileData).OpenEndTime)

		GuildMapFileDatas[k].(*GuildMapFileData).OpenWeekDayInt32 = utils.GetInt32FromString3(v.(*GuildMapFileData).OpenWeekDay, ",")
		//		_, tt := time.Parse(format, v.(*GuildMapFileData).OpenStartTime)
		//		GuildMapFileDatas[k].(*GuildMapFileData).StartTime = tt

	}
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

//公会等级
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

//检查进入地图条件 如果不能进入则返回空nil
func CheckGotoGuildMap(id int32, guildlevel int32) *GuildMapFileData {
	mapfiledata := GetGuildMapFileData(id)
	if mapfiledata == nil {
		return nil
	}
	if mapfiledata.IsOpen != 1 {
		return nil
	}
	if mapfiledata.NeedGuildLevel > guildlevel {
		return nil
	}
	nowtime := time.Now()
	nowtime_today, _ := time.Parse("15:04:05", nowtime.Format("15:04:05"))

	if mapfiledata.StartTime.After(nowtime_today) || nowtime_today.After(mapfiledata.EndTime) {
		//log.Info("nowtime_today:")
		return nil
	}

	for _, v := range mapfiledata.OpenWeekDayInt32 {
		if nowtime.Weekday() == time.Weekday(v%7) {

			return mapfiledata
		}
	}

	return nil

}

//公会地图
func GetGuildMapFileData(id int32) *GuildMapFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (GuildMapFileDatas[int(id)])
	if re == nil {
		log.Info("not find GetGuildMapFileData:%d", id)
		return nil
	}
	return (GuildMapFileDatas[int(id)]).(*GuildMapFileData)
}

//单位配置文件数据
type GuildMapFileData struct {
	//配置文件数据
	ID             int32  //
	OpenMonthDay   int32  //在月份的几号开启	-1表示所有 10表示10号
	OpenWeekDay    string //在一周中的星期几开启 -1表示所有 5表示星期五
	OpenStartTime  string //开始时间 字符串
	OpenEndTime    string //结束时间 字符串
	NeedGuildLevel int32  //需要的公会等级
	NextSceneID    int32  //场景ID
	X              float32
	Y              float32

	IsOpen int32 //总开关 1表示开 其他表示关 关闭了就看不到了

	StartTime        time.Time //开始时间日期格式
	EndTime          time.Time //结束时间日期格式
	OpenWeekDayInt32 []int32   //开放周期

}
