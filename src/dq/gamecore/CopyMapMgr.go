package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/timer"
	"dq/utils"
	"sync"
	"time"
)

//公会拍卖物品
var (
	CopyMapMgrObj = &CopyMapMgr{}
)

//匹配进副本
var CopyMapSceneIDLock = new(sync.RWMutex)
var CopyMapSceneID = int32(10000)

func GetCopyMapSceneID() int32 {
	CopyMapSceneIDLock.Lock()
	defer CopyMapSceneIDLock.Unlock()
	var re = CopyMapSceneID
	CopyMapSceneID++
	if CopyMapSceneID >= 100000 {
		CopyMapSceneID = int32(10000)
	}
	return re
}

//副本玩家信息
type CopyMapPlayer struct {
	PlayerInfo *Player
	//State          int32 //状态 1可以匹配 2匹配中 3表示副本进行中
	//CopyMapSceneId int32 //进行中的副本场景唯一ID
	PiPeiCopyMapId int32 //正在进行匹配的副本ID
}

type CMMgrFunc interface {
	PiPeiFuBen(players []*CopyMapPlayer, cmfid int32)
}

//副本系统管理器
type CopyMapMgr struct {
	CopyMapPlayerPool *utils.BeeMap //副本匹配池
	OperateLock       *sync.RWMutex //同步操作锁

	Server CMMgrFunc
	//时间到 倒计时
	UpdateTimer *timer.Timer
}

//初始化
func (this *CopyMapMgr) Init(server CMMgrFunc) {
	log.Info("----------CopyMapMgr Init---------")
	this.Server = server
	this.CopyMapPlayerPool = utils.NewBeeMap()
	this.OperateLock = new(sync.RWMutex)
	this.UpdateTimer = timer.AddRepeatCallback(time.Second*3, this.Update)
}

func (this *CopyMapMgr) Close() {
	log.Info("----------CopyMapMgr Close---------")
	if this.UpdateTimer != nil {
		this.UpdateTimer.Cancel()
		this.UpdateTimer = nil
	}

}

//取消匹配
func (this *CopyMapMgr) CancelPiPei(player *Player) {
	if player == nil {
		return
	}
	this.CopyMapPlayerPool.Delete(player.Characterid)
	msg := &protomsg.SC_ShowPiPeiInfo{}
	msg.PiPeiState = 1
	player.SendMsgToClient("SC_ShowPiPeiInfo", msg)
}

//玩家匹配副本
func (this *CopyMapMgr) JionPiPei(player *Player, copymapid int32) {
	if player == nil {
		return
	}

	unit := player.MainUnit
	if unit == nil {
		return
	}
	//
	if unit.RemainCopyMapTimes <= 0 {
		//次数不够
		player.SendNoticeWordToClient(44)
		return
	}
	//等级不够
	if conf.CheckGotoCopyMap(copymapid, unit.Level) == nil {
		player.SendNoticeWordToClient(45)
		return
	}

	cmp := &CopyMapPlayer{}
	cmp.PlayerInfo = player
	cmp.PiPeiCopyMapId = copymapid

	this.CopyMapPlayerPool.Set(player.Characterid, cmp)
	msg := &protomsg.SC_ShowPiPeiInfo{}
	msg.PiPeiState = 2
	player.SendMsgToClient("SC_ShowPiPeiInfo", msg)
}

//获取所有副本信息
func (this *CopyMapMgr) GetCopyMapsInfo(player *Player) *protomsg.SC_GetCopyMapsInfo {
	if player == nil {
		return nil
	}

	copymapplayer := this.CopyMapPlayerPool.Get(player.Characterid)

	msg := &protomsg.SC_GetCopyMapsInfo{}
	msg.Maps = make([]*protomsg.CopyMapInfo, 0)
	for _, v := range conf.CopyMapFileDatas {
		one := &protomsg.CopyMapInfo{}
		one.ID = v.(*conf.CopyMapFileData).ID
		one.NeedLevel = v.(*conf.CopyMapFileData).NeedLevel
		one.NextSceneID = v.(*conf.CopyMapFileData).NextSceneID
		one.PlayerCount = v.(*conf.CopyMapFileData).PlayerCount
		one.State = 1
		if copymapplayer != nil {
			if copymapplayer.(*CopyMapPlayer).PiPeiCopyMapId == v.(*conf.CopyMapFileData).ID {
				one.State = 2 //匹配中
			}
		}
		msg.Maps = append(msg.Maps, one)
	}
	msg.RemainPlayTimes = 5
	unit := player.MainUnit
	if unit != nil {
		msg.RemainPlayTimes = unit.RemainCopyMapTimes
	}

	return msg
}

//更新
func (this *CopyMapMgr) Update() {
	CopyMapPlayerPoolItems := this.CopyMapPlayerPool.Items()

	for _, v1 := range conf.CopyMapFileDatas {
		cmfid := v1.(*conf.CopyMapFileData).ID //检查此ID副本的匹配情况
		allcmplayer := make([]*CopyMapPlayer, 0)
		for _, v := range CopyMapPlayerPoolItems {
			if v.(*CopyMapPlayer).PiPeiCopyMapId == cmfid {
				allcmplayer = append(allcmplayer, v.(*CopyMapPlayer))
				if len(allcmplayer) >= int(v1.(*conf.CopyMapFileData).PlayerCount) { //当人数满足情况了，就成功

					if this.Server != nil {
						this.Server.PiPeiFuBen(allcmplayer, v1.(*conf.CopyMapFileData).NextSceneID)
					}

					//匹配成功后把角色从匹配池里删除
					for _, v2 := range allcmplayer {

						mainunit := v2.PlayerInfo.MainUnit
						if mainunit != nil {
							mainunit.SubRemainCopyMapTimes(1)
						}

						msg := &protomsg.SC_ShowPiPeiInfo{}
						msg.PiPeiState = 1
						v2.PlayerInfo.SendMsgToClient("SC_ShowPiPeiInfo", msg)
						this.CopyMapPlayerPool.Delete(v2.PlayerInfo.Characterid)
					}
					allcmplayer = make([]*CopyMapPlayer, 0)
					return
				}
			}
		}
	}

}
