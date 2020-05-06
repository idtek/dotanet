package gamecore

import (
	"dq/log"
	"dq/protobuf"

	//"dq/timer"
	"dq/conf"
	"dq/db"
	"dq/utils"
	"strconv"
	"sync"
	"time"
)

var (
	GuildManagerObj = &GuildManager{}
)

//type ServerInterface interface {
//	GetPlayerByID(id int32) *Player
//}

//公会成员信息
type GuildCharacterInfo struct {
	protomsg.GuildChaInfo
	GuildId   int32  //公会ID
	GuildName string //公会名字

}

func NewGuildCharacterInfo(characterinfo *db.DB_CharacterInfo) *GuildCharacterInfo {

	guild := GuildManagerObj.Guilds.Get(characterinfo.GuildId)
	if guild == nil {
		return nil
	}

	guildchainfo := &GuildCharacterInfo{}
	//重新设置公会成员信息
	guild.(*GuildInfo).CharactersMap.Set(characterinfo.Characterid, guildchainfo)
	guildchainfo.Uid = characterinfo.Uid
	guildchainfo.Characterid = characterinfo.Characterid
	guildchainfo.GuildId = characterinfo.GuildId
	guildchainfo.GuildName = guild.(*GuildInfo).Name
	guildchainfo.Name = characterinfo.Name
	guildchainfo.Level = characterinfo.Level
	guildchainfo.Typeid = characterinfo.Typeid
	guildchainfo.PinLevel = characterinfo.GuildPinLevel
	guildchainfo.PinExperience = characterinfo.GuildPinExperience
	guildchainfo.Post = characterinfo.GuildPost
	postdata := conf.GetGuildPostFileData(characterinfo.GuildPost)
	if postdata != nil {
		guildchainfo.PostName = postdata.Name
	}
	pinleveldata := conf.GetGuildPinLevelFileData(characterinfo.GuildPinLevel)
	if pinleveldata != nil {
		guildchainfo.PinLevelName = pinleveldata.Name
		guildchainfo.PinMaxExperience = pinleveldata.UpgradeEx
	}

	return guildchainfo
}

//公会信息
type GuildInfo struct {
	db.DB_GuildInfo
	CharactersMap            *utils.BeeMap //公会成员
	RequestJoinCharactersMap *utils.BeeMap //请求加入公会角色
	AuctionMap               *utils.BeeMap //正在拍卖的商品
	conf.GuildLevelFileData
	//UpgradeEx                int32         //最大经验值 (升级需要的经验值)
}

//公会管理器
type GuildManager struct {
	Guilds      *utils.BeeMap //当前服务器组队信息
	OperateLock *sync.RWMutex //同步操作锁
	Server      ServerInterface

	//地图信息
	MapInfo *protomsg.SC_GetGuildMapsInfo
}

//初始化
func (this *GuildManager) Init(server ServerInterface) {
	log.Info("----------GuildManager Init---------")
	this.Guilds = utils.NewBeeMap()
	this.Server = server
	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

}

//检查是否有同名的公会存在
func (this *GuildManager) CheckName(name string) bool {
	items := this.Guilds.Items()
	for _, v := range items {
		//重名了
		if v.(*GuildInfo).Name == name {
			return true
		}
	}
	return false

}

//获取所有公会排名UI信息
func (this GuildManager) GetGuildRankInfo() *protomsg.SC_GetGuildRankInfo {
	protoallguilds := &protomsg.SC_GetGuildRankInfo{}
	//公会排名信息
	protoallguilds.Guilds = make([]*protomsg.GuildShortInfo, 0)
	allguilds := this.Guilds.Items()
	for _, v := range allguilds {
		//最多显示前20名
		if v == nil || v.(*GuildInfo).Rank >= 20 {
			continue
		}

		one := this.GuildInfo2ProtoGuildShortInfo(v.(*GuildInfo))
		protoallguilds.Guilds = append(protoallguilds.Guilds, one)
	}

	//地图信息
	protoallguilds.MapInfo = &protomsg.GuildMapInfo{}
	mapdata := conf.GetGuildMapFileData(10) //工会战地图ID为10
	if mapdata != nil {
		protoallguilds.MapInfo.ID = mapdata.ID
		protoallguilds.MapInfo.OpenMonthDay = mapdata.OpenMonthDay
		protoallguilds.MapInfo.OpenWeekDay = mapdata.OpenWeekDay
		protoallguilds.MapInfo.OpenStartTime = mapdata.OpenStartTime
		protoallguilds.MapInfo.OpenEndTime = mapdata.OpenEndTime
		protoallguilds.MapInfo.NeedGuildLevel = mapdata.NeedGuildLevel
		protoallguilds.MapInfo.NextSceneID = mapdata.NextSceneID
	}

	return protoallguilds
}

//获取所有公会简短信息
func (this GuildManager) GetAllGuildsInfo() *protomsg.SC_GetAllGuildsInfo {
	protoallguilds := &protomsg.SC_GetAllGuildsInfo{}
	protoallguilds.CreatePriceType = int32(conf.Conf.NormalInfo.CreateGuildPriceType)
	protoallguilds.CreatePrice = int32(conf.Conf.NormalInfo.CreateGuildPrice)
	protoallguilds.Guilds = make([]*protomsg.GuildShortInfo, 0)
	allguilds := this.Guilds.Items()
	for _, v := range allguilds {
		one := this.GuildInfo2ProtoGuildShortInfo(v.(*GuildInfo))
		protoallguilds.Guilds = append(protoallguilds.Guilds, one)
	}
	return protoallguilds
}

//本公会信息转成proto的公会简短信息
func (this *GuildManager) GuildInfo2ProtoGuildShortInfo(guild *GuildInfo) *protomsg.GuildShortInfo {
	guildBaseInfo := &protomsg.GuildShortInfo{}
	guildBaseInfo.ID = guild.Id
	guildBaseInfo.Name = guild.Name
	guildBaseInfo.Level = guild.Level
	guildBaseInfo.Experience = guild.Experience
	guildBaseInfo.MaxExperience = guild.UpgradeEx
	guildBaseInfo.CharacterCount = int32(guild.CharactersMap.Size())
	guildBaseInfo.MaxCount = guild.MaxCount
	guildBaseInfo.PresidentName = ""
	guildBaseInfo.Notice = guild.Notice
	guildBaseInfo.Rank = guild.Rank
	president := guild.CharactersMap.Get(guild.PresidentCharacterid)
	if president != nil {
		guildBaseInfo.PresidentName = president.(*GuildCharacterInfo).Name
	}
	guildBaseInfo.Joinaudit = guild.Joinaudit
	guildBaseInfo.Joinlevellimit = guild.Joinlevellimit
	return guildBaseInfo
}

//获取申请列表信息
func (this *GuildManager) GetJoinGuildPlayer(id int32) *protomsg.SC_GetJoinGuildPlayer {
	guildinfo := &protomsg.SC_GetJoinGuildPlayer{}
	guild1 := this.Guilds.Get(id)
	if guild1 == nil {
		return nil
	}
	guild := guild1.(*GuildInfo)

	//公会申请成员信息
	guildinfo.RequestCharacters = make([]*protomsg.GuildChaInfo, 0)
	rechaitems := guild.RequestJoinCharactersMap.Items()
	for _, v := range rechaitems {
		one := &v.(*GuildCharacterInfo).GuildChaInfo
		guildinfo.RequestCharacters = append(guildinfo.RequestCharacters, one)
	}

	return guildinfo
}

//获取公会信息
func (this *GuildManager) GetGuildInfo(id int32) *protomsg.SC_GetGuildInfo {
	guildinfo := &protomsg.SC_GetGuildInfo{}
	guild1 := this.Guilds.Get(id)
	if guild1 == nil {
		return nil
	}
	guild := guild1.(*GuildInfo)

	//公会信息
	guildinfo.GuildBaseInfo = this.GuildInfo2ProtoGuildShortInfo(guild)
	//公会成员信息
	guildinfo.Characters = make([]*protomsg.GuildChaInfo, 0)
	chaitems := guild.CharactersMap.Items()
	for _, v := range chaitems {
		one := &v.(*GuildCharacterInfo).GuildChaInfo
		guildinfo.Characters = append(guildinfo.Characters, one)
	}
	//	//公会申请成员信息
	//	guildinfo.RequestCharacters = make([]*protomsg.GuildChaInfo, 0)
	//	rechaitems := guild.RequestJoinCharactersMap.Items()
	//	for _, v := range rechaitems {
	//		one := &v.(*GuildCharacterInfo).GuildChaInfo
	//		guildinfo.RequestCharacters = append(guildinfo.RequestCharacters, one)
	//	}

	return guildinfo
}

//设置公会排名
func (this *GuildManager) SetGuildRank(guildid int32, rank int32) {
	guild1 := this.Guilds.Get(guildid)
	if guild1 == nil {
		return
	}
	guild1.(*GuildInfo).Rank = rank
}

//清除所有公会排名
func (this *GuildManager) ReSetGuildRank() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	guilditems := this.Guilds.Items()
	for _, guild := range guilditems {
		guild.(*GuildInfo).Rank = 10000
	}
}

//创建公会
func (this *GuildManager) CreateGuild(name string) *GuildInfo {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if this.CheckName(name) == true {
		return nil
	}

	guild := &GuildInfo{}
	guild.Createday = time.Now().Format("2006-01-02")
	guild.Name = name
	guild.Level = 1
	leveldata := conf.GetGuildLevelFileData(guild.Level)
	if leveldata != nil {
		guild.GuildLevelFileData = *leveldata
	}

	guild.Notice = "欢迎来到(" + name + ")大家庭!"
	guild.Joinaudit = 0
	guild.Joinlevellimit = 1
	guild.CharactersMap = utils.NewBeeMap()
	guild.RequestJoinCharactersMap = utils.NewBeeMap()
	guild.AuctionMap = utils.NewBeeMap()
	//数据库创建信息获得ID
	_, id := db.DbOne.CreateGuild(name, guild.Createday)
	if id < 0 {
		return nil
	}
	guild.Id = id                              //设置公会ID
	guild.Rank = int32(this.Guilds.Size()) + 1 //设置公会排名
	//把公会加入列表
	this.Guilds.Set(guild.Id, guild)

	return guild

}

//改变职位
func (this *GuildManager) ChangePost(player *Player, data *protomsg.CS_ChangePost, targetplayer *Player) bool {
	if player == nil || data == nil {
		return false
	}
	myguild := player.MyGuild
	if myguild == nil {
		return false
	}

	//找到当前公会
	guild1 := this.Guilds.Get(myguild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return false
	}
	guild := guild1.(*GuildInfo)
	postdata := conf.GetGuildPostFileData(myguild.Post)
	if postdata == nil || postdata.PostWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return false
	}
	ischangepost := false
	//申请的玩家角色
	character := targetplayer
	if character == nil { //离线
		guildtargetcharacter := guild.CharactersMap.Get(data.Characterid)
		if guildtargetcharacter == nil {
			//公会里找不到
			return false
		}

		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(data.Characterid, &players)
		if len(players) <= 0 {
			//数据库中找不到该用户
			return false
		}
		//存档数据库
		if players[0].GuildPost != data.Post {
			players[0].GuildPost = data.Post
			ischangepost = true
			db.DbOne.SaveCharacter(players[0])
			//--
			guildtargetcharacter.(*GuildCharacterInfo).Post = data.Post
			postdata := conf.GetGuildPostFileData(data.Post)
			if postdata != nil {
				guildtargetcharacter.(*GuildCharacterInfo).PostName = postdata.Name
			}
		}

	} else { //在线
		targetguild := character.MyGuild
		if targetguild != nil {
			if targetguild.GuildChaInfo.Post != data.Post {
				targetguild.GuildChaInfo.Post = data.Post
				postdata := conf.GetGuildPostFileData(data.Post)
				if postdata != nil {
					targetguild.PostName = postdata.Name
				}
				ischangepost = true
			}

		}
	}

	return ischangepost
}

//编辑公告
func (this *GuildManager) EditorGuildNotice(player *Player, data *protomsg.CS_EditorGuildNotice) bool {
	if player == nil || data == nil {
		return false
	}
	myguild := player.MyGuild
	if myguild == nil {
		return false
	}

	//找到当前公会
	guild1 := this.Guilds.Get(myguild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return false
	}
	guild := guild1.(*GuildInfo)
	postdata := conf.GetGuildPostFileData(myguild.Post)
	if postdata == nil || postdata.NoticeWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return false
	} else {
		guild.Notice = data.Notice
	}

	return true
}

//Operate
//
func (this *GuildManager) GuildOperate(player *Player, data *protomsg.CS_GuildOperate) bool {
	if player == nil || data == nil {
		return false
	}
	myguild := player.MyGuild
	if myguild == nil {
		return false
	}

	//找到当前公会
	guild1 := this.Guilds.Get(myguild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return false
	}
	guild := guild1.(*GuildInfo)

	postdata := conf.GetGuildPostFileData(myguild.Post)
	if data.Code == 1 { //退出公会
		if postdata == nil || postdata.ExitWriteAble != 1 {
			//没有权限 31
			player.SendNoticeWordToClient(31)
			return false
		} else {
			//
			player.MyGuild = nil
			//把数据存入公会中
			guild.CharactersMap.Delete(player.Characterid)
			//存档
			this.SaveDBGuildInfo(guild)
		}
	} else if data.Code == 2 { //解散公会
		if postdata == nil || postdata.DismissWriteAble != 1 {
			//没有权限 31
			player.SendNoticeWordToClient(31)
			return false
		} else {
			//---------在线的玩家 踢出公会--不在线的玩家上线后会找不到公会自动退出

			allplayer := guild.CharactersMap.Items()
			for _, v := range allplayer {
				if v == nil {
					continue
				}
				player1 := this.Server.GetPlayerByChaID(v.(*GuildCharacterInfo).Characterid)
				if player1 != nil {
					player1.MyGuild = nil
				}

			}
			//从内存中删除公会
			this.Guilds.Delete(guild.Id)
			//从数据库中删除公会
			db.DbOne.DeleteGuild(guild.Id)
		}
	}
	return true
}

//踢人
func (this *GuildManager) DeleteGuildPlayer(player *Player, data *protomsg.CS_DeleteGuildPlayer, targetplayer *Player) {
	myguild := player.MyGuild
	if player == nil || data == nil || myguild == nil || player == targetplayer {
		return
	}

	//回复玩家加入公会
	postdata := conf.GetGuildPostFileData(myguild.Post)
	if postdata == nil || postdata.DeletePlayerWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return
	}
	//找到当前公会
	guild1 := this.Guilds.Get(myguild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//申请的玩家角色
	character := targetplayer
	if character == nil { //离线
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(data.Characterid, &players)
		if len(players) <= 0 {
			//找不到该用户
			return
		}
		//职位没有高于被T的人。是不允许的
		if myguild.Post <= players[0].GuildPost {
			//没有权限 31
			player.SendNoticeWordToClient(31)
			return
		}
		//存档数据库
		players[0].GuildId = 0
		players[0].GuildPinLevel = int32(1)
		players[0].GuildPinExperience = int32(0)
		players[0].GuildPost = int32(1)
		db.DbOne.SaveCharacter(players[0])

	} else { //在线
		targetguild := character.MyGuild
		if targetguild != nil {
			if myguild.Post <= targetguild.Post {
				//没有权限 31
				player.SendNoticeWordToClient(31)
				return
			}

			//对方已经有公会了 32
			character.MyGuild = nil
		}
	}

	//把数据存入公会中
	guild.CharactersMap.Delete(data.Characterid)

	//存档
	this.SaveDBGuildInfo(guild)
}

//回复加入公会的申请
func (this *GuildManager) ResponseJoinGuild(player *Player, data *protomsg.CS_ResponseJoinGuildPlayer, targetplayer *Player) {
	if player == nil || data == nil || player.MyGuild == nil {
		return
	}
	//回复玩家加入公会
	postdata := conf.GetGuildPostFileData(player.MyGuild.Post)
	if postdata == nil || postdata.ResponseJoinPlayerWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return
	}
	//找到当前公会
	guild1 := this.Guilds.Get(player.MyGuild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//删除申请请求
	chainfo := guild.RequestJoinCharactersMap.Get(data.Characterid)
	if chainfo == nil {
		//找不到该玩家
		return
	}

	guild.RequestJoinCharactersMap.Delete(data.Characterid)

	//如果是拒绝就到此为止
	if data.Result != 1 {
		return
	}

	//如果超过上限就到此为止
	if guild.CharactersMap.Size() >= int(guild.MaxCount) {
		//公会成员已经达到数量上限
		return
	}

	//申请的玩家角色
	character := targetplayer
	if character == nil { //离线
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(data.Characterid, &players)
		if len(players) <= 0 {
			//找不到该用户
			return
		}
		if players[0].GuildId > 0 {
			//对方已经有公会了 32
			player.SendNoticeWordToClient(32)
			return
		}

		//存档数据库
		players[0].GuildId = player.MyGuild.GuildId
		players[0].GuildPinLevel = int32(1)
		players[0].GuildPinExperience = int32(0)
		players[0].GuildPost = int32(1)
		db.DbOne.SaveCharacter(players[0])
		//把数据存入公会中
		guildchainfo := &GuildCharacterInfo{}
		guildchainfo.Uid = players[0].Uid
		guildchainfo.Characterid = players[0].Characterid
		guildchainfo.Name = players[0].Name
		guildchainfo.Level = players[0].Level
		guildchainfo.Typeid = players[0].Typeid
		guildchainfo.GuildId = players[0].GuildId
		guildchainfo.PinLevel = players[0].GuildPinLevel
		guildchainfo.PinExperience = players[0].GuildPinExperience
		guildchainfo.Post = players[0].GuildPost

		postdata := conf.GetGuildPostFileData(players[0].GuildPost)
		if postdata != nil {
			guildchainfo.PostName = postdata.Name
		}
		pinleveldata := conf.GetGuildPinLevelFileData(players[0].GuildPinLevel)
		if pinleveldata != nil {
			guildchainfo.PinLevelName = pinleveldata.Name
			guildchainfo.PinMaxExperience = pinleveldata.UpgradeEx
		}

		guild.CharactersMap.Set(guildchainfo.Characterid, guildchainfo)

	} else { //在线
		if character.MyGuild != nil {
			//对方已经有公会了 32
			player.SendNoticeWordToClient(32)
		} else {
			//角色创建的公会信息
			character.NewAddGuildInfo(player.MyGuild.GuildId, 1)
		}
	}

	//存档
	this.SaveDBGuildInfo(guild)

	//message CS_ResponseJoinGuildPlayer{
	//    int32 Characterid = 1;
	//    int32 Result = 2; //1表示同意 其他表示不同意
	//}
}

//申请加入公会
func (this *GuildManager) RequestJoinGuild(player *Player, guildid int32) {
	if player == nil || guildid <= 0 {
		return
	}
	if player.MyGuild != nil {
		//已经有公会了 不能加入
		player.SendNoticeWordToClient(28)
		return
	}
	guild1 := this.Guilds.Get(guildid)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//角色信息
	guildchainfo := &GuildCharacterInfo{}
	guildchainfo.Uid = player.Uid
	guildchainfo.Characterid = player.Characterid
	if player.MainUnit != nil {
		guildchainfo.Name = player.MainUnit.Name
		guildchainfo.Level = player.MainUnit.Level
		guildchainfo.Typeid = player.MainUnit.TypeID
	}

	guild.RequestJoinCharactersMap.Set(player.Characterid, guildchainfo)

	player.SendNoticeWordToClient(30)

	log.Info("----------RequestJoinGuild--")
}

//增加公会经验
func (this *GuildManager) AddGuildExp(addexp int32, guildid int32) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	guild1 := this.Guilds.Get(guildid)
	if guild1 == nil {
		//不存在该公会
		return
	}
	guild := guild1.(*GuildInfo)
	if guild.Level >= conf.GuildMaxLevel {
		//已经最大等级
		return
	}

	guild.Experience += addexp
	//升级pin
	if guild.Experience >= guild.UpgradeEx {
		guild.Experience = 0
		guild.Level++

		leveldata := conf.GetGuildLevelFileData(guild.Level)
		if leveldata != nil {
			guild.GuildLevelFileData = *leveldata
		}

		//存档
		this.SaveDBGuildInfo(guild)
	}

}

//添加公会拍卖装备
func (this *GuildManager) AddAuctionItem(guildid int32, itemid int32, itemlevel int32, receivecharacters *utils.BeeMap) bool {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if receivecharacters == nil {
		return false
	}

	guild1 := this.Guilds.Get(guildid)
	if guild1 == nil {
		//不存在该公会
		return false
	}
	guild := guild1.(*GuildInfo)

	//------
	//	Id                int32 `json:"id"`
	//	Guildid           int32 `json:"guildid"` //公会ID
	//	ItemID            int32 `json:"itemid"`
	//	Level             int32 `json:"level"`
	//	PriceType         int32 `json:"pricetype"`         //价格类型 1金币 2砖石
	//	Price             int32 `json:"price"`             //价格
	//	BidderCharacterid int32 `json:"bidderCharacterid"` //竞拍者角色ID
	//	Receivecharacters int32 `json:"receivecharacters"` //参与分红者的ID
	//	Remaintime        int32 `json:"remaintime"`        //剩余时间(秒)
	auctioninfo := &db.DB_AuctionInfo{}
	auctioninfo.Guildid = guildid
	auctioninfo.ItemID = itemid
	auctioninfo.Level = itemlevel
	auctioninfo.PriceType = int32(conf.Conf.NormalInfo.AuctionPriceType)
	auctioninfo.Price = int32(conf.Conf.NormalInfo.AuctionFirstPrice)
	auctioninfo.BidderCharacterid = -1
	auctioninfo.Remaintime = int32(conf.Conf.NormalInfo.AuctionTime)

	receivecharacter := make([]int32, 0)
	chaitems := receivecharacters.Items()
	for k, _ := range chaitems {
		if guild.CharactersMap.Check(k.(int32)) == true {
			receivecharacter = append(receivecharacter, k.(int32))
		}
	}

	//加入拍卖行
	AuctionManagerObj.NewAuctionItem(auctioninfo, receivecharacter)

	guild.AuctionMap.Set(auctioninfo.Id, auctioninfo.Id)

	//存档
	this.SaveDBGuildInfo(guild)

	return true
}

//从数据库载入数据
func (this *GuildManager) LoadDataFromDB() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	commoditys := make([]db.DB_GuildInfo, 0)
	db.DbOne.GetGuilds(&commoditys)
	for _, v := range commoditys {
		//log.Info("----------ExchangeManager load %d %v", v.Id, &commoditys[k])
		guild := &GuildInfo{}
		guild.DB_GuildInfo = v

		leveldata := conf.GetGuildLevelFileData(guild.Level)
		if leveldata != nil {
			guild.GuildLevelFileData = *leveldata
		}

		guild.CharactersMap = utils.NewBeeMap()
		guild.RequestJoinCharactersMap = utils.NewBeeMap()
		guild.AuctionMap = utils.NewBeeMap()

		//解析公会成员数据
		allguildids := utils.GetInt32FromString3(v.Characters, ";")
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterids(allguildids, &players)
		for _, v1 := range players {
			guildchainfo := &GuildCharacterInfo{}
			guildchainfo.Uid = v1.Uid
			guildchainfo.Characterid = v1.Characterid
			guildchainfo.Name = v1.Name
			guildchainfo.Level = v1.Level
			guildchainfo.Typeid = v1.Typeid
			guildchainfo.GuildId = v1.GuildId
			guildchainfo.PinLevel = v1.GuildPinLevel
			guildchainfo.PinExperience = v1.GuildPinExperience
			guildchainfo.Post = v1.GuildPost

			postdata := conf.GetGuildPostFileData(v1.GuildPost)
			if postdata != nil {
				guildchainfo.PostName = postdata.Name
			}
			pinleveldata := conf.GetGuildPinLevelFileData(v1.GuildPinLevel)
			if pinleveldata != nil {
				guildchainfo.PinLevelName = pinleveldata.Name
				guildchainfo.PinMaxExperience = pinleveldata.UpgradeEx
			}

			guild.CharactersMap.Set(guildchainfo.Characterid, guildchainfo)
		}
		//解析请求加入公会的角色

		requestallguildids := utils.GetInt32FromString3(v.RequestJoinCharacters, ";")
		requestplayers := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterids(requestallguildids, &requestplayers)
		for _, v1 := range requestplayers {
			guildchainfo := &GuildCharacterInfo{}
			guildchainfo.Uid = v1.Uid
			guildchainfo.Characterid = v1.Characterid
			guildchainfo.Name = v1.Name
			guildchainfo.Level = v1.Level
			guildchainfo.Typeid = v1.Typeid

			guild.RequestJoinCharactersMap.Set(guildchainfo.Characterid, guildchainfo)
		}
		auctionmap := utils.GetInt32FromString3(v.Auction, ";")
		for _, v := range auctionmap {
			guild.AuctionMap.Set(v, v)
		}

		this.Guilds.Set(v.Id, guild)

	}

}

//进入公会地图
func (this *GuildManager) GotoGuildMap(player *Player, mapid int32) *protomsg.SC_GotoGuildMap {
	re := &protomsg.SC_GotoGuildMap{}
	if player == nil {
		return re
	}
	myguild := player.MyGuild
	if myguild == nil {
		return re
	}
	guild1 := this.Guilds.Get(myguild.GuildId)
	if guild1 == nil {
		//不存在该公会
		return re
	}
	guild := guild1.(*GuildInfo)

	mapfiledata := conf.CheckGotoGuildMap(mapid, guild.DB_GuildInfo.Level)
	if mapfiledata == nil {
		player.SendNoticeWordToClient(35)
		return re
	}
	re.Result = 1

	//进入成功
	doorway := conf.DoorWay{}

	doorway.NextX = mapfiledata.X
	doorway.NextY = mapfiledata.Y
	doorway.NextSceneID = mapfiledata.NextSceneID
	mainunit := player.MainUnit
	if mainunit != nil {
		oldscene := mainunit.InScene
		oldscene.HuiChengPlayer.Set(player, &doorway)
	}

	return re

}

//获取公会地图信息 只获取类型为1的 普通地图
func (this *GuildManager) GetGuildMapsInfo() *protomsg.SC_GetGuildMapsInfo {
	if this.MapInfo != nil {
		return this.MapInfo
	}

	data := &protomsg.SC_GetGuildMapsInfo{}
	data.Maps = make([]*protomsg.GuildMapInfo, 0)

	for _, v := range conf.GuildMapFileDatas {
		mapdata := v.(*conf.GuildMapFileData)
		if mapdata.IsOpen != 1 || mapdata.MapType != 1 {
			continue
		}
		one := &protomsg.GuildMapInfo{}

		one.ID = mapdata.ID
		one.OpenMonthDay = mapdata.OpenMonthDay
		one.OpenWeekDay = mapdata.OpenWeekDay
		one.OpenStartTime = mapdata.OpenStartTime
		one.OpenEndTime = mapdata.OpenEndTime
		one.NeedGuildLevel = mapdata.NeedGuildLevel
		one.NextSceneID = mapdata.NextSceneID

		data.Maps = append(data.Maps, one)

	}
	this.MapInfo = data
	return this.MapInfo
}

//获取公会拍卖物品
func (this *GuildManager) GetAuctionItems(player *Player) *protomsg.SC_GetAuctionItems {
	if player == nil || player.MyGuild == nil {
		return nil
	}
	guild1 := this.Guilds.Get(player.MyGuild.GuildId)
	if guild1 == nil {
		//不存在该公会
		return nil
	}
	guild := guild1.(*GuildInfo)

	data := &protomsg.SC_GetAuctionItems{}
	data.Items = make([]*protomsg.AuctionItem, 0)
	items := guild.AuctionMap.Items()
	for _, v := range items {
		if v == nil {
			continue
		}
		itemone1 := AuctionManagerObj.Commoditys.Get(v.(int32))
		if itemone1 == nil {
			continue
		}
		itemone := itemone1.(*AuctionInfo)

		d1 := &protomsg.AuctionItem{}
		d1.ID = itemone.Id
		d1.ItemID = itemone.ItemID
		d1.PriceType = itemone.PriceType
		d1.Price = itemone.Price
		d1.Level = itemone.Level
		d1.BidderCharacterName = ""
		bidderplayer := guild.CharactersMap.Get(itemone.BidderCharacterid)
		if bidderplayer != nil {
			d1.BidderCharacterName = bidderplayer.(*GuildCharacterInfo).Name
		}

		d1.RemainTime = itemone.Remaintime

		d1.ReceivecharactersName = make([]string, 0)
		for _, v1 := range itemone.ReceiveCharactersMap {

			playerone := guild.CharactersMap.Get(v1)
			if playerone == nil {
				continue
			}
			d1.ReceivecharactersName = append(d1.ReceivecharactersName, playerone.(*GuildCharacterInfo).Name)
		}

		data.Items = append(data.Items, d1)
	}

	return data

}

func (this *GuildManager) GetGuild(id int32) *GuildInfo {
	guild1 := this.Guilds.Get(id)
	if guild1 == nil {
		//不存在该公会
		return nil
	}
	guild := guild1.(*GuildInfo)
	return guild
}

//type GuildInfo struct {
//	db.DB_GuildInfo
//	CharactersMap            *utils.BeeMap //公会成员
//	RequestJoinCharactersMap *utils.BeeMap //请求加入公会角色
//}
func (this *GuildManager) SaveDBGuildInfo(guild *GuildInfo) {
	if guild == nil {
		return
	}
	chaitems := guild.CharactersMap.Items()
	guild.Characters = ""
	for _, item := range chaitems {
		guild.Characters += strconv.Itoa(int(item.(*GuildCharacterInfo).Characterid)) + ";"
	}

	joinitems := guild.RequestJoinCharactersMap.Items()
	guild.RequestJoinCharacters = ""
	for _, item := range joinitems {
		guild.RequestJoinCharacters += strconv.Itoa(int(item.(*GuildCharacterInfo).Characterid)) + ";"
	}

	auctionitems := guild.AuctionMap.Items()
	guild.Auction = ""
	for _, item := range auctionitems {
		guild.Auction += strconv.Itoa(int(item.(int32))) + ";"
	}
	//db.DbOne.DeleteGuild
	db.DbOne.SaveGuild(guild.DB_GuildInfo)
}

func (this *GuildManager) Close() {
	//存入数据库
	log.Info("GuildManager save")
	guilditems := this.Guilds.Items()
	for _, guild := range guilditems {
		this.SaveDBGuildInfo(guild.(*GuildInfo))
	}
}
