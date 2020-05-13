package gamescene1

import (
	"dq/conf"
	"dq/datamsg"

	//"dq/db"
	"dq/log"
	"dq/network"
	"net"
	"time"

	"dq/db"
	"dq/utils"

	//"dq/cyward"
	"dq/gamecore"
	"dq/protobuf"

	//"dq/timer"
	//"dq/vec2d"
	"dq/wordsfilter"
	"io"

	//"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/golang/protobuf/proto"
)

//游戏部分
type GameScene1Agent struct {
	conn network.Conn

	handles map[string]func(data *protomsg.MsgBase)

	ServerName string
	Scenes     *utils.BeeMap
	Players    *utils.BeeMap
	Characters *utils.BeeMap

	wgScene sync.WaitGroup

	IsClose bool
}

func (a *GameScene1Agent) GetConnectId() int32 {

	return 0
}
func (a *GameScene1Agent) GetModeType() string {
	return ""
}

func (a *GameScene1Agent) ceshi() {
	fenhong := utils.NewBeeMap()
	fenhong.Set(int32(32), int32(32))
	fenhong.Set(int32(35), int32(35))
	fenhong.Set(int32(36), int32(36))
	fenhong.Set(int32(40), int32(40))
	fenhong.Set(int32(41), int32(41))
	gamecore.GuildManagerObj.AddAuctionItem(7, 63, 1, fenhong)
}

func (a *GameScene1Agent) Init() {

	a.IsClose = false

	//初始化 组队信息
	gamecore.TeamManagerObj.Init(a)
	//初始化 交易所信息
	gamecore.ExchangeManagerObj.Init(a)
	//初始化公会拍卖信息
	gamecore.AuctionManagerObj.Init(a)
	//初始化 公会信息
	gamecore.GuildManagerObj.Init(a)
	//初始化数据管理器
	gamecore.GameCoreDataManagerObj.Init()
	//初始化副本系统
	gamecore.CopyMapMgrObj.Init(a)

	//-------测试--------
	//a.ceshi()

	//----------------

	a.ServerName = datamsg.GameScene1

	a.Scenes = utils.NewBeeMap()
	a.Players = utils.NewBeeMap()
	a.Characters = utils.NewBeeMap()

	a.handles = make(map[string]func(data *protomsg.MsgBase))
	a.handles["MsgUserEnterScene"] = a.DoMsgUserEnterScene
	a.handles["Disconnect"] = a.DoDisconnect

	a.handles["CS_PlayerMove"] = a.DoPlayerMove

	a.handles["CS_PlayerAttack"] = a.DoPlayerAttack
	a.handles["CS_PlayerSkill"] = a.DoPlayerSkill

	a.handles["CS_GetUnitInfo"] = a.DoGetUnitInfo
	a.handles["CS_GetCharacterSimpleInfo"] = a.DoGetCharacterSimpleInfo

	a.handles["CS_GetItemExtraInfo"] = a.DoGetItemExtraInfo
	a.handles["CS_GetBagInfo"] = a.DoGetBagInfo
	a.handles["CS_ChangeItemPos"] = a.DoChangeItemPos
	a.handles["CS_DestroyItem"] = a.DoDestroyItem
	a.handles["CS_SystemHuiShouItem"] = a.DoSystemHuiShouItem

	a.handles["CS_PlayerUpgradeSkill"] = a.DoPlayerUpgradeSkill
	a.handles["CS_ChangeAttackMode"] = a.DoChangeAttackMode

	a.handles["CS_LodingScene"] = a.DoLodingScene
	a.handles["CS_UseAI"] = a.DoUseAI

	a.handles["CS_LookVedioSucc"] = a.DoLookVedioSucc
	//队伍
	a.handles["CS_OrganizeTeam"] = a.DoOrganizeTeam
	a.handles["CS_ResponseOrgTeam"] = a.DoResponseOrgTeam
	a.handles["CS_OutTeam"] = a.DoOutTeam

	//商店
	a.handles["CS_GetStoreData"] = a.DoGetStoreData
	a.handles["CS_BuyCommodity"] = a.DoBuyCommodity
	//立即复活
	a.handles["CS_QuickRevive"] = a.DoQuickRevive

	//聊天信息
	a.handles["CS_ChatInfo"] = a.DoChatInfo

	//好友相关
	a.handles["CS_AddFriendRequest"] = a.DoAddFriendRequest
	a.handles["CS_RemoveFriend"] = a.DoRemoveFriend
	a.handles["CS_AddFriendResponse"] = a.DoAddFriendResponse
	a.handles["CS_GetFriendsList"] = a.DoGetFriendsList

	//邮件相关
	a.handles["CS_GetMailsList"] = a.DoGetMailsList
	a.handles["CS_GetMailInfo"] = a.DoGetMailInfo
	a.handles["CS_GetMailRewards"] = a.DoGetMailRewards
	a.handles["CS_DeleteNoRewardMails"] = a.DoDeleteNoRewardMails

	//交易所相关
	a.handles["CS_GetExchangeShortCommoditys"] = a.DoGetExchangeShortCommoditys
	a.handles["CS_GetExchangeDetailedCommoditys"] = a.DoGetExchangeDetailedCommoditys
	a.handles["CS_BuyExchangeCommodity"] = a.DoBuyExchangeCommodity
	a.handles["CS_ShelfExchangeCommodity"] = a.DoShelfExchangeCommodity
	a.handles["CS_GetSellUIInfo"] = a.DoGetSellUIInfo
	a.handles["CS_UnShelfExchangeCommodity"] = a.DoUnShelfExchangeCommodity
	a.handles["CS_GetWorldAuctionItems"] = a.DoGetWorldAuctionItems
	a.handles["CS_NewPriceWorldAuctionItem"] = a.DoNewPriceWorldAuctionItem

	//公会相关
	a.handles["CS_GetAllGuildsInfo"] = a.DoGetAllGuildsInfo
	a.handles["CS_CreateGuild"] = a.DoCreateGuild
	a.handles["CS_JoinGuild"] = a.DoJoinGuild
	a.handles["CS_GetGuildInfo"] = a.DoGetGuildInfo
	a.handles["CS_GetJoinGuildPlayer"] = a.DoGetJoinGuildPlayer
	a.handles["CS_ResponseJoinGuildPlayer"] = a.DoResponseJoinGuildPlayer
	a.handles["CS_DeleteGuildPlayer"] = a.DoDeleteGuildPlayer
	a.handles["CS_GetAuctionItems"] = a.DoGetAuctionItems
	a.handles["CS_NewPriceAuctionItem"] = a.DoNewPriceAuctionItem
	a.handles["CS_GuildOperate"] = a.DoGuildOperate
	a.handles["CS_GetGuildMapsInfo"] = a.DoGetGuildMapsInfo
	a.handles["CS_GotoGuildMap"] = a.DoGotoGuildMap
	a.handles["CS_EditorGuildNotice"] = a.DoEditorGuildNotice
	a.handles["CS_ChangePost"] = a.DoChangePost
	a.handles["CS_GetGuildRankInfo"] = a.DoGetGuildRankInfo
	a.handles["CS_GetGuildRankBattleInfo"] = a.DoGetGuildRankBattleInfo

	//活动地图
	a.handles["CS_GetActivityMapsInfo"] = a.DoGetActivityMapsInfo
	a.handles["CS_GetMapInfo"] = a.DoGetMapInfo
	a.handles["CS_GotoActivityMap"] = a.DoGotoActivityMap
	a.handles["CS_GetDuoBaoInfo"] = a.DoGetDuoBaoInfo

	//副本
	a.handles["CS_GetCopyMapsInfo"] = a.DoGetCopyMapsInfo
	a.handles["CS_CopyMapPiPei"] = a.DoCopyMapPiPei
	a.handles["CS_CopyMapCancel"] = a.DoCopyMapCancel

	//创建场景
	allscene := conf.GetAllScene()
	for _, v := range allscene {
		log.Info("scene:%d  %s", v.(*conf.SceneFileData).TypeID, v.(*conf.SceneFileData).ScenePath)
		if v.(*conf.SceneFileData).IsOpen != 1 || v.(*conf.SceneFileData).InitOpen != 1 {
			continue
		}
		//log.Info("scene succ:%d ", v.(*conf.SceneFileData).TypeID)
		a.CreateScene(v.(*conf.SceneFileData), -1)
		time.Sleep(time.Duration(33/len(allscene)) * time.Millisecond)
	}

	//自己的更新
	a.wgScene.Add(1)
	go func() {
		a.Update()
		a.wgScene.Done()
	}()

	a.ShowData2Http()
}

func (a *GameScene1Agent) TestGoScene(sceneid int32, player *gamecore.Player) {
	if player == nil {
		return
	}
	//玩家进入地图
	//进入新地图
	doorway := conf.DoorWay{}
	doorway.NextX = 15
	doorway.NextY = 15
	doorway.NextSceneID = sceneid
	mainunit := player.MainUnit
	if mainunit != nil {
		oldscene := mainunit.InScene
		if oldscene != nil {
			oldscene.HuiChengPlayer.Set(player, &doorway)
		}

	}
}

func (a *GameScene1Agent) PiPeiFuBen(players []*gamecore.CopyMapPlayer, cmfid int32) {

	var newid = gamecore.GetCopyMapSceneID() //获取唯一ID

	fubenscenetypeid := cmfid
	scenefile := conf.GetSceneFileData(fubenscenetypeid)
	if scenefile == nil {
		//如果场景文件不存在
		return
	}
	newscene := a.CreateScene(scenefile, newid)
	if newscene == nil {
		return
	}

	//玩家进入地图
	//进入新地图
	doorway := conf.DoorWay{}
	doorway.NextX = 15
	doorway.NextY = 15
	doorway.NextSceneID = newid

	for _, v := range players {
		player := v.PlayerInfo
		mainunit := player.MainUnit
		if mainunit != nil {
			oldscene := mainunit.InScene
			if oldscene != nil {
				oldscene.HuiChengPlayer.Set(player, &doorway)
			}
		}
		//把小组ID都设为1
		newscene.SetSceneCharacterGroups(player.Characterid, 1)
	}

}

//创建场景 typeid为设置给场景的唯一ID
func (a *GameScene1Agent) CreateScene(scenefile *conf.SceneFileData, typeid int32) *gamecore.Scene {
	if scenefile == nil {
		return nil
	}

	if typeid <= 0 {
		typeid = scenefile.TypeID
	}

	scene := gamecore.CreateScene(scenefile, a, a)
	scene.TypeID = typeid
	a.Scenes.Set(typeid, scene)
	a.wgScene.Add(1)
	go func() {
		scene.Update()
		a.Scenes.Delete(typeid)
		a.wgScene.Done()
	}()

	return scene
}

//查看数据
func (a *GameScene1Agent) ShowData2Http() {
	//http://127.0.0.1:9999/sd
	//http://119.23.8.72:9999/sd
	httpserver := &http.Server{Addr: ":9999", Handler: nil}

	http.HandleFunc("/sd", func(w http.ResponseWriter, r *http.Request) {

		playercount := a.Players.Size()
		io.WriteString(w, "playercount:"+strconv.Itoa(int(playercount)))
	})

	go httpserver.ListenAndServe()
}

//自己的更新
func (a *GameScene1Agent) Update() {
	for {
		a.CheckSceneCloseAndOpen()
		time.Sleep(time.Duration(time.Second * 2))
		if a.IsClose {
			break
		}
	}
}

//检测地图开启与关闭
func (a *GameScene1Agent) CheckSceneCloseAndOpen() {

	scenemap := make(map[*gamecore.Scene]bool)

	//活动地图
	for _, v := range conf.ActivityMapFileDatas {
		if v == nil {
			continue
		}
		//如果场景地图不存在
		scenefile := conf.GetSceneFileData(v.(*conf.ActivityMapFileData).NextSceneID)
		if scenefile == nil {
			continue
		}

		//id 和 等级
		mapdata := conf.CheckActivitySceneStart_End(v.(*conf.ActivityMapFileData).ID)

		if mapdata == true { //如果可以进入地图  就开启地图
			//onescnee.(*gamecore.Scene).SetCleanPlayer(false)
			onescnee := a.Scenes.Get(v.(*conf.ActivityMapFileData).NextSceneID)
			if onescnee == nil { //
				onescnee = a.CreateScene(scenefile, -1)

			}
			if onescnee != nil {
				scenemap[onescnee.(*gamecore.Scene)] = false
			}

		} else { //如果不可以进入就关闭地图
			//onescnee.(*gamecore.Scene).SetCleanPlayer(true)
			onescnee := a.Scenes.Get(v.(*conf.ActivityMapFileData).NextSceneID)
			if onescnee == nil { //
				continue
			}
			if _, ok := scenemap[onescnee.(*gamecore.Scene)]; ok == false {
				scenemap[onescnee.(*gamecore.Scene)] = true
			}
		}

	}
	//公会地图
	for _, v := range conf.GuildMapFileDatas {
		if v == nil {
			continue
		}
		//如果场景地图不存在
		scenefile := conf.GetSceneFileData(v.(*conf.GuildMapFileData).NextSceneID)
		if scenefile == nil {
			continue
		}
		//id 和 等级
		mapdata := conf.CheckGuildSceneStart_End(v.(*conf.GuildMapFileData).ID)

		if mapdata == true { //如果可以进入地图  就开启地图
			//onescnee.(*gamecore.Scene).SetCleanPlayer(false)
			onescnee := a.Scenes.Get(v.(*conf.GuildMapFileData).NextSceneID)
			if onescnee == nil { //
				onescnee = a.CreateScene(scenefile, -1)

			}
			if onescnee != nil {
				scenemap[onescnee.(*gamecore.Scene)] = false
			}

		} else { //如果不可以进入就关闭地图
			//onescnee.(*gamecore.Scene).SetCleanPlayer(true)
			onescnee := a.Scenes.Get(v.(*conf.GuildMapFileData).NextSceneID)
			if onescnee == nil { //
				continue
			}
			if _, ok := scenemap[onescnee.(*gamecore.Scene)]; ok == false {
				scenemap[onescnee.(*gamecore.Scene)] = true
			}
		}

	}

	for k, v := range scenemap {
		k.SetCleanPlayer(v)
	}
}

//
func (a *GameScene1Agent) DoDisconnect(data *protomsg.MsgBase) {

	log.Info("GameScene1Agent---------DoDisconnect")

	player := a.Players.Get(data.Uid)
	if player != nil {
		//退出之前的场景
		if player.(*gamecore.Player).ConnectId == data.ConnectId {

			log.Info("---------DoDisconnect--delete")

			gamecore.TeamManagerObj.LeaveTeam(player.(*gamecore.Player))
			curscene := player.(*gamecore.Player).CurScene
			if curscene != nil {
				curscene.PlayerWillLeave(player.(*gamecore.Player)) //离开前的处理
			}

			player.(*gamecore.Player).SaveDB()

			player.(*gamecore.Player).OutScene()
			a.Players.Delete(data.Uid)
			a.Characters.Delete(player.(*gamecore.Player).Characterid)
			//存档 数据库

		} else {
			log.Info("---------DoDisconnect--ConnectId fail")
		}

	}

	//LoginOut
	t1 := protomsg.MsgBase{
		ModeType:  datamsg.LoginMode,
		MsgType:   "LoginOut",
		Uid:       data.Uid,
		ConnectId: data.ConnectId,
	}
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, nil))

}
func (a *GameScene1Agent) PlayerChangeScene(player *gamecore.Player, doorway conf.DoorWay) {
	if player == nil {
		return
	}
	dbdata := player.GetDBData()
	if dbdata == nil {
		return
	}
	dbdata.X = doorway.NextX
	dbdata.Y = doorway.NextY
	//h2 := &protomsg.MsgUserEnterScene{}

	h2 := &protomsg.MsgUserEnterScene{
		Uid:            player.Uid,
		ConnectId:      player.ConnectId,
		SrcServerName:  "",
		DestServerName: datamsg.GameScene1, //
		SceneID:        doorway.NextSceneID,
		Datas:          utils.Struct2Bytes(dbdata), //数据库中的角色信息
	}
	a.DoUserEnterScene(h2)
}
func (a *GameScene1Agent) GetPlayerByID(uid int32) *gamecore.Player {
	player := a.Players.Get(uid)
	if player == nil {
		return nil
	}
	return player.(*gamecore.Player)
}
func (a *GameScene1Agent) GetPlayerByChaID(uid int32) *gamecore.Player {
	player := a.Characters.Get(uid)
	if player == nil {
		return nil
	}
	return player.(*gamecore.Player)
}

func (a *GameScene1Agent) DoHuiChengData(old []byte) []byte {
	//改变坐标
	characterinfo := db.DB_CharacterInfo{}
	utils.Bytes2Struct(old, &characterinfo)
	characterinfo.X = float32(utils.RandInt64(70, 80))
	characterinfo.Y = float32(utils.RandInt64(70, 80))
	return utils.Struct2Bytes(characterinfo)
}

func (a *GameScene1Agent) DoUserEnterScene(h2 *protomsg.MsgUserEnterScene) {
	if h2 == nil {
		return
	}

	//	characterinfo := db.DB_CharacterInfo{}
	//	utils.Bytes2Struct(h2.Datas, &characterinfo)
	//	log.Info("---------datas:%v", characterinfo)

	//如果目的地服务器是本服务器
	if h2.DestServerName == a.ServerName {

		scene := a.Scenes.Get(h2.SceneID)
		log.Info("enter scene :%d", h2.SceneID)
		if scene == nil || scene.(*gamecore.Scene).Quit == true {
			log.Info("no scene :%d", h2.SceneID)
			//自动回城
			h2.SceneID = conf.HePingShiJieID //回到安全区
			scene = a.Scenes.Get(h2.SceneID)
			//改变坐标
			h2.Datas = a.DoHuiChengData(h2.Datas)
		}

		player := a.Players.Get(h2.Uid)
		if player == nil {
			player = gamecore.CreatePlayer(h2.Uid, h2.ConnectId, -1)
			player.(*gamecore.Player).ServerAgent = a
			a.Players.Set(player.(*gamecore.Player).Uid, player)

		} else {
			//			//重新连接
			//			if player.(*gamecore.Player).ConnectId != h2.ConnectId {
			//				player.(*gamecore.Player).ConnectId = h2.ConnectId
			//				player.(*gamecore.Player).ClearShowData()
			//			}

		}

		//退出之前的场景
		player.(*gamecore.Player).OutScene()

		//进入新场景
		//hepingshijie := a.Scenes.Get(conf.HePingShiJieID)
		if player.(*gamecore.Player).GoInScene(scene.(*gamecore.Scene), h2.Datas) == false {
			h2.SceneID = conf.HePingShiJieID //回到安全区
			scene = a.Scenes.Get(h2.SceneID)
			//改变坐标
			h2.Datas = a.DoHuiChengData(h2.Datas)
			player.(*gamecore.Player).GoInScene(scene.(*gamecore.Scene), h2.Datas)
			log.Info("enter scene faild :%d", h2.SceneID)
		}
		//重新设置坐标 InitPosition

		a.Characters.Set(player.(*gamecore.Player).Characterid, player)

		//发送场景信息给玩家
		msg := &protomsg.SC_NewScene{}
		msg.Name = scene.(*gamecore.Scene).SceneName
		msg.LogicFps = int32(scene.(*gamecore.Scene).SceneFrame)
		msg.CurFrame = scene.(*gamecore.Scene).CurFrame
		msg.ServerName = a.ServerName
		//msg.SceneID = scene.(*gamecore.Scene).TypeID
		msg.SceneID = scene.(*gamecore.Scene).DataFileID
		msg.TimeHour = int32(time.Now().Hour())
		msg.TimeMinute = int32(time.Now().Minute())
		msg.TimeSecond = int32(time.Now().Second())
		player.(*gamecore.Player).SendMsgToClient("SC_NewScene", msg)

		log.Info("SendMsgToClient SC_NewScene")

	}

}

func (a *GameScene1Agent) DoMsgUserEnterScene(data *protomsg.MsgBase) {

	log.Info("---------DoMsgUserEnterScene:playercount:%d", a.Players.Size())
	h2 := &protomsg.MsgUserEnterScene{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	a.DoUserEnterScene(h2)

}

//切换攻击模式
func (a *GameScene1Agent) DoChangeAttackMode(data *protomsg.MsgBase) {

	log.Info("---------DoChangeAttackMode")
	h2 := &protomsg.CS_ChangeAttackMode{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).ChangeAttackMode(h2)

}

//升级技能
//DoPlayerUpgradeSkill
func (a *GameScene1Agent) DoPlayerUpgradeSkill(data *protomsg.MsgBase) {

	log.Info("---------DoPlayerUpgradeSkill")
	h2 := &protomsg.CS_PlayerUpgradeSkill{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).UpgradeSkill(h2)

}

//a.handles["CS_SystemHuiShouItem"] = a.DoSystemHuiShouItem
func (a *GameScene1Agent) DoSystemHuiShouItem(data *protomsg.MsgBase) {

	log.Info("---------CS_SystemHuiShouItem")
	h2 := &protomsg.CS_SystemHuiShouItem{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).SystemHuiShouItem(h2)

	a.SendBagInfo(player.(*gamecore.Player))

}

func (a *GameScene1Agent) DoDestroyItem(data *protomsg.MsgBase) {

	log.Info("---------CS_DestroyItem")
	h2 := &protomsg.CS_DestroyItem{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).DestroyItem(h2)

	a.SendBagInfo(player.(*gamecore.Player))

}

//DoChangeItemPos
func (a *GameScene1Agent) DoChangeItemPos(data *protomsg.MsgBase) {

	log.Info("---------DoChangeItemPos")
	h2 := &protomsg.CS_ChangeItemPos{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).ChangeItemPos(h2)

	a.SendUnitInfo(player.(*gamecore.Player).MainUnit, player.(*gamecore.Player))
	a.SendBagInfo(player.(*gamecore.Player))

}

func (a *GameScene1Agent) SendBagInfo(player *gamecore.Player) {
	if player == nil {
		return
	}
	msg := &protomsg.SC_BagInfo{}
	msg.Equips = make([]*protomsg.UnitEquip, 0)
	for _, v := range player.BagInfo {
		if v != nil {
			equip := &protomsg.UnitEquip{}
			equip.Pos = v.Index
			equip.TypdID = v.TypeID
			equip.Level = v.Level
			item := conf.GetItemData(v.TypeID)
			if item != nil {
				equip.PriceType = item.PriceType
				equip.Price = item.Price
			}

			msg.Equips = append(msg.Equips, equip)
		}
	}

	player.SendMsgToClient("SC_BagInfo", msg)
}

//a.handles["CS_GetItemExtraInfo"] = a.DoGetItemExtraInfo
func (a *GameScene1Agent) DoGetItemExtraInfo(data *protomsg.MsgBase) {

	log.Info("---------DoGetItemExtraInfo")
	h2 := &protomsg.CS_GetItemExtraInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%d", h2.TypeId)
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	itemdata := conf.GetItemData(h2.TypeId)
	if itemdata == nil {
		return
	}
	msg := &protomsg.SC_GetItemExtraInfo{}
	msg.TypeId = h2.TypeId
	msg.Exception = itemdata.Exception
	msg.ExceptionParam = itemdata.ExceptionParam
	player.(*gamecore.Player).SendMsgToClient("SC_GetItemExtraInfo", msg)
}

//DoGetBagInfo
func (a *GameScene1Agent) DoGetBagInfo(data *protomsg.MsgBase) {

	log.Info("---------DoGetBagInfo")
	h2 := &protomsg.CS_GetBagInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%d", h2.UnitID)
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	a.SendBagInfo(player.(*gamecore.Player))

}

func (a *GameScene1Agent) SendUnitInfo(unit *gamecore.Unit, player *gamecore.Player) {
	unitdata := &protomsg.UnitBoardDatas{}
	unitdata.ID = unit.ID
	unitdata.Name = unit.Name
	unitdata.AttributeStrength = unit.AttributeStrength
	unitdata.AttributeAgility = unit.AttributeAgility
	unitdata.AttributeIntelligence = unit.AttributeIntelligence
	unitdata.Attack = unit.Attack
	unitdata.AttackSpeed = unit.AttackSpeed
	unitdata.AttackRange = unit.AttackRange
	unitdata.MoveSpeed = float32(unit.MoveSpeed)
	unitdata.MagicScale = unit.MagicScale
	unitdata.MPRegain = unit.MPRegain
	unitdata.PhysicalAmaor = unit.PhysicalAmaor
	unitdata.PhysicalResist = unit.PhysicalResist
	unitdata.MagicAmaor = unit.MagicAmaor
	unitdata.StatusAmaor = unit.StatusAmaor
	unitdata.Dodge = unit.Dodge
	unitdata.HPRegain = unit.HPRegain
	unitdata.AttributePrimary = int32(unit.AttributePrimary)
	unitdata.DropItems = unit.NPCItemDropInfo
	unitdata.RemainExperience = unit.RemainExperience
	//道具栏
	unitdata.Equips = make([]*protomsg.UnitEquip, 0)
	for k, v := range unit.Items {
		equip := &protomsg.UnitEquip{}
		equip.Pos = int32(k)
		if v != nil {
			equip.TypdID = v.TypeID
			equip.Level = v.Level
		} else {
			equip.TypdID = 0
		}
		unitdata.Equips = append(unitdata.Equips, equip)
	}

	msg := &protomsg.SC_UnitInfo{}
	msg.UnitData = unitdata
	player.SendMsgToClient("SC_UnitInfo", msg)
}

//a.handles["CS_GetCharacterSimpleInfo"] = a.DoGetCharacterSimpleInfo
func (a *GameScene1Agent) DoGetCharacterSimpleInfo(data *protomsg.MsgBase) {

	log.Info("---------CS_GetCharacterSimpleInfo")
	h2 := &protomsg.CS_GetCharacterSimpleInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	chadata := &db.DB_CharacterInfo{}

	character := a.Characters.Get(h2.CharacterID)
	if character == nil {
		//从数据库中读取数据
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(h2.CharacterID, &players)
		if len(players) <= 0 {
			//找不到该用户
			return
		}
		chadata = &players[0]
	} else {
		chadata = character.(*gamecore.Player).GetDBData()
	}
	if chadata == nil {
		return
	}

	//读取配置文件
	confdata := conf.GetUnitFileData(chadata.Typeid)
	if confdata == nil {
		return
	}

	msg := &protomsg.SC_GetCharacterSimpleInfo{}
	msg.CharacterID = chadata.Characterid
	msg.Name = chadata.Name
	msg.Level = chadata.Level
	msg.ModeType = confdata.ModeType
	msg.EquipItems = make([]string, gamecore.UnitEquitCount)
	msg.EquipItems[0] = chadata.Item1
	msg.EquipItems[1] = chadata.Item2
	msg.EquipItems[2] = chadata.Item3
	msg.EquipItems[3] = chadata.Item4
	msg.EquipItems[4] = chadata.Item5
	msg.EquipItems[5] = chadata.Item6
	msg.Skills = chadata.Skill
	msg.LastLoginDate = chadata.GetExperienceDay
	player.(*gamecore.Player).SendMsgToClient("SC_GetCharacterSimpleInfo", msg)

}

//DoGetUnitInfo
func (a *GameScene1Agent) DoGetUnitInfo(data *protomsg.MsgBase) {

	log.Info("---------DoGetUnitInfo")
	h2 := &protomsg.CS_GetUnitInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%d", h2.UnitID)
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	curscene := player.(*gamecore.Player).CurScene
	if curscene == nil {
		return
	}
	unit := curscene.FindUnitByID(h2.UnitID)
	if unit == nil {
		return
	}
	a.SendUnitInfo(unit, player.(*gamecore.Player))
}

//DoPlayerSkill
func (a *GameScene1Agent) DoPlayerSkill(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerSkill")
	h2 := &protomsg.CS_PlayerSkill{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//log.Info("---------%v  %f  %f", h2, h2.X, h2.Y)

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).SkillCmd(h2)

}

func (a *GameScene1Agent) DoPlayerAttack(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerAttack")
	h2 := &protomsg.CS_PlayerAttack{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//log.Info("---------%v", h2)

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).AttackCmd(h2)

}

func (a *GameScene1Agent) DoOrganizeTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_OrganizeTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player1 := a.Players.Get(h2.Player1)
	player2 := a.Players.Get(h2.Player2)
	if player1 == nil || player2 == nil {
		return
	}
	if h2.Player1 == h2.Player2 {
		return
	}

	gamecore.TeamManagerObj.OrganizeTeam(player1.(*gamecore.Player), player2.(*gamecore.Player))
}
func (a *GameScene1Agent) DoResponseOrgTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ResponseOrgTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.TeamManagerObj.ResponseOrgTeam(h2, player.(*gamecore.Player))
}

//商店
func (a *GameScene1Agent) DoGetStoreData(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetStoreData{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).SendMsgToClient("SC_StoreData", conf.GetStoreData2SC_StoreData())

}
func (a *GameScene1Agent) DoBuyCommodity(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_BuyCommodity{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	//商品信息
	cominfo := conf.GetStoreFileData(h2.TypeID)
	player.(*gamecore.Player).BuyItem(cominfo)
}

//好友相关 请求把目标加为好友
func (a *GameScene1Agent) DoAddFriendRequest(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_AddFriendRequest{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	friend := a.Players.Get(h2.Uid)
	if friend == nil {
		player.(*gamecore.Player).MyFriends.AddFriendRequest(h2, nil)
	} else {
		player.(*gamecore.Player).MyFriends.AddFriendRequest(h2, friend.(*gamecore.Player))
	}

}
func (a *GameScene1Agent) DoRemoveFriend(data *protomsg.MsgBase) {
	//	h2 := &protomsg.CS_RemoveFriend{}
	//	err := proto.Unmarshal(data.Datas, h2)
	//	if err != nil {
	//		log.Info(err.Error())
	//		return
	//	}
	//	player := a.Players.Get(data.Uid)
	//	if player == nil {
	//		return
	//	}

}
func (a *GameScene1Agent) CheckOnline(uid int32, characterid int32) bool {
	p1 := a.Players.Get(uid)
	if p1 != nil && p1.(*gamecore.Player).Characterid == characterid {
		return true
	}

	return false
}

//回复好友请求
func (a *GameScene1Agent) DoAddFriendResponse(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_AddFriendResponse{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	friend := a.Players.Get(h2.FriendInfo.Uid)
	if friend == nil {
		player.(*gamecore.Player).MyFriends.AddFriendResponse(h2, nil)
	} else {
		player.(*gamecore.Player).MyFriends.AddFriendResponse(h2, friend.(*gamecore.Player))
	}

	//重新发送好友信息
	d1 := player.(*gamecore.Player).MyFriends.GetSCData()
	for k, v := range d1.Friends {
		if a.CheckOnline(v.Uid, v.Characterid) == true {
			d1.Friends[k].State = 1
		} else {
			d1.Friends[k].State = 2
		}

	}

	player.(*gamecore.Player).SendMsgToClient("SC_GetFriendsList", d1)

}
func (a *GameScene1Agent) DoGetFriendsList(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetFriendsList{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	//*protomsg.SC_GetFriendsList
	d1 := player.(*gamecore.Player).MyFriends.GetSCData()
	for k, v := range d1.Friends {
		if a.CheckOnline(v.Uid, v.Characterid) == true {
			d1.Friends[k].State = 1
		} else {
			d1.Friends[k].State = 2
		}

	}
	player.(*gamecore.Player).SendMsgToClient("SC_GetFriendsList", d1)
}

//副本
//a.handles["CS_GetCopyMapsInfo"] = a.DoGetCopyMapsInfo
//a.handles["CS_CopyMapPiPei"] = a.DoCopyMapPiPei
//	a.handles["CS_CopyMapCancel"] = a.DoCopyMapCancel
func (a *GameScene1Agent) DoCopyMapPiPei(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_CopyMapPiPei{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	gamecore.CopyMapMgrObj.JionPiPei(player.(*gamecore.Player), h2.CopyMapID)

	msg := gamecore.CopyMapMgrObj.GetCopyMapsInfo(player.(*gamecore.Player))
	player.(*gamecore.Player).SendMsgToClient("SC_GetCopyMapsInfo", msg)

}
func (a *GameScene1Agent) DoCopyMapCancel(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_CopyMapCancel{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.CopyMapMgrObj.CancelPiPei(player.(*gamecore.Player))

	msg := gamecore.CopyMapMgrObj.GetCopyMapsInfo(player.(*gamecore.Player))
	player.(*gamecore.Player).SendMsgToClient("SC_GetCopyMapsInfo", msg)

}
func (a *GameScene1Agent) DoGetCopyMapsInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetCopyMapsInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	msg := gamecore.CopyMapMgrObj.GetCopyMapsInfo(player.(*gamecore.Player))
	player.(*gamecore.Player).SendMsgToClient("SC_GetCopyMapsInfo", msg)

}

//活动地图
//	a.handles["CS_GetActivityMapsInfo"] = a.DoGetActivityMapsInfo
//	a.handles["CS_GetMapInfo"] = a.DoGetMapInfo
//a.handles["CS_GotoActivityMap"] = a.DoGotoActivityMap
//a.handles["CS_GetDuoBaoInfo"] = a.DoGetDuoBaoInfo
func (a *GameScene1Agent) DoGetDuoBaoInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetDuoBaoInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	//场景信息 夺宝奇兵
	mapdata := conf.GetActivityMapFileData(conf.ActivityDuoBaoMapID) //
	if mapdata == nil {
		return
	}
	scenedata := gamecore.GameCoreDataManagerObj.SceneDropDatas.Get(mapdata.NextSceneID)
	if scenedata == nil {
		return
	}
	scenefile := conf.GetSceneFileData(mapdata.NextSceneID)
	if scenefile == nil {
		return
	}

	msg := &protomsg.SC_GetDuoBaoInfo{}
	msg.MapGoInInfo = conf.GetProtoMsgActivityMapsInfo(mapdata)
	//地图信息 掉落信息
	msg.MapInfo = &protomsg.SC_GetMapInfo{}
	msg.MapInfo.SceneID = mapdata.NextSceneID
	msg.MapInfo.BossFreshTime = scenedata.(*gamecore.SceneDropData).BossFreshTime
	msg.MapInfo.DropItems = make([]int32, 0)
	items := scenedata.(*gamecore.SceneDropData).DropItems.Items()
	for _, v := range items {
		msg.MapInfo.DropItems = append(msg.MapInfo.DropItems, v.(int32))
	}
	msg.Minute = 5
	params := utils.GetInt32FromString3(scenefile.ExceptionParam, ",")
	if len(params) >= 1 {
		msg.Minute = params[0] / 60
	}
	player.(*gamecore.Player).SendMsgToClient("SC_GetDuoBaoInfo", msg)

	//测试场景
	//a.PiPeiFuBen(player.(*gamecore.Player))
	//a.TestGoScene(3008, player.(*gamecore.Player))

}
func (a *GameScene1Agent) DoGotoActivityMap(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GotoActivityMap{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	level := player.(*gamecore.Player).GetLevel()
	mapdata := conf.CheckGotoActivityMap(h2.ID, level)
	if mapdata == nil {
		player.(*gamecore.Player).SendNoticeWordToClient(37)
		return
	}
	if player.(*gamecore.Player).BuyItemSubMoneyLock(mapdata.PriceType, mapdata.Price) == false {
		player.(*gamecore.Player).SendNoticeWordToClient(mapdata.PriceType)
		return
	}
	//进入新地图
	doorway := conf.DoorWay{}
	doorway.NextX = mapdata.X
	doorway.NextY = mapdata.Y
	doorway.NextSceneID = mapdata.NextSceneID
	mainunit := player.(*gamecore.Player).MainUnit
	if mainunit != nil {
		oldscene := mainunit.InScene
		oldscene.HuiChengPlayer.Set(player, &doorway)
	}
	//进入成功
	re := &protomsg.SC_GotoActivityMap{}
	re.Result = 1
	player.(*gamecore.Player).SendMsgToClient("SC_GotoActivityMap", re)

}
func (a *GameScene1Agent) DoGetActivityMapsInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetActivityMapsInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	//msg := conf.SC_GetActivityMapsInfoMsg
	player.(*gamecore.Player).SendMsgToClient("SC_GetActivityMapsInfo", conf.SC_GetActivityMapsInfoMsg)

}
func (a *GameScene1Agent) DoGetMapInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMapInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	//	scene := a.Scenes.Get(h2.SceneID)
	//	if scene == nil {
	//		return
	//	}
	scenedata := gamecore.GameCoreDataManagerObj.SceneDropDatas.Get(h2.SceneID)
	if scenedata == nil {
		return
	}

	msg := &protomsg.SC_GetMapInfo{}
	msg.SceneID = h2.SceneID
	msg.BossFreshTime = scenedata.(*gamecore.SceneDropData).BossFreshTime
	msg.DropItems = make([]int32, 0)
	items := scenedata.(*gamecore.SceneDropData).DropItems.Items()
	for _, v := range items {
		msg.DropItems = append(msg.DropItems, v.(int32))
	}
	player.(*gamecore.Player).SendMsgToClient("SC_GetMapInfo", msg)
}

//公会相关
//	a.handles["CS_GetAllGuildsInfo"] = a.DoGetAllGuildsInfo
//	a.handles["CS_CreateGuild"] = a.DoCreateGuild
//	a.handles["CS_JoinGuild"] = a.DoJoinGuild
//	a.handles["CS_GetGuildInfo"] = a.DoGetGuildInfo
//a.handles["CS_GetJoinGuildPlayer"] = a.DoGetJoinGuildPlayer
//a.handles["CS_ResponseJoinGuildPlayer"] = a.DoResponseJoinGuildPlayer
//	a.handles["CS_DeleteGuildPlayer"] = a.DoDeleteGuildPlayer
//a.handles["CS_GetAuctionItems"] = a.DoGetAuctionItems
//	a.handles["CS_NewPriceAuctionItem"] = a.DoNewPriceAuctionItem
//a.handles["CS_GuildOperate"] = a.DoGuildOperate
//a.handles["CS_GetGuildMapsInfo"] = a.DoGetGuildMapsInfo
//	a.handles["CS_GotoGuildMap"] = a.DoGotoGuildMap
//a.handles["CS_EditorGuildNotice"] = a.DoEditorGuildNotice
//a.handles["CS_ChangePost"] = a.DoChangePost
//a.handles["CS_GetGuildRankInfo"] = a.DoGetGuildRankInfo
//	a.handles["CS_GetGuildRankBattleInfo"] = a.DoGetGuildRankBattleInfo
func (a *GameScene1Agent) DoGetGuildRankInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetGuildRankInfo{} //获取公会排名界面信息
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	msg := gamecore.GuildManagerObj.GetGuildRankInfo()
	player.(*gamecore.Player).SendMsgToClient("SC_GetGuildRankInfo", msg)

}
func (a *GameScene1Agent) DoGetGuildRankBattleInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetGuildRankBattleInfo{} //获取公会战 角色击杀信息
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	//场景信息 公会战场景
	mapdata := conf.GetGuildMapFileData(conf.GuildBattleMapID) //
	if mapdata == nil {
		return
	}
	scene := a.Scenes.Get(mapdata.NextSceneID)

	if scene == nil || scene.(*gamecore.Scene).Quit == true {
		//已经结束或未开始
		player.(*gamecore.Player).SendNoticeWordToClient(40)
		return
	}
	killdata := scene.(*gamecore.Scene).KillData
	if killdata == nil {
		//无数据
		return
	}
	msg := &protomsg.SC_GetGuildRankBattleInfo{}
	msg.AllCha = make([]*protomsg.GuildRankBattleChaInfo, 0)
	allguilds := killdata.Items()
	for _, v := range allguilds {
		//最多显示前20名
		if v == nil {
			continue
		}
		one := &protomsg.GuildRankBattleChaInfo{}
		one.Characterid = v.(*gamecore.SceneStatisticsCharacterInfo).Characterid
		one.Name = v.(*gamecore.SceneStatisticsCharacterInfo).Name
		one.Level = v.(*gamecore.SceneStatisticsCharacterInfo).Level
		one.KillCount = v.(*gamecore.SceneStatisticsCharacterInfo).KillCount
		one.DeathCount = v.(*gamecore.SceneStatisticsCharacterInfo).DeathCount
		one.GuildId = v.(*gamecore.SceneStatisticsCharacterInfo).GuildId
		one.GuildName = v.(*gamecore.SceneStatisticsCharacterInfo).GuildName
		one.Typeid = v.(*gamecore.SceneStatisticsCharacterInfo).Typeid
		msg.AllCha = append(msg.AllCha, one)
	}
	player.(*gamecore.Player).SendMsgToClient("SC_GetGuildRankBattleInfo", msg)
}
func (a *GameScene1Agent) DoChangePost(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ChangePost{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	targetplayer := a.Characters.Get(h2.Characterid)
	ischange := false
	if targetplayer == nil {
		ischange = gamecore.GuildManagerObj.ChangePost(player.(*gamecore.Player), h2, nil)
	} else {
		ischange = gamecore.GuildManagerObj.ChangePost(player.(*gamecore.Player), h2, targetplayer.(*gamecore.Player))
	}
	//发生了改变
	if ischange {
		//操作成功 返回公会信息
		myguild := player.(*gamecore.Player).MyGuild
		if myguild != nil {
			msg := gamecore.GuildManagerObj.GetGuildInfo(myguild.GuildId)
			player.(*gamecore.Player).SendMsgToClient("SC_GetGuildInfo", msg)
		}
	}
}

func (a *GameScene1Agent) DoEditorGuildNotice(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_EditorGuildNotice{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	log.Info("DoEditorGuildNotice:%s", h2.Notice)
	//过滤非法字符
	h2.Notice = wordsfilter.WF.DoReplace(h2.Notice)

	if gamecore.GuildManagerObj.EditorGuildNotice(player.(*gamecore.Player), h2) == true {
		//操作成功 返回公会信息
		myguild := player.(*gamecore.Player).MyGuild
		if myguild != nil {
			msg := gamecore.GuildManagerObj.GetGuildInfo(myguild.GuildId)
			player.(*gamecore.Player).SendMsgToClient("SC_GetGuildInfo", msg)
		}

	}

}

func (a *GameScene1Agent) DoGotoGuildMap(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GotoGuildMap{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	msg := gamecore.GuildManagerObj.GotoGuildMap(player.(*gamecore.Player), h2.ID)
	player.(*gamecore.Player).SendMsgToClient("SC_GotoGuildMap", msg)

}

func (a *GameScene1Agent) DoGetGuildMapsInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetGuildMapsInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	msg := gamecore.GuildManagerObj.GetGuildMapsInfo()
	player.(*gamecore.Player).SendMsgToClient("SC_GetGuildMapsInfo", msg)
}

func (a *GameScene1Agent) DoGuildOperate(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GuildOperate{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	if gamecore.GuildManagerObj.GuildOperate(player.(*gamecore.Player), h2) == true {
		//操作成功 返回所有公会信息
		msg := gamecore.GuildManagerObj.GetAllGuildsInfo()
		player.(*gamecore.Player).SendMsgToClient("SC_GetAllGuildsInfo", msg)
	}

}

//获取公会拍卖物品
func (a *GameScene1Agent) DoGetAuctionItems(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetAuctionItems{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	msg := gamecore.GuildManagerObj.GetAuctionItems(player.(*gamecore.Player))
	if msg != nil {
		player.(*gamecore.Player).SendMsgToClient("SC_GetAuctionItems", msg)
	}
}

//出价
func (a *GameScene1Agent) DoNewPriceAuctionItem(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_NewPriceAuctionItem{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.AuctionManagerObj.NewPrice(h2.Price, h2.ID, player.(*gamecore.Player))

	//重新返回数据
	msg := gamecore.GuildManagerObj.GetAuctionItems(player.(*gamecore.Player))
	if msg != nil {
		player.(*gamecore.Player).SendMsgToClient("SC_GetAuctionItems", msg)
	}
}
func (a *GameScene1Agent) DoDeleteGuildPlayer(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_DeleteGuildPlayer{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	targetplayer := a.Characters.Get(h2.Characterid)

	if targetplayer == nil {
		gamecore.GuildManagerObj.DeleteGuildPlayer(player.(*gamecore.Player), h2, nil)
	} else {
		gamecore.GuildManagerObj.DeleteGuildPlayer(player.(*gamecore.Player), h2, targetplayer.(*gamecore.Player))
	}

	if player.(*gamecore.Player).MyGuild != nil {
		msg := gamecore.GuildManagerObj.GetGuildInfo(player.(*gamecore.Player).MyGuild.GuildId)
		player.(*gamecore.Player).SendMsgToClient("SC_GetGuildInfo", msg)
	}

}
func (a *GameScene1Agent) DoResponseJoinGuildPlayer(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ResponseJoinGuildPlayer{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	targetplayer := a.Characters.Get(h2.Characterid)

	if targetplayer == nil {
		gamecore.GuildManagerObj.ResponseJoinGuild(player.(*gamecore.Player), h2, nil)
	} else {
		gamecore.GuildManagerObj.ResponseJoinGuild(player.(*gamecore.Player), h2, targetplayer.(*gamecore.Player))
	}

	if player.(*gamecore.Player).MyGuild != nil {
		msg := gamecore.GuildManagerObj.GetJoinGuildPlayer(player.(*gamecore.Player).MyGuild.GuildId)
		player.(*gamecore.Player).SendMsgToClient("SC_GetJoinGuildPlayer", msg)
	}
}
func (a *GameScene1Agent) DoGetJoinGuildPlayer(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetJoinGuildPlayer{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	//如果id小于等于0 则为获取所有公会信息
	if player.(*gamecore.Player).MyGuild != nil {
		msg := gamecore.GuildManagerObj.GetJoinGuildPlayer(player.(*gamecore.Player).MyGuild.GuildId)
		player.(*gamecore.Player).SendMsgToClient("SC_GetJoinGuildPlayer", msg)
	}

}
func (a *GameScene1Agent) DoGetAllGuildsInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetAllGuildsInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	msg := gamecore.GuildManagerObj.GetAllGuildsInfo()
	player.(*gamecore.Player).SendMsgToClient("SC_GetAllGuildsInfo", msg)

}
func (a *GameScene1Agent) DoCreateGuild(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_CreateGuild{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).CreateGuild(h2)

	//创建成功 发送公会信息给玩家
	if player.(*gamecore.Player).MyGuild != nil {
		msg := gamecore.GuildManagerObj.GetGuildInfo(player.(*gamecore.Player).MyGuild.GuildId)
		player.(*gamecore.Player).SendMsgToClient("SC_GetGuildInfo", msg)
	}

}

//申请加入公会
func (a *GameScene1Agent) DoJoinGuild(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_JoinGuild{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	log.Info("----------DoJoinGuild--%d", h2.ID)

	gamecore.GuildManagerObj.RequestJoinGuild(player.(*gamecore.Player), h2.ID)

}

//获取公会信息
func (a *GameScene1Agent) DoGetGuildInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetGuildInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	//如果id小于等于0 则为获取所有公会信息
	if h2.ID <= 0 {
		msg := gamecore.GuildManagerObj.GetAllGuildsInfo()
		player.(*gamecore.Player).SendMsgToClient("SC_GetAllGuildsInfo", msg)
	} else {
		msg := gamecore.GuildManagerObj.GetGuildInfo(h2.ID)
		player.(*gamecore.Player).SendMsgToClient("SC_GetGuildInfo", msg)
	}

}

//交易所相关
//a.handles["CS_GetExchangeShortCommoditys"] = a.DoGetExchangeShortCommoditys
//	a.handles["CS_GetExchangeDetailedCommoditys"] = a.DoGetExchangeDetailedCommoditys
//	a.handles["CS_BuyExchangeCommodity"] = a.DoBuyExchangeCommodity
//	a.handles["CS_ShelfExchangeCommodity"] = a.DoShelfExchangeCommodity
//a.handles["CS_GetSellUIInfo"] = a.DoGetSellUIInfo
//a.handles["CS_UnShelfExchangeCommodity"] = a.DoUnShelfExchangeCommodity
//a.handles["CS_GetWorldAuctionItems"] = a.DoGetWorldAuctionItems
//a.handles["CS_NewPriceWorldAuctionItem"] = a.DoNewPriceWorldAuctionItem
//出价
func (a *GameScene1Agent) DoNewPriceWorldAuctionItem(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_NewPriceWorldAuctionItem{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.AuctionManagerObj.NewPrice(h2.Price, h2.ID, player.(*gamecore.Player))

	//重新返回数据
	msg := gamecore.AuctionManagerObj.GetWorldAuctionItems(player.(*gamecore.Player))
	player.(*gamecore.Player).SendMsgToClient("SC_GetWorldAuctionItems", msg)
}
func (a *GameScene1Agent) DoGetWorldAuctionItems(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetWorldAuctionItems{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	msg := gamecore.AuctionManagerObj.GetWorldAuctionItems(player.(*gamecore.Player))
	player.(*gamecore.Player).SendMsgToClient("SC_GetWorldAuctionItems", msg)
}
func (a *GameScene1Agent) DoUnShelfExchangeCommodity(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_UnShelfExchangeCommodity{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.ExchangeManagerObj.UnShelfExchangeCommodity_Lock(h2.ID)
	msg := player.(*gamecore.Player).GetSellUIInfo()
	player.(*gamecore.Player).SendMsgToClient("SC_GetSellUIInfo", msg)
}
func (a *GameScene1Agent) DoGetSellUIInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetSellUIInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	msg := player.(*gamecore.Player).GetSellUIInfo()

	player.(*gamecore.Player).SendMsgToClient("SC_GetSellUIInfo", msg)
}
func (a *GameScene1Agent) DoShelfExchangeCommodity(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ShelfExchangeCommodity{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	ok, item := player.(*gamecore.Player).ShelfBagItem2Exchange(h2.BagPos)
	if !ok || item == nil {
		return
	}
	//	ItemID            int32   `json:"itemid"`
	//	Level             int32   `json:"level"`
	//	PriceType         int32   `json:"pricetype"`         //价格类型 1金币 2砖石
	//	Price             int32   `json:"price"`             //价格
	//	SellerUid         int32   `json:"sellerUid"`         //卖家UID(账号ID)
	//	SellerCharacterid int32   `json:"sellerCharacterid"` //卖家角色ID
	dataitem := &db.DB_PlayerItemTransactionInfo{}
	dataitem.ItemID = item.TypeID
	dataitem.Level = item.Level
	dataitem.PriceType = h2.PriceType
	dataitem.Price = h2.Price
	dataitem.SellerUid = data.Uid
	dataitem.SellerCharacterid = player.(*gamecore.Player).Characterid
	//data *db.DB_PlayerItemTransactionInfo
	gamecore.ExchangeManagerObj.ShelfExchangeCommodity(dataitem)

	//发送玩家售卖UI信息
	msg := player.(*gamecore.Player).GetSellUIInfo()
	player.(*gamecore.Player).SendMsgToClient("SC_GetSellUIInfo", msg)

}
func (a *GameScene1Agent) DoBuyExchangeCommodity(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_BuyExchangeCommodity{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	commodityv := gamecore.ExchangeManagerObj.Commoditys.Get(h2.ID)
	if commodityv == nil {
		return
	}
	commodity := commodityv.(*db.DB_PlayerItemTransactionInfo)
	ok := gamecore.ExchangeManagerObj.BuyExchangeCommodity(h2, player.(*gamecore.Player))
	if ok {
		//购买成功后重新发送界面信息
		mail := gamecore.ExchangeManagerObj.GetExchangeDetailedCommoditys(commodity.ItemID)
		player.(*gamecore.Player).SendMsgToClient("SC_GetExchangeDetailedCommoditys", mail)
	}

}
func (a *GameScene1Agent) DoGetExchangeDetailedCommoditys(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetExchangeDetailedCommoditys{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	mail := gamecore.ExchangeManagerObj.GetExchangeDetailedCommoditys(h2.ItemID)

	player.(*gamecore.Player).SendMsgToClient("SC_GetExchangeDetailedCommoditys", mail)
}
func (a *GameScene1Agent) DoGetExchangeShortCommoditys(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetExchangeShortCommoditys{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	mail := gamecore.ExchangeManagerObj.GetExchangeShortCommoditys()

	player.(*gamecore.Player).SendMsgToClient("SC_GetExchangeShortCommoditys", mail)
}

//邮件信息
//邮件相关
//	a.handles["CS_GetMailsList"] = a.DoGetMailsList
//	a.handles["CS_GetMailInfo"] = a.DoGetMailInfo
//	a.handles["CS_GetMailRewards"] = a.DoGetMailRewards
//a.handles["CS_DeleteNoRewardMails"] = a.DoDeleteNoRewardMails
func (a *GameScene1Agent) DoDeleteNoRewardMails(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_DeleteNoRewardMails{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	player.(*gamecore.Player).MyMails.DeleteNoRewardMails()
	mails := player.(*gamecore.Player).MyMails.GetMailsList()

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailsList", mails)
}
func (a *GameScene1Agent) DoGetMailRewards(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailRewards{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mail := player.(*gamecore.Player).MyMails.GetMailRewards(h2.Id)

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailRewards", mail)
}
func (a *GameScene1Agent) DoGetMailInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mail := player.(*gamecore.Player).MyMails.GetOneMailInfo(h2.Id)

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailInfo", mail)
}
func (a *GameScene1Agent) DoGetMailsList(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailsList{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mails := player.(*gamecore.Player).MyMails.GetMailsList()

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailsList", mails)
}

//发送消息给全服玩家
func (a *GameScene1Agent) SendMsg2QuanFu(msgtype string, msg proto.Message) {
	allplayer := a.Players.Items()
	for _, v := range allplayer {
		if v == nil {
			continue
		}
		v.(*gamecore.Player).SendMsgToClient(msgtype, msg)
	}
}

//聊天信息
func (a *GameScene1Agent) DoChatInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ChatInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	mainunit := player.(*gamecore.Player).MainUnit
	if mainunit == nil {
		return
	}
	characterid := player.(*gamecore.Player).Characterid

	//过滤非法字符
	h2.Content = wordsfilter.WF.DoReplace(h2.Content)

	//

	////聊天频道 1附近 2全服 3私聊 4队伍 5公会
	if h2.Channel == 1 {

		//
		if player.(*gamecore.Player).SendChatCheck() == false {
			player.(*gamecore.Player).SendNoticeWordToClient(39)
			return
		}

		if player.(*gamecore.Player).CurScene == nil {
			return
		}
		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.SrcCharacterID = characterid
		msg.Content = h2.Content //内容过滤
		allplayer := player.(*gamecore.Player).CurScene.GetAllPlayerUseLock()
		for _, v := range allplayer {
			if v == nil {
				continue
			}
			v.SendMsgToClient("SC_ChatInfo", msg)
		}

	} else if h2.Channel == 2 {

		//收费
		if player.(*gamecore.Player).SendChatCheck() == false {
			player.(*gamecore.Player).SendNoticeWordToClient(39)
			return
		}

		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.SrcCharacterID = characterid
		msg.Content = h2.Content //内容过滤
		a.SendMsg2QuanFu("SC_ChatInfo", msg)
		//		allplayer := a.Players.Items()
		//		for _, v := range allplayer {
		//			if v == nil {
		//				continue
		//			}
		//			v.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)
		//		}
	} else if h2.Channel == 3 {
		destplayer := a.Players.Get(h2.DestPlayerUID)
		if destplayer == nil {
			//未找到当前玩家
			return
		}

		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.SrcCharacterID = characterid
		msg.DestPlayerUID = h2.DestPlayerUID
		msg.Content = h2.Content //内容过滤
		player.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)
		destplayer.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)

	} else if h2.Channel == 4 {
		team := gamecore.TeamManagerObj.GetTeam(player.(*gamecore.Player))
		if team == nil {
			return
		}
		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.SrcCharacterID = characterid
		msg.Content = h2.Content //内容过滤
		allplayer := team.Players.Items()
		for _, v := range allplayer {
			if v == nil {
				continue
			}
			v.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)
		}
	} else if h2.Channel == 5 {
		if player.(*gamecore.Player).MyGuild == nil {
			return
		}
		guild := gamecore.GuildManagerObj.GetGuild(player.(*gamecore.Player).MyGuild.GuildId)
		if guild == nil {
			return
		}

		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.SrcCharacterID = characterid
		msg.Content = h2.Content //内容过滤
		allplayer := guild.CharactersMap.Items()
		for _, v := range allplayer {
			if v == nil {
				continue
			}
			player1 := a.GetPlayerByChaID(v.(*gamecore.GuildCharacterInfo).Characterid)
			if player1 != nil {
				player1.SendMsgToClient("SC_ChatInfo", msg)
			}

		}
	}

}

func (a *GameScene1Agent) DoQuickRevive(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_QuickRevive{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*gamecore.Player).MainUnit != nil {
		player.(*gamecore.Player).MainUnit.QuickRevive = h2
	}
}

func (a *GameScene1Agent) DoLookVedioSucc(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_LookVedioSucc{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*gamecore.Player).MainUnit != nil {
		player.(*gamecore.Player).MainUnit.LookViewGetDiamond = h2
	}

}

//a.handles["CS_OutTeam"] = a.DoOutTeam
func (a *GameScene1Agent) DoOutTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_OutTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player1 := a.Players.Get(h2.OutPlayerUID)
	if player1 == nil {
		return
	}
	if player == player1 {
		gamecore.TeamManagerObj.LeaveTeam(player1.(*gamecore.Player))
	} else {
		gamecore.TeamManagerObj.OutTeam(player.(*gamecore.Player), player1.(*gamecore.Player))
	}

}

//a.handles["CS_UseAI"] = a.DoUseAI
func (a *GameScene1Agent) DoUseAI(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")
	h2 := &protomsg.CS_UseAI{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).UseAI(h2.AIid)

}

//CS_LodingScene
func (a *GameScene1Agent) DoLodingScene(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).IsLoadedSceneSucc = true

}

func (a *GameScene1Agent) DoPlayerMove(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")
	h2 := &protomsg.CS_PlayerMove{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//log.Info("---------%v", h2)

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).MoveCmd(h2)

}

//
func (a *GameScene1Agent) Run() {

	a.Init()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go a.doMessage(data)

	}
}

func (a *GameScene1Agent) doMessage(data []byte) {
	//log.Info("----------game5g----readmsg---------")
	h1 := &protomsg.MsgBase{}
	err := proto.Unmarshal(data, h1)
	if err != nil {
		log.Info("--error")
	} else {

		//log.Info("--MsgType:" + h1.MsgType)
		if f, ok := a.handles[h1.MsgType]; ok {
			f(h1)
		}

	}

}

func (a *GameScene1Agent) OnClose() {

	a.IsClose = true

	scenes := a.Scenes.Items()
	for _, v := range scenes {
		v.(*gamecore.Scene).Close()
	}
	gamecore.CopyMapMgrObj.Close()
	gamecore.TeamManagerObj.Close()

	gamecore.GuildManagerObj.Close()

	gamecore.AuctionManagerObj.Close()

	gamecore.ExchangeManagerObj.Close()
	gamecore.GameCoreDataManagerObj.Close()

	a.wgScene.Wait()

	//存储玩家数据

	log.Debug("GameScene1Agent OnClose")
}

func (a *GameScene1Agent) WriteMsg(msg interface{}) {

}
func (a *GameScene1Agent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *GameScene1Agent) RegisterToGate() {
	t2 := protomsg.MsgRegisterToGate{
		ModeType: a.ServerName,
	}

	t1 := protomsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *GameScene1Agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *GameScene1Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *GameScene1Agent) Close() {
	a.conn.Close()
}

func (a *GameScene1Agent) Destroy() {
	a.conn.Destroy()
}
