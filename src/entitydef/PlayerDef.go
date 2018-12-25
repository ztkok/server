package entitydef

import "zeus/iserver"
import "protoMsg"

// PlayerDef 自动生成的属性包装代码
type PlayerDef struct {
	ip iserver.IEntityProps
}

// SetPropsSetter 设置接口
func (p *PlayerDef) SetPropsSetter(ip iserver.IEntityProps) {
	p.ip = ip
}

// SetBackPackProp 设置 BackPackProp
func (p *PlayerDef) SetBackPackProp(v *protoMsg.BackPackProp) {
	p.ip.SetProp("BackPackProp", v)
}

// SetBackPackPropDirty 设置BackPackProp被修改
func (p *PlayerDef) SetBackPackPropDirty() {
	p.ip.PropDirty("BackPackProp")
}

// GetBackPackProp 获取 BackPackProp
func (p *PlayerDef) GetBackPackProp() *protoMsg.BackPackProp {
	v := p.ip.GetProp("BackPackProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.BackPackProp)
}

// SetWeaponEquipInGame 设置 WeaponEquipInGame
func (p *PlayerDef) SetWeaponEquipInGame(v *protoMsg.WeaponInGame) {
	p.ip.SetProp("WeaponEquipInGame", v)
}

// SetWeaponEquipInGameDirty 设置WeaponEquipInGame被修改
func (p *PlayerDef) SetWeaponEquipInGameDirty() {
	p.ip.PropDirty("WeaponEquipInGame")
}

// GetWeaponEquipInGame 获取 WeaponEquipInGame
func (p *PlayerDef) GetWeaponEquipInGame() *protoMsg.WeaponInGame {
	v := p.ip.GetProp("WeaponEquipInGame")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.WeaponInGame)
}

// SetPicture 设置 Picture
func (p *PlayerDef) SetPicture(v string) {
	p.ip.SetProp("Picture", v)
}

// SetPictureDirty 设置Picture被修改
func (p *PlayerDef) SetPictureDirty() {
	p.ip.PropDirty("Picture")
}

// GetPicture 获取 Picture
func (p *PlayerDef) GetPicture() string {
	v := p.ip.GetProp("Picture")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetHeadProp 设置 HeadProp
func (p *PlayerDef) SetHeadProp(v *protoMsg.HeadProp) {
	p.ip.SetProp("HeadProp", v)
}

// SetHeadPropDirty 设置HeadProp被修改
func (p *PlayerDef) SetHeadPropDirty() {
	p.ip.PropDirty("HeadProp")
}

// GetHeadProp 获取 HeadProp
func (p *PlayerDef) GetHeadProp() *protoMsg.HeadProp {
	v := p.ip.GetProp("HeadProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.HeadProp)
}

// SetDayDiamGet 设置 DayDiamGet
func (p *PlayerDef) SetDayDiamGet(v uint64) {
	p.ip.SetProp("DayDiamGet", v)
}

// SetDayDiamGetDirty 设置DayDiamGet被修改
func (p *PlayerDef) SetDayDiamGetDirty() {
	p.ip.PropDirty("DayDiamGet")
}

// GetDayDiamGet 获取 DayDiamGet
func (p *PlayerDef) GetDayDiamGet() uint64 {
	v := p.ip.GetProp("DayDiamGet")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetMaxHP 设置 MaxHP
func (p *PlayerDef) SetMaxHP(v uint32) {
	p.ip.SetProp("MaxHP", v)
}

// SetMaxHPDirty 设置MaxHP被修改
func (p *PlayerDef) SetMaxHPDirty() {
	p.ip.PropDirty("MaxHP")
}

// GetMaxHP 获取 MaxHP
func (p *PlayerDef) GetMaxHP() uint32 {
	v := p.ip.GetProp("MaxHP")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetChracterMapDataInfo 设置 ChracterMapDataInfo
func (p *PlayerDef) SetChracterMapDataInfo(v *protoMsg.ChracterMapDataInfo) {
	p.ip.SetProp("ChracterMapDataInfo", v)
}

// SetChracterMapDataInfoDirty 设置ChracterMapDataInfo被修改
func (p *PlayerDef) SetChracterMapDataInfoDirty() {
	p.ip.PropDirty("ChracterMapDataInfo")
}

// GetChracterMapDataInfo 获取 ChracterMapDataInfo
func (p *PlayerDef) GetChracterMapDataInfo() *protoMsg.ChracterMapDataInfo {
	v := p.ip.GetProp("ChracterMapDataInfo")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.ChracterMapDataInfo)
}

// SetTodayOnlineTime 设置 TodayOnlineTime
func (p *PlayerDef) SetTodayOnlineTime(v int64) {
	p.ip.SetProp("TodayOnlineTime", v)
}

// SetTodayOnlineTimeDirty 设置TodayOnlineTime被修改
func (p *PlayerDef) SetTodayOnlineTimeDirty() {
	p.ip.PropDirty("TodayOnlineTime")
}

// GetTodayOnlineTime 获取 TodayOnlineTime
func (p *PlayerDef) GetTodayOnlineTime() int64 {
	v := p.ip.GetProp("TodayOnlineTime")
	if v == nil {
		return int64(0)
	}

	return v.(int64)
}

// SetName 设置 Name
func (p *PlayerDef) SetName(v string) {
	p.ip.SetProp("Name", v)
}

// SetNameDirty 设置Name被修改
func (p *PlayerDef) SetNameDirty() {
	p.ip.PropDirty("Name")
}

// GetName 获取 Name
func (p *PlayerDef) GetName() string {
	v := p.ip.GetProp("Name")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetIsWearingGilley 设置 IsWearingGilley
func (p *PlayerDef) SetIsWearingGilley(v uint32) {
	p.ip.SetProp("IsWearingGilley", v)
}

// SetIsWearingGilleyDirty 设置IsWearingGilley被修改
func (p *PlayerDef) SetIsWearingGilleyDirty() {
	p.ip.PropDirty("IsWearingGilley")
}

// GetIsWearingGilley 获取 IsWearingGilley
func (p *PlayerDef) GetIsWearingGilley() uint32 {
	v := p.ip.GetProp("IsWearingGilley")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetDayCoinGetTime 设置 DayCoinGetTime
func (p *PlayerDef) SetDayCoinGetTime(v uint64) {
	p.ip.SetProp("DayCoinGetTime", v)
}

// SetDayCoinGetTimeDirty 设置DayCoinGetTime被修改
func (p *PlayerDef) SetDayCoinGetTimeDirty() {
	p.ip.PropDirty("DayCoinGetTime")
}

// GetDayCoinGetTime 获取 DayCoinGetTime
func (p *PlayerDef) GetDayCoinGetTime() uint64 {
	v := p.ip.GetProp("DayCoinGetTime")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetSlowMove 设置 SlowMove
func (p *PlayerDef) SetSlowMove(v uint8) {
	p.ip.SetProp("SlowMove", v)
}

// SetSlowMoveDirty 设置SlowMove被修改
func (p *PlayerDef) SetSlowMoveDirty() {
	p.ip.PropDirty("SlowMove")
}

// GetSlowMove 获取 SlowMove
func (p *PlayerDef) GetSlowMove() uint8 {
	v := p.ip.GetProp("SlowMove")
	if v == nil {
		return uint8(0)
	}

	return v.(uint8)
}

// SetSubRotation1 设置 SubRotation1
func (p *PlayerDef) SetSubRotation1(v *protoMsg.Vector3) {
	p.ip.SetProp("SubRotation1", v)
}

// SetSubRotation1Dirty 设置SubRotation1被修改
func (p *PlayerDef) SetSubRotation1Dirty() {
	p.ip.PropDirty("SubRotation1")
}

// GetSubRotation1 获取 SubRotation1
func (p *PlayerDef) GetSubRotation1() *protoMsg.Vector3 {
	v := p.ip.GetProp("SubRotation1")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.Vector3)
}

// SetWatchable 设置 Watchable
func (p *PlayerDef) SetWatchable(v uint32) {
	p.ip.SetProp("Watchable", v)
}

// SetWatchableDirty 设置Watchable被修改
func (p *PlayerDef) SetWatchableDirty() {
	p.ip.PropDirty("Watchable")
}

// GetWatchable 获取 Watchable
func (p *PlayerDef) GetWatchable() uint32 {
	v := p.ip.GetProp("Watchable")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetPackWearInGame 设置 PackWearInGame
func (p *PlayerDef) SetPackWearInGame(v *protoMsg.WearInGame) {
	p.ip.SetProp("PackWearInGame", v)
}

// SetPackWearInGameDirty 设置PackWearInGame被修改
func (p *PlayerDef) SetPackWearInGameDirty() {
	p.ip.PropDirty("PackWearInGame")
}

// GetPackWearInGame 获取 PackWearInGame
func (p *PlayerDef) GetPackWearInGame() *protoMsg.WearInGame {
	v := p.ip.GetProp("PackWearInGame")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.WearInGame)
}

// SetGameEnter 设置 GameEnter
func (p *PlayerDef) SetGameEnter(v string) {
	p.ip.SetProp("GameEnter", v)
}

// SetGameEnterDirty 设置GameEnter被修改
func (p *PlayerDef) SetGameEnterDirty() {
	p.ip.PropDirty("GameEnter")
}

// GetGameEnter 获取 GameEnter
func (p *PlayerDef) GetGameEnter() string {
	v := p.ip.GetProp("GameEnter")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetIsInTank 设置 IsInTank
func (p *PlayerDef) SetIsInTank(v uint8) {
	p.ip.SetProp("IsInTank", v)
}

// SetIsInTankDirty 设置IsInTank被修改
func (p *PlayerDef) SetIsInTankDirty() {
	p.ip.PropDirty("IsInTank")
}

// GetIsInTank 获取 IsInTank
func (p *PlayerDef) GetIsInTank() uint8 {
	v := p.ip.GetProp("IsInTank")
	if v == nil {
		return uint8(0)
	}

	return v.(uint8)
}

// SetGoodsParachuteID 设置 GoodsParachuteID
func (p *PlayerDef) SetGoodsParachuteID(v uint32) {
	p.ip.SetProp("GoodsParachuteID", v)
}

// SetGoodsParachuteIDDirty 设置GoodsParachuteID被修改
func (p *PlayerDef) SetGoodsParachuteIDDirty() {
	p.ip.PropDirty("GoodsParachuteID")
}

// GetGoodsParachuteID 获取 GoodsParachuteID
func (p *PlayerDef) GetGoodsParachuteID() uint32 {
	v := p.ip.GetProp("GoodsParachuteID")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetVehicleProp 设置 VehicleProp
func (p *PlayerDef) SetVehicleProp(v *protoMsg.VehicleProp) {
	p.ip.SetProp("VehicleProp", v)
}

// SetVehiclePropDirty 设置VehicleProp被修改
func (p *PlayerDef) SetVehiclePropDirty() {
	p.ip.PropDirty("VehicleProp")
}

// GetVehicleProp 获取 VehicleProp
func (p *PlayerDef) GetVehicleProp() *protoMsg.VehicleProp {
	v := p.ip.GetProp("VehicleProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.VehicleProp)
}

// SetGoodsRoleModel 设置 GoodsRoleModel
func (p *PlayerDef) SetGoodsRoleModel(v uint32) {
	p.ip.SetProp("GoodsRoleModel", v)
}

// SetGoodsRoleModelDirty 设置GoodsRoleModel被修改
func (p *PlayerDef) SetGoodsRoleModelDirty() {
	p.ip.PropDirty("GoodsRoleModel")
}

// GetGoodsRoleModel 获取 GoodsRoleModel
func (p *PlayerDef) GetGoodsRoleModel() uint32 {
	v := p.ip.GetProp("GoodsRoleModel")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetDayBraveGetTime 设置 DayBraveGetTime
func (p *PlayerDef) SetDayBraveGetTime(v uint64) {
	p.ip.SetProp("DayBraveGetTime", v)
}

// SetDayBraveGetTimeDirty 设置DayBraveGetTime被修改
func (p *PlayerDef) SetDayBraveGetTimeDirty() {
	p.ip.PropDirty("DayBraveGetTime")
}

// GetDayBraveGetTime 获取 DayBraveGetTime
func (p *PlayerDef) GetDayBraveGetTime() uint64 {
	v := p.ip.GetProp("DayBraveGetTime")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetAccessToken 设置 AccessToken
func (p *PlayerDef) SetAccessToken(v string) {
	p.ip.SetProp("AccessToken", v)
}

// SetAccessTokenDirty 设置AccessToken被修改
func (p *PlayerDef) SetAccessTokenDirty() {
	p.ip.PropDirty("AccessToken")
}

// GetAccessToken 获取 AccessToken
func (p *PlayerDef) GetAccessToken() string {
	v := p.ip.GetProp("AccessToken")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetBraveCoin 设置 BraveCoin
func (p *PlayerDef) SetBraveCoin(v uint64) {
	p.ip.SetProp("BraveCoin", v)
}

// SetBraveCoinDirty 设置BraveCoin被修改
func (p *PlayerDef) SetBraveCoinDirty() {
	p.ip.PropDirty("BraveCoin")
}

// GetBraveCoin 获取 BraveCoin
func (p *PlayerDef) GetBraveCoin() uint64 {
	v := p.ip.GetProp("BraveCoin")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetQQVIP 设置 QQVIP
func (p *PlayerDef) SetQQVIP(v uint8) {
	p.ip.SetProp("QQVIP", v)
}

// SetQQVIPDirty 设置QQVIP被修改
func (p *PlayerDef) SetQQVIPDirty() {
	p.ip.PropDirty("QQVIP")
}

// GetQQVIP 获取 QQVIP
func (p *PlayerDef) GetQQVIP() uint8 {
	v := p.ip.GetProp("QQVIP")
	if v == nil {
		return uint8(0)
	}

	return v.(uint8)
}

// SetRoleModel 设置 RoleModel
func (p *PlayerDef) SetRoleModel(v uint32) {
	p.ip.SetProp("RoleModel", v)
}

// SetRoleModelDirty 设置RoleModel被修改
func (p *PlayerDef) SetRoleModelDirty() {
	p.ip.PropDirty("RoleModel")
}

// GetRoleModel 获取 RoleModel
func (p *PlayerDef) GetRoleModel() uint32 {
	v := p.ip.GetProp("RoleModel")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetDayDiamGetTime 设置 DayDiamGetTime
func (p *PlayerDef) SetDayDiamGetTime(v uint64) {
	p.ip.SetProp("DayDiamGetTime", v)
}

// SetDayDiamGetTimeDirty 设置DayDiamGetTime被修改
func (p *PlayerDef) SetDayDiamGetTimeDirty() {
	p.ip.PropDirty("DayDiamGetTime")
}

// GetDayDiamGetTime 获取 DayDiamGetTime
func (p *PlayerDef) GetDayDiamGetTime() uint64 {
	v := p.ip.GetProp("DayDiamGetTime")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetHP 设置 HP
func (p *PlayerDef) SetHP(v uint32) {
	p.ip.SetProp("HP", v)
}

// SetHPDirty 设置HP被修改
func (p *PlayerDef) SetHPDirty() {
	p.ip.PropDirty("HP")
}

// GetHP 获取 HP
func (p *PlayerDef) GetHP() uint32 {
	v := p.ip.GetProp("HP")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetAimPos 设置 AimPos
func (p *PlayerDef) SetAimPos(v *protoMsg.Vector3) {
	p.ip.SetProp("AimPos", v)
}

// SetAimPosDirty 设置AimPos被修改
func (p *PlayerDef) SetAimPosDirty() {
	p.ip.PropDirty("AimPos")
}

// GetAimPos 获取 AimPos
func (p *PlayerDef) GetAimPos() *protoMsg.Vector3 {
	v := p.ip.GetProp("AimPos")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.Vector3)
}

// SetLevel 设置 Level
func (p *PlayerDef) SetLevel(v uint32) {
	p.ip.SetProp("Level", v)
}

// SetLevelDirty 设置Level被修改
func (p *PlayerDef) SetLevelDirty() {
	p.ip.PropDirty("Level")
}

// GetLevel 获取 Level
func (p *PlayerDef) GetLevel() uint32 {
	v := p.ip.GetProp("Level")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetExp 设置 Exp
func (p *PlayerDef) SetExp(v uint32) {
	p.ip.SetProp("Exp", v)
}

// SetExpDirty 设置Exp被修改
func (p *PlayerDef) SetExpDirty() {
	p.ip.PropDirty("Exp")
}

// GetExp 获取 Exp
func (p *PlayerDef) GetExp() uint32 {
	v := p.ip.GetProp("Exp")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetPayOS 设置 PayOS
func (p *PlayerDef) SetPayOS(v string) {
	p.ip.SetProp("PayOS", v)
}

// SetPayOSDirty 设置PayOS被修改
func (p *PlayerDef) SetPayOSDirty() {
	p.ip.PropDirty("PayOS")
}

// GetPayOS 获取 PayOS
func (p *PlayerDef) GetPayOS() string {
	v := p.ip.GetProp("PayOS")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetCoin 设置 Coin
func (p *PlayerDef) SetCoin(v uint64) {
	p.ip.SetProp("Coin", v)
}

// SetCoinDirty 设置Coin被修改
func (p *PlayerDef) SetCoinDirty() {
	p.ip.PropDirty("Coin")
}

// GetCoin 获取 Coin
func (p *PlayerDef) GetCoin() uint64 {
	v := p.ip.GetProp("Coin")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetBodyProp 设置 BodyProp
func (p *PlayerDef) SetBodyProp(v *protoMsg.BodyProp) {
	p.ip.SetProp("BodyProp", v)
}

// SetBodyPropDirty 设置BodyProp被修改
func (p *PlayerDef) SetBodyPropDirty() {
	p.ip.PropDirty("BodyProp")
}

// GetBodyProp 获取 BodyProp
func (p *PlayerDef) GetBodyProp() *protoMsg.BodyProp {
	v := p.ip.GetProp("BodyProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.BodyProp)
}

// SetFastrun 设置 Fastrun
func (p *PlayerDef) SetFastrun(v uint8) {
	p.ip.SetProp("Fastrun", v)
}

// SetFastrunDirty 设置Fastrun被修改
func (p *PlayerDef) SetFastrunDirty() {
	p.ip.PropDirty("Fastrun")
}

// GetFastrun 获取 Fastrun
func (p *PlayerDef) GetFastrun() uint8 {
	v := p.ip.GetProp("Fastrun")
	if v == nil {
		return uint8(0)
	}

	return v.(uint8)
}

// SetSpeedRate 设置 SpeedRate
func (p *PlayerDef) SetSpeedRate(v float32) {
	p.ip.SetProp("SpeedRate", v)
}

// SetSpeedRateDirty 设置SpeedRate被修改
func (p *PlayerDef) SetSpeedRateDirty() {
	p.ip.PropDirty("SpeedRate")
}

// GetSpeedRate 获取 SpeedRate
func (p *PlayerDef) GetSpeedRate() float32 {
	v := p.ip.GetProp("SpeedRate")
	if v == nil {
		return float32(0)
	}

	return v.(float32)
}

// SetNickName 设置 NickName
func (p *PlayerDef) SetNickName(v string) {
	p.ip.SetProp("NickName", v)
}

// SetNickNameDirty 设置NickName被修改
func (p *PlayerDef) SetNickNameDirty() {
	p.ip.PropDirty("NickName")
}

// GetNickName 获取 NickName
func (p *PlayerDef) GetNickName() string {
	v := p.ip.GetProp("NickName")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetLoginTime 设置 LoginTime
func (p *PlayerDef) SetLoginTime(v int64) {
	p.ip.SetProp("LoginTime", v)
}

// SetLoginTimeDirty 设置LoginTime被修改
func (p *PlayerDef) SetLoginTimeDirty() {
	p.ip.PropDirty("LoginTime")
}

// GetLoginTime 获取 LoginTime
func (p *PlayerDef) GetLoginTime() int64 {
	v := p.ip.GetProp("LoginTime")
	if v == nil {
		return int64(0)
	}

	return v.(int64)
}

// SetOutsideWeapon 设置 OutsideWeapon
func (p *PlayerDef) SetOutsideWeapon(v uint32) {
	p.ip.SetProp("OutsideWeapon", v)
}

// SetOutsideWeaponDirty 设置OutsideWeapon被修改
func (p *PlayerDef) SetOutsideWeaponDirty() {
	p.ip.PropDirty("OutsideWeapon")
}

// GetOutsideWeapon 获取 OutsideWeapon
func (p *PlayerDef) GetOutsideWeapon() uint32 {
	v := p.ip.GetProp("OutsideWeapon")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetSubRotation2 设置 SubRotation2
func (p *PlayerDef) SetSubRotation2(v *protoMsg.Vector3) {
	p.ip.SetProp("SubRotation2", v)
}

// SetSubRotation2Dirty 设置SubRotation2被修改
func (p *PlayerDef) SetSubRotation2Dirty() {
	p.ip.PropDirty("SubRotation2")
}

// GetSubRotation2 获取 SubRotation2
func (p *PlayerDef) GetSubRotation2() *protoMsg.Vector3 {
	v := p.ip.GetProp("SubRotation2")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.Vector3)
}

// SetGender 设置 Gender
func (p *PlayerDef) SetGender(v string) {
	p.ip.SetProp("Gender", v)
}

// SetGenderDirty 设置Gender被修改
func (p *PlayerDef) SetGenderDirty() {
	p.ip.PropDirty("Gender")
}

// GetGender 获取 Gender
func (p *PlayerDef) GetGender() string {
	v := p.ip.GetProp("Gender")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetVeteran 设置 Veteran
func (p *PlayerDef) SetVeteran(v uint32) {
	p.ip.SetProp("Veteran", v)
}

// SetVeteranDirty 设置Veteran被修改
func (p *PlayerDef) SetVeteranDirty() {
	p.ip.PropDirty("Veteran")
}

// GetVeteran 获取 Veteran
func (p *PlayerDef) GetVeteran() uint32 {
	v := p.ip.GetProp("Veteran")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetFriendsNum 设置 FriendsNum
func (p *PlayerDef) SetFriendsNum(v uint32) {
	p.ip.SetProp("FriendsNum", v)
}

// SetFriendsNumDirty 设置FriendsNum被修改
func (p *PlayerDef) SetFriendsNumDirty() {
	p.ip.PropDirty("FriendsNum")
}

// GetFriendsNum 获取 FriendsNum
func (p *PlayerDef) GetFriendsNum() uint32 {
	v := p.ip.GetProp("FriendsNum")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetSkillEffectList 设置 SkillEffectList
func (p *PlayerDef) SetSkillEffectList(v *protoMsg.SkillEffectList) {
	p.ip.SetProp("SkillEffectList", v)
}

// SetSkillEffectListDirty 设置SkillEffectList被修改
func (p *PlayerDef) SetSkillEffectListDirty() {
	p.ip.PropDirty("SkillEffectList")
}

// GetSkillEffectList 获取 SkillEffectList
func (p *PlayerDef) GetSkillEffectList() *protoMsg.SkillEffectList {
	v := p.ip.GetProp("SkillEffectList")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.SkillEffectList)
}

// SetBrick 设置 Brick
func (p *PlayerDef) SetBrick(v uint64) {
	p.ip.SetProp("Brick", v)
}

// SetBrickDirty 设置Brick被修改
func (p *PlayerDef) SetBrickDirty() {
	p.ip.PropDirty("Brick")
}

// GetBrick 获取 Brick
func (p *PlayerDef) GetBrick() uint64 {
	v := p.ip.GetProp("Brick")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetOnlineTime 设置 OnlineTime
func (p *PlayerDef) SetOnlineTime(v int64) {
	p.ip.SetProp("OnlineTime", v)
}

// SetOnlineTimeDirty 设置OnlineTime被修改
func (p *PlayerDef) SetOnlineTimeDirty() {
	p.ip.PropDirty("OnlineTime")
}

// GetOnlineTime 获取 OnlineTime
func (p *PlayerDef) GetOnlineTime() int64 {
	v := p.ip.GetProp("OnlineTime")
	if v == nil {
		return int64(0)
	}

	return v.(int64)
}

// SetBrick2 设置 Brick2
func (p *PlayerDef) SetBrick2(v uint64) {
	p.ip.SetProp("Brick2", v)
}

// SetBrick2Dirty 设置Brick2被修改
func (p *PlayerDef) SetBrick2Dirty() {
	p.ip.PropDirty("Brick2")
}

// GetBrick2 获取 Brick2
func (p *PlayerDef) GetBrick2() uint64 {
	v := p.ip.GetProp("Brick2")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetMedalData 设置 MedalData
func (p *PlayerDef) SetMedalData(v *protoMsg.MedalDataList) {
	p.ip.SetProp("MedalData", v)
}

// SetMedalDataDirty 设置MedalData被修改
func (p *PlayerDef) SetMedalDataDirty() {
	p.ip.PropDirty("MedalData")
}

// GetMedalData 获取 MedalData
func (p *PlayerDef) GetMedalData() *protoMsg.MedalDataList {
	v := p.ip.GetProp("MedalData")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.MedalDataList)
}

// SetParachuteID 设置 ParachuteID
func (p *PlayerDef) SetParachuteID(v uint32) {
	p.ip.SetProp("ParachuteID", v)
}

// SetParachuteIDDirty 设置ParachuteID被修改
func (p *PlayerDef) SetParachuteIDDirty() {
	p.ip.PropDirty("ParachuteID")
}

// GetParachuteID 获取 ParachuteID
func (p *PlayerDef) GetParachuteID() uint32 {
	v := p.ip.GetProp("ParachuteID")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetComradePoints 设置 ComradePoints
func (p *PlayerDef) SetComradePoints(v uint64) {
	p.ip.SetProp("ComradePoints", v)
}

// SetComradePointsDirty 设置ComradePoints被修改
func (p *PlayerDef) SetComradePointsDirty() {
	p.ip.PropDirty("ComradePoints")
}

// GetComradePoints 获取 ComradePoints
func (p *PlayerDef) GetComradePoints() uint64 {
	v := p.ip.GetProp("ComradePoints")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetTheme 设置 Theme
func (p *PlayerDef) SetTheme(v uint32) {
	p.ip.SetProp("Theme", v)
}

// SetThemeDirty 设置Theme被修改
func (p *PlayerDef) SetThemeDirty() {
	p.ip.PropDirty("Theme")
}

// GetTheme 获取 Theme
func (p *PlayerDef) GetTheme() uint32 {
	v := p.ip.GetProp("Theme")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetBrick1 设置 Brick1
func (p *PlayerDef) SetBrick1(v uint64) {
	p.ip.SetProp("Brick1", v)
}

// SetBrick1Dirty 设置Brick1被修改
func (p *PlayerDef) SetBrick1Dirty() {
	p.ip.PropDirty("Brick1")
}

// GetBrick1 获取 Brick1
func (p *PlayerDef) GetBrick1() uint64 {
	v := p.ip.GetProp("Brick1")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetLogoutTime 设置 LogoutTime
func (p *PlayerDef) SetLogoutTime(v int64) {
	p.ip.SetProp("LogoutTime", v)
}

// SetLogoutTimeDirty 设置LogoutTime被修改
func (p *PlayerDef) SetLogoutTimeDirty() {
	p.ip.PropDirty("LogoutTime")
}

// GetLogoutTime 获取 LogoutTime
func (p *PlayerDef) GetLogoutTime() int64 {
	v := p.ip.GetProp("LogoutTime")
	if v == nil {
		return int64(0)
	}

	return v.(int64)
}

// SetMvSpeed 设置 MvSpeed
func (p *PlayerDef) SetMvSpeed(v float32) {
	p.ip.SetProp("MvSpeed", v)
}

// SetMvSpeedDirty 设置MvSpeed被修改
func (p *PlayerDef) SetMvSpeedDirty() {
	p.ip.PropDirty("MvSpeed")
}

// GetMvSpeed 获取 MvSpeed
func (p *PlayerDef) GetMvSpeed() float32 {
	v := p.ip.GetProp("MvSpeed")
	if v == nil {
		return float32(0)
	}

	return v.(float32)
}

// SetDayBraveGet 设置 DayBraveGet
func (p *PlayerDef) SetDayBraveGet(v uint64) {
	p.ip.SetProp("DayBraveGet", v)
}

// SetDayBraveGetDirty 设置DayBraveGet被修改
func (p *PlayerDef) SetDayBraveGetDirty() {
	p.ip.PropDirty("DayBraveGet")
}

// GetDayBraveGet 获取 DayBraveGet
func (p *PlayerDef) GetDayBraveGet() uint64 {
	v := p.ip.GetProp("DayBraveGet")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetGunSight 设置 GunSight
func (p *PlayerDef) SetGunSight(v uint64) {
	p.ip.SetProp("GunSight", v)
}

// SetGunSightDirty 设置GunSight被修改
func (p *PlayerDef) SetGunSightDirty() {
	p.ip.PropDirty("GunSight")
}

// GetGunSight 获取 GunSight
func (p *PlayerDef) GetGunSight() uint64 {
	v := p.ip.GetProp("GunSight")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetGamerType 设置 GamerType
func (p *PlayerDef) SetGamerType(v uint32) {
	p.ip.SetProp("GamerType", v)
}

// SetGamerTypeDirty 设置GamerType被修改
func (p *PlayerDef) SetGamerTypeDirty() {
	p.ip.PropDirty("GamerType")
}

// GetGamerType 获取 GamerType
func (p *PlayerDef) GetGamerType() uint32 {
	v := p.ip.GetProp("GamerType")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetPlayerLogin 设置 PlayerLogin
func (p *PlayerDef) SetPlayerLogin(v *protoMsg.PlayerLogin) {
	p.ip.SetProp("PlayerLogin", v)
}

// SetPlayerLoginDirty 设置PlayerLogin被修改
func (p *PlayerDef) SetPlayerLoginDirty() {
	p.ip.PropDirty("PlayerLogin")
}

// GetPlayerLogin 获取 PlayerLogin
func (p *PlayerDef) GetPlayerLogin() *protoMsg.PlayerLogin {
	v := p.ip.GetProp("PlayerLogin")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.PlayerLogin)
}

// SetDayCoinGet 设置 DayCoinGet
func (p *PlayerDef) SetDayCoinGet(v uint64) {
	p.ip.SetProp("DayCoinGet", v)
}

// SetDayCoinGetDirty 设置DayCoinGet被修改
func (p *PlayerDef) SetDayCoinGetDirty() {
	p.ip.PropDirty("DayCoinGet")
}

// GetDayCoinGet 获取 DayCoinGet
func (p *PlayerDef) GetDayCoinGet() uint64 {
	v := p.ip.GetProp("DayCoinGet")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetDiam 设置 Diam
func (p *PlayerDef) SetDiam(v uint64) {
	p.ip.SetProp("Diam", v)
}

// SetDiamDirty 设置Diam被修改
func (p *PlayerDef) SetDiamDirty() {
	p.ip.PropDirty("Diam")
}

// GetDiam 获取 Diam
func (p *PlayerDef) GetDiam() uint64 {
	v := p.ip.GetProp("Diam")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetHeadWearInGame 设置 HeadWearInGame
func (p *PlayerDef) SetHeadWearInGame(v *protoMsg.WearInGame) {
	p.ip.SetProp("HeadWearInGame", v)
}

// SetHeadWearInGameDirty 设置HeadWearInGame被修改
func (p *PlayerDef) SetHeadWearInGameDirty() {
	p.ip.PropDirty("HeadWearInGame")
}

// GetHeadWearInGame 获取 HeadWearInGame
func (p *PlayerDef) GetHeadWearInGame() *protoMsg.WearInGame {
	v := p.ip.GetProp("HeadWearInGame")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.WearInGame)
}

type IPlayerDef interface {
	SetBackPackProp(v *protoMsg.BackPackProp)
	SetBackPackPropDirty()
	GetBackPackProp() *protoMsg.BackPackProp
	SetWeaponEquipInGame(v *protoMsg.WeaponInGame)
	SetWeaponEquipInGameDirty()
	GetWeaponEquipInGame() *protoMsg.WeaponInGame
	SetPicture(v string)
	SetPictureDirty()
	GetPicture() string
	SetHeadProp(v *protoMsg.HeadProp)
	SetHeadPropDirty()
	GetHeadProp() *protoMsg.HeadProp
	SetDayDiamGet(v uint64)
	SetDayDiamGetDirty()
	GetDayDiamGet() uint64
	SetMaxHP(v uint32)
	SetMaxHPDirty()
	GetMaxHP() uint32
	SetChracterMapDataInfo(v *protoMsg.ChracterMapDataInfo)
	SetChracterMapDataInfoDirty()
	GetChracterMapDataInfo() *protoMsg.ChracterMapDataInfo
	SetTodayOnlineTime(v int64)
	SetTodayOnlineTimeDirty()
	GetTodayOnlineTime() int64
	SetName(v string)
	SetNameDirty()
	GetName() string
	SetIsWearingGilley(v uint32)
	SetIsWearingGilleyDirty()
	GetIsWearingGilley() uint32
	SetDayCoinGetTime(v uint64)
	SetDayCoinGetTimeDirty()
	GetDayCoinGetTime() uint64
	SetSlowMove(v uint8)
	SetSlowMoveDirty()
	GetSlowMove() uint8
	SetSubRotation1(v *protoMsg.Vector3)
	SetSubRotation1Dirty()
	GetSubRotation1() *protoMsg.Vector3
	SetWatchable(v uint32)
	SetWatchableDirty()
	GetWatchable() uint32
	SetPackWearInGame(v *protoMsg.WearInGame)
	SetPackWearInGameDirty()
	GetPackWearInGame() *protoMsg.WearInGame
	SetGameEnter(v string)
	SetGameEnterDirty()
	GetGameEnter() string
	SetIsInTank(v uint8)
	SetIsInTankDirty()
	GetIsInTank() uint8
	SetGoodsParachuteID(v uint32)
	SetGoodsParachuteIDDirty()
	GetGoodsParachuteID() uint32
	SetVehicleProp(v *protoMsg.VehicleProp)
	SetVehiclePropDirty()
	GetVehicleProp() *protoMsg.VehicleProp
	SetGoodsRoleModel(v uint32)
	SetGoodsRoleModelDirty()
	GetGoodsRoleModel() uint32
	SetDayBraveGetTime(v uint64)
	SetDayBraveGetTimeDirty()
	GetDayBraveGetTime() uint64
	SetAccessToken(v string)
	SetAccessTokenDirty()
	GetAccessToken() string
	SetBraveCoin(v uint64)
	SetBraveCoinDirty()
	GetBraveCoin() uint64
	SetQQVIP(v uint8)
	SetQQVIPDirty()
	GetQQVIP() uint8
	SetRoleModel(v uint32)
	SetRoleModelDirty()
	GetRoleModel() uint32
	SetDayDiamGetTime(v uint64)
	SetDayDiamGetTimeDirty()
	GetDayDiamGetTime() uint64
	SetHP(v uint32)
	SetHPDirty()
	GetHP() uint32
	SetAimPos(v *protoMsg.Vector3)
	SetAimPosDirty()
	GetAimPos() *protoMsg.Vector3
	SetLevel(v uint32)
	SetLevelDirty()
	GetLevel() uint32
	SetExp(v uint32)
	SetExpDirty()
	GetExp() uint32
	SetPayOS(v string)
	SetPayOSDirty()
	GetPayOS() string
	SetCoin(v uint64)
	SetCoinDirty()
	GetCoin() uint64
	SetBodyProp(v *protoMsg.BodyProp)
	SetBodyPropDirty()
	GetBodyProp() *protoMsg.BodyProp
	SetFastrun(v uint8)
	SetFastrunDirty()
	GetFastrun() uint8
	SetSpeedRate(v float32)
	SetSpeedRateDirty()
	GetSpeedRate() float32
	SetNickName(v string)
	SetNickNameDirty()
	GetNickName() string
	SetLoginTime(v int64)
	SetLoginTimeDirty()
	GetLoginTime() int64
	SetOutsideWeapon(v uint32)
	SetOutsideWeaponDirty()
	GetOutsideWeapon() uint32
	SetSubRotation2(v *protoMsg.Vector3)
	SetSubRotation2Dirty()
	GetSubRotation2() *protoMsg.Vector3
	SetGender(v string)
	SetGenderDirty()
	GetGender() string
	SetVeteran(v uint32)
	SetVeteranDirty()
	GetVeteran() uint32
	SetFriendsNum(v uint32)
	SetFriendsNumDirty()
	GetFriendsNum() uint32
	SetSkillEffectList(v *protoMsg.SkillEffectList)
	SetSkillEffectListDirty()
	GetSkillEffectList() *protoMsg.SkillEffectList
	SetBrick(v uint64)
	SetBrickDirty()
	GetBrick() uint64
	SetOnlineTime(v int64)
	SetOnlineTimeDirty()
	GetOnlineTime() int64
	SetBrick2(v uint64)
	SetBrick2Dirty()
	GetBrick2() uint64
	SetMedalData(v *protoMsg.MedalDataList)
	SetMedalDataDirty()
	GetMedalData() *protoMsg.MedalDataList
	SetParachuteID(v uint32)
	SetParachuteIDDirty()
	GetParachuteID() uint32
	SetComradePoints(v uint64)
	SetComradePointsDirty()
	GetComradePoints() uint64
	SetTheme(v uint32)
	SetThemeDirty()
	GetTheme() uint32
	SetBrick1(v uint64)
	SetBrick1Dirty()
	GetBrick1() uint64
	SetLogoutTime(v int64)
	SetLogoutTimeDirty()
	GetLogoutTime() int64
	SetMvSpeed(v float32)
	SetMvSpeedDirty()
	GetMvSpeed() float32
	SetDayBraveGet(v uint64)
	SetDayBraveGetDirty()
	GetDayBraveGet() uint64
	SetGunSight(v uint64)
	SetGunSightDirty()
	GetGunSight() uint64
	SetGamerType(v uint32)
	SetGamerTypeDirty()
	GetGamerType() uint32
	SetPlayerLogin(v *protoMsg.PlayerLogin)
	SetPlayerLoginDirty()
	GetPlayerLogin() *protoMsg.PlayerLogin
	SetDayCoinGet(v uint64)
	SetDayCoinGetDirty()
	GetDayCoinGet() uint64
	SetDiam(v uint64)
	SetDiamDirty()
	GetDiam() uint64
	SetHeadWearInGame(v *protoMsg.WearInGame)
	SetHeadWearInGameDirty()
	GetHeadWearInGame() *protoMsg.WearInGame
}
