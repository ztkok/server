package main

import (
	"excel"
	"math/rand"
	"protoMsg"
	"time"
	"zeus/iserver"
	"zeus/linmath"

	log "github.com/cihub/seelog"
)

/*
 攻击处理
*/

var isShowAttackLog bool

// iAttacker 攻击者
type iAttacker interface {
	iserver.ISpaceEntity

	GetAttack(targetPos linmath.Vector3, headshot bool) uint32
	isHeadShot(defender iDefender) bool
	GetKillNum() uint32
	IncrKillNum()
	GetName() string
	GetInsignia() string
	GetUserTeamID() uint64
	GetInUseWeapon() uint32
	IncrHeadShotNum()
	AddEffectHarm(effect uint32)
}

// iDefender 防御者
type iDefender interface {
	iserver.ISpaceEntity

	isInTank() bool

	UnderAttack(entityID uint64)
	Death(injuredType uint32, attackid uint64)

	GetInsignia() string
	GetUserTeamID() uint64
	GetName() string
	GetHP() uint32
	SetHP(v uint32)
	AddHp(num uint32)
	GetVehicleProp() *protoMsg.VehicleProp
	GetHeadProp() *protoMsg.HeadProp
	SetHeadProp(v *protoMsg.HeadProp)
	SetHeadPropDirty()
	GetBodyProp() *protoMsg.BodyProp
	SetBodyProp(v *protoMsg.BodyProp)
	SetBodyPropDirty()
	GetState() uint8

	DisposeSubHp(injuredInfo InjuredInfo)
	SetDieNotify(proto *protoMsg.DieNotifyRet)
}

// dongReduceDam 护盾减伤
func dongReduceDam(defender iDefender, attack uint32) uint32 {
	defenderUser, ok := defender.(*RoomUser)
	if !ok {
		return 0
	}

	defend := defenderUser.SkillData.skillEffectDam[SE_Shield]
	if defend <= 0 {
		return 0
	}

	param := &SkillEffectValue{}
	param.effectID = SE_Shield
	param.value = uint64(attack)
	GetSkillEffect(SE_Shield).End(defenderUser, param)

	if attack <= defend {
		return attack
	}

	return defend
}

// AttackHandle 攻击处理
func AttackHandle(attacker iAttacker, defender iDefender, Ishead uint32) {
	if attacker.GetID() == defender.GetID() {
		log.Warn("Can't attack self, id: ", attacker.GetID())
		return
	}

	var headshot bool
	if Ishead == AttackPos_Head && attacker.isHeadShot(defender) {
		headshot = true
	}
	attack := attacker.GetAttack(defender.GetPos(), headshot)

	// 攻击到护盾上
	if Ishead == AttackPos_Dong {
		attack -= dongReduceDam(defender, attack)
	}

	if 0 == attack {
		log.Warn("Damage value is zero")
		return
	}

	oldattack := attack
	attack -= damageHandle(attacker, defender, attack, headshot)
	if isShowAttackLog {
		log.Info(attacker.GetName(), "attack", defender.GetName(), ",  calculate damage, original damage: ", oldattack, " after damage: ", attack)
	}

	var injuredTyp uint32

	if headshot {
		defender.DisposeSubHp(InjuredInfo{num: attack, injuredType: headShotAttack, isHeadshot: true, attackid: attacker.GetID(), attackdbid: attacker.GetDBID()})
		injuredTyp = headShotAttack
	} else {
		gunid := attacker.GetInUseWeapon()
		var attackType uint32
		if gunid == 0 {
			attackType = fistAttack
		} else {
			attackType = gunAttack
		}
		defender.DisposeSubHp(InjuredInfo{num: attack, injuredType: attackType, isHeadshot: false, attackid: attacker.GetID(), attackdbid: attacker.GetDBID()})
		injuredTyp = gunAttack
	}

	if attack > 0 {
		attacker.CastRPCToMe("DefenderSubHpNotify", defender.GetID(), injuredTyp)
	}

	defender.UnderAttack(attacker.GetID())
}

func damageHandle(attacker iAttacker, defender iDefender, attack uint32, headshot bool) uint32 {
	var ret uint32
	if headshot {
		prop := defender.GetHeadProp()
		if prop != nil {
			base, ok := excel.GetItem(uint64(prop.Itemid))
			if !ok {
				return ret
			}

			reduce := uint32(base.Reducerate) * attack / 100
			if prop.Reducedam > reduce {
				prop.Reducedam -= reduce
				ret += reduce
				if isShowAttackLog {
					log.Info(attacker.GetName(), "attack", defender.GetName(), ",  helmet reduce damage, ret: ", ret)
				}
			} else {
				ret += prop.Reducedam
				if isShowAttackLog {
					log.Info(attacker.GetName(), "attack", defender.GetName(), ", helmet crash after reduce damage, ret: ", ret)
				}
				prop.Baseid = 0
				prop.Itemid = 0
				prop.Reducedam = 0
			}
			defender.SetHeadProp(prop)
			defender.SetHeadPropDirty()
		}
	} else {
		prop := defender.GetBodyProp()
		if prop != nil {
			base, ok := excel.GetItem(uint64(prop.Baseid))
			if !ok {
				return ret
			}

			reduce := uint32(base.Reducerate) * attack / 100
			if prop.Reducedam > reduce {
				prop.Reducedam -= reduce
				ret += reduce
				if isShowAttackLog {
					log.Info(attacker.GetName(), "attack", defender.GetName(), ", body armor reduce damage, ret: ", ret)
				}
			} else {
				ret += prop.Reducedam
				if isShowAttackLog {
					log.Info(attacker.GetName(), "attack", defender.GetName(), ", body armor crash reduce damage, ret: ", ret)
				}
				prop.Baseid = 0
				prop.Reducedam = 0
			}
			defender.SetBodyProp(prop)
			defender.SetBodyPropDirty()
		}
	}

	return ret
}

func canReduceDam(baseid uint32) bool {
	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return false
	}

	rand.Seed(time.Now().UnixNano())
	rate := rand.Intn(100) + 1
	return uint64(rate) <= base.Reducerate
}
