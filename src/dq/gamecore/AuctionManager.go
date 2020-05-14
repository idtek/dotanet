package gamecore

import (
	"dq/conf"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/timer"
	"dq/utils"
	"math"
	"strconv"
	"sync"
	"time"
)

//公会拍卖物品
var (
	AuctionManagerObj = &AuctionManager{}
)

//公会信息
type AuctionInfo struct {
	db.DB_AuctionInfo
	ReceiveCharactersMap []int32 //分红成员

}

type AuctionManager struct {
	Commoditys *utils.BeeMap //当前服务器交易所的商品

	OperateLock *sync.RWMutex //同步操作锁
	Server      ServerInterface
	//时间到 倒计时
	UpdateTimer *timer.Timer
}

//从数据库载入数据
func (this *AuctionManager) LoadDataFromDB() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	commoditys := make([]db.DB_AuctionInfo, 0)
	db.DbOne.GetAuction(&commoditys)
	for _, v := range commoditys {
		//log.Info("----------ExchangeManager load %d %v", v.Id, &commoditys[k])
		receivecha := utils.GetInt32FromString3(v.Receivecharacters, ";")
		this.Commoditys.Set(v.Id, &AuctionInfo{v, receivecha})
	}

}

//初始化
func (this *AuctionManager) Init(server ServerInterface) {
	log.Info("----------AuctionManager Init---------")
	this.Server = server
	this.Commoditys = utils.NewBeeMap()
	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

	this.UpdateTimer = timer.AddRepeatCallback(time.Second*1, this.Update)
}

//上架商品到世界拍卖行
func (this *AuctionManager) NewAuctionItem2World(player *Player, guildid int32, itemid int32, itemlevel int32, receivecharacters *utils.BeeMap) {
	if player == nil {
		return
	}

	auctioninfo := &db.DB_AuctionInfo{}
	auctioninfo.Guildid = guildid
	auctioninfo.ItemID = itemid
	auctioninfo.Level = itemlevel
	auctioninfo.PriceType = int32(conf.Conf.NormalInfo.AuctionPriceType)
	auctioninfo.Price = int32(conf.Conf.NormalInfo.AuctionFirstPrice)
	auctioninfo.BidderCharacterid = -1
	auctioninfo.Remaintime = int32(conf.Conf.NormalInfo.AuctionTime)
	auctioninfo.BidderType = 1 ////出价者类型 1表示所有人 2表示参与分红的人

	receivecharacter := make([]int32, 0)
	receivecharactername := make([]string, 0)
	chaitems := receivecharacters.Items()
	for k, _ := range chaitems {

		player1 := this.Server.GetPlayerByChaID(k.(int32))
		if player1 == nil {
			continue
		}
		mainunit := player1.MainUnit
		if mainunit == nil {
			continue
		}
		receivecharacter = append(receivecharacter, k.(int32))
		receivecharactername = append(receivecharactername, mainunit.Name)
	}

	//加入拍卖行
	AuctionManagerObj.NewAuctionItem(auctioninfo, receivecharacter, receivecharactername)
}

//上架商品(本函数未删除玩家背包里的道具)
func (this *AuctionManager) NewAuctionItem(data *db.DB_AuctionInfo, receivecha []int32, chanames []string) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	//分红的人
	data.Receivecharacters = ""
	for _, v := range receivecha {
		data.Receivecharacters += strconv.Itoa(int(v)) + ";"
	}
	data.ReceiveCharactersName = ""
	for _, v := range chanames {
		data.ReceiveCharactersName += v + ";"
	}

	db.DbOne.CreateAndSaveAuction(data)

	this.Commoditys.Set(data.Id, &AuctionInfo{*data, receivecha})
}

//存档
func (this *AuctionManager) SaveDbOneLock(data *AuctionInfo) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	this.SaveDbOneNoLock(data)
}
func (this *AuctionManager) SaveDbOneNoLock(data *AuctionInfo) {

	if data == nil {
		return
	}

	data.Receivecharacters = ""
	for _, v := range data.ReceiveCharactersMap {
		data.Receivecharacters += strconv.Itoa(int(v)) + ";"
	}
	db.DbOne.SaveAuction(data.DB_AuctionInfo)
}

//结算
func (this *AuctionManager) AuctionOver(commodity *AuctionInfo) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	if commodity == nil {
		return
	}

	//检查是否还存在
	if this.Commoditys.Check(commodity.Id) == false {
		return
	}
	//成功
	if commodity.BidderCharacterid > 0 {
		//给竞拍者发道具
		oldplayer := this.Server.GetPlayerByChaID(commodity.BidderCharacterid)
		Create_AuctionSucc_Mail(commodity.ItemID, commodity.Level, commodity.BidderCharacterid, oldplayer)
	}

	guild1 := GuildManagerObj.Guilds.Get(commodity.Guildid)
	if guild1 == nil {
		//不存在该公会
		//给分红者分钱
		for _, v := range commodity.ReceiveCharactersMap {
			oldplayer := this.Server.GetPlayerByChaID(v)
			//世界拍卖行 平分
			huode := 1.0 / float64(len(commodity.ReceiveCharactersMap))
			getmoney := int32(math.Ceil((float64(commodity.Price) * huode)))
			log.Info("--fenhong-no guild-%f-", float64(commodity.Price))

			Create_AuctionFenHong_Mail(commodity.PriceType, getmoney, v, oldplayer)
		}
		//本地删除该道具
		this.Commoditys.Delete(commodity.Id)
		//数据库删除
		db.DbOne.DeleteAuction(commodity.Id)
		return
	}

	//公会拍卖行
	guild := guild1.(*GuildInfo)

	//all 计算分红的钱
	allchareceive := utils.NewBeeMap()
	allbilie := 0.0
	for _, v := range commodity.ReceiveCharactersMap {

		one := guild.CharactersMap.Get(v)
		if one == nil {
			allchareceive.Set(v, float64(1.0))
			allbilie += 1
			continue
		}
		onecha := one.(*GuildCharacterInfo)
		pinleveldata := conf.GetGuildPinLevelFileData(onecha.PinLevel)
		if pinleveldata == nil {
			allbilie += 1
			allchareceive.Set(v, float64(1.0))
			continue
		}
		allbilie += float64(pinleveldata.Receive)
		allchareceive.Set(v, float64(pinleveldata.Receive))
	}

	//给分红者分钱
	for _, v := range commodity.ReceiveCharactersMap {
		oldplayer := this.Server.GetPlayerByChaID(v)
		//根据品级来分红
		huode := allchareceive.Get(v).(float64) / allbilie
		getmoney := int32(math.Ceil((float64(commodity.Price) * huode)))
		log.Info("--fenhong--%f--%f--%f", allchareceive.Get(v).(float64), allbilie, float64(commodity.Price))

		Create_AuctionFenHong_Mail(commodity.PriceType, getmoney, v, oldplayer)
	}
	//本地删除该道具
	this.Commoditys.Delete(commodity.Id)
	//数据库删除
	db.DbOne.DeleteAuction(commodity.Id)
	//删除
	guild.AuctionMap.Delete(commodity.Id)

}

//新报价
func (this *AuctionManager) NewPrice(price int32, id int32, player *Player) bool {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if player == nil {
		return false
	}

	commodity1 := this.Commoditys.Get(id)
	if commodity1 == nil {
		//未找到该商品
		player.SendNoticeWordToClient(33)
		return false
	}
	commodity := commodity1.(*AuctionInfo)
	if price <= commodity.Price {
		//报价低于当前价格
		player.SendNoticeWordToClient(34)
		return false
	}
	if commodity.BidderType == 2 { //只有分红的人可以参与拍卖
		isreceivecha := false
		for _, v := range commodity.ReceiveCharactersMap {
			if v == player.Characterid {
				isreceivecha = true
				break
			}
		}
		if isreceivecha == false {
			player.SendNoticeWordToClient(46)
			return false
		}
	}
	//
	if player.BuyItemSubMoney(commodity.PriceType, price) == false {
		//当前没有这么多钱
		//货币不足
		player.SendNoticeWordToClient(commodity.PriceType)
		return false
	}

	//成功
	if commodity.BidderCharacterid > 0 {
		//返回竞拍的钱
		oldplayer := this.Server.GetPlayerByChaID(commodity.BidderCharacterid)
		Create_AuctionFail_Mail(commodity.PriceType, commodity.Price, commodity.BidderCharacterid, oldplayer)
	}

	//重新改写竞拍价格
	commodity.Price = price
	commodity.BidderCharacterid = player.Characterid
	mainunit := player.MainUnit
	if mainunit != nil {
		commodity.BidderCharacterName = mainunit.Name
	}

	//如果倒计时小于30秒 则重新刷新为30秒
	if commodity.Remaintime <= 30 {
		commodity.Remaintime = 30
	}
	//存档
	this.SaveDbOneNoLock(commodity)
	return true
}

func (this *AuctionManager) Close() {
	log.Info("----------AuctionManager Close---------")
	if this.UpdateTimer != nil {
		this.UpdateTimer.Cancel()
		this.UpdateTimer = nil
	}
	teams := this.Commoditys.Items()
	for _, v := range teams {
		this.SaveDbOneLock(v.(*AuctionInfo))
	}
}

//获取世界拍卖行
func (this *AuctionManager) GetWorldAuctionItems(player *Player) *protomsg.SC_GetWorldAuctionItems {
	if player == nil {
		return nil
	}

	data := &protomsg.SC_GetWorldAuctionItems{}
	data.Items = make([]*protomsg.AuctionItem, 0)
	items := this.Commoditys.Items()
	for _, v := range items {
		if v == nil {
			continue
		}
		itemone := v.(*AuctionInfo)
		if itemone.Guildid != -1 {
			continue
		}

		d1 := &protomsg.AuctionItem{}
		d1.ID = itemone.Id
		d1.ItemID = itemone.ItemID
		d1.PriceType = itemone.PriceType
		d1.Price = itemone.Price
		d1.Level = itemone.Level
		d1.BidderCharacterName = itemone.BidderCharacterName
		//		bidderplayer := guild.CharactersMap.Get(itemone.BidderCharacterid)
		//		if bidderplayer != nil {
		//			d1.BidderCharacterName = bidderplayer.(*GuildCharacterInfo).Name
		//		}

		d1.RemainTime = itemone.Remaintime
		d1.BidderType = itemone.BidderType

		d1.ReceivecharactersName = utils.GetStringFromString3(itemone.ReceiveCharactersName, ";")

		data.Items = append(data.Items, d1)
	}

	return data
}

//更新
func (this *AuctionManager) Update() {
	//	this.OperateLock.Lock()
	//	defer this.OperateLock.Unlock()

	//检查玩家上架的物品是否结束
	teams := this.Commoditys.Items()
	for k, v := range teams {
		if v == nil {
			continue
		}

		teams[k].(*AuctionInfo).Remaintime -= 1
		if teams[k].(*AuctionInfo).Remaintime <= 0 {
			//时间到了就结束 并 分红
			this.AuctionOver(v.(*AuctionInfo))
			continue
		}

	}
}
