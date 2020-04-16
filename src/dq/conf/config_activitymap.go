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
	"dq/protobuf"
	"dq/utils"
	"time"
)

var (
	ActivityMapFileDatas = make(map[interface{}]interface{})
	//发送给客户端的数据
	SC_GetActivityMapsInfoMsg = &protomsg.SC_GetActivityMapsInfo{}
)

//场景配置文件
func LoadActivityFileData() {

	_, ActivityMapFileDatas = utils.ReadXlsxData("bin/conf/activitymap.xlsx", (*ActivityMapFileData)(nil))
	format := "15:04:05"
	for k, v := range ActivityMapFileDatas {
		ActivityMapFileDatas[k].(*ActivityMapFileData).StartTime, _ = time.Parse(format, v.(*ActivityMapFileData).OpenStartTime)
		ActivityMapFileDatas[k].(*ActivityMapFileData).EndTime, _ = time.Parse(format, v.(*ActivityMapFileData).OpenEndTime)

		ActivityMapFileDatas[k].(*ActivityMapFileData).OpenWeekDayInt32 = utils.GetInt32FromString3(v.(*ActivityMapFileData).OpenWeekDay, ",")

	}
	SC_GetActivityMapsInfoMsg = GetActivityMapsInfo2SC_GetActivityMapsInfo()
}

//检查进入地图条件 如果不能进入则返回空nil
func CheckGotoActivityMap(id int32, level int32) *ActivityMapFileData {
	mapfiledata := GetActivityMapFileData(id)
	if mapfiledata == nil {
		return nil
	}
	if mapfiledata.IsOpen != 1 {
		return nil
	}
	if mapfiledata.NeedLevel > level {
		return nil
	}
	nowtime := time.Now()
	nowtime_today, _ := time.Parse("15:04:05", nowtime.Format("15:04:05"))

	if mapfiledata.StartTime.After(nowtime_today) || nowtime_today.After(mapfiledata.EndTime) {
		//log.Info("nowtime_today:")
		return nil
	}

	for _, v := range mapfiledata.OpenWeekDayInt32 {
		if nowtime.Weekday() == time.Weekday(v) {

			return mapfiledata
		}
	}

	return nil

}

//活动地图
func GetActivityMapFileData(id int32) *ActivityMapFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (ActivityMapFileDatas[int(id)])
	if re == nil {
		log.Info("not find GetActivityMapFileData:%d", id)
		return nil
	}
	return (ActivityMapFileDatas[int(id)]).(*ActivityMapFileData)
}

//GetActivityMapsInfo
func GetActivityMapsInfo2SC_GetActivityMapsInfo() *protomsg.SC_GetActivityMapsInfo {
	re := &protomsg.SC_GetActivityMapsInfo{}
	re.Maps = make([]*protomsg.ActivityMapInfo, 0)
	for _, v := range ActivityMapFileDatas {
		one := &protomsg.ActivityMapInfo{}
		if v.(*ActivityMapFileData).IsOpen != 1 {
			continue
		}

		one.ID = v.(*ActivityMapFileData).ID
		one.OpenWeekDay = v.(*ActivityMapFileData).OpenWeekDay
		one.OpenStartTime = v.(*ActivityMapFileData).OpenStartTime
		one.OpenEndTime = v.(*ActivityMapFileData).OpenEndTime
		one.NeedLevel = v.(*ActivityMapFileData).NeedLevel
		one.NextSceneID = v.(*ActivityMapFileData).NextSceneID
		one.PriceType = v.(*ActivityMapFileData).PriceType
		one.Price = v.(*ActivityMapFileData).Price
		re.Maps = append(re.Maps, one)
	}

	return re
}

//单位配置文件数据
type ActivityMapFileData struct {
	//配置文件数据
	ID            int32  //
	OpenWeekDay   string //在一周中的星期几开启 -1表示所有 5表示星期五
	OpenStartTime string //开始时间 字符串
	OpenEndTime   string //结束时间 字符串
	NeedLevel     int32  //需要的等级
	NextSceneID   int32  //场景ID
	X             float32
	Y             float32
	PriceType     int32 //价格类型 10000金币 10001砖石
	Price         int32 //价格

	IsOpen int32 //总开关 1表示开 其他表示关 关闭了就看不到了

	StartTime        time.Time //开始时间日期格式
	EndTime          time.Time //结束时间日期格式
	OpenWeekDayInt32 []int32   //开放周期

}
