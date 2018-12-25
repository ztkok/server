package entitydef

import "zeus/iserver"
import "protoMsg"

// AIDef 自动生成的属性包装代码
type AIDef struct {
	ip iserver.IEntityProps
}

// SetPropsSetter 设置接口
func (p *AIDef) SetPropsSetter(ip iserver.IEntityProps) {
	p.ip = ip
}

// SetIsWearingGilley 设置 IsWearingGilley
func (p *AIDef) SetIsWearingGilley(v uint32) {
	p.ip.SetProp("IsWearingGilley", v)
}

// SetIsWearingGilleyDirty 设置IsWearingGilley被修改
func (p *AIDef) SetIsWearingGilleyDirty() {
	p.ip.PropDirty("IsWearingGilley")
}

// GetIsWearingGilley 获取 IsWearingGilley
func (p *AIDef) GetIsWearingGilley() uint32 {
	v := p.ip.GetProp("IsWearingGilley")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetHeadProp 设置 HeadProp
func (p *AIDef) SetHeadProp(v *protoMsg.HeadProp) {
	p.ip.SetProp("HeadProp", v)
}

// SetHeadPropDirty 设置HeadProp被修改
func (p *AIDef) SetHeadPropDirty() {
	p.ip.PropDirty("HeadProp")
}

// GetHeadProp 获取 HeadProp
func (p *AIDef) GetHeadProp() *protoMsg.HeadProp {
	v := p.ip.GetProp("HeadProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.HeadProp)
}

// SetRole 设置 Role
func (p *AIDef) SetRole(v uint32) {
	p.ip.SetProp("Role", v)
}

// SetRoleDirty 设置Role被修改
func (p *AIDef) SetRoleDirty() {
	p.ip.PropDirty("Role")
}

// GetRole 获取 Role
func (p *AIDef) GetRole() uint32 {
	v := p.ip.GetProp("Role")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetMaxHP 设置 MaxHP
func (p *AIDef) SetMaxHP(v uint32) {
	p.ip.SetProp("MaxHP", v)
}

// SetMaxHPDirty 设置MaxHP被修改
func (p *AIDef) SetMaxHPDirty() {
	p.ip.PropDirty("MaxHP")
}

// GetMaxHP 获取 MaxHP
func (p *AIDef) GetMaxHP() uint32 {
	v := p.ip.GetProp("MaxHP")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetBodyProp 设置 BodyProp
func (p *AIDef) SetBodyProp(v *protoMsg.BodyProp) {
	p.ip.SetProp("BodyProp", v)
}

// SetBodyPropDirty 设置BodyProp被修改
func (p *AIDef) SetBodyPropDirty() {
	p.ip.PropDirty("BodyProp")
}

// GetBodyProp 获取 BodyProp
func (p *AIDef) GetBodyProp() *protoMsg.BodyProp {
	v := p.ip.GetProp("BodyProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.BodyProp)
}

// SetName 设置 Name
func (p *AIDef) SetName(v string) {
	p.ip.SetProp("Name", v)
}

// SetNameDirty 设置Name被修改
func (p *AIDef) SetNameDirty() {
	p.ip.PropDirty("Name")
}

// GetName 获取 Name
func (p *AIDef) GetName() string {
	v := p.ip.GetProp("Name")
	if v == nil {
		return ""
	}

	return v.(string)
}

// SetGamerType 设置 GamerType
func (p *AIDef) SetGamerType(v uint32) {
	p.ip.SetProp("GamerType", v)
}

// SetGamerTypeDirty 设置GamerType被修改
func (p *AIDef) SetGamerTypeDirty() {
	p.ip.PropDirty("GamerType")
}

// GetGamerType 获取 GamerType
func (p *AIDef) GetGamerType() uint32 {
	v := p.ip.GetProp("GamerType")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetRoleModel 设置 RoleModel
func (p *AIDef) SetRoleModel(v uint32) {
	p.ip.SetProp("RoleModel", v)
}

// SetRoleModelDirty 设置RoleModel被修改
func (p *AIDef) SetRoleModelDirty() {
	p.ip.PropDirty("RoleModel")
}

// GetRoleModel 获取 RoleModel
func (p *AIDef) GetRoleModel() uint32 {
	v := p.ip.GetProp("RoleModel")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetChracterMapDataInfo 设置 ChracterMapDataInfo
func (p *AIDef) SetChracterMapDataInfo(v *protoMsg.ChracterMapDataInfo) {
	p.ip.SetProp("ChracterMapDataInfo", v)
}

// SetChracterMapDataInfoDirty 设置ChracterMapDataInfo被修改
func (p *AIDef) SetChracterMapDataInfoDirty() {
	p.ip.PropDirty("ChracterMapDataInfo")
}

// GetChracterMapDataInfo 获取 ChracterMapDataInfo
func (p *AIDef) GetChracterMapDataInfo() *protoMsg.ChracterMapDataInfo {
	v := p.ip.GetProp("ChracterMapDataInfo")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.ChracterMapDataInfo)
}

// SetHP 设置 HP
func (p *AIDef) SetHP(v uint32) {
	p.ip.SetProp("HP", v)
}

// SetHPDirty 设置HP被修改
func (p *AIDef) SetHPDirty() {
	p.ip.PropDirty("HP")
}

// GetHP 获取 HP
func (p *AIDef) GetHP() uint32 {
	v := p.ip.GetProp("HP")
	if v == nil {
		return uint32(0)
	}

	return v.(uint32)
}

// SetVehicleProp 设置 VehicleProp
func (p *AIDef) SetVehicleProp(v *protoMsg.VehicleProp) {
	p.ip.SetProp("VehicleProp", v)
}

// SetVehiclePropDirty 设置VehicleProp被修改
func (p *AIDef) SetVehiclePropDirty() {
	p.ip.PropDirty("VehicleProp")
}

// GetVehicleProp 获取 VehicleProp
func (p *AIDef) GetVehicleProp() *protoMsg.VehicleProp {
	v := p.ip.GetProp("VehicleProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.VehicleProp)
}

// SetBackPackProp 设置 BackPackProp
func (p *AIDef) SetBackPackProp(v *protoMsg.BackPackProp) {
	p.ip.SetProp("BackPackProp", v)
}

// SetBackPackPropDirty 设置BackPackProp被修改
func (p *AIDef) SetBackPackPropDirty() {
	p.ip.PropDirty("BackPackProp")
}

// GetBackPackProp 获取 BackPackProp
func (p *AIDef) GetBackPackProp() *protoMsg.BackPackProp {
	v := p.ip.GetProp("BackPackProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.BackPackProp)
}

type IAIDef interface {
	SetIsWearingGilley(v uint32)
	SetIsWearingGilleyDirty()
	GetIsWearingGilley() uint32
	SetHeadProp(v *protoMsg.HeadProp)
	SetHeadPropDirty()
	GetHeadProp() *protoMsg.HeadProp
	SetRole(v uint32)
	SetRoleDirty()
	GetRole() uint32
	SetMaxHP(v uint32)
	SetMaxHPDirty()
	GetMaxHP() uint32
	SetBodyProp(v *protoMsg.BodyProp)
	SetBodyPropDirty()
	GetBodyProp() *protoMsg.BodyProp
	SetName(v string)
	SetNameDirty()
	GetName() string
	SetGamerType(v uint32)
	SetGamerTypeDirty()
	GetGamerType() uint32
	SetRoleModel(v uint32)
	SetRoleModelDirty()
	GetRoleModel() uint32
	SetChracterMapDataInfo(v *protoMsg.ChracterMapDataInfo)
	SetChracterMapDataInfoDirty()
	GetChracterMapDataInfo() *protoMsg.ChracterMapDataInfo
	SetHP(v uint32)
	SetHPDirty()
	GetHP() uint32
	SetVehicleProp(v *protoMsg.VehicleProp)
	SetVehiclePropDirty()
	GetVehicleProp() *protoMsg.VehicleProp
	SetBackPackProp(v *protoMsg.BackPackProp)
	SetBackPackPropDirty()
	GetBackPackProp() *protoMsg.BackPackProp
}
