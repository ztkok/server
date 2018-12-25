package main

import (
	"common"
	"excel"
	"strconv"
	"strings"
	"zeus/dbservice"
	"zeus/linmath"
)

// GmMgr Gm命令
type GmMgr struct {
	user *RoomUser

	isValidate bool
	// cmds 命令集合
	cmds map[string](func(map[string]string))
}

// NewGmMgr 获取Gm管理器
func NewGmMgr(user *RoomUser) *GmMgr {
	gm := &GmMgr{
		user:       user,
		isValidate: true,
		cmds:       make(map[string](func(map[string]string))),
	}
	gm.init()

	return gm
}

// init 初始化管理器
func (gm *GmMgr) init() {
	gm.cmds["AddHp"] = gm.AddHp                     // 添加玩家血量
	gm.cmds["SetSpeedRate"] = gm.SetSpeedRate       // 设置人物速度
	gm.cmds["FlushProp"] = gm.FlushProp             // 刷新道具
	gm.cmds["TransToNpc"] = gm.TransToNpc           // 随机传送至Npc附近
	gm.cmds["AttackLogSwitch"] = gm.AttackLogSwitch // 攻击日志是否显示开关
	gm.cmds["Goto"] = gm.Goto                       // 移动到坐标
	gm.cmds["Validate"] = gm.SetValidate            // 设置验证开关
	gm.cmds["SetDoorState"] = gm.SetDoorState       // 设置门的状态
	gm.cmds["DropBox"] = gm.DropBox                 // 空投箱
}

// exec 执行命令
func (gm *GmMgr) exec(paras string) {
	grade, err := dbservice.Account(gm.user.GetDBID()).GetGrade()
	if err != nil {
		gm.user.Error("exec failed, GetGrade err: ", err)
		return
	}
	if grade != 0 {
		gm.user.Warn("exec failed, grade is not ok")
		return
	}

	pairSet := strings.Split(paras, " ")

	pairMap := make(map[string]string)
	for _, pair := range pairSet {
		paraSet := strings.Split(pair, "=")
		if len(paraSet) != 2 {
			continue
		}

		pairMap[paraSet[0]] = paraSet[1]
	}

	if cmdStr, ok := pairMap["rcmd"]; ok {
		if cmd, ok := gm.cmds[cmdStr]; ok {
			cmd(pairMap)
		} else {
			gm.user.Warn("cmds don't contain ", cmdStr)
		}
	} else {
		gm.user.Warn("client params contains no cmds")
	}

	gm.user.Info("exec gm cmds, params: ", paras)
}

// AddHp 添加血量
func (gm *GmMgr) AddHp(paras map[string]string) {
	gm.user.Info("AddHp")

	if strValue, ok := paras["value"]; ok {
		value, err := strconv.Atoi(strValue)
		if err == nil {
			if value >= 0 {
				if gm.user.stateMgr.GetState() != RoomPlayerBaseState_WillDie {
					gm.user.SetHP(gm.user.GetHP() + uint32(value))
				}
			} else {
				gm.user.DisposeSubHp(InjuredInfo{num: uint32(-value), injuredType: losthp, isHeadshot: false})
			}
			return
		}
	}
	gm.user.Info("AddHp failed, params: ", paras, " State(10被击倒):", gm.user.stateMgr.GetState())
}

// SetValidate 服务器验证开关
func (gm *GmMgr) SetValidate(paras map[string]string) {
	gm.user.Info("SetValidate")

	if v, ok := paras["validate"]; ok {
		if v == "0" {
			gm.isValidate = false
		} else {
			gm.isValidate = true
		}
	}

	gm.user.Info("SetValidate isValidate: ", gm.isValidate)
}

// SetSpeedRate 设置速度速率
func (gm *GmMgr) SetSpeedRate(paras map[string]string) {
	gm.user.Info("SetSpeedRate")

	if strValue, ok := paras["rate"]; ok {
		value, err := strconv.ParseFloat(strValue, 32)
		if err == nil {
			gm.user.SetSpeedRate(float32(value))
			return
		}
	}

	gm.user.Info("SetSpeedRate failed, params: ", paras)
}

// FlushProp 刷新道具
func (gm *GmMgr) FlushProp(paras map[string]string) {
	gm.user.Info("FlushProp")

	var itemid int64
	var err error
	if idStr, ok := paras["id"]; ok {
		itemid, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return
		}
	} else {
		return
	}

	num := common.StringToUint32(paras["num"])

	pos, err := gm.user.getRandPoint(gm.user.GetPos(), 3)
	if err != nil {
		return
	}

	space := gm.user.GetSpace().(*Scene)

	base, ok := excel.GetItem(uint64(itemid))
	if !ok {
		gm.user.Warn("Item id not exist, id: ", itemid)
		return
	}

	if base.Type == ItemTypeCar {
		entityid := space.GetEntityTempID()
		sceneitem := &SceneItem{
			id:       entityid,
			pos:      gm.user.GetPos(),
			itemid:   uint32(itemid),
			item:     NewItem(uint32(itemid), 1),
			haveobjs: GetRefreshItemMgr(space).dropVehicleWeapons(uint64(itemid)),
		}

		waterLevel := space.mapdata.Water_height
		posy, err := space.GetHeight(pos.X, pos.Z)
		if err != nil || posy <= waterLevel {
			sceneitem.pos.Y = waterLevel + 1
		}

		space.mapitem[sceneitem.id] = sceneitem
		space.AddEntity("Vehicle", entityid, 0, "", false, true)
		gm.user.Info("AddEntity success, vehicleid: ", entityid)

	} else {
		waterLevel := space.mapdata.Water_height
		posy, err := space.GetHeight(pos.X, pos.Z)
		if err != nil || posy <= waterLevel {
			pos.Y = waterLevel
		}

		if num == 0 || num > 50 {
			num = 1
		}

		for i := uint32(0); i < num; i++ {
			space.refreshitem.dropItemByID(uint32(itemid), pos)
		}
	}

	gm.user.Info("Flush Prop success, params: ", paras, " pos: ", pos)
}

// TransToNpc 传送至一个npc旁
func (gm *GmMgr) TransToNpc(paras map[string]string) {
	gm.user.Info("TransToNpc")

	space := gm.user.GetSpace().(*Scene)

	aiSum := 0
	userSum := 0
	liveSum := 0
	for id, isLive := range space.members {
		if isLive == false {
			continue
		}
		liveSum++

		ai, ok := space.GetEntity(id).(*RoomAI)
		if ok {
			aiSum++
			pos, err := gm.user.getRandPoint(ai.GetPos(), 5)
			if err == nil {
				gm.user.SetPos(pos)
				gm.user.Info("TransToNpc success, pos: ", pos)
				break
			} else {
				gm.user.Error("getRandPoint err: ", err, " ai.GetPos(): ", ai.GetPos(), " RandPos: ", pos)
			}
			continue
		}

		_, ok = space.GetEntity(id).(*RoomUser)
		if ok {
			userSum++
			//gm.user.Info("TransToNpc 玩家信息：", "name:", user.GetName(), " user.GetPos():", user.GetPos())
			continue
		}

		gm.user.Info("TransToNpc id :", id)
	}

	gm.user.Info("TransToNpc ", " len(space.members):", len(space.members),
		" aiSum:", aiSum, " spaceID:", space.GetID(), " userSum:", userSum, " liveSum:", liveSum)
}

// AttackLogSwitch 显示攻击日志开关
func (gm *GmMgr) AttackLogSwitch(paras map[string]string) {
	gm.user.Info("AttackLogSwitch")

	isShowAttackLog = !isShowAttackLog
}

// Goto 移动到坐标
func (gm *GmMgr) Goto(paras map[string]string) {
	gm.user.Info("Goto")

	if strValue, ok := paras["pos"]; ok {
		value := strings.Split(strValue, ",")
		if len(value) == 2 {
			x, xerr := strconv.ParseFloat(value[0], 32)
			z, zerr := strconv.ParseFloat(value[1], 32)
			y, yerr := gm.user.GetSpace().GetHeight(float32(x), float32(z))
			if xerr != nil || yerr != nil || zerr != nil {
				gm.user.Warn("Get coord failed")
				return
			}

			gm.user.SetPos(linmath.NewVector3(float32(x), y, float32(z)))
			gm.user.Infof("Set pos: %+v\n", linmath.NewVector3(float32(x), y, float32(z)))

			space := gm.user.GetSpace().(*Scene)
			if space == nil {
				return
			}

			for _, v := range gm.user.watchers {
				if targetUser, ok := space.GetEntity(v).(IRoomChracter); ok {
					targetUser.SetPos(linmath.NewVector3(float32(x), y, float32(z)))
				}
			}
		}
	}
}

// SetDoorState 设置门的状态
func (gm *GmMgr) SetDoorState(paras map[string]string) {

	var id uint64
	var state uint64
	var err error

	if idValue, ok := paras["id"]; ok {
		id, err = strconv.ParseUint(idValue, 10, 64)
		if err != nil {
			gm.user.Error("SetDoorState failed, ParseUint err: ", err, " id: ", paras["id"])
			return
		}
	}

	if stateValue, ok := paras["state"]; ok {
		state, err = strconv.ParseUint(stateValue, 10, 32)
		if err != nil {
			gm.user.Error("SetDoorState failed, ParseUint err: ", err, " state: ", paras["state"])
			return
		}
	}

	space := gm.user.GetSpace().(*Scene)
	space.doorMgr.SetDoorState(gm.user, id, uint32(state))

	gm.user.Info("SetDoorState success, params: ", paras)
}

// DropBox 空投箱
func (gm *GmMgr) DropBox(paras map[string]string) {
	space := gm.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	GetRefreshItemMgr(space).GmRefreshBox(gm.user.GetPos())
}
