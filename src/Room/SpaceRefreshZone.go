package main

import (
	"common"
	"excel"
	"math"
	"math/rand"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"

	log "github.com/cihub/seelog"
)

//GetRefreshZoneMgr 刷新区域
func GetRefreshZoneMgr(scene *Scene) *RefreshZoneMgr {
	if scene.refreshzone == nil {
		zonemgr := &RefreshZoneMgr{tb: scene}
		zonemgr.refreshcount = 0
		zonemgr.bombAreas = make(map[uint32]*BombAreaInfo)
		zonemgr.bombCounter = 1
		scene.refreshzone = zonemgr
	}

	return scene.refreshzone
}

//RefreshZoneMgr 刷新区域
type RefreshZoneMgr struct {
	tb *Scene

	bombAreas   map[uint32]*BombAreaInfo
	bombCounter uint32

	refreshstatus uint32
	refreshcount  uint64
	refreshtime   int64

	shrinkinterval int64
	shrinkcount    uint32

	safecenterinit linmath.Vector3
	nextsafecenter linmath.Vector3
	nextsaferadius float32
	pointA         linmath.Vector3
	pointB         linmath.Vector3

	tmpsafecenter linmath.Vector3 // 娱乐模式小圈的中心点

	shrinkdam uint32
}

type BombAreaInfo struct {
	id                uint32 //轰炸区id
	caller            string //召唤者名字
	color             uint32 //名字颜色
	bombcenter        linmath.Vector3
	bombradius        float32
	bombendtime       uint64
	bombrefreshtime   uint64
	bombrefreshstatus uint32
}

//GetRefreshCount 刷新次数
func (sf *RefreshZoneMgr) GetRefreshCount() uint64 {
	return sf.refreshcount
}

//BombCenter 轰炸区中心
func (sf *RefreshZoneMgr) BombCenter(id uint32) linmath.Vector3 {
	return sf.bombAreas[id].bombcenter
}

//BombRadius 轰炸区半径
func (sf *RefreshZoneMgr) BombRadius(id uint32) float32 {
	return sf.bombAreas[id].bombradius
}

//SetRefreshStatus 设置刷新状态
func (sf *RefreshZoneMgr) SetRefreshStatus(state uint32) {
	sf.refreshstatus = state
}

//Update 更新
func (sf *RefreshZoneMgr) Update() {
	now := time.Now().Unix()
	sf.RefreshZone(now)
}

//callUpBombArea 召唤空袭
func (sf *RefreshZoneMgr) callUpBombArea(caller string, color uint32) {
	sf.bombCounter++
	sf.bombAreas[sf.bombCounter] = &BombAreaInfo{
		id:                sf.bombCounter,
		caller:            caller,
		color:             color,
		bombrefreshtime:   uint64(time.Now().UnixNano() / (1000 * 1000)),
		bombrefreshstatus: common.Bomb_Status_Init,
	}

	sf.tb.Info("callUpBombArea: ", caller)
}

//RefreshBomb 刷新轰炸区
func (sf *RefreshZoneMgr) RefreshBomb() {
	now := uint64(time.Now().UnixNano() / (1000 * 1000))
	tb := sf.tb

	for id, area := range sf.bombAreas {
		var (
			base excel.MapruleData
			ok   bool
		)

		if id == 1 {
			base, ok = excel.GetMaprule(sf.refreshcount)
		} else {
			base, ok = excel.GetMaprule(20001)
		}

		if !ok {
			sf.tb.Error("Invalid map refresh rule id")
			return
		}

		if area.bombrefreshstatus == common.Bomb_Status_Init && now >= area.bombrefreshtime+uint64(base.Bombrefresh*1000) {
			// log.Debug("轰炸区时间")
			area.bombrefreshstatus = common.Bomb_Status_ShrinkBegin
			area.bombrefreshtime = now + uint64(base.Delaybombdam*1000)
			area.bombendtime = now + uint64(base.Bombtime*1000)

			area.bombradius = base.Bombradius

			radius := area.bombradius
			if sf.nextsaferadius > radius {
				radius = sf.nextsaferadius - radius
			}
			area.bombcenter = tb.GetCircleRamdomPos(sf.nextsafecenter, radius)
			if area.bombradius != 0 {
				tb.chatNotify(1, "轰炸区刷新了")
				tb.rpcBombRefresh(area.bombcenter, area.bombradius, area.id, area.caller, area.color)
			}
		} else if area.bombrefreshstatus == common.Bomb_Status_ShrinkBegin && now >= area.bombrefreshtime+uint64(base.Bombspeed*1000) {
			area.bombrefreshtime = now
			if area.bombrefreshtime >= area.bombendtime {
				area.bombrefreshstatus = common.Bomb_Status_Disappear
				tb.rpcBombDisapear(area.id)
				if id != 1 {
					delete(sf.bombAreas, area.id)
				}
			} else {
				if area.bombradius != 0 {
					dampos := tb.GetCircleRamdomPos(area.bombcenter, area.bombradius)
					sf.BombDam(dampos, base.Bombdamrradius, uint32(base.Bombdam))
					tb.rpcBombDam(dampos, base.Bombdamrradius, area.id)
				}
			}
		}
	}
}

//BombDam 轰炸伤害
func (sf *RefreshZoneMgr) BombDam(pos linmath.Vector3, r float32, dam uint32) {
	tb := sf.tb
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			if common.Distance(user.GetPos(), pos) < r {
				if !user.isInBuilding() {
					// log.Error(user.GetID(), "不在房子里减血", dam)
					user.DisposeSubHp(InjuredInfo{num: dam, injuredType: bomb, isHeadshot: false})
				} else {
					// log.Error(user.GetID(), "角色在房子里面")
				}
			}
		}
	})

	tb.TravsalEntity("AI", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomAI); ok {
			if common.Distance(user.GetPos(), pos) < r {
				user.DisposeSubHp(InjuredInfo{num: dam, injuredType: bomb, isHeadshot: false})
			}
		}
	})
}

//InitZone 初始区域
func (sf *RefreshZoneMgr) InitZone(now int64) {
	sf.refreshcount = sf.tb.mapdata.Id*1000 + 1
	sf.refreshstatus = common.Status_Init
	sf.refreshtime = now

	sf.bombAreas[1] = &BombAreaInfo{
		id:                1,
		bombrefreshtime:   uint64(time.Now().UnixNano() / (1000 * 1000)),
		bombrefreshstatus: common.Bomb_Status_Init,
	}
	// log.Error("@@@@@@@轰炸区时间", sf.bombrefreshtime, time.Now().Unix())

	sf.InitPoint()
	sf.GeneratePoint()
	sf.tb.refreshSpecialBox(sf.refreshcount)
}

//RefreshZone 刷新区域
func (sf *RefreshZoneMgr) RefreshZone(now int64) {
	tb := sf.tb
	base, ok := excel.GetMaprule(sf.refreshcount)
	if !ok {
		return
	}

	sf.shrinkdam = uint32(base.Shrinkdam)
	sf.RefreshBomb()

	if sf.refreshstatus == common.Status_Init && now >= sf.refreshtime+int64(base.Shrinkarea) {
		sf.refreshstatus = common.Status_ShrinkBegin
		sf.refreshtime = now
		sf.shrinkinterval = int64(base.Shrinktime / base.Shrinkcount)
		if sf.shrinkinterval == 0 {
			sf.shrinkinterval = 2
		}
		sf.shrinkcount = 0

		tb.chatNotify(1, "禁令区开始蔓延")
		if sf.refreshcount == 1 {
			tb.zoneNotify(3, tb.safecenter, tb.saferadius, uint32(sf.shrinkinterval))
		}
	} else if sf.refreshstatus == common.Status_ShrinkBegin && now >= sf.refreshtime+sf.shrinkinterval {
		sf.refreshtime = now
		sf.shrinkcount++

		if sf.shrinkcount >= uint32(base.Shrinkcount) {
			sf.refreshstatus = common.Status_Init
			sf.refreshcount++
			changebase, ok := excel.GetMaprule(sf.refreshcount)
			if !ok {
				return
			}

			GetRefreshItemMgr(tb).SetRefreshBox(now)

			sf.bombAreas[1].bombrefreshstatus = common.Bomb_Status_Init
			sf.bombAreas[1].bombrefreshtime = uint64(time.Now().UnixNano() / (1000 * 1000))

			tb.safecenter = sf.nextsafecenter
			sf.safecenterinit = tb.safecenter
			tb.saferadius = sf.nextsaferadius

			radiusrate := float32(changebase.Radiusrate) / 100.0
			radius := float32(radiusrate * tb.saferadius)
			sf.nextsafecenter = tb.GetCircleRamdomPos(tb.safecenter, tb.saferadius-radius)
			sf.nextsaferadius = radius
			// log.Info("下次安全区半径", radius)

			sf.GeneratePoint()

			//刷圈之前 通知客户端清理空投
			tb.TravsalEntity("Player", func(e iserver.IEntity) {
				if e != nil {
					if user, ok := e.(*RoomUser); ok {
						user.RPC(iserver.ServerTypeClient, "ClearDropBoxes")
					}
				}
			})

			sf.tb.refreshSpecialBox(sf.refreshcount)
			tb.chatNotify(1, "目标区域出现")
			tb.zoneNotify(3, tb.safecenter, tb.saferadius, uint32(sf.shrinkinterval))
			tb.zoneNotify(2, sf.nextsafecenter, sf.nextsaferadius, uint32(sf.shrinkinterval))
			sf.rpcShrinkRefresh()
		} else {
			sf.ShrinkPoint(&base)
			tb.zoneNotify(0, tb.safecenter, tb.saferadius, uint32(sf.shrinkinterval))
		}
	}
}

//GeneratePoint 生成点
func (sf *RefreshZoneMgr) GeneratePoint() {
	tb := sf.tb
	distacne := common.Distance(tb.safecenter, sf.nextsafecenter)
	rate := (distacne + sf.nextsaferadius) / sf.nextsaferadius
	rate *= -1.0
	x := common.GetDefiniteequinox(tb.safecenter.X, sf.nextsafecenter.X, rate)
	z := common.GetDefiniteequinox(tb.safecenter.Z, sf.nextsafecenter.Z, rate)
	sf.pointA = linmath.Vector3{
		X: x,
		Y: 0,
		Z: z,
	}

	rate = tb.saferadius / (tb.saferadius - distacne)
	rate *= -1.0
	x = common.GetDefiniteequinox(tb.safecenter.X, sf.nextsafecenter.X, rate)
	z = common.GetDefiniteequinox(tb.safecenter.Z, sf.nextsafecenter.Z, rate)
	sf.pointB = linmath.Vector3{
		X: x,
		Y: 0,
		Z: z,
	}
}

//InitPoint 初始点
func (sf *RefreshZoneMgr) InitPoint() {
	tb := sf.tb

	tb.zoneNotify(3, tb.safecenter, tb.saferadius, 0)
	tb.zoneNotify(2, sf.nextsafecenter, sf.nextsaferadius, 0)
	log.Debug("初始化刷新安全区域", tb.safecenter, tb.saferadius, sf.refreshcount)
	sf.rpcShrinkRefresh()
}

func (sf *RefreshZoneMgr) rpcShrinkRefresh() {
	base, ok := excel.GetMaprule(sf.refreshcount)
	if !ok {
		return
	}

	sf.tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "ShrinkRefresh", uint64(base.Shrinkarea))
		}
	})
}

//ShrinkPoint 收缩
func (sf *RefreshZoneMgr) ShrinkPoint(base *excel.MapruleData) {
	tb := sf.tb
	rate := float32(sf.shrinkcount) / float32(uint32(base.Shrinkcount)-sf.shrinkcount)
	x := common.GetDefiniteequinox(sf.safecenterinit.X, sf.nextsafecenter.X, rate)
	z := common.GetDefiniteequinox(sf.safecenterinit.Z, sf.nextsafecenter.Z, rate)
	tb.safecenter = linmath.Vector3{
		X: x,
		Y: 0,
		Z: z,
	}

	rate = float32(uint32(base.Shrinkcount)-sf.shrinkcount) / float32(sf.shrinkcount)
	x = common.GetDefiniteequinox(sf.pointA.X, sf.pointB.X, rate)
	z = common.GetDefiniteequinox(sf.pointA.Z, sf.pointB.Z, rate)
	temp := linmath.Vector3{
		X: x,
		Y: 0,
		Z: z,
	}
	tb.saferadius = common.Distance(tb.safecenter, temp)
}

//InBombArea 是否为炸弹区域
func (sf *RefreshZoneMgr) InBombArea(pos linmath.Vector3) bool {
	for _, area := range sf.bombAreas {
		dist := area.bombcenter.Sub(pos).Len()
		if dist < 0 {
			dist *= -1
		}

		if dist <= area.bombradius {
			return true
		}
	}

	return false
}

//genSafeZonePoint 生成安全区中心初始点
func (sf *RefreshZoneMgr) genSafeZonePoint() {
	tb := sf.tb
	tb.safecenter = linmath.Vector3{
		X: tb.mapdata.Width / 2,
		Y: 0,
		Z: tb.mapdata.Height / 2,
	}
	dx := float64(tb.safecenter.X)
	dz := float64(tb.safecenter.Z)
	tb.saferadius = float32(math.Sqrt(float64(dx*dx + dz*dz)))
	sf.safecenterinit = tb.safecenter

	poslist := strings.Split(tb.mapdata.Safe_zone, ";")
	rand.Seed(time.Now().UnixNano())
	posrand := rand.Intn(len(poslist))
	zonepos := strings.Split(poslist[posrand], ",")
	// 娱乐模式 初始安全区刷新方式和 其他模式 不同
	if tb.GetMatchMode() != common.MatchModeArcade && tb.GetMatchMode() != common.MatchModeTankWar {
		cx := common.StringToInt(zonepos[0])
		cy := common.StringToInt(zonepos[1])
		cz := common.StringToInt(zonepos[2])
		sf.tmpsafecenter = linmath.Vector3{
			X: float32(cx),
			Y: float32(cy),
			Z: float32(cz),
		}
		sf.nextsafecenter = linmath.Vector3{
			X: float32(cx),
			Y: float32(cy),
			Z: float32(cz),
		}
		sf.nextsaferadius = tb.mapdata.Saferadius
	} else {
		angle := rand.Float64()
		cosX := float32(math.Cos(2*angle*math.Pi)) * tb.mapdata.Saferadius
		sinZ := float32(math.Sin(2*angle*math.Pi)) * tb.mapdata.Saferadius

		cx := common.StringToInt(zonepos[0])
		cy := common.StringToInt(zonepos[1])
		cz := common.StringToInt(zonepos[2])
		sf.tmpsafecenter = linmath.Vector3{
			X: float32(cx),
			Y: float32(cy),
			Z: float32(cz),
		}
		sf.nextsafecenter = linmath.Vector3{
			X: float32(cx) + cosX,
			Y: float32(cy),
			Z: float32(cz) + sinZ,
		}
		sf.nextsaferadius = 2 * tb.mapdata.Saferadius
	}

	log.Debug("tb.GetMatchMode():", tb.GetMatchMode(), " sf.nextsafecenter:", sf.nextsafecenter, " sf.tmpsafecenter:", sf.tmpsafecenter, " sf.nextsaferadius:", sf.nextsaferadius)
}
