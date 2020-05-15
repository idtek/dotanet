package gamecore

import (
	"dq/db"
	"dq/log"
	"dq/timer"
	"dq/utils"
	"sort"
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
type CharacterBattleInfoList []*CharacterBattleInfo

func (p CharacterBattleInfoList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p CharacterBattleInfoList) Len() int      { return len(p) }
func (p CharacterBattleInfoList) Less(i, j int) bool {
	if p[i].Score == p[j].Score {
		if p[i].MvpCount == p[j].MvpCount {
			return p[i].WinCount > p[j].WinCount
		}
		return p[i].MvpCount > p[j].MvpCount
	}

	return p[i].Score > p[j].Score
}

type BattleMgr struct {
	Characters *utils.BeeMap //当前服务器交易所的商品

	OperateLock *sync.RWMutex //同步操作锁

	ChangeBattlesInfo []*db.DB_BattleInfo //改变的角色竞技场信息

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
func (this *BattleMgr) ChangeData(playerchaid int32, name string, typeid int32, addwin int32, addlose int32, adddrew int32, addmvp int32, addfmvp int32) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if playerchaid <= 0 {
		return
	}

	characterbattleinfo := &CharacterBattleInfo{}
	cb := this.Characters.Get(playerchaid)
	if cb == nil {
		characterbattleinfo.Characterid = playerchaid
		characterbattleinfo.Name = name
		characterbattleinfo.Typeid = typeid
		characterbattleinfo.WinCount = 0
		characterbattleinfo.LoseCount = 0
		characterbattleinfo.DrewCount = 0
		characterbattleinfo.MvpCount = 0
		characterbattleinfo.FMvpCount = 0
		characterbattleinfo.Score = BattleInitScore

		//数据库创建信息
		db.DbOne.CreateCharacterBattleInfo(playerchaid)
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

	//改变信息 每10秒入库一次
	this.ChangeBattlesInfo = append(this.ChangeBattlesInfo, &characterbattleinfo.DB_BattleInfo)

}

//获取玩家竞技场信息
func (this *BattleMgr) GetCharacterBattleScore(chaid int32) int32 {
	cb := this.Characters.Get(chaid)
	if cb == nil {

		return BattleInitScore
	}
	return cb.(*CharacterBattleInfo).Score
}

//初始化
func (this *BattleMgr) Init() {
	log.Info("----------BattleMgr Init---------")
	this.Characters = utils.NewBeeMap()
	this.ChangeBattlesInfo = make([]*db.DB_BattleInfo, 0)

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
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	changecount := len(this.ChangeBattlesInfo)
	if changecount <= 0 {
		return
	}

	//入库改变的信息
	db.DbOne.SaveCharacterBattleInfo(this.ChangeBattlesInfo)
	this.ChangeBattlesInfo = make([]*db.DB_BattleInfo, 0)
	//排序
	items2 := this.Characters.Items()
	psort := make(CharacterBattleInfoList, 0)
	for _, v := range items2 {
		psort = append(psort, v.(*CharacterBattleInfo))
	}
	sort.Sort(psort)
}
