package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/vec2d"
	"strconv"
)

//夺宝奇兵信息
type SceneDuoBaoInfo struct {
	OpenBoxTime       float32 //打开宝箱需要的持有时间
	OpenBoxRemainTime float32 //打开宝箱剩余时间
	BoxTypeId         int32   //宝箱道具ID
	BoxParent         *Player //宝箱的寄宿者(就是宝箱在哪个玩家身上)
	Parent            *Scene
}

func CreateSceneDuoBaoInfo(time float32, boxtypeid int32, scene *Scene) *SceneDuoBaoInfo {
	re := &SceneDuoBaoInfo{}
	re.OpenBoxTime = time
	re.OpenBoxRemainTime = time
	re.BoxParent = nil
	re.Parent = scene
	re.BoxTypeId = boxtypeid
	return re
}

//处理玩家掉落宝箱
func (this *SceneDuoBaoInfo) DoPlayerLostBox(player *Player) {
	if player == this.BoxParent {
		player.RemoveBagItem(this.BoxTypeId)
		//掉落
		drop := make([]int32, 0)
		drop = append(drop, this.BoxTypeId)
		pos := vec2d.Vec2{X: 50, Y: 50}
		mainunit := player.MainUnit
		if mainunit != nil && mainunit.Body != nil {
			pos = mainunit.Body.Position
		}
		this.Parent.CreateSceneItems(drop, pos, 0, nil)

		this.PlayerLostBox()
	}
}

//玩家丢失了宝箱
func (this *SceneDuoBaoInfo) PlayerLostBox() {
	this.BoxParent = nil
	this.OpenBoxRemainTime = this.OpenBoxTime
}

//检查玩家是否拾取了宝箱
func (this *SceneDuoBaoInfo) CheckPlayerGetBox(player *Player, typeid int32) {
	if player == nil {
		return
	}
	if typeid == this.BoxTypeId {
		this.PlayerGetBox(player)
	}
}

//玩家获得了宝箱
func (this *SceneDuoBaoInfo) PlayerGetBox(player *Player) {
	this.BoxParent = player
	this.OpenBoxRemainTime = this.OpenBoxTime
}

//结束(当前玩家获得宝箱中的东西)
func (this *SceneDuoBaoInfo) DoOver() {
	if this.BoxParent == nil {
		return
	}
	log.Info("-SceneDuoBaoInfo-DoOver-")
	//打开宝箱
	typeid := this.BoxParent.UseBagItemFromTypeid(this.BoxTypeId)
	if typeid > 0 {
		unit1 := this.BoxParent.MainUnit
		if unit1 != nil {
			itemdata := conf.GetItemData(typeid)
			if itemdata == nil {
				return
			}
			//this.SendNoticeWordToClient(25, itemdata.ItemName, "lv."+strconv.Itoa(int(level)))
			this.Parent.SendNoticeWordToQuanFuPlayer(42, unit1.Name, itemdata.ItemName, "lv.1")
		}
	}
	this.PlayerLostBox()
}

func (this *SceneDuoBaoInfo) Update(dt float32) {
	if this.BoxParent == nil {
		return
	}
	lastremaintime := this.OpenBoxRemainTime
	this.OpenBoxRemainTime -= (dt)
	if this.OpenBoxRemainTime <= 0 {
		this.DoOver()
		return
	}

	//每隔60秒提示
	if int32(lastremaintime)/60 != int32(this.OpenBoxRemainTime)/60 {
		unit1 := this.BoxParent.MainUnit
		if unit1 != nil && unit1.Body != nil {
			this.Parent.SendNoticeWordToAllPlayer(41, unit1.Name+"("+strconv.Itoa(int(unit1.Body.Position.X))+","+strconv.Itoa(int(unit1.Body.Position.Y))+")")
		}
	}

	unit := this.BoxParent.MainUnit
	if unit != nil && unit.IsDisappear() {
		this.DoPlayerLostBox(this.BoxParent)
		return
	}
}
