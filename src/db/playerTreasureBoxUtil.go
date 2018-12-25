package db

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/cihub/seelog"
)

const (
	treasureBoxPrefix = "TreasureBoxInfoTable"
)

// TreasureBoxUtil
type TreasureBoxUtil struct {
	uid uint64
	id  uint32 //宝箱id
}

// TreasureBoxInfo 宝箱信息
type TreasureBoxInfo struct {
	Id         uint32 //宝箱id
	ActStartTm int64  //活动开始时间
	Time       int64  //上次领取时间
	TotalNum   uint32 //总领取次数
	CycNum     uint32 //周期领取次数
}

func (treasureBoxInfo *TreasureBoxInfo) String() string {
	return fmt.Sprintf("%+v\n", *treasureBoxInfo)
}

func PlayerTreasureBoxUtil(uid uint64, id uint32) *TreasureBoxUtil {
	return &TreasureBoxUtil{
		uid: uid,
		id:  id, //宝箱id
	}
}

func (t *TreasureBoxUtil) key() string {
	return fmt.Sprintf("%s:%d", treasureBoxPrefix, t.id)
}

// Clear 清空已领取过信息
func (t *TreasureBoxUtil) Clear() {
	hDEL(t.key(), t.uid)
}

// IsGet 是否已领取过
func (t *TreasureBoxUtil) IsGet() bool {
	return hExists(t.key(), t.uid)
}

// SetTreasureBoxInfo 设置宝箱领取信息
func (t *TreasureBoxUtil) SetTreasureBoxInfo(info *TreasureBoxInfo) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("SetTreasureBoxInfo error = ", err)
		return false
	}

	hSet(t.key(), t.uid, string(d))
	return true
}

// GetTreasureBoxInfo 获取宝箱领取信息
func (t *TreasureBoxUtil) GetTreasureBoxInfo(info *TreasureBoxInfo) error {
	v := hGet(t.key(), t.uid)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Warn("GetTreasureBoxInfo Failed to Unmarshal ", err)
		return errors.New("unmarshal error")
	}

	return nil
}
