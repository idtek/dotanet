package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/log"
	"dq/protobuf"
	"strconv"

	//"dq/timer"
	"dq/db"
	"dq/utils"
	"dq/vec2d"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

type ReCreateUnit struct {
	//场景中的NPC 死亡后重新创建信息
	ReCreateInfo *conf.Unit
	DeathTime    float64 //死亡时间
}
type ChangeSceneFunc interface {
	PlayerChangeScene(player *Player, doorway conf.DoorWay)
	SendMsg2QuanFu(msgtype string, msg proto.Message)
}

//场景统计角色信息
type SceneStatisticsCharacterInfo struct {
	Characterid int32  //角色id
	Name        string //角色名字
	Typeid      int32  //英雄类型
	GuildId     int32  //公会ID
	GuildName   string //公会名字
	Level       int32  //等级
	GroupID     int32  //groupid
	MainUnit    *Unit  //单位
	//统计数据
	KillCount  int32 //击杀次数
	DeathCount int32 //死亡次数

}

//场景统计 (击杀死亡)
type SceneStatistics struct {
	KillData *utils.BeeMap //击杀数据 key:Characterid value:SceneStatisticsCharacterInfo
}

//初始化击杀记录的玩家信息
func (this *SceneStatistics) InitPlayerKill(player *Player) {
	if player == nil || this.KillData == nil {
		return
	}
	mainunit := player.MainUnit
	if mainunit == nil {
		return
	}
	killdata1 := this.KillData.Get(player.Characterid)

	if killdata1 == nil {
		killdata1 = &SceneStatisticsCharacterInfo{}
		killdata1.(*SceneStatisticsCharacterInfo).Characterid = player.Characterid
		killdata1.(*SceneStatisticsCharacterInfo).Name = mainunit.Name
		killdata1.(*SceneStatisticsCharacterInfo).Level = mainunit.Level
		killdata1.(*SceneStatisticsCharacterInfo).Typeid = mainunit.TypeID
		killdata1.(*SceneStatisticsCharacterInfo).GroupID = player.GroupID

		killguild := player.MyGuild
		if killguild != nil {
			killdata1.(*SceneStatisticsCharacterInfo).GuildId = killguild.GuildId
			killdata1.(*SceneStatisticsCharacterInfo).GuildName = killguild.GuildName
		}
		killdata1.(*SceneStatisticsCharacterInfo).MainUnit = mainunit
		killdata1.(*SceneStatisticsCharacterInfo).KillCount = 0
		killdata1.(*SceneStatisticsCharacterInfo).DeathCount = 0

		this.KillData.Set(player.Characterid, killdata1)
	}

}

//击杀动作(不需要加锁 场景逻辑循坏里执行)
func (this *SceneStatistics) KillAction(killer *Unit, bekiller *Unit) {
	if this.KillData == nil {
		return
	}

	if killer == nil || bekiller == nil {
		return
	}
	if killer.MyPlayer == nil || bekiller.MyPlayer == nil {
		return
	}
	//击杀者信息
	killdata1 := this.KillData.Get(killer.MyPlayer.Characterid)
	if killdata1 == nil {
		log.Info("无效击杀")
		return
	}
	killdata1.(*SceneStatisticsCharacterInfo).KillCount += 1
	killdata1.(*SceneStatisticsCharacterInfo).MainUnit = killer

	//被击杀者信息
	bekilldata1 := this.KillData.Get(bekiller.MyPlayer.Characterid)
	if bekilldata1 == nil {
		log.Info("无效被击杀")
		return
	}
	bekilldata1.(*SceneStatisticsCharacterInfo).DeathCount += 1
	bekilldata1.(*SceneStatisticsCharacterInfo).MainUnit = bekiller

	log.Info("kill action:%d", killdata1.(*SceneStatisticsCharacterInfo).KillCount)

}

//清除击杀数据
func (this *SceneStatistics) KillClean() {
	this.KillData = utils.NewBeeMap()
}

type Scene struct {
	SceneStatistics
	DuoBaoQiBing       *SceneDuoBaoInfo //夺宝奇兵相关
	conf.SceneFileData                  //场景文件信息
	DataFileID         int32            //数据文件ID 不唯一 对应excel表里的typeid
	FirstUpdateTime    int64            //上次更新时间
	MoveCore           *cyward.WardCore //移动核心
	SceneName          string           //场景名字
	CurFrame           int32            //当前帧

	EveryTimeDoRemainTime float32        //每秒钟干得事的剩余时间
	DoorWays              []conf.DoorWay //传送门

	ReCreateUnitInfo   map[*ReCreateUnit]*ReCreateUnit //重新创建NPC信息
	NeedCreateBossInfo *utils.BeeMap                   //需要创建的boss信息
	NoramlUnitCount    int32                           //当前场景中的普通单位数量

	CharacterGroups *utils.BeeMap //角色小组信息

	Players   map[int32]*Player           //游戏中所有的玩家
	Units     map[int32]*Unit             //游戏中所有的单位
	ZoneUnits map[utils.SceneZone][]*Unit //区域中的单位

	Bullets     map[int32]*Bullet             //游戏中所有的子弹
	ZoneBullets map[utils.SceneZone][]*Bullet //区域中的的子弹

	SceneItems     map[int32]*SceneItem             //游戏场景中所有的道具
	ZoneSceneItems map[utils.SceneZone][]*SceneItem //区域中的的道具

	//Halos          map[int32]*Halo             //游戏中所有的光环
	Halos          *utils.BeeMap
	ZoneHalos      map[utils.SceneZone][]*Halo //区域中的的光环
	CanRemoveHalos map[int32]*Halo             //可以被删除的halo 比如击杀单位后 halo无效

	NextAddUnit    *utils.BeeMap //下一帧需要增加的单位
	NextRemoveUnit *utils.BeeMap //下一帧需要删除的单位

	HuiChengPlayer *utils.BeeMap //回城的玩家

	NextAddPlayer    *utils.BeeMap //下一帧需要增加的玩家
	NextRemovePlayer *utils.BeeMap //下一帧需要删除的玩家

	//地图信息
	//DropItems     *utils.BeeMap //本地图会掉落的道具
	//BossFreshTime int32         //boss刷新时间 秒为单位

	ChangeScene ChangeSceneFunc
	Sverver     ServerInterface
	Quit        bool //是否退出
	CleanPlayer bool //清除玩家到和平世界
	SceneFrame  int32

	playerlock *sync.RWMutex //玩家操作同步操作锁
}

func CreateScene(data *conf.SceneFileData, parent ChangeSceneFunc, server ServerInterface) *Scene {
	scene := &Scene{}
	scene.ChangeScene = parent
	scene.Sverver = server
	scene.SceneFileData = *data
	scene.DataFileID = data.TypeID
	scene.SceneName = data.ScenePath
	scene.Quit = false
	scene.CleanPlayer = false

	scene.playerlock = new(sync.RWMutex)
	scene.Init()
	return scene
}

//初始化
func (this *Scene) Init() {
	this.SceneFrame = 22
	this.CurFrame = 0
	this.EveryTimeDoRemainTime = 1
	this.NoramlUnitCount = 0

	this.FirstUpdateTime = time.Now().UnixNano()

	this.NextAddUnit = utils.NewBeeMap()
	this.NextRemoveUnit = utils.NewBeeMap()
	this.HuiChengPlayer = utils.NewBeeMap()

	this.NextAddPlayer = utils.NewBeeMap()
	this.NextRemovePlayer = utils.NewBeeMap()

	this.NeedCreateBossInfo = utils.NewBeeMap()

	this.CharacterGroups = utils.NewBeeMap()

	//this.DropItems = utils.NewBeeMap()

	//
	this.ReCreateUnitInfo = make(map[*ReCreateUnit]*ReCreateUnit)

	this.Players = make(map[int32]*Player)
	this.Units = make(map[int32]*Unit)
	this.ZoneUnits = make(map[utils.SceneZone][]*Unit)

	this.Bullets = make(map[int32]*Bullet)
	this.ZoneBullets = make(map[utils.SceneZone][]*Bullet)

	this.SceneItems = make(map[int32]*SceneItem)
	this.ZoneSceneItems = make(map[utils.SceneZone][]*SceneItem)

	//this.Halos = make(map[int32]*Halo)
	this.Halos = utils.NewBeeMap()
	this.ZoneHalos = make(map[utils.SceneZone][]*Halo)
	this.CanRemoveHalos = make(map[int32]*Halo)

	this.DoStartException()

	scenedata := conf.GetSceneData(this.ScenePath)

	//场景碰撞区域
	this.MoveCore = cyward.CreateWardCore()
	for _, v := range scenedata.Collides {
		//log.Info("Collide %v", v)
		if v.IsRect == true {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			r := vec2d.Vec2{v.Width, v.Height}
			b1 := this.MoveCore.CreateBody(pos, r, 1, 3)
			b1.Name = v.Name
		} else {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			b1 := this.MoveCore.CreateBodyPolygon(pos, v.Points, 1, 3)
			b1.Name = v.Name
		}
	}
	//场景分区数据 创建100个单位
	createunitdata := conf.GetCreateUnitData(this.CreateUnit)
	for _, v := range createunitdata.Units {
		v.ReCreateTime = v.ReCreateTime //敌人单位+30秒重生时间
		oneunit := this.CreateUnitByConf(v)
		if oneunit != nil {
			//保存会掉落的道具
			dropitems := oneunit.GetDropItems()
			for _, v1 := range dropitems {
				GameCoreDataManagerObj.AddDrop(this.TypeID, v1)
			}
			//boss刷新时间
			if oneunit.UnitType == 4 {
				GameCoreDataManagerObj.SetBossFreshTime(this.TypeID, int32(v.ReCreateTime))
				//如果是副本则需要等普通怪全部死亡后才刷新BOSS
				if this.CreateBossRule == 1 {
					this.NextRemoveUnit.Set(oneunit.ID, oneunit)
					this.NeedCreateBossInfo.Set(&v, &v)
				}
			}

		}

		//log.Info("createunity :%v", v)
	}
	//传送门显示
	this.DoorWays = createunitdata.DoorWays
	for _, v := range this.DoorWays {
		halo := NewHalo(v.HaloTypeID, 1)
		halo.SetParent(nil)
		halo.IsActive = false
		halo.Position = vec2d.Vec2{X: v.X, Y: v.Y}
		if halo != nil {
			this.AddHalo(halo)
		}
	}

	//创建道具
	//this.CreateSceneItem(1, vec2d.Vec2{X: 34, Y: 84})
	//创建道具
	//this.CreateSceneItem(2, vec2d.Vec2{X: 28, Y: 84})

	//	for i := 0; i < 4; i++ {
	//		//创建英雄
	//		hero1 := CreateUnit(this, 5+int32(i))
	//		hero1.AttackMode = 3 //全体攻击模式
	//		hero1.SetAI(NewNormalAI(hero1))
	//		//设置移动核心body
	//		pos1 := vec2d.Vec2{float64(2 + i*2), float64(5)}
	//		r1 := vec2d.Vec2{hero1.CollisionR, hero1.CollisionR}
	//		hero1.Body = this.MoveCore.CreateBody(pos1, r1, 0, 1)
	//		this.Units[hero1.ID] = hero1
	//	}

	//创建英雄2
	//	hero2 := CreateUnit(this, 15)
	//	hero2.AttackMode = 1 //和平攻击模式
	//	hero2.SetAI(NewNormalAI(hero2))
	//	//hero2.AddSkill(52, 4)
	//	//设置移动核心body
	//	pos2 := vec2d.Vec2{float64(15), float64(15)}
	//	r2 := vec2d.Vec2{hero2.CollisionR, hero2.CollisionR}
	//	hero2.Body = this.MoveCore.CreateBody(pos2, r2, 0, 2)
	//	this.Units[hero2.ID] = hero2

}

//设置角色小组ID
func (this *Scene) SetSceneCharacterGroups(characterid int32, groupid int32) {
	this.CharacterGroups.Set(characterid, groupid)
}

//检查是否要创建boss
func (this *Scene) CheckCreateBoss() {
	//如果是副本则需要等普通怪全部死亡后才刷新BOSS
	if this.CreateBossRule == 0 || this.NoramlUnitCount > 0 || this.NeedCreateBossInfo.Size() <= 0 {
		return
	}
	items := this.NeedCreateBossInfo.Items()
	for _, v := range items {
		this.CreateUnitByConf(*(v.(*conf.Unit)))
	}
	this.NeedCreateBossInfo.DeleteAll()
}

//场景开始 处理异常情况
func (this *Scene) DoStartException() {
	if this.Exception == 0 {
		return
	}
	switch this.Exception {
	case 1: //工会战
		{
			this.KillClean() //记录击杀数据
		}
		break
	case 2: //夺宝奇兵
		params := utils.GetInt32FromString3(this.ExceptionParam, ",")
		if len(params) < 2 {
			return
		}
		time := params[0]   //时间
		typeid := params[1] //道具ID
		this.DuoBaoQiBing = CreateSceneDuoBaoInfo(float32(time), typeid, this)
		break
	case 3: //竞技场
		{
			this.KillClean() //记录击杀数据
		}
		break
	default:
	}
}

type GuildRankInfo struct {
	GuildId    int32
	KillCount  int32
	DeathCount int32
	Score      int32
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type GuildRankInfoList []*GuildRankInfo

func (p GuildRankInfoList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p GuildRankInfoList) Len() int      { return len(p) }
func (p GuildRankInfoList) Less(i, j int) bool {
	if p[i].Score == p[j].Score {
		if p[i].KillCount == p[j].KillCount {
			return p[i].GuildId > p[j].GuildId
		}
		return p[i].KillCount > p[j].KillCount
	}

	return p[i].Score > p[j].Score
}

//获取竞技场面板信息
func (this *Scene) GetBattleInfo() *protomsg.SC_GetBattleHeroInfo {
	if this.Exception != 3 || this.KillData == nil { //竞技场
		return nil
	}
	msg := &protomsg.SC_GetBattleHeroInfo{}
	msg.Group1 = make([]*protomsg.BattleOverPlayerOneInfo, 0)
	msg.Group2 = make([]*protomsg.BattleOverPlayerOneInfo, 0)
	msg.WinnerGroup = 0 // 0表示 没有结束 1表示队伍1胜利 2表示队伍2胜利
	killdataitems := this.KillData.Items()
	for chaid, killdata1 := range killdataitems {
		if killdata1 == nil {
			continue
		}
		mainunit := killdata1.(*SceneStatisticsCharacterInfo).MainUnit
		if mainunit == nil {
			continue
		}
		player := killdata1.(*SceneStatisticsCharacterInfo).MainUnit.MyPlayer
		if player == nil {
			continue
		}

		one := &protomsg.BattleOverPlayerOneInfo{}
		one.Characterid = chaid.(int32)
		one.Score = player.BattleScore
		one.Name = mainunit.Name
		one.Typeid = mainunit.TypeID
		one.KillCount = killdata1.(*SceneStatisticsCharacterInfo).KillCount
		one.DeathCount = killdata1.(*SceneStatisticsCharacterInfo).DeathCount
		one.Level = killdata1.(*SceneStatisticsCharacterInfo).Level
		chadata := player.GetLastDBData()
		one.EquipItems = make([]string, UnitEquitCount)
		one.EquipItems[0] = chadata.Item1
		one.EquipItems[1] = chadata.Item2
		one.EquipItems[2] = chadata.Item3
		one.EquipItems[3] = chadata.Item4
		one.EquipItems[4] = chadata.Item5
		one.EquipItems[5] = chadata.Item6

		if player.GroupID == 1 {
			msg.Group1 = append(msg.Group1, one)
		} else if player.GroupID == 2 {
			msg.Group2 = append(msg.Group2, one)
		}

	}

	return msg
}

//场景结束玩家离开之前 处理异常情况
func (this *Scene) DoEndExceptionBeforePlayerLeave() {
	if this.Exception == 0 {
		return
	}
	switch this.Exception {
	case 3: //竞技场
		//一次击杀得1分 一次死亡-1分 按分数排名
		if this.KillData == nil {
			return
		}
		killdataitems := this.KillData.Items()
		group1score := int32(0)
		group2score := int32(0)
		group1maxscore := int32(-10000)
		group2maxscore := int32(-10000)
		group1maxscorechaid := int32(-10000)
		group2maxscorechaid := int32(-10000)
		group1 := make([]*SceneStatisticsCharacterInfo, 0)
		group2 := make([]*SceneStatisticsCharacterInfo, 0)

		for _, v := range killdataitems { //
			if v == nil {
				continue
			}
			sdata := v.(*SceneStatisticsCharacterInfo)
			onescore := sdata.KillCount - sdata.DeathCount
			if sdata.GroupID == 1 {
				group1score += onescore
				if onescore > group1maxscore {
					group1maxscore = onescore
					group1maxscorechaid = sdata.Characterid
				}
				group1 = append(group1, sdata)
			} else if sdata.GroupID == 2 {
				group2score += onescore
				if onescore > group2maxscore {
					group2maxscore = onescore
					group2maxscorechaid = sdata.Characterid
				}
				group2 = append(group2, sdata)
			}

		}

		paramarr := utils.GetInt32FromString3(this.ExceptionParam, ",")
		wingetexp := int32(0)
		if len(paramarr) > 0 {
			wingetexp = paramarr[0]
		}

		//Players
		//(playerchaid , name , typeid , addwin , addlose , adddrew , addmvp , addfmvp )
		winnergroup := int32(0)
		if group1score > group2score {
			winnergroup = 1
			for _, v := range group1 { //胜利
				d1 := v
				if d1.Characterid == group1maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 1, 0, 0, 1, 0) //MVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 1, 0, 0, 0, 0)
				}
				if d1.MainUnit != nil {
					d1.MainUnit.AddExperience(wingetexp)
				}
			}
			for _, v := range group2 { //失败
				d1 := v
				if d1.Characterid == group2maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 1, 0, 0, 1) //fMVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 1, 0, 0, 0)
				}
			}
		} else if group1score == group2score {
			winnergroup = 3
			for _, v := range group1 { //平局
				d1 := v
				if d1.Characterid == group1maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 0, 1, 1, 0) //MVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 0, 1, 0, 0)
				}
			}
			for _, v := range group2 { //平局
				d1 := v
				if d1.Characterid == group2maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 0, 1, 1, 0) //fMVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 0, 1, 0, 0)
				}
			}
		} else {
			winnergroup = 2
			for _, v := range group2 { //胜利
				d1 := v
				if d1.Characterid == group2maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 1, 0, 0, 1, 0) //MVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 1, 0, 0, 0, 0)
				}
				if d1.MainUnit != nil {
					d1.MainUnit.AddExperience(wingetexp)
				}
			}
			for _, v := range group1 { //失败
				d1 := v
				if d1.Characterid == group1maxscorechaid {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 1, 0, 0, 1) //fMVP
				} else {
					BattleMgrObj.ChangeData(d1.Characterid, d1.Name, d1.Typeid, 0, 1, 0, 0, 0)
				}
			}
		}
		//GroupID
		//胜利失败显示
		overmsg := this.GetBattleInfo()
		overmsg.WinnerGroup = winnergroup
		for _, v := range killdataitems { //
			if v == nil {
				continue
			}
			sdata := v.(*SceneStatisticsCharacterInfo)
			//给所有玩家发生胜利失败情况
			if sdata.MainUnit.MyPlayer != nil && overmsg != nil {
				sdata.MainUnit.MyPlayer.SendMsgToClient("SC_GetBattleHeroInfo", overmsg)
			}
		}
		break
	default:
	}
}

//场景结束 处理异常情况
func (this *Scene) DoEndException() {
	if this.Exception == 0 {
		return
	}
	switch this.Exception {
	case 1: //工会战
		{
			//处理公会排名 一次击杀得1分 一次死亡-1分 按分数排名
			if this.KillData == nil {
				return
			}
			killdataitems := this.KillData.Items()
			guildmapdata := make(map[int32]*GuildRankInfo)
			for _, v := range killdataitems { //计算所有参与的公会的得分
				if v == nil {
					continue
				}
				sdata := v.(*SceneStatisticsCharacterInfo)
				guildid := sdata.GuildId
				if guildid <= 0 {
					continue
				}
				if gri, ok := guildmapdata[guildid]; ok {
					gri.Score += sdata.KillCount - sdata.DeathCount
					gri.KillCount += sdata.KillCount
					gri.DeathCount += sdata.DeathCount
				} else {
					gri := &GuildRankInfo{}
					gri.GuildId = guildid
					gri.Score = sdata.KillCount - sdata.DeathCount
					gri.KillCount = sdata.KillCount
					gri.DeathCount = sdata.DeathCount
					guildmapdata[guildid] = gri
				}

			}
			//--处理排序--
			p := make(GuildRankInfoList, len(guildmapdata))
			i := 0
			for _, v := range guildmapdata {
				p[i] = v
				i++
			}
			sort.Sort(p)

			//清除之前的排序
			GuildManagerObj.ReSetGuildRank()
			//获取公会经验值
			guildexparr := utils.GetInt32FromString3(this.ExceptionParam, ",")

			//设置排序
			for k, v := range p {
				log.Info("-----------rank:%d  %d", v.GuildId, k)
				GuildManagerObj.SetGuildRank(v.GuildId, int32(k)+1)
				if int32(k) == 0 {
					//获得第一名时通知全服玩家的 跑马灯
					oneguild := GuildManagerObj.GetGuild(p[0].GuildId)
					if oneguild != nil {
						this.SendNoticeWordToQuanFuPlayer(48, oneguild.Name)
					}
				}
				//获取公会经验
				if len(guildexparr) > 0 {
					getexp := int32(0)
					if len(guildexparr) <= k {
						getexp = guildexparr[len(guildexparr)-1]
					} else {
						getexp = guildexparr[k]
					}

					GuildManagerObj.AddGuildExp(getexp, v.GuildId)
				}

			}
		}
	case 2: //夺宝奇兵
		if this.DuoBaoQiBing != nil {

		}
		break
	case 3: //竞技场

		break
	default:
	}
}

func (this *Scene) CreateUnitByConf(v conf.Unit) *Unit {
	unit := CreateUnit(this, v.TypeID)
	unit.SetAI(NewNormalAI(unit))
	pos := vec2d.Vec2{v.X, v.Y}
	r := vec2d.Vec2{unit.CollisionR, unit.CollisionR}
	unit.Body = this.MoveCore.CreateBody(pos, r, 0, 1)
	unit.Body.Direction = vec2d.Vec2{X: 0, Y: 1}
	unit.Body.Direction.Rotate(-v.Rotation)
	unit.SetReCreateInfo(&v)
	//log.Info("Collide %v  %f", unit.Body.Direction, v.Rotation)
	this.Units[unit.ID] = unit
	//添加场景BUFF
	//unit.AddBuffFromStr("236", 1, unit)
	return unit
}

func (this *Scene) UnitBlink(unit interface{}) {
	//	if unit != nil {
	//		targetpos := vec2d.Vec2{X: 5, Y: 5}
	//		unit.(*Unit).Body.BlinkToPos(targetpos)
	//	}
}

//通过ID查找单位
func (this *Scene) FindUnitByID(id int32) *Unit {
	return this.Units[id]
}

//获取可视范围内的所有单位
func (this *Scene) FindVisibleUnitsByPos(pos vec2d.Vec2) []*Unit {

	units := make([]*Unit, 0)

	zones := utils.GetVisibleZones((pos.X), (pos.Y))
	//遍历可视区域
	for _, vzone := range zones {
		if _, ok := this.ZoneUnits[vzone]; ok {
			//遍历区域中的单位
			for _, unit := range this.ZoneUnits[vzone] {
				units = append(units, unit)

			}
		}
	}

	return units
}

//获取可视范围内的所有单位
func (this *Scene) FindVisibleUnits(my *Unit) []*Unit {

	units := make([]*Unit, 0)

	zones := utils.GetVisibleZones((my.Body.Position.X), (my.Body.Position.Y))
	//遍历可视区域
	for _, vzone := range zones {
		if _, ok := this.ZoneUnits[vzone]; ok {
			//遍历区域中的单位
			for _, unit := range this.ZoneUnits[vzone] {
				units = append(units, unit)

			}
		}
	}

	return units
}

//获取可视范围内的所有玩家
//func (this *Scene) FindVisiblePlayers(my *Unit) []*Player {

//	units := make([]*Unit, 0)

//	zones := utils.GetVisibleZones((my.Body.Position.X), (my.Body.Position.Y))
//	//遍历可视区域
//	for _, vzone := range zones {
//		if _, ok := this.ZoneUnits[vzone]; ok {
//			//遍历区域中的单位
//			for _, unit := range this.ZoneUnits[vzone] {
//				units = append(units, unit)

//			}
//		}
//	}

//	return units
//}

//

//玩家将要主动离开
func (this *Scene) PlayerWillLeave(player *Player) {
	if this.DuoBaoQiBing != nil {
		this.DuoBaoQiBing.DoPlayerLostBox(player)
	}
}

//传送到和平城
func (this *Scene) HuiCheng(player *Player) {
	if this.ChangeScene == nil || player == nil {
		return
	}
	if this.HuiChengMode == 0 {
		this.HuiChengHePingShiJie(player)
	} else if this.HuiChengMode == 1 {
		doorway := conf.DoorWay{}
		doorway.NextX = -1
		doorway.NextY = -1
		doorway.NextSceneID = this.TypeID

		this.HuiChengPlayer.Set(player, &doorway)
	}

}

//传送到和平城
func (this *Scene) HuiChengHePingShiJie(player *Player) {
	if this.ChangeScene == nil || player == nil {
		return
	}
	doorway := conf.DoorWay{}
	doorway.NextX = float32(utils.RandInt64(70, 80))
	doorway.NextY = float32(utils.RandInt64(70, 80))
	doorway.NextSceneID = conf.HePingShiJieID

	this.HuiChengPlayer.Set(player, &doorway)
}

func (this *Scene) DoHuiCheng() {
	if this.ChangeScene == nil {
		return
	}
	allhuicheng := this.HuiChengPlayer.Items()
	for k, v := range allhuicheng {
		this.PlayerWillLeave(k.(*Player))
		this.ChangeScene.PlayerChangeScene(k.(*Player), *v.(*conf.DoorWay))
		this.HuiChengPlayer.Delete(k)
	}
}

//传送门检查
func (this *Scene) DoDoorWay() {
	if this.ChangeScene == nil {
		return
	}

	for _, v := range this.DoorWays {
		pos := vec2d.Vec2{X: v.X, Y: v.Y}
		for _, player := range this.Players {
			if player != nil && player.MainUnit != nil &&
				player.MainUnit.Body != nil {

				mypos := player.MainUnit.Body.Position
				subpos := vec2d.Sub(mypos, pos)
				distanse := subpos.Length()
				if distanse <= v.R {
					//传送到其他场景
					log.Info("chuan song :%v", v)
					if player.MainUnit.Level >= v.NeedLevel {
						this.PlayerWillLeave(player)
						this.ChangeScene.PlayerChangeScene(player, v)
					} else {
						player.SendNoticeWordToClient(9, strconv.Itoa(int(v.NeedLevel)))
					}

				}
			}
		}
	}
}

//
func (this *Scene) EveryTimeDo(dt float32) {

	this.EveryTimeDoRemainTime -= float32(dt)
	if this.EveryTimeDoRemainTime <= 0 {
		//do
		this.EveryTimeDoRemainTime += 1
		this.DoReCreateUnit()
		this.DoDoorWay()

		this.CheckCreateBoss()

		//道具
		this.UpdateSceneItem(dt)
		this.DoRemoveSceneItem()

	}
}

//发送消息给全服
func (this *Scene) SendNoticeWordToQuanFuPlayer(typeid int32, param ...string) {
	msg := &protomsg.SC_NoticeWords{}
	msg.TypeID = typeid
	msg.P = param
	if this.ChangeScene != nil {
		this.ChangeScene.SendMsg2QuanFu("SC_NoticeWords", msg)
	}
}

//发送消息给当前场景的所有玩家
func (this *Scene) SendNoticeWordToAllPlayer(typeid int32, param ...string) {
	//遍历所有玩家
	for _, player := range this.Players {
		v := player.MainUnit
		if v == nil {
			continue
		}
		player.SendNoticeWordToClientP(typeid, param)
	}
}

//检查游戏开始多少秒后没有玩家就关闭
func (this *Scene) CheckStartCloseTime() {
	if this.NoPlayerCloseTime == -1 {
		return
	}
	if float32(this.NoPlayerCloseTime) <= float32(this.CurFrame)*(1.0/float32(this.SceneFrame)) {
		if len(this.Players) <= 0 && this.NextAddPlayer.Size() <= 0 {
			this.Close()
		}
	}
}

func (this *Scene) Update() {

	log.Info("Update start %d  %d", this.TypeID, this.DataFileID)
	t1 := utils.GetCurTimeOfSecond()
	log.Info("t1:%f", (t1))
	for {
		//log.Info("Update loop")
		//t1 := time.Now().UnixNano()
		//log.Info("main time:%d", (t1)/1e6)

		time1 := utils.GetCurTimeOfSecond()
		this.EveryTimeDo(1 / float32(this.SceneFrame))

		//夺宝奇兵
		if this.DuoBaoQiBing != nil {
			this.DuoBaoQiBing.Update(1 / float32(this.SceneFrame))
		}

		this.DoHuiCheng()

		this.DoRemoveBullet()
		this.DoRemoveHalo()
		this.DoAddAndRemoveUnit()

		time2 := utils.GetCurTimeOfSecond()
		this.DoLogic()
		time3 := utils.GetCurTimeOfSecond()
		this.UpdateHalo(1 / float32(this.SceneFrame))
		//time4 := utils.GetCurTimeOfSecond()
		this.UpdateBullet(1 / float32(this.SceneFrame))
		time5 := utils.GetCurTimeOfSecond()

		this.DoMove()
		time6 := utils.GetCurTimeOfSecond()
		this.DoZone()
		time7 := utils.GetCurTimeOfSecond()
		this.DoSendData()
		time8 := utils.GetCurTimeOfSecond()

		if time8-time1 >= 0.03 {
			log.Info("time:%f %f %f %f %f  ", time2-time1, time3-time2,
				time6-time5, time8-time7, time8-time1)
		}

		this.CurFrame++
		//清除玩家到和平世界
		this.DoCleanPlayer()

		//检查关闭场景
		this.CheckStartCloseTime()
		//处理分区

		if this.Quit {
			this.DoAddAndRemoveUnit()

			this.DoEndException()
			break
		}

		this.DoSleep()

	}
	log.Info("Scene Quit:%d  %d", this.TypeID, this.DataFileID)
	//t2 := time.Now().UnixNano()
	//log.Info("t2:%d   delta:%d    frame:%d", (t2)/1e6, (t2-t1)/1e6, this.CurFrame)

}

//同步数据
func (this *Scene) DoSendData() {

	time1 := utils.GetCurTimeOfSecond()
	//生成单位的 客户端 显示数据
	for _, v := range this.Units {
		v.FreshClientDataSub()
		v.FreshClientData()
	}
	//生成子弹的 客户端 显示数据
	for _, v := range this.Bullets {
		v.FreshClientDataSub()
		v.FreshClientData()
	}

	//生成光环的 客户端 显示数据
	haloitems := this.Halos.Items()
	for _, v := range haloitems {
		v.(*Halo).FreshClientDataSub()
		v.(*Halo).FreshClientData()
	}
	var unitcount = 0
	var bulletcount = 0
	var sceneitemcount = 0
	var halotcount = 0
	time2 := utils.GetCurTimeOfSecond()
	//遍历所有玩家
	for _, player := range this.Players {
		v := player.MainUnit
		if v == nil {
			continue
		}
		myguild := player.MyGuild
		//场景中的同工会玩家
		if myguild != nil {
			for _, player1 := range this.Players {
				otherguild := player1.MyGuild
				otherunit := player1.MainUnit
				if otherguild == nil || otherunit == nil {
					continue
				}
				if myguild.GuildId == otherguild.GuildId {
					player.AddUnitData(otherunit)
				}
			}
		}
		//场景中的同小组玩家
		if player.GroupID > 0 {
			for _, player1 := range this.Players {
				otherunit := player1.MainUnit
				if otherunit == nil {
					continue
				}
				if player.GroupID == player1.GroupID {
					player.AddUnitData(otherunit)
				}
			}
		}

		zones := utils.GetVisibleZones((v.Body.Position.X), (v.Body.Position.Y))
		//遍历可视区域
		for _, vzone := range zones {
			//可视区域的单位
			if _, ok := this.ZoneUnits[vzone]; ok {
				//遍历区域中的单位
				for _, unit := range this.ZoneUnits[vzone] {
					player.AddUnitData(unit)
					unitcount++
				}
			}
			if _, ok := this.ZoneBullets[vzone]; ok {
				//遍历区域中的单位
				for _, bullet := range this.ZoneBullets[vzone] {
					player.AddBulletData(bullet)
					bulletcount++
				}
			}
			//ZoneSceneItems
			if _, ok := this.ZoneSceneItems[vzone]; ok {
				//遍历区域中的单位
				for _, sceneitem := range this.ZoneSceneItems[vzone] {
					player.AddSceneItemData(sceneitem)
					sceneitemcount++
				}
			}
			//
			if _, ok := this.ZoneHalos[vzone]; ok {
				//遍历区域中的单位
				for _, halo := range this.ZoneHalos[vzone] {
					player.AddHaloData(halo)
					halotcount++
				}
			}
		}
		//player.Update(this.CurFrame)
	}
	time3 := utils.GetCurTimeOfSecond()
	//遍历所有玩家
	for _, player := range this.Players {
		v := player.MainUnit
		if v == nil {
			continue
		}
		player.Update(this.CurFrame)
	}

	time4 := utils.GetCurTimeOfSecond()
	if time4-time1 > 0.015 {
		log.Info("DoSendData:t1:%f   t2:%f  t3:%f ", time4-time3, time3-time2, time2-time1)
	}

}

//处理移动
func (this *Scene) DoMove() {
	this.MoveCore.Update(1 / float64(this.SceneFrame))
}

//处理分区
func (this *Scene) DoZone() {
	//单位分区
	this.ZoneUnits = make(map[utils.SceneZone][]*Unit)
	for _, v := range this.Units {

		zone := utils.GetSceneZone((v.Body.Position.X), (v.Body.Position.Y))
		this.ZoneUnits[zone] = append(this.ZoneUnits[zone], v)

	}
	//子弹分区
	this.ZoneBullets = make(map[utils.SceneZone][]*Bullet)
	for _, v := range this.Bullets {

		zone := utils.GetSceneZone((v.Position.X), (v.Position.Y))
		this.ZoneBullets[zone] = append(this.ZoneBullets[zone], v)
	}
	//道具分区
	this.ZoneSceneItems = make(map[utils.SceneZone][]*SceneItem)
	for _, v := range this.SceneItems {

		zone := utils.GetSceneZone((v.Position.X), (v.Position.Y))
		this.ZoneSceneItems[zone] = append(this.ZoneSceneItems[zone], v)
	}
	//光环分区
	this.ZoneHalos = make(map[utils.SceneZone][]*Halo)
	haloitems := this.Halos.Items()
	for _, v := range haloitems {

		zone := utils.GetSceneZone((v.(*Halo).Position.X), (v.(*Halo).Position.Y))
		this.ZoneHalos[zone] = append(this.ZoneHalos[zone], v.(*Halo))
	}
}

//处理单位逻辑
func (this *Scene) DoLogic() {
	this.NoramlUnitCount = 0
	//处理单位逻辑
	for _, v := range this.Units {
		v.PreUpdate(1 / float64(this.SceneFrame))
		if v.UnitType == 2 { //普通单位
			this.NoramlUnitCount++
		}
	}
	for _, v := range this.Units {
		v.Update(1 / float64(this.SceneFrame))
	}
}

//处理sleep
func (this *Scene) DoSleep() {
	sencond := time.Second
	onetime := int64(1 / float64(this.SceneFrame) * float64(sencond))
	t1 := time.Now().UnixNano()

	nexttime := this.FirstUpdateTime + onetime*int64(this.CurFrame)

	sleeptime := nexttime - t1

	//log.Info("main time:%d    %d", (t1-this.LastUpdateTime)/1e6, onetime/1e6)
	//	if this.TypeID == 1 {
	//		log.Info("sleep :%d   ", sleeptime/1e6)
	//	}

	if sleeptime > 0 {
		time.Sleep(time.Duration(sleeptime))
	} else {

	}

}

//增加光环
func (this *Scene) AddHalo(halo *Halo) {
	if halo == nil {
		return
	}
	//this.Halos[halo.ID] = halo
	this.Halos.Set(halo.ID, halo)

	if halo.KilledInvalid == 1 {
		this.CanRemoveHalos[halo.ID] = halo
	}

	//log.Info("------AddHalo----%d", halo.TypeID)
}

//删除光环
func (this *Scene) RemoveHalo(id int32) {
	//delete(this.Halos, id)
	this.Halos.Delete(id)
	delete(this.CanRemoveHalos, id)
}

//获取光环
func (this *Scene) ForbiddenHalo(id int32, isForbidden bool) {
	halo := this.Halos.Get(id)
	if halo != nil {
		halo.(*Halo).IsForbidden = isForbidden
	}
}

//删除击杀单位后无效光环
func (this *Scene) RemoveHaloForKilled(parent *Unit) {
	for k, v := range this.CanRemoveHalos {
		if v.Parent == parent && v.KilledInvalid == 1 {
			delete(this.CanRemoveHalos, k)
			//delete(this.Halos, k)
			this.Halos.Delete(k)
			return
		}
	}
}

//删除时间结束的光环
func (this *Scene) DoRemoveHalo() {
	//ZoneBullets
	haloitems := this.Halos.Items()
	for k, v := range haloitems {
		if v.(*Halo).IsDone() {
			//log.Info("------DoRemoveHalo----%d", v.TypeID)
			//delete(this.Halos, k)
			this.Halos.Delete(k)
			delete(this.CanRemoveHalos, k.(int32))
		}
	}
}

//更新光环
func (this *Scene) UpdateHalo(dt float32) {
	haloitems := this.Halos.Items()
	for _, v := range haloitems {
		v.(*Halo).Update(dt)
	}
}

//单位死亡后创建道具
func (this Scene) CreateSceneItems(typeid []int32, centerpos vec2d.Vec2, isguilddrop int32, attackunit *utils.BeeMap) {
	//CheckItemCollision
	positions := make([]vec2d.Vec2, 8)
	size := 1.0
	positions[0] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y + size}
	positions[1] = vec2d.Vec2{X: centerpos.X, Y: centerpos.Y + size}
	positions[2] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y + size}

	positions[3] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y}
	positions[4] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y}

	positions[5] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y - size}
	positions[6] = vec2d.Vec2{X: centerpos.X, Y: centerpos.Y - size}
	positions[7] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y - size}

	createitemid := 0
	startindex := rand.Intn(8)
	for i := 0; i < 8; i++ {
		posindex := startindex + i
		if posindex >= 8 {
			posindex = 0
		}
		if this.MoveCore.CheckItemCollision(positions[posindex]) == true {
			this.CreateSceneItem(typeid[createitemid], positions[posindex], isguilddrop, attackunit)
			createitemid++
			if createitemid >= len(typeid) {
				break
			}
		}
	}
}

//创建场景道具
func (this *Scene) CreateSceneItem(typeid int32, pos vec2d.Vec2, isguilddrop int32, attackunit *utils.BeeMap) {
	sceneitem := NewSceneItem(typeid, pos, isguilddrop, attackunit)
	this.AddSceneItem(sceneitem)
}

//增加场景道具
func (this *Scene) AddSceneItem(sceneitem *SceneItem) {
	this.SceneItems[sceneitem.ID] = sceneitem
}

//增加子弹
func (this *Scene) AddBullet(bullet *Bullet) {
	this.Bullets[bullet.ID] = bullet
}

//获取子弹
func (this *Scene) GetBulletByID(id int32) *Bullet {
	re := this.Bullets[id]
	return re
}

//删除子弹
func (this *Scene) DoRemoveBullet() {
	//ZoneBullets
	for k, v := range this.Bullets {
		if v.IsDone() {
			delete(this.Bullets, k)
		}
	}
}

//更新子弹
func (this *Scene) UpdateBullet(dt float32) {
	for _, v := range this.Bullets {
		v.Update(dt)
	}
}

//删除道具
func (this *Scene) DoRemoveSceneItem() {
	//ZoneBullets
	for k, v := range this.SceneItems {
		if v.IsDone() {
			delete(this.SceneItems, k)
		}
	}
}

//更新道具
func (this *Scene) UpdateSceneItem(dt float32) {
	for _, v := range this.SceneItems {
		v.Update()
		if v.IsDone() == false {
			//遍历所有玩家
			for _, player := range this.Players {
				unit := player.MainUnit
				if unit == nil || unit.Body == nil {
					continue
				}
				//				if player.CanSelectSceneItem() == false {
				//					continue
				//				}
				//LengthSquared
				dir := vec2d.Sub(unit.Body.Position, v.Position)
				if dir.LengthSquared() <= 1 {
					if v.IsAuctionItemDrop == 1 && player.MyGuild != nil {
						//拾取到公会拍卖行
						GuildManagerObj.AddAuctionItem(player.MyGuild.GuildId, v.TypeID, 1, v.AttackUnits)
						v.BeSelect()
						break
					} else if v.IsAuctionItemDrop == 2 { //世界拍卖行
						AuctionManagerObj.NewAuctionItem2World(player, -1, v.TypeID, 1, v.AttackUnits)
						v.BeSelect()
						break
					} else {
						//拾取到背包
						if player.SelectSceneItem(v) == true {
							v.BeSelect()
							if this.DuoBaoQiBing != nil {
								this.DuoBaoQiBing.CheckPlayerGetBox(player, v.TypeID)
							}

							break
						}
					}

				}
			}
		}
	}
}

//增加单位
func (this *Scene) DoAddAndRemoveUnit() {
	//增加
	itemsadd := this.NextAddUnit.Items()
	for k, v := range itemsadd {

		player := v.(*Unit).MyPlayer
		if player != nil {
			if player.IsLoadedSceneSucc == false {
				continue
			}
		}

		log.Info("----add player!!!!")

		if v.(*Unit).Body != nil {
			this.MoveCore.RemoveBody(v.(*Unit).Body)
			v.(*Unit).Body = nil
		}

		//设置移动核心body
		pos := vec2d.Vec2{0, 0}
		r := vec2d.Vec2{v.(*Unit).CollisionR, v.(*Unit).CollisionR}
		v.(*Unit).Body = this.MoveCore.CreateBody(pos, r, 0, 1)
		v.(*Unit).Body.Tag = int(v.(*Unit).ID)
		//出现位置小于0 则随机一个位置
		if v.(*Unit).InitPosition.X <= 0 || v.(*Unit).InitPosition.Y <= 0 {
			x := utils.GetRandomFloatTwoNum(this.StartX, this.EndX)
			y := utils.GetRandomFloatTwoNum(this.StartY, this.EndY)
			v.(*Unit).Body.BlinkToPos(vec2d.Vec2{X: float64(x), Y: float64(y)}, 0)
		} else {
			v.(*Unit).Body.BlinkToPos(v.(*Unit).InitPosition, 0)
		}

		this.Units[k.(int32)] = v.(*Unit)

		//添加场景BUFF 英雄且非幻象
		if v.(*Unit).UnitType == 1 && v.(*Unit).IsMirrorImage != 1 {
			v.(*Unit).AddBuffFromStr(this.SceneBuff, 1, v.(*Unit))
		}

		this.NextAddUnit.Delete(k)

		//this.Players[]
	}
	//this.NextAddUnit.DeleteAll()

	//删除
	itemsremove := this.NextRemoveUnit.Items()
	for k, v := range itemsremove {
		this.MoveCore.RemoveBody(v.(*Unit).Body)
		//v.(*Unit).Body
		//this.Units.Delete(k)
		v.(*Unit).OnDestroy()
		delete(this.Units, k.(int32))

		this.NextRemoveUnit.Delete(k)
	}
	//this.NextRemoveUnit.DeleteAll()

	//增加玩家
	playeradd := this.NextAddPlayer.Items()
	for k, v := range playeradd {

		player := v.(*Player)
		if player != nil {
			if player.IsLoadedSceneSucc == false || player.MainUnit == nil || player.MainUnit.Body == nil {
				continue
			}
		}
		this.playerlock.Lock()
		this.Players[k.(int32)] = v.(*Player)
		this.playerlock.Unlock()

		//初始化击杀数据
		this.InitPlayerKill(v.(*Player))

		this.NextAddPlayer.Delete(k)
	}
	//this.NextAddPlayer.DeleteAll()

	//删除玩家
	playerremove := this.NextRemovePlayer.Items()
	for k, _ := range playerremove {
		this.playerlock.Lock()
		delete(this.Players, k.(int32))
		this.playerlock.Unlock()

		this.NextRemovePlayer.Delete(k)
	}
	//this.NextRemovePlayer.DeleteAll()
}

//安全锁的情况下获取所有玩家
func (this *Scene) GetAllPlayerUseLock() map[int32]*Player {
	this.playerlock.RLock()
	r := make(map[int32]*Player)
	for k, v := range this.Players {
		r[k] = v
	}
	this.playerlock.RUnlock()
	return r
}

//场景结束  把玩家转移到 和平地图
func (this *Scene) SetCleanPlayer(b bool) {
	this.CleanPlayer = b

}

//
func (this *Scene) DoCleanPlayer() {
	//CloseTime 时间到了
	if this.CleanPlayer == true || (this.CloseTime != -1 && float32(this.CloseTime) <= float32(this.CurFrame)*(1.0/float32(this.SceneFrame))) {
		this.Close()
		if this.DuoBaoQiBing != nil {
			this.DuoBaoQiBing.DoOver()
		}
		this.DoEndExceptionBeforePlayerLeave()
		//当前玩家回城
		for _, player := range this.Players {
			this.HuiChengHePingShiJie(player)
		}

		this.DoHuiCheng()
	}
	//

	//

}

func (this *Scene) Close() {
	this.Quit = true
}

//设置小组ID

//玩家进入
func (this *Scene) PlayerGoin(player *Player, characterinfo *db.DB_CharacterInfo) bool {
	//if player.MainUnit == nil {
	if this.CleanPlayer == true || this.Quit == true {
		return false
	}
	if this.ForceAttackMode != 0 {
		characterinfo.AttackMode = this.ForceAttackMode
	}

	//设置角色的小组ID
	groupid := this.CharacterGroups.Get(player.Characterid)
	if groupid != nil {
		player.GroupID = groupid.(int32)
	} else {
		player.GroupID = 0
	}

	player.MainUnit = CreateUnitByPlayer(this, player, characterinfo)
	//player.Characterid =
	//}

	this.NextAddUnit.Set(player.MainUnit.ID, player.MainUnit)

	this.NextAddPlayer.Set(player.Uid, player)
	return true

}

//处理重新创建NPC
func (this *Scene) DoReCreateUnit() {
	curtime := utils.GetCurTimeOfSecond()
	for k, v := range this.ReCreateUnitInfo {

		if curtime-v.DeathTime >= v.ReCreateInfo.ReCreateTime {
			this.CreateUnitByConf(*(v.ReCreateInfo))
			delete(this.ReCreateUnitInfo, k)
		}
	}
}

func (this *Scene) RemoveUnit(unit *Unit) {
	if unit == nil {
		return

	}
	//删除主单位
	this.NextRemoveUnit.Set(unit.ID, unit)
	if unit.ReCreateInfo != nil {
		rc := &ReCreateUnit{}
		rc.ReCreateInfo = unit.ReCreateInfo
		rc.DeathTime = utils.GetCurTimeOfSecond()
		this.ReCreateUnitInfo[rc] = rc
		//this.ReCreateUnitInfo = append(this.ReCreateUnitInfo)
	}

}

//玩家退出
func (this *Scene) PlayerGoout(player *Player) {
	//删除主单位
	//this.NextRemoveUnit.Set(player.MainUnit.ID, player.MainUnit)

	//掉落夺宝奇兵的宝藏

	this.RemoveUnit(player.MainUnit)
	if player.MainUnit != nil {
		this.NextAddUnit.Delete(player.MainUnit.ID)
	}

	items := player.OtherUnit.Items()
	for _, v := range items {
		this.RemoveUnit(v.(*Unit))
		this.NextAddUnit.Delete(v.(*Unit).ID)
	}

	this.NextRemovePlayer.Set(player.Uid, player)
	this.NextAddPlayer.Delete(player.Uid)

}
