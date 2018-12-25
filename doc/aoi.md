## AOI 范围实体的进出变化通知

1. space.Entity的主循环中调用updatePosCoord(e.pos)

```go
// OnLoop 循环调用
func (e *Entity) OnLoop() {
	e.updateAOI()
	e.resetState()
	e.Entity.OnLoop()
	e.updatePosCoord(e.pos)
	e.updateState()
}
```

2. 实体调用场景实体的UpdateCoord(e)更新自己的AOI状态

```go
func (e *Entity) updatePosCoord(pos linmath.Vector3) {

	if e.needUpdateAOI {
		s := e.GetSpace()
		if s != nil {
			s.UpdateCoord(e)
		}

		e.lastAOIPos = pos
		e.needUpdateAOI = false
	}
}
```

3. 场景实体的UpdateCoord实现是TileCoord

```go
//UpdateCoord 更新坐标位置
func (c *TileCoord) UpdateCoord(n iserver.ICoordEntity) {

	//视野远的
	if n.IsAOITrigger() || !n.IsNearAOILayer() {
		c.farTiles.update(n)
	}

	//视野近的
	if n.IsAOITrigger() || n.IsNearAOILayer() {
		c.nearTiles.update(n)
	}
}
```

4. 九宫格系统update(n)

```go
// Update 更新位置
func (t *Tiles) update(n iserver.ICoordEntity) {
	pos := newCoordPos(n.GetPos())

	if !t.isValidPos(pos) {
		t.remove(n)
		return
	}

	info, ok := t.entites[n.GetID()]
	if !ok {
		t.add(n, pos)
	} else {
		t.move(info, pos)
	}
}
```

```go
func (t *Tiles) move(info *CoordInfo, pos CoordPos) {

	nt := t.getTower(pos)
	if info.tower == nt || nt == nil {
		//还在原来的格子或者新格子异常为空了就不管了
		return
	}

	info.tower.moveTo(info.entity, nt)
	info.tower = nt
}
```

```go
func (t *Tower) moveTo(n iserver.ICoordEntity, nt *Tower) {

	deltaX := nt.gridX - t.gridX
	deltaY := nt.gridY - t.gridY

	if deltaX == 0 && deltaY == 0 {
		return
	}

	//不是移到相邻的格子，通知九宫格所有玩家
	if deltaX > 1 || deltaX < -1 || deltaY > 1 || deltaY < -1 {
		t.remove(n)	//从原来的格子中移除实体信息
		nt.add(n) //加到新格子里面
		return
	}

	//从原来的格子中移除实体信息
	t.removeFromList(n)
	//给反方向的格子通知一下离开
	t.travsalNeighour(t.getInvertDir(t.getDir(deltaX, deltaY)), func(tt *Tower) { tt.notifyTowerRemove(n) })

	//给移动方向上的通知一下增加
	nt.travsalNeighour(t.getDir(deltaX, deltaY), func(tt *Tower) { tt.notifyTowerAdd(n) })
	//加到新格子里面
	nt.addToList(n)
}
```

5. 离开aoi会调用OnEntityLeaveAOI

```go
//OnEntityLeaveAOI 实体离开AOI范围
func (e *Entity) OnEntityLeaveAOI(o iserver.ICoordEntity) {
	// 当o在我的额外关注列表中时, 不触发真正的LeaveAOI
	if extWatch, ok := e.extWatchList[o.GetID()]; ok {
		extWatch.isInAOI = false
		return
	}

	if e._isWatcher {
		//标记离开
		e.aoies = append(e.aoies, AOIInfo{false, o})
	}

	if o.IsWatcher() {
		e.watcherNums--
	}
}
```

6. 进入aoi范围调用OnEntityEnterAOI

```go
//OnEntityEnterAOI 实体进入AOI范围
func (e *Entity) OnEntityEnterAOI(o iserver.ICoordEntity) {
	// 当o在我的额外关注列表中时, 不触发真正的EnterAOI, 只是打个标记
	if extWatch, ok := e.extWatchList[o.GetID()]; ok {
		extWatch.isInAOI = true
		return
	}

	if e._isWatcher {
		//标记进入
		e.aoies = append(e.aoies, AOIInfo{true, o})
	}

	if o.IsWatcher() {
		e.watcherNums++
	}
}
```

7. AOI通知

```go
// OnLoop 循环调用
func (e *Entity) OnLoop() {
	e.updateAOI()	
	e.resetState()
	e.Entity.OnLoop()
	e.updatePosCoord(e.pos)
	e.updateState()
}
```

```go
//通知一下其它实体进入离开aoi范围的消息 
func (e *Entity) updateAOI() {

	if len(e.aoies) != 0 && e._isWatcher {
		msg := msgdef.NewEntityAOISMsg()
		for i := 0; i < len(e.aoies); i++ {
			info := e.aoies[i]

			ip := info.entity.(iAOIPacker)

			var data []byte

			if info.isEnter {
				num, propBytes := ip.GetAOIProp()
				m := &msgdef.EnterAOI{
					EntityID:   ip.GetID(),
					EntityType: ip.GetType(),
					State:      ip.GetStatePack(),
					PropNum:    uint16(num),
					Properties: propBytes,
				}

				data = make([]byte, m.Size()+1)
				data[0] = 1
				m.MarshalTo(data[1:])

			} else {
				m := &msgdef.LeaveAOI{
					EntityID: ip.GetID(),
				}

				data = make([]byte, m.Size()+1)
				data[0] = 0
				m.MarshalTo(data[1:])
			}

			msg.AddData(data)
		}

		e.PostToClient(msg)
		e.aoies = e.aoies[0:0]
	}
}
```


## AOI范围其它消息广播

1. 广播消息基本都是通过场景实体的TravsalAOI来实现的

```go
// 根据远近视野区分调用对应的九宫格系统的遍历
func (c *TileCoord) TravsalAOI(n iserver.ICoordEntity, cb func(iserver.ICoordEntity)) {

	if n.IsAOITrigger() || !n.IsNearAOILayer() {
		c.farTiles.TravsalAOI(n, cb)
	}

	if n.IsAOITrigger() || n.IsNearAOILayer() {
		c.nearTiles.TravsalAOI(n, cb)
	}
}
```

```go
// 遍历九个格子的玩家
func (t *Tiles) TravsalAOI(n iserver.ICoordEntity, cb func(iserver.ICoordEntity)) {
	pos := newCoordPos(n.GetPos())
	tt := t.getTower(pos)
	if tt == nil {
		return
	}

	tt.TravsalAOI(cb)
	t.travsalNeighour(tt, towerDir_All, func(nt *Tower) { nt.TravsalAOI(cb) })
}
```

```go
// 每个格子里的玩家遍历 
func (t *Tower) TravsalAOI(cb func(iserver.ICoordEntity)) {

	for e := t.aoiEntities.Front(); e != nil; e = e.Next() {
		ii := e.Value.(iserver.ICoordEntity)

		if ii.IsNearAOILayer() == t.tiles.isNearLayer && ii.IsAOITrigger() {
			cb(ii)
		}
	}
}
```

	比如 CastMsgToAllClient(msg)
```go
// CastMsgToAllClient 发送消息给所有关注我的客户端
func (e *Entity) CastMsgToAllClient(msg msgdef.IMsg) {
	e.GetSpace().TravsalAOI(e, func(ia iserver.ICoordEntity) {
		if ise, ok := ia.(iserver.IEntityStateGetter); ok {
			if ise.GetEntityState() != iserver.Entity_State_Loop {
				return
			}

			if ie, ok := ia.(iserver.IEntity); ok {
				ie.Post(iserver.ServerTypeClient, msg)
			}
		}
	})

	e.TravsalExtWatchs(func(o *extWatchEntity) {
		if ise, ok := o.entity.(iserver.IEntityStateGetter); ok {
			if ise.GetEntityState() != iserver.Entity_State_Loop {
				return
			}

			if ie, ok := o.entity.(iserver.IEntity); ok {
				ie.Post(iserver.ServerTypeClient, msg)
			}
		}
	})
}
```