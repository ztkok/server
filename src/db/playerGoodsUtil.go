package db

import (
	"encoding/json"
	"errors"
	"fmt"

	"protoMsg"

	log "github.com/cihub/seelog"
)

const (
	goodsPrefix = "OwnGoods"
)

// GoodsInfo 商品信息
type GoodsInfo struct {
	Id         uint32
	Time       int64 //获得时间
	EndTime    int64 //过期时间
	State      uint32
	Sum        uint32
	Used       uint32 // 0 卸下 1 装备
	Preference uint32 // 0 非偏好， 1 偏好
}

func (info *GoodsInfo) ToProto() *protoMsg.OwnGoodsItem {
	return &protoMsg.OwnGoodsItem{
		Id:         info.Id,
		State:      info.State,
		Used:       info.Used,
		Sum:        info.Sum,
		Preference: info.Preference,
		EndTime:    info.EndTime,
	}
}

// GoodsUtil
type GoodsUtil struct {
	uid uint64
}

func PlayerGoodsUtil(uid uint64) *GoodsUtil {
	return &GoodsUtil{
		uid: uid,
	}
}

func (p *GoodsUtil) key() string {
	return fmt.Sprintf("%s:%d", goodsPrefix, p.uid)
}

// GetAllGoodsInfo 获取所有商品信息
func (p *GoodsUtil) GetAllGoodsInfo() []*GoodsInfo {

	ret := make([]*GoodsInfo, 0)

	goodsList := hGetAll(p.key())

	for _, goods := range goodsList {

		var d *GoodsInfo
		if unErr := json.Unmarshal([]byte(goods), &d); unErr != nil {
			log.Error(unErr)
			continue
		}

		ret = append(ret, d)
	}

	return ret
}

// AddGoodsInfo 添加商品信息
func (p *GoodsUtil) AddGoodsInfo(info *GoodsInfo) bool {
	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("AddGoodsInfo error = ", err)
		return false
	}

	hSet(p.key(), info.Id, string(d))
	return true
}

// IsOwnGoods 是否已拥有商品
func (p *GoodsUtil) IsOwnGoods(id uint32) bool {
	return hExists(p.key(), id)
}

// IsGoodsEnough 判断商品数量是否大于等于一个最小值
func (p *GoodsUtil) IsGoodsEnough(id uint32, least uint32) bool {
	if least == 0 {
		return true
	}

	info, _ := p.GetGoodsInfo(id)
	if info != nil {
		return info.Sum >= least
	}

	return false
}

// DelGoods 删除商品
func (p *GoodsUtil) DelGoods(id uint32) {
	hDEL(p.key(), id)
}

var GoodsNotExist = errors.New("id is not exist")

// GetGoodsInfo 获取物品信息
func (p *GoodsUtil) GetGoodsInfo(id uint32) (*GoodsInfo, error) {

	if !hExists(p.key(), id) {
		return nil, GoodsNotExist
	}

	var d *GoodsInfo
	v := hGet(p.key(), id)
	if err := json.Unmarshal([]byte(v), &d); err != nil {
		log.Warn("GetGoodsInfo Failed to Unmarshal ", err)
		return nil, err
	}

	return d, nil

}
