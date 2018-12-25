package main

import (
	"entitydef"
	"math/rand"
	"time"
	"zeus/linmath"
	"zeus/space"

	log "github.com/cihub/seelog"
)

// SpaceItem 道具
type SpaceItem struct {
	entitydef.ItemDef
	space.TinyEntity

	item       *Item
	haveobjs   map[uint32]*Item
	cangettime int64
}

// Init 初始化底层回调
func (item *SpaceItem) Init(initParam interface{}) {
	// item.SetMarker(false) //不需要记录谁关注了自己
	// item.SetEnable(false) //不需要DoMsg

	space := item.GetSpace().(*Scene)
	mapitem, ok := space.mapitem[item.GetID()]
	if !ok {
		log.Error("Init failed, can't get item, id: ", item.GetID())
		return
	}

	if mapitem.item == nil {
		log.Error("Init failed, item is nil, id: ",item.GetID())
		return
	}

	// item.SetMarkAOIRange(mapitem.item.base.MarkAoiRange)
	// item.SetWatchAOIRange(0)

	item.item = mapitem.item
	item.haveobjs = make(map[uint32]*Item)
	for k, v := range mapitem.haveobjs {
		item.haveobjs[k] = v
	}
	item.haveobjs = mapitem.haveobjs

	item.cangettime = mapitem.cangettime
	item.Setbaseid(mapitem.itemid)
	item.Setnum(mapitem.item.count)
	item.Setreducedam(mapitem.item.reducedam)
	item.SetPos(mapitem.pos)

	rota := mapitem.dir
	if rota.IsEqual(linmath.Vector3_Zero()) {
		rand.Seed(time.Now().UnixNano())
		rota.Y = rand.Float32() * 360
		if item.item.base.Type <= ItemTypeWeapon {
			rota.Z = 90.0
		}
	}
	item.SetRota(rota)
	//log.Info("Item inited")
}

// Destroy 销毁的时候底层回调
// func (item *SpaceItem) Destroy() {
// 	log.Info("销毁道具", item)
// }

// CreateNewEntityState 创建一个新的状态快照，由框架层调用
func (item *SpaceItem) CreateNewEntityState() space.IEntityState {
	return NewRoomItemState()
}
