package main

import (
	"common"
	"excel"
	"time"
	"zeus/iserver"
)

//Energy 能量条
type Energy struct {
	user        *RoomUser
	num         uint32
	lastaddhp   int64
	lastsube    int64
	addhpupdate bool
}

//InitEnergy 初始能量
func InitEnergy(user *RoomUser) *Energy {
	return &Energy{user: user}
}

func (sf *Energy) addEnergy(num uint32) {
	system, ok := excel.GetSystem(uint64(common.System_EnergyMax))
	if !ok {
		return
	}
	max := uint32(system.Value)

	sf.num += num
	if sf.num > max {
		sf.num = max
	}

	sf.refreshData()
}

func (sf *Energy) subEnergy(num uint32) {
	if sf.num <= num {
		sf.num = 0
	} else {
		sf.num -= num
	}

	sf.refreshData()
}

func (sf *Energy) clearEnergy() {
	system, ok := excel.GetSystem(uint64(common.System_EnergyMax))
	if !ok {
		return
	}
	max := uint32(system.Value)

	sf.subEnergy(max)
}

func (sf *Energy) refreshData() {
	system, ok := excel.GetSystem(uint64(common.System_EnergyMax))
	if !ok {
		return
	}
	max := uint32(system.Value)

	var addspeed float32
	var curid uint64
	var curmax uint64
	for _, v := range excel.GetEnergyMap() {
		if uint32(v.Minvalue) <= sf.num && sf.num <= uint32(v.Value) {
			addspeed = v.Addspeed
			curid = v.Id
			curmax = v.Value
			break
		}
	}

	if addspeed != 0 {
		sf.user.RPC(iserver.ServerTypeClient, "setEnergy", uint64(sf.num), uint64(max), uint64(1))

		base, ok := excel.GetEnergy(curid - 1)
		if ok && addspeed > base.Addspeed && uint64(sf.num) > base.Value && curmax > base.Value {
			addspeed = float32(addspeed-base.Addspeed)/float32(curmax-base.Value)*float32(uint64(sf.num)-base.Value) + base.Addspeed
		}
	} else {
		sf.user.RPC(iserver.ServerTypeClient, "setEnergy", uint64(sf.num), uint64(max), uint64(0))
	}
	sf.user.SetSpeedRate(float32(addspeed + 1.0))
	sf.user.Debug("增加能量值速度", addspeed, " 当前能量值:", sf.num)
}

func (sf *Energy) update() {
	if sf.num == 0 {
		return
	}
	now := time.Now().Unix()
	for _, v := range excel.GetEnergyMap() {
		if uint32(v.Minvalue) <= sf.num && sf.num <= uint32(v.Value) {
			if !sf.addhpupdate && now == sf.lastaddhp+int64(v.Addinterval)-2 {
				sf.addhpupdate = true
				sf.user.RPC(iserver.ServerTypeClient, "EnergeAddHp", uint32(v.Addhp))
			}

			if now >= sf.lastaddhp+int64(v.Addinterval) {
				sf.lastaddhp = now
				sf.addhpupdate = false
				sf.user.DisposeAddHp(uint32(v.Addhp))
			}

			if now >= sf.lastsube+int64(v.Subinterval) {
				sf.lastsube = now
				sf.subEnergy(uint32(v.Subenergy))
			}

			break
		}
	}
}
