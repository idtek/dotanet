package gamecore

import (
	"dq/db"
	"dq/log"
	"dq/timer"
	"dq/utils"
	"sync"
	"time"
)

//竞技场信息
var (
	BattleMgrObj       = &BattleMgr{}
	BattleInitScore    = int32(3000) //初始分
	BattleWinAddScore  = int32(25)   //赢一次加分
	BattleLoseAddScore = int32(-25)  //输一次加分
	BattleDrewAddScore = int32(1)    //平一次加分
)

//角色竞技场信息
type CharacterBattleInfo struct {
	db.DB_BattleInfo
}

type BattleMgr struct {
	Characters *utils.BeeMap //当前服务器交易所的商品

	OperateLock *sync.RWMutex //同步操作锁

	//时间到 倒计时
	UpdateTimer *timer.Timer
}

//从数据库载入数据
func (this *BattleMgr) LoadDataFromDB() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	commoditys := make([]db.DB_BattleInfo, 0)
	db.DbOne.GetBattle(&commoditys)
	for _, v := range commoditys {
		//log.Info("----------ExchangeManager load %d %v", v.Id, &commoditys[k])

		this.Characters.Set(v.Characterid, &CharacterBattleInfo{v})
	}

}

//更改数据
func (this *BattleMgr) ChangeData(player *Player, addwin int32, addlose int32, adddrew int32, addmvp int32, addfmvp int32) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if player == nil {
		return
	}
	unit := player.MainUnit
	if unit == nil {
		return
	}
	characterbattleinfo := &CharacterBattleInfo{}
	cb := this.Characters.Get(player.Characterid)
	if cb == nil {
		characterbattleinfo.Characterid = player.Characterid
		characterbattleinfo.Name = unit.Name
		characterbattleinfo.Typeid = unit.TypeID
		characterbattleinfo.WinCount = 0
		characterbattleinfo.LoseCount = 0
		characterbattleinfo.DrewCount = 0
		characterbattleinfo.MvpCount = 0
		characterbattleinfo.FMvpCount = 0
		characterbattleinfo.Score = BattleInitScore
	} else {
		characterbattleinfo = cb.(*CharacterBattleInfo)
	}
	characterbattleinfo.WinCount += addwin
	characterbattleinfo.LoseCount += addlose
	characterbattleinfo.DrewCount += adddrew
	characterbattleinfo.MvpCount += addmvp
	characterbattleinfo.FMvpCount += addfmvp
	if addwin > 0 {
		characterbattleinfo.Score += BattleWinAddScore
	}
	if addlose > 0 {
		characterbattleinfo.Score += BattleLoseAddScore
	}
	if adddrew > 0 {
		characterbattleinfo.Score += BattleDrewAddScore
	}

	this.Characters.Set(characterbattleinfo.Characterid, characterbattleinfo)

}

//初始化
func (this *BattleMgr) Init(server ServerInterface) {
	log.Info("----------BattleMgr Init---------")
	this.Characters = utils.NewBeeMap()
	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

	this.UpdateTimer = timer.AddRepeatCallback(time.Second*10, this.Update)
}

func (this *BattleMgr) Close() {
	log.Info("----------BattleMgr Close---------")
	if this.UpdateTimer != nil {
		this.UpdateTimer.Cancel()
		this.UpdateTimer = nil
	}
}

//更新
func (this *BattleMgr) Update() {
	//	this.OperateLock.Lock()
	//	defer this.OperateLock.Unlock()

}
