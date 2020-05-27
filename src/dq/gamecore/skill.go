package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"strconv"
)

type Skill struct {
	conf.SkillData //技能数据

	Level            int32   //技能当前等级
	RemainCDTime     float32 //技能cd 剩余时间
	AttackAutoActive int32   //攻击时自动释放 是否激活 1:激活 2:否

	RemainVisibleTime float32 //剩余显示时间
	RemainSkillCount  int32   //剩余点数

	RemainTriggerTime float32 //剩余触发的时间

	Parent *Unit //载体

	Param1 int32 //参数1

}

//激活与不激活
func (this *Skill) DoActive() {
	if this.AttackAutoActive == 1 {
		this.AttackAutoActive = 2
	} else {
		this.AttackAutoActive = 1
	}
}

//获取魔法消耗
func (this *Skill) GetManaCost() int32 {
	if this.OtherManaCostType == 0 {
		return this.ManaCost
	} else if this.OtherManaCostType == 1 { //最大魔法百分比
		if this.Parent == nil {
			return this.ManaCost
		}
		v := this.ManaCost + int32(float32(this.Parent.MAX_MP)*this.OtherManaCostVal)
		return v
	}
	return this.ManaCost
}

//设置显示
func (this *Skill) SetVisible(visible int32) {
	this.Visible = visible
	if visible == 1 {
		this.RemainVisibleTime = this.VisibleTime
	}
}

//设置子弹属性
func (this *Skill) SetBulletProperty(b *Bullet, unit *Unit) {
	if b == nil {
		return
	}

	b.SetNormalHurtRatio(this.NormalHurt)
	b.SetProjectileMode(this.BulletModeType, this.BulletSpeed)
	//技能增强
	if this.HurtType == 2 {
		hurtvalue := (this.HurtValue + int32(float32(this.HurtValue)*unit.MagicScale))
		b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: hurtvalue})
	} else {
		b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: this.HurtValue})
	}
	b.MagicValueHurt2PhisicHurtCR = this.MagicValueHurt2PhisicHurtCR
	b.AddTargetBuff(this.TargetBuff, this.Level)
	b.AddTargetHalo(this.TargetHalo, this.Level)
	b.SkillID = this.TypeID
	b.SkillLevel = this.Level
	b.Exception = this.Exception
	b.ExceptionParam = this.ExceptionParam
	b.EveryDoHurtChangeHurtCR = this.EveryDoHurtChangeHurtCR
	b.UseUnitProjectilePos = this.UseUnitProjectilePos
	//召唤信息
	b.BulletCallUnitInfo = BulletCallUnitInfo{this.CallUnitInfo, this.Level}
	if this.AwaysHurt == 1 {
		b.IsDoHurtOnMove = 1
	}
	//弹射
	b.SetEjection(this.EjectionCount, this.EjectionRange, this.EjectionDecay, this.EjectionRepeat)
	//伤害范围 和目标关系
	b.SetRange(this.HurtRange)
	b.UnitTargetTeam = this.UnitTargetTeam
	//强制移动
	//if this.ForceMoveType == 1 {
	b.SetForceMove(this.ForceMoveTime, this.ForceMoveSpeedSize, this.ForceMoveLevel, this.ForceMoveType, this.ForceMoveBuff)
	//}
	b.SwitchedPlaces = this.SwitchedPlaces
	b.DestForceAttackSrc = this.DestForceAttackSrc
	b.FreshSkillTime = this.FreshSkillTime
	//加血
	if this.AddHPTarget == 2 {
		b.SetAddHP(this.AddHPType, this.AddHPValue)
	}
	b.AddMPValue += this.AddMPValue
	b.PhysicalHurtAddHP += this.PhysicalHurtAddHP
	b.MagicHurtAddHP += this.MagicHurtAddHP

	b.SetPathHalo(this.PathHalo, this.PathHaloMinTime)

	b.ClearLevel = this.TargetClearLevel //设置驱散等级

}

//创建子弹
func (this *Skill) CreateBullet(unit *Unit, data *protomsg.CS_PlayerSkill) []*Bullet {
	var bullets = make([]*Bullet, 0)
	if unit == nil || data == nil {
		return bullets
	}
	//
	//自身
	var b *Bullet = nil
	if this.CastTargetType == 1 {

		b = NewBullet1(unit, unit)

	} else if this.CastTargetType == 2 { //目标单位

		targetunit := unit.InScene.FindUnitByID(data.TargetUnitID)
		b = NewBullet1(unit, targetunit)

	} else if this.CastTargetType == 3 || this.CastTargetType == 5 { //目的点
		b = NewBullet2(unit, vec2d.Vec2{float64(data.X), float64(data.Y)})
	}
	//施法目标范围
	if this.CastTargetRange > 0 {
		allunit := unit.InScene.FindVisibleUnitsByPos(vec2d.Vec2{b.DestPos.X, b.DestPos.Y})
		for _, v := range allunit {
			if v.IsDisappear() {
				continue
			}
			//			if this.UnitTargetTeam == 1 && unit.CheckIsEnemy(v) == true {
			//				continue
			//			}
			//			if this.UnitTargetTeam == 2 && unit.CheckIsEnemy(v) == false {
			//				continue
			//			}

			if unit.CheckUnitTargetTeam(v, this.UnitTargetTeam) == false {
				continue
			}

			if v.CheckUnitTargetCamp(this.UnitTargetCamp) == false {
				continue
			}

			//检测是否在范围内
			if v.Body == nil || this.CastTargetRange <= 0 {
				continue
			}
			dis := float32(vec2d.Distanse(unit.Body.Position, v.Body.Position))
			//log.Info("-----------------dis:%f", dis)
			if dis <= this.CastTargetRange {
				b = NewBullet1(unit, v)
				this.SetBulletProperty(b, unit)
				bullets = append(bullets, b)
			}
		}
	} else {

		if this.BulletCount > 1 && this.CastTargetType == 1 {

			for i := int32(0); i < this.BulletCount; i++ {
				dir := vec2d.Vec2{float64(0), float64(1)}
				dir.Normalize()
				dir.MulToFloat64(float64(this.CastRange))
				dir.Rotate(float64(int32(360) / this.BulletCount * i))
				dir.Collect(&unit.Body.Position)
				b = NewBullet2(unit, vec2d.Vec2{float64(dir.X), float64(dir.Y)})
				this.SetBulletProperty(b, unit)
				bullets = append(bullets, b)
			}

		} else {
			this.SetBulletProperty(b, unit)
			bullets = append(bullets, b)
		}

	}
	return bullets
}

func (this *Skill) UpdateException() {
	if this.Exception == 0 {
		return
	}
	switch this.Exception {
	//	case 5: //帕克幻象发球
	//		{
	//			if this.Parent == nil || this.Parent.IsDisappear() {
	//				return
	//			}
	//			bullet := this.Parent.InScene.GetBulletByID(this.Param1)
	//			if bullet == nil {

	//				//检查关联
	//				if this.VisibleRelationSkillID > 0 {
	//					skilldata1, ok1 := this.Parent.Skills[this.VisibleRelationSkillID]
	//					if ok1 {
	//						this.Visible = 2
	//						skilldata1.Visible = 1
	//					}
	//				}
	//			}

	//		}
	default:
	}
}

//
func (this *Skill) AddSkillCount() {
	this.RemainCDTime = this.Cooldown - this.Parent.MagicCD*this.Cooldown
	this.RemainSkillCount++
	if this.RemainSkillCount > this.SkillCount {
		this.RemainSkillCount = this.SkillCount
	}
}

//

//更新
func (this *Skill) Update(dt float64) {
	//CD时间减少
	if this.RemainSkillCount < this.SkillCount {
		this.RemainCDTime -= float32(dt)
		if this.RemainCDTime <= 0 {
			this.RemainCDTime = 0
			this.AddSkillCount()
		}
	}
	//
	if this.RemainSkillCount > 0 {
		//4:每秒钟触发(龙心buff)
		if this.CastType == 2 && this.TriggerTime == 4 {

			this.RemainTriggerTime -= float32(dt)
			if this.RemainTriggerTime <= 0 {
				this.RemainTriggerTime += 1.0
				//skilldata.MyBuff, skilldata.Level, this
				this.Parent.AddBuffFromStr(this.MyBuff, this.Level, this.Parent)
			}
		}
	}

	if this.Visible == 1 {
		if this.RemainVisibleTime > 0 {
			this.RemainVisibleTime -= float32(dt)
			if this.RemainVisibleTime <= 0 {
				//检查关联
				if this.VisibleRelationSkillID > 0 {
					skilldata1, ok1 := this.Parent.Skills[this.VisibleRelationSkillID]
					if ok1 {
						this.SetVisible(2)
						skilldata1.SetVisible(1)
					}
				}
			}
		}

	}
	this.UpdateException()
}

//检查cd 返回true表示可以使用
func (this *Skill) CheckCDTime() bool {
	if this.RemainSkillCount > 0 {
		return true
	}
	return false
}

//同步cd 把本技能的CD同步为 参数技能的CD
func (this *Skill) SameCD(skill *Skill) {
	if skill == nil {
		return
	}

	this.RemainCDTime = skill.RemainCDTime
	this.RemainSkillCount = skill.RemainSkillCount

}

//重置CD时间
func (this *Skill) ResetCDTime(time float32) {
	this.RemainCDTime = time
	if time > 0 {
		this.RemainSkillCount = 0
	}
}

//被攻击打断技能CD
func (this *Skill) DoBeHurt() {

	//log.Info("---------%d----%f   %f", this.TypeID, this.BeHurtStopTime, this.RemainCDTime)
	if this.BeHurtStopTime > 0 {

		if this.BeHurtStopTime > this.RemainCDTime || this.RemainSkillCount >= 1 {
			this.RemainCDTime = this.BeHurtStopTime
			this.RemainSkillCount = 0
		}

		//log.Info("-------------%f", this.BeHurtStopTime)
	}
}

//使技能重新可用 (刷新球功能)
func (this *Skill) FreshSkill() {
	this.RemainCDTime = 0
	this.RemainSkillCount = this.SkillCount
}

//使用技能后刷新CD
func (this *Skill) FreshCDTime(time float32) {

	if this.RemainSkillCount == this.SkillCount {
		this.RemainCDTime = time
	}
	this.RemainSkillCount--

}

//返回数据库字符串
func (this *Skill) ToDBString() string {
	return strconv.Itoa(int(this.TypeID)) + "," + strconv.Itoa(int(this.Level)) + "," + strconv.FormatFloat(float64(this.RemainCDTime), 'f', 4, 32) + "," + strconv.Itoa(int(this.RemainSkillCount))
}

func NewOneSkill(skillid int32, skilllevel int32, unit *Unit) *Skill {
	sk := &Skill{}
	skdata := conf.GetSkillData(skillid, skilllevel)
	if skdata == nil {
		log.Error("NewUnitSkills %d  %d", skillid, skilllevel)
		return nil
	}
	sk.SkillData = *skdata
	sk.Level = skilllevel
	sk.RemainCDTime = 0
	sk.AttackAutoActive = 1
	sk.Parent = unit
	sk.RemainSkillCount = sk.SkillCount

	return sk
}

//通过数据库数据和单位基本数据创建技能 (1,2,0) ID,LEVEL,CD剩余时间 剩余点数
func NewUnitSkills(dbdata []string, unitskilldata string, unit *Unit) map[int32]*Skill {
	re := make(map[int32]*Skill)

	//单位基本技能
	skillids := utils.GetInt32FromString2(unitskilldata)
	for _, v := range skillids {
		sk := &Skill{}
		skdata := conf.GetSkillData(v, 1)
		if skdata == nil {
			log.Error("NewUnitSkills %d  %d", v, 1)
			continue
		}
		sk.SkillData = *skdata
		//sk.SkillData.Index = int32(k)

		//log.Info("skill index:%d", sk.SkillData.Index)
		sk.Level = sk.InitLevel
		sk.RemainCDTime = 0
		sk.Parent = unit
		sk.RemainSkillCount = sk.SkillCount
		re[sk.TypeID] = sk
	}
	//数据库技能
	for _, v := range dbdata {

		oneskilldbdata := utils.GetFloat32FromString2(v)
		if len(oneskilldbdata) < 2 {
			continue
		}
		skillid := int32(oneskilldbdata[0])
		skilllevel := int32(oneskilldbdata[1])
		skillcd := float32(0)
		if len(oneskilldbdata) >= 3 {
			skillcd = oneskilldbdata[2]
		}

		sk := &Skill{}
		skdata := conf.GetSkillData(skillid, skilllevel)
		if skdata == nil {
			log.Error("NewUnitSkills %d  %d", skillid, skilllevel)
			continue

		}
		sk.SkillData = *skdata
		sk.Level = skilllevel
		sk.RemainCDTime = skillcd
		sk.AttackAutoActive = 1
		sk.RemainSkillCount = sk.SkillCount
		if len(oneskilldbdata) >= 4 {
			sk.RemainSkillCount = int32(oneskilldbdata[3])
		}

		//sk.RemainCDTime = 10.0
		if initskill, ok := re[sk.TypeID]; ok {
			sk.Index = initskill.Index
			sk.Parent = unit
			re[sk.TypeID] = sk
		}

	}

	return re
}
