package gamecore

import (
	"dq/log"
	"dq/utils"
	"sync"
)

//公会拍卖物品
var (
	GameCoreDataManagerObj = &GameCoreDataManager{}
)

type SceneDropData struct {
	//地图信息
	DropItems     *utils.BeeMap //本地图会掉落的道具
	BossFreshTime int32         //boss刷新时间 秒为单位
}

type GameCoreDataManager struct {
	SceneDropDatas *utils.BeeMap //当前服务器交易所的商品

	OperateLock *sync.RWMutex //同步操作锁

}

//初始化
func (this *GameCoreDataManager) Init() {
	log.Info("----------GameCoreDataManager Init---------")
	this.SceneDropDatas = utils.NewBeeMap()
	this.OperateLock = new(sync.RWMutex)

}

//添加掉落
func (this *GameCoreDataManager) AddDrop(sceneid int32, dropitem int32) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	scenedropitem := this.SceneDropDatas.Get(sceneid)
	if scenedropitem == nil {
		one := &SceneDropData{}
		one.BossFreshTime = 0
		one.DropItems = utils.NewBeeMap()
		one.DropItems.Set(dropitem, dropitem)
		this.SceneDropDatas.Set(sceneid, one)
		return
	}
	scenedropitem.(*SceneDropData).DropItems.Set(dropitem, dropitem)
}

//设置boss刷新时间
func (this *GameCoreDataManager) SetBossFreshTime(sceneid int32, freshtime int32) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	scenedropitem := this.SceneDropDatas.Get(sceneid)
	if scenedropitem == nil {
		one := &SceneDropData{}
		one.BossFreshTime = freshtime
		one.DropItems = utils.NewBeeMap()
		this.SceneDropDatas.Set(sceneid, one)
		return
	}
	scenedropitem.(*SceneDropData).BossFreshTime = freshtime
}

func (this *GameCoreDataManager) Close() {
	log.Info("----------GameCoreDataManager Close---------")

}
