package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/cihub/seelog"
)

// AchievementUtil 成就系统
type AchievementUtil struct {
	uid uint64
}

// PlayerAchievementUtil 成就系统
func PlayerAchievementUtil(uid uint64) *AchievementUtil {
	return &AchievementUtil{
		uid: uid,
	}
}

// SetInit 设置旧数据处理标记
func (util *AchievementUtil) SetInit() {
	hSet(util.infoKey(), "init", 1)
}

// IsInit 是否完成过旧数据处理
func (util *AchievementUtil) IsInit() bool {
	return hExists(util.infoKey(), "init")
}

// SetLevel 设置成就等级
func (util *AchievementUtil) SetLevel(level uint32) {
	hSet(util.infoKey(), "level", level)
}

// SetExp 设置成就经验
func (util *AchievementUtil) SetExp(exp uint32) {
	hSet(util.infoKey(), "exp", exp)
}

// SetShow 设置展示中的成就
func (util *AchievementUtil) SetShow(used []uint32) {
	data, err := json.Marshal(used)
	if err != nil {
		log.Debug("SetAchieveShow Error ", err)
		return
	}
	hSet(util.infoKey(), "used", string(data))
}

// GetLevelInfo 获取成就等级 经验
func (util *AchievementUtil) GetLevelInfo() (uint32, uint32) {
	info := hMGet(util.infoKey(), []string{"level", "exp"})

	level, err := strconv.Atoi(info["level"])
	if err != nil {
		level = 1
	}
	exp, err := strconv.Atoi(info["exp"])
	if err != nil {
		exp = 0
	}

	return uint32(level), uint32(exp)
}

// GetShow 获取当前展示中的成就
func (util *AchievementUtil) GetShow() []uint32 {
	value := hGet(util.infoKey(), "used")
	ret := []uint32{0, 0, 0}
	json.Unmarshal([]byte(value), &ret)
	return ret
}

// AchieveInfo 成就完成数据
type AchieveInfo struct {
	Id    uint64   `json:"id"`    // 成就id
	Stamp []uint64 `json:"stamp"` // 成就获得的时间
	Flag  uint32   `json:"flag"`  //标记此成就是否是新获得的 1:新获得	0:旧成就
}

// AddAchieve 添加成就
func (util *AchievementUtil) AddAchieve(info []*AchieveInfo) {
	for _, i := range info {
		d, err := json.Marshal(i)
		if err != nil {
			log.Debug("achieve error ", err)
			return
		}
		hSet(util.listKey(), i.Id, string(d))
	}

}

// GetAchieveInfo 获取当前的成就进度
func (util *AchievementUtil) GetAchieveInfo() map[uint64]*AchieveInfo {
	list := hGetAll(util.listKey())
	ret := map[uint64]*AchieveInfo{}
	for _, data := range list {
		info := &AchieveInfo{}
		err := json.Unmarshal([]byte(data), info)
		if err != nil {
			log.Debug("unmarsha Achieve error ", err)
			continue
		}
		ret[info.Id] = info
	}
	return ret
}

// IsGetAchieve 是否达成
func (util *AchievementUtil) IsGetAchieve(id uint32) bool {
	return hExists(util.listKey(), id)
}

// AddReward 记录奖励领取
func (util *AchievementUtil) AddReward(level uint32) {
	hSet(util.rewardKey(), level, time.Now().Unix())
}

// IsGetReward 是否已领取奖励
func (util *AchievementUtil) IsGetReward(level uint32) bool {
	return hExists(util.rewardKey(), level)
}

// GetReward 获取奖励领取记录
func (util *AchievementUtil) GetReward() []uint32 {
	value := hKeys(util.rewardKey())
	var ret []uint32
	for _, v := range value {
		d, e := strconv.ParseUint(v, 10, 64)
		if e != nil {
			log.Debug("GetReward Error ", e)
		}
		ret = append(ret, uint32(d))
	}
	return ret
}

// AddAchievementData 添加成就条件的记录
func (util *AchievementUtil) AddAchievementData(id uint32, add interface{}) float64 {
	return hIncrByFloat(util.dataKey(), id, add)
}

// GetAchievementData 获取成就条件的当前进度
func (util *AchievementUtil) GetAchievementData(id uint32) uint32 {
	value := hGet(util.dataKey(), id)
	v, e := strconv.Atoi(value)
	if e != nil {
		log.Debug("GetAchievementData Error ")
		return 0
	}
	return uint32(v)
}

// GetAllAchievementData 获取所有成就条件的当前进度
func (util *AchievementUtil) GetAllAchievementData() map[uint64]float64 {
	value := hGetAll(util.dataKey())
	ret := map[uint64]float64{}
	for k, v := range value {
		k1, e := strconv.ParseUint(k, 10, 64)
		if e != nil {
			continue
		}
		v1, e := strconv.ParseFloat(v, 32)
		if e != nil {
			continue
		}
		ret[k1] = v1
	}
	return ret
}

// dataKey 成就条件数据记录
func (util *AchievementUtil) dataKey() string {
	return fmt.Sprintf("%s:%d", "PlayerAchieveData", util.uid)
}

// infoKey 玩家的成就的整体数据 等级 经验 etc
func (util *AchievementUtil) infoKey() string {
	return fmt.Sprintf("%s:%d", "PlayerAchieveInfo", util.uid)
}

// listKey 玩家的成就达成记录
func (util *AchievementUtil) listKey() string {
	return fmt.Sprintf("%s:%d", "PlayerAchieveList", util.uid)
}

// rewardKey 玩家的成就奖励领取
func (util *AchievementUtil) rewardKey() string {
	return fmt.Sprintf("%s:%d", "PlayerAchieveReward", util.uid)
}
