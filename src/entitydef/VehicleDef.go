package entitydef

import "zeus/iserver"
import "protoMsg"

// VehicleDef 自动生成的属性包装代码
type VehicleDef struct {
	ip iserver.IEntityProps
}

// SetPropsSetter 设置接口
func (p *VehicleDef) SetPropsSetter(ip iserver.IEntityProps) {
	p.ip = ip
}

// SetVehicleProp 设置 VehicleProp
func (p *VehicleDef) SetVehicleProp(v *protoMsg.VehicleProp) {
	p.ip.SetProp("VehicleProp", v)
}

// SetVehiclePropDirty 设置VehicleProp被修改
func (p *VehicleDef) SetVehiclePropDirty() {
	p.ip.PropDirty("VehicleProp")
}

// GetVehicleProp 获取 VehicleProp
func (p *VehicleDef) GetVehicleProp() *protoMsg.VehicleProp {
	v := p.ip.GetProp("VehicleProp")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.VehicleProp)
}

// SetVehiclePhysics 设置 VehiclePhysics
func (p *VehicleDef) SetVehiclePhysics(v *protoMsg.VehiclePhysics) {
	p.ip.SetProp("VehiclePhysics", v)
}

// SetVehiclePhysicsDirty 设置VehiclePhysics被修改
func (p *VehicleDef) SetVehiclePhysicsDirty() {
	p.ip.PropDirty("VehiclePhysics")
}

// GetVehiclePhysics 获取 VehiclePhysics
func (p *VehicleDef) GetVehiclePhysics() *protoMsg.VehiclePhysics {
	v := p.ip.GetProp("VehiclePhysics")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.VehiclePhysics)
}

// Setownerid 设置 ownerid
func (p *VehicleDef) Setownerid(v uint64) {
	p.ip.SetProp("ownerid", v)
}

// SetowneridDirty 设置ownerid被修改
func (p *VehicleDef) SetowneridDirty() {
	p.ip.PropDirty("ownerid")
}

// Getownerid 获取 ownerid
func (p *VehicleDef) Getownerid() uint64 {
	v := p.ip.GetProp("ownerid")
	if v == nil {
		return uint64(0)
	}

	return v.(uint64)
}

// SetSubRotation1 设置 SubRotation1
func (p *VehicleDef) SetSubRotation1(v *protoMsg.Vector3) {
	p.ip.SetProp("SubRotation1", v)
}

// SetSubRotation1Dirty 设置SubRotation1被修改
func (p *VehicleDef) SetSubRotation1Dirty() {
	p.ip.PropDirty("SubRotation1")
}

// GetSubRotation1 获取 SubRotation1
func (p *VehicleDef) GetSubRotation1() *protoMsg.Vector3 {
	v := p.ip.GetProp("SubRotation1")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.Vector3)
}

// SetSubRotation2 设置 SubRotation2
func (p *VehicleDef) SetSubRotation2(v *protoMsg.Vector3) {
	p.ip.SetProp("SubRotation2", v)
}

// SetSubRotation2Dirty 设置SubRotation2被修改
func (p *VehicleDef) SetSubRotation2Dirty() {
	p.ip.PropDirty("SubRotation2")
}

// GetSubRotation2 获取 SubRotation2
func (p *VehicleDef) GetSubRotation2() *protoMsg.Vector3 {
	v := p.ip.GetProp("SubRotation2")
	if v == nil {
		return nil
	}

	return v.(*protoMsg.Vector3)
}

type IVehicleDef interface {
	SetVehicleProp(v *protoMsg.VehicleProp)
	SetVehiclePropDirty()
	GetVehicleProp() *protoMsg.VehicleProp
	SetVehiclePhysics(v *protoMsg.VehiclePhysics)
	SetVehiclePhysicsDirty()
	GetVehiclePhysics() *protoMsg.VehiclePhysics
	Setownerid(v uint64)
	SetowneridDirty()
	Getownerid() uint64
	SetSubRotation1(v *protoMsg.Vector3)
	SetSubRotation1Dirty()
	GetSubRotation1() *protoMsg.Vector3
	SetSubRotation2(v *protoMsg.Vector3)
	SetSubRotation2Dirty()
	GetSubRotation2() *protoMsg.Vector3
}
