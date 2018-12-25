package main

import (
	"math"
	"protoMsg"
	"zeus/iserver"
	"zeus/linmath"

	"excel"

	"time"

	log "github.com/cihub/seelog"
)

// ControlerAI ai控制器
type ControlerAI struct {
	ai *RoomAI
}

// NewControlerAI 获取ai控制器
func NewControlerAI(ai *RoomAI) *ControlerAI {
	ctrl := &ControlerAI{
		ai: ai,
	}
	return ctrl
}

func (ctrl *ControlerAI) Loop() {
	ctrl.Detection()
}

func (ctrl *ControlerAI) Detection() {
	targetID := ctrl.searchEnemy()
	if targetID == 0 {
		return
	}
	//ctrl.ai.GetSpace().RemoveDelayCallByObj(ctrl.ai)
	if defender, ok := ctrl.ai.GetEntities().GetEntity(targetID).(iDefender); ok {
		// 设置玩家ai旋转角度
		ctrl.setAttRota(defender)

		if scene, ok := ctrl.ai.GetSpace().(*Scene); ok {
			if !scene.ISceneData.canAttack(ctrl.ai, targetID) {
				return
			}
		}
		ctrl.ai.defender = defender
		num := 1
		gunData, ok := excel.GetGun(uint64(ctrl.ai.GetInUseWeapon()))
		if ok {
			n := int(gunData.Bmax - gunData.Bmin)
			if n > 0 {
				num = ctrl.ai.r.Intn(n) + int(gunData.Bmin)
			}
		}
		t := time.Duration(gunData.Shootinterval1*1000) * time.Millisecond
		ctrl.doAttackTarget()
		for i := 1; i < num; i++ {
			ctrl.ai.GetSpace().AddDelayCallByObj(ctrl.ai, ctrl.doAttackTarget, time.Duration(i)*t)
		}
	}
}

// searchEnemy 搜索敌人
func (ctrl *ControlerAI) searchEnemy() (id uint64) {

	space := ctrl.ai.GetSpace().(*Scene)
	if space == nil {
		log.Error("searchEnemy failed, can not get space")
		return
	}

	sliID := make([]uint64, 0)

	aiPos := ctrl.ai.GetPos()

	ctrl.ai.GetSpace().TravsalAOI(ctrl.ai, func(o iserver.ICoordEntity) {
		target, ok := o.(iserver.ISpaceEntity)
		if !ok {
			return
		}

		if ctrl.ai.GetID() == target.GetID() {
			return
		}

		if aiPos.Sub(target.GetPos()).Len() > 30 {
			return
		}

		if target.GetType() != "Player" {
			return
		}

		u, ok := target.GetRealPtr().(*RoomUser)
		if !ok {
			return
		}

		if u.GetBaseState() == RoomPlayerBaseState_Watch || u.GetBaseState() == RoomPlayerBaseState_Dead {
			return
		}

		if u.isInTank() {
			return
		}

		// 射线检测

		// 目标点
		targetPos := target.GetPos()
		targetPos.Y += 1

		// 攻击点
		origionPos := ctrl.ai.GetPos()
		origionPos.Y += 1

		// 射线方向
		direction := targetPos.Sub(origionPos)
		direction.Normalize()

		// 检测障碍物
		rayDistance := origionPos.Sub(targetPos).Len()
		_, _, _, hit, _ := space.Raycast(origionPos, direction, rayDistance, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
		if hit {
			//log.Error("检测到障碍物")
			return
		}

		// 射线检测
		caldis, canattack, err := space.SphereRaycast(targetPos, 0.1, origionPos, direction, ctrl.ai.GetWeaponDistance())
		if !canattack {
			log.Error("Ray test failed, err: ", err, " distance: ", caldis, " gun distance: ", ctrl.ai.GetWeaponDistance(), " origionPos:", origionPos, " rota:", direction, " defender:", targetPos)
			return
		}

		sliID = append(sliID, target.GetID())
	})

	if len(sliID) != 0 {
		return sliID[0]
	}

	return 0
}

// setAttRota 设置攻击旋转角度
func (ctrl *ControlerAI) setAttRota(defender iDefender) {
	oo := defender.GetPos().Sub(ctrl.ai.GetPos())
	rota := linmath.Vector3{
		X: 0,
		Y: float32(180 / math.Pi * math.Acos(float64(oo.Z/oo.Len()))),
		Z: 0,
	}
	if oo.X < 0 {
		rota.Y = -rota.Y
	}
	ctrl.ai.SetRota(rota)
}

// 攻击目标
func (ctrl *ControlerAI) doAttackTarget() {
	// 目标死亡,攻击完成
	if ctrl.ai.defender == nil {
		return
	}

	// 通知开枪
	ctrl.notifyShootState(true)

	randValue := ctrl.ai.r.Float32()
	if randValue >= 1-ctrl.ai.aiData.AccFix {
		AttackHandle(ctrl.ai, ctrl.ai.defender, AttackPos_Body)

		// 目标点
		targetPos := ctrl.ai.defender.GetPos()
		targetPos.Y += 1

		origionPos := ctrl.ai.GetPos()
		origionPos.Y += 1

		// 射线方向
		direction := targetPos.Sub(origionPos)

		msg := &protoMsg.AttackReq{}
		msg.Defendid = ctrl.ai.defender.GetID()
		msg.Dir = &protoMsg.Vector3{X: direction.X, Y: direction.Y, Z: direction.Z}
		msg.Distance = ctrl.ai.GetWeaponDistance()
		msg.Firetime = 0
		msg.Ishead = false
		msg.Attackid = ctrl.ai.GetID()
		msg.Origion = &protoMsg.Vector3{X: origionPos.X, Y: origionPos.Y, Z: origionPos.Z}
		ctrl.ai.CastMsgToAllClientExceptMe(msg)
	}

	ctrl.notifyShootState(false)
}

// notifyShootState 通知射击状态
func (ctrl *ControlerAI) notifyShootState(state bool) {
	shootMsg := &protoMsg.ShootReq{}
	shootMsg.Attackid = ctrl.ai.GetID()
	shootMsg.Issuc = state
	ctrl.ai.CastRPCToAllClientExceptMe("ShootReq", shootMsg)
}
