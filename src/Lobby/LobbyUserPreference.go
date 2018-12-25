package main

import (
	"common"
	"db"
	"math/rand"
	"protoMsg"
	"zeus/iserver"

	"excel"
)

// initPreferenceInfo 初始化偏好信息
func (user *LobbyUser) initPreferenceInfo() {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		user.Error("preItemRandomCreate GetPreferenceList err:", err)
		return
	}

	msgList := &protoMsg.PreferenceList{}
	// 随机刷新偏好道具(角色和伞包)
	for i := 1; i <= 2; i++ {
		msgInfo := &protoMsg.PreferenceInfo{}
		msgInfo.Id = uint32(i)
		msgInfo.Start = append(msgInfo.Start, info.Start[uint32(i)][1])
		msgInfo.PreType = append(msgInfo.PreType, info.PreType[uint32(i)][1])

		msgList.Info = append(msgList.Info, msgInfo)
	}
	// 随机刷新偏好道具(背包和头盔)
	for i := 5; i <= 6; i++ {
		msgInfo := &protoMsg.PreferenceInfo{}
		msgInfo.Id = uint32(i)

		for j := 1; j <= 3; j++ {
			msgInfo.Start = append(msgInfo.Start, info.Start[uint32(i)][uint32(j)])
			msgInfo.PreType = append(msgInfo.PreType, info.PreType[uint32(i)][uint32(j)])
		}

		msgList.Info = append(msgList.Info, msgInfo)
	}

	user.Debug("initPreferenceInfo msgList:", msgList)
	if err := user.RPC(iserver.ServerTypeClient, "InitPreferenceInfo", msgList); err != nil {
		user.Error("err:", err)
	}
}

// setPreferenceSwitch 设置随机偏好是否启用
func (user *LobbyUser) setPreferenceSwitch(typ, level uint32, state bool) {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		user.Error("setPreferenceSwitch GetPreferenceList err:", err)
		return
	}

	ret := true // 设置是否成功标志
	if state {
		state = false
		ret = false
		for _, info := range db.PlayerGoodsUtil(user.GetDBID()).GetAllGoodsInfo() {
			if info == nil {
				continue
			}

			goodsConfig, ok := excel.GetStore(uint64(info.Id))
			if !ok {
				user.Error("GetStore does't exist, info.Id: ", info.Id)
				continue
			}
			if goodsConfig.State == 3 { //免费道具
				continue
			}

			if goodsConfig.Type != uint64(typ) {
				continue
			}

			switch typ {
			case 1, 2:
				state = true
				ret = true
				break
			case 5, 6:
				itemData, ok := excel.GetItem(goodsConfig.RelationID)
				if !ok {
					user.Error("GetItem does't exist, goodsConfig.RelationID: ", goodsConfig.RelationID)
					continue
				}
				if itemData.Subtype == uint64(level) {
					state = true
					ret = true
					break
				}
			}
		}
	}

	if info.Start[typ][level] != state {
		start := make(map[uint32]bool)
		if typ == 5 || typ == 6 {
			for i := 1; i <= 3; i++ {
				if uint32(i) != level {
					start[uint32(i)] = info.Start[typ][uint32(i)]
				} else {
					start[level] = state
				}
			}
		} else {
			start[level] = state
		}
		info.Start[typ] = start
		if err := util.SetPreferenceList(info); err != nil {
			user.Error("setPreferenceSwitch SetPreferenceList err:", err)
			return
		}
	}

	user.Debug("OpenRandomPreferenceRsp state:", state, " ret:", ret)
	if err := user.RPC(iserver.ServerTypeClient, "OpenRandomPreferenceRsp", ret, state, typ, level); err != nil {
		user.Error(err)
	}
}

// allOrPartPreSwitch 全体还是部分偏好设置开关
func (user *LobbyUser) allOrPartPreSwitch(typ, level, result uint32) (bool, uint32) {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		user.Error("allOrPartPreSwitch GetPreferenceList err:", err)
		return false, info.PreType[typ][level]
	}

	if result == 1 && len(info.MayaItemList[typ][level]) == 0 {
		return false, info.PreType[typ][level]
	}

	preType := make(map[uint32]uint32)
	if typ == 5 || typ == 6 {
		for i := 1; i <= 3; i++ {
			if uint32(i) != level {
				preType[uint32(i)] = info.PreType[typ][uint32(i)]
			} else {
				preType[level] = result
			}
		}
	} else {
		preType[level] = result
	}
	info.PreType[typ] = preType
	if err := util.SetPreferenceList(info); err != nil {
		user.Error("allOrPartPreSwitch SetPreferenceList err:", err)
		return false, info.PreType[typ][level]
	}

	return true, preType[level]
}

// getAllItem 将拥有的道具进行分类
func (user *LobbyUser) getAllItem() map[uint32]map[uint32][]uint32 {
	roleItemList := make(map[uint32][]uint32)
	packItemList := make(map[uint32][]uint32)
	bagItemList := make(map[uint32][]uint32)
	headItemList := make(map[uint32][]uint32)
	allMayaItemList := make(map[uint32]map[uint32][]uint32)

	for _, info := range db.PlayerGoodsUtil(user.GetDBID()).GetAllGoodsInfo() {
		if info == nil {
			continue
		}

		goodsConfig, ok := excel.GetStore(uint64(info.Id))
		if !ok {
			user.Error("GetStore does't exist, info.Id: ", info.Id)
			continue
		}
		if goodsConfig.State == 3 { //免费道具
			continue
		}

		switch goodsConfig.Type {
		case 1: //1: 角色
			roleItemList[1] = append(roleItemList[1], info.Id)
		case 2: //2：伞包
			packItemList[1] = append(packItemList[1], info.Id)
		case 5: //5：背包
			itemData, ok := excel.GetItem(goodsConfig.RelationID)
			if !ok {
				user.Error("GetItem does't exist, goodsConfig.RelationID: ", goodsConfig.RelationID)
				continue
			}

			bagItemList[uint32(itemData.Subtype)] = append(bagItemList[uint32(itemData.Subtype)], info.Id)
		case 6: //6: 头盔
			itemData, ok := excel.GetItem(goodsConfig.RelationID)
			if !ok {
				user.Error("GetItem does't exist, goodsConfig.RelationID: ", goodsConfig.RelationID)
				continue
			}

			headItemList[uint32(itemData.Subtype)] = append(headItemList[uint32(itemData.Subtype)], info.Id)
		}
	}

	allMayaItemList[1] = roleItemList
	allMayaItemList[2] = packItemList
	allMayaItemList[5] = bagItemList
	allMayaItemList[6] = headItemList

	return allMayaItemList
}

// preItemRandomCreate 随机生成偏好道具
func (user *LobbyUser) preItemRandomCreate(typ, level uint32) {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		user.Error("preItemRandomCreate GetPreferenceList err:", err)
		return
	}

	if !info.Start[typ][level] {
		user.Info("preItemRandomCreate close!")
		return
	}

	switch info.PreType[typ][level] {
	case 0: //TODO:全体随机
		allMayaItemList := user.getAllItem()
		user.randomCreatItem(typ, level, info.Start, allMayaItemList)
	case 1: //偏好随机
		user.randomCreatItem(typ, level, info.Start, info.MayaItemList)
	default:
		user.Error("preItemRandomCreate PreType err! PreType:", info.PreType[typ][level])
		return
	}
}

// randomCreatItem 随机生成道具
func (user *LobbyUser) randomCreatItem(typ, level uint32, start map[uint32]map[uint32]bool, allMayaItemList map[uint32]map[uint32][]uint32) {
	if !start[typ][level] {
		return
	}

	switch typ {
	case 1: //1:角色
		l := len(allMayaItemList[typ][1])
		if l > 0 {
			index := 0
			for {
				index = rand.Intn(l)
				if len(allMayaItemList[typ][1]) == 1 || user.GetRoleModel() != allMayaItemList[typ][1][index] {
					break
				}
			}
			user.SetRoleModel(allMayaItemList[typ][1][index])
			user.SetGoodsRoleModel(allMayaItemList[typ][1][index])

			if user.teamMgrProxy != nil && user.GetTeamID() != 0 {
				if err := user.teamMgrProxy.RPC(common.ServerTypeMatch, "SetRoleModel", user.GetID(), user.GetTeamID(), allMayaItemList[typ][1][index]); err != nil {
					user.Error("RPC SetRoleModel err: ", err)
				}
			}
		}
	case 2: //2：伞包
		l := len(allMayaItemList[typ][1])
		if l > 0 {
			index := 0
			for {
				index = rand.Intn(l)
				if len(allMayaItemList[typ][1]) == 1 || user.GetParachuteID() != allMayaItemList[typ][1][index] {
					break
				}
			}
			user.SetParachuteID(allMayaItemList[typ][1][index])
			user.SetGoodsParachuteID(allMayaItemList[typ][1][index])
		}
	case 5: //5：背包
		msg := user.GetPackWearInGame()
		user.randomMayaItem(typ, level, msg, allMayaItemList)
	case 6: //6:头盔
		msg := user.GetHeadWearInGame()
		user.randomMayaItem(typ, level, msg, allMayaItemList)
	default:
		user.Error("typ err! typ:", typ)
		return
	}
}

// randomMayaItem 随机生成幻化道具
func (user *LobbyUser) randomMayaItem(typ, level uint32, msg *protoMsg.WearInGame, allMayaItemList map[uint32]map[uint32][]uint32) {
	if msg == nil {
		return
	}

	l := len(allMayaItemList[typ][level])
	if l > 0 {
		for {
			index := rand.Intn(l)
			switch level {
			case 1:
				if len(allMayaItemList[typ][level]) == 1 || msg.First != allMayaItemList[typ][level][index] {
					user.storeMgr.useMayaItem(msg.First, 0)
					user.storeMgr.useMayaItem(allMayaItemList[typ][level][index], 1)
					return
				}
			case 2:
				if len(allMayaItemList[typ][level]) == 1 || msg.Second != allMayaItemList[typ][level][index] {
					user.storeMgr.useMayaItem(msg.Second, 0)
					user.storeMgr.useMayaItem(allMayaItemList[typ][level][index], 1)
					return
				}
			case 3:
				if len(allMayaItemList[typ][level]) == 1 || msg.Third != allMayaItemList[typ][level][index] {
					user.storeMgr.useMayaItem(msg.Third, 0)
					user.storeMgr.useMayaItem(allMayaItemList[typ][level][index], 1)
					return
				}
			default:
				user.Error("level err! level:", level)
				return
			}
		}
	}
}

// setItemPreference 设置某物品为偏好物品
func (user *LobbyUser) setItemPreference(typ, id, preference uint32) bool {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		user.Error("setItemPreference GetPreferenceList err:", err)
		return false
	}

	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		user.Error("GetStore does't exist, id: ", id)
		return false
	}
	if goodsConfig.State == 3 { //免费道具
		return false
	}

	var itemlist []uint32
	switch goodsConfig.Type {
	case 1, 2: //1:角色，2：伞包
		for _, v := range info.MayaItemList[typ][1] {
			if v == id {
				continue
			}
			itemlist = append(itemlist, v)
		}
		if preference == 1 {
			itemlist = append(itemlist, id)
		}

		mayaItemList := make(map[uint32][]uint32)
		mayaItemList[1] = itemlist
		info.MayaItemList[typ] = mayaItemList

		if preference == 0 && len(itemlist) == 0 {
			preType := make(map[uint32]uint32)
			preType[1] = 0
			info.PreType[typ] = preType
			if err := user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), typ, uint32(1)); err != nil {
				user.Error(err)
			}
		}
	case 5, 6: //5：背包，6:头盔
		itemData, ok := excel.GetItem(goodsConfig.RelationID)
		if !ok {
			user.Error("GetItem does't exist, goodsConfig.RelationID: ", goodsConfig.RelationID)
			return false
		}

		for _, v := range info.MayaItemList[typ][uint32(itemData.Subtype)] {
			if v == id {
				continue
			}
			itemlist = append(itemlist, v)
		}
		if preference == 1 {
			itemlist = append(itemlist, id)
		}

		mayaItemList := make(map[uint32][]uint32)
		for i := 1; i <= 3; i++ {
			if uint64(i) != itemData.Subtype {
				mayaItemList[uint32(i)] = info.MayaItemList[typ][uint32(i)]
			} else {
				mayaItemList[uint32(itemData.Subtype)] = itemlist
			}
		}
		info.MayaItemList[typ] = mayaItemList

		if preference == 0 && len(itemlist) == 0 {
			preType := make(map[uint32]uint32)
			for i := 1; i <= 3; i++ {
				if uint32(i) != uint32(itemData.Subtype) {
					preType[uint32(i)] = info.PreType[typ][uint32(i)]
				} else {
					preType[uint32(itemData.Subtype)] = 0
				}
			}
			info.PreType[typ] = preType
			if err := user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), typ, uint32(itemData.Subtype)); err != nil {
				user.Error(err)
			}
		}
	default:
		user.Error("typ err! typ:", typ)
		return false
	}

	if err := util.SetPreferenceList(info); err != nil {
		user.Error("setItemPreference SetPreferenceList err:", err)
		return false
	}

	return true
}
