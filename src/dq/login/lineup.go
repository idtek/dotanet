package login

import (
	"sync"
)

//排队数据
type LineUp struct {
	Uid          int32 //玩家的UID
	ConnectId    int32
	SendGameData []byte //发送给游戏服务器的数据
}

type LineUpMgr struct {
	LineUpPlayers []*LineUp //排队的玩家

	LineUpLock *sync.RWMutex //锁
}

func NewLineUpMgr() *LineUpMgr {
	re := &LineUpMgr{}
	re.Init()

	return re
}

//返回前面还有多少人在排队
func (this *LineUpMgr) Push(data *LineUp) int32 {
	this.LineUpLock.Lock()
	defer this.LineUpLock.Unlock()
	if data == nil {
		return 0
	}
	this.LineUpPlayers = append(this.LineUpPlayers, data)
	return int32(len(this.LineUpPlayers)) - 1
}
func (this *LineUpMgr) Pop() *LineUp {
	this.LineUpLock.Lock()
	defer this.LineUpLock.Unlock()
	if len(this.LineUpPlayers) <= 0 {
		return nil
	}
	re := this.LineUpPlayers[0]
	this.LineUpPlayers = this.LineUpPlayers[1:]
	return re
}

//取消排队
func (this *LineUpMgr) Cancel(uid int32) {
	this.LineUpLock.Lock()
	defer this.LineUpLock.Unlock()
	for k, v := range this.LineUpPlayers {
		if v == nil {
			continue
		}
		if v.Uid == uid {
			this.LineUpPlayers = append(this.LineUpPlayers[:k], this.LineUpPlayers[k+1:]...)
			break
		}
	}
}

//获取还有多少人在我前面
func (this *LineUpMgr) GetFrontMeCount(uid int32) int32 {
	this.LineUpLock.RLock()
	defer this.LineUpLock.RUnlock()
	for k, v := range this.LineUpPlayers {
		if v == nil {
			continue
		}
		if v.Uid == uid {
			return int32(k)
		}
	}
	return int32(len(this.LineUpPlayers))
}

func (this *LineUpMgr) Init() {
	this.LineUpLock = new(sync.RWMutex)
	this.LineUpPlayers = make([]*LineUp, 0)
}
