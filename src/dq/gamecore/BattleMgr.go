package gamecore

import (
	"dq/db"
	"dq/log"
	"dq/protobuf"
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
	Rank int32
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
	Characters *utils.BeeMap //竞技场角色

	OperateLock *sync.RWMutex //同步操作锁

	ChangeBattlesInfo []*db.DB_BattleInfo //改变的角色竞技场信息

	CharacterSort CharacterBattleInfoList //竞技场角色排序

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

		this.Characters.Set(v.Characterid, &CharacterBattleInfo{v, -1})
	}

	this.SortNoLock()

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

	this.CharacterSort = make(CharacterBattleInfoList, 0)

	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

	this.UpdateTimer = timer.AddRepeatCallback(time.Second*10, this.Update)
}

//不加锁的排序
func (this *BattleMgr) SortNoLock() {
	//排序
	items2 := this.Characters.Items()
	this.CharacterSort = make(CharacterBattleInfoList, 0)
	for _, v := range items2 {
		this.CharacterSort = append(this.CharacterSort, v.(*CharacterBattleInfo))
	}
	sort.Sort(this.CharacterSort)
	//log.Info("-------Rank len:%d", int32(len(this.CharacterSort)))

	for rank, v := range this.CharacterSort {
		one := this.Characters.Get(v.Characterid)
		if one != nil {
			one.(*CharacterBattleInfo).Rank = int32(rank) + int32(1)
			//log.Info("-------Rank:%d", one.(*CharacterBattleInfo).Rank)
		}
	}
}

func (this *BattleMgr) GetRank(player *Player, data *protomsg.CS_GetBattleRankInfo) *protomsg.SC_GetBattleRankInfo {
	this.OperateLock.RLock()
	defer this.OperateLock.RUnlock()

	if player == nil || data == nil || data.RankCount <= 0 || data.RankStart < 0 {
		return nil
	}

	endindex := data.RankStart + data.RankCount
	//log.Info("-------GetRank:%v  %d  %d", data, endindex, int32(len(this.CharacterSort)))
	msg := &protomsg.SC_GetBattleRankInfo{}
	msg.RankInfo = make([]*protomsg.BattleRankOneInfo, 0)
	for i := data.RankStart; i < endindex && i < int32(len(this.CharacterSort)); i++ {
		rankone := &protomsg.BattleRankOneInfo{}
		rankone.Characterid = this.CharacterSort[i].Characterid
		rankone.Name = this.CharacterSort[i].Name
		rankone.Typeid = this.CharacterSort[i].Typeid
		rankone.Rank = i + 1
		rankone.Score = this.CharacterSort[i].Score
		msg.RankInfo = append(msg.RankInfo, rankone)
	}

	mydata := this.Characters.Get(player.Characterid)
	if mydata == nil {
		msg.MyRankInfo = &protomsg.BattleRankOneInfo{}
		msg.MyRankInfo.Characterid = player.Characterid
		mainunit := player.MainUnit
		if mainunit != nil {
			msg.MyRankInfo.Name = mainunit.Name
			msg.MyRankInfo.Typeid = mainunit.TypeID
		}
		msg.MyRankInfo.Rank = -1
		msg.MyRankInfo.Score = BattleInitScore
	} else {
		msg.MyRankInfo = &protomsg.BattleRankOneInfo{}
		msg.MyRankInfo.Characterid = mydata.(*CharacterBattleInfo).Characterid
		msg.MyRankInfo.Name = mydata.(*CharacterBattleInfo).Name
		msg.MyRankInfo.Typeid = mydata.(*CharacterBattleInfo).Typeid
		msg.MyRankInfo.Rank = mydata.(*CharacterBattleInfo).Rank
		msg.MyRankInfo.Score = mydata.(*CharacterBattleInfo).Score
	}

	return msg
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

	this.SortNoLock()
}
