package db

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

// InsigniaUtil 玩家勋章
type InsigniaUtil struct {
	uid uint64
}

// PlayerInsigniaUtil 玩家勋章
func PlayerInsigniaUtil(uid uint64) *InsigniaUtil {
	return &InsigniaUtil{
		uid: uid,
	}
}

func (util *InsigniaUtil) key() string {
	return fmt.Sprintf("%s:%d", "PlayerInsignia", util.uid)
}
func (util *InsigniaUtil) AddInsignia(id uint64) {
	hIncrBy(util.key(), id, 1)
}
func (util *InsigniaUtil) SetInsignia(info map[uint32]uint32) {
	if len(info) == 0 {
		return
	}
	args := redis.Args{}.Add(util.key())
	for k, v := range info {
		args = args.Add(k)
		args = args.Add(v)
	}
	hMSet(args)
}

//GetInsignia 获取当前的勋章信息
func (util *InsigniaUtil) GetInsignia() map[uint32]uint32 {
	info := hGetAll(util.key())
	ret := map[uint32]uint32{}
	for k, v := range info {
		id, e := strconv.Atoi(k)
		if e != nil {
			continue
		}
		num, e := strconv.Atoi(v)
		if e != nil {
			continue
		}
		ret[uint32(id)] = uint32(num)
	}
	return ret
}

//IsExistsInsignia 是否存在改勋章
func (util *InsigniaUtil) IsExistsInsignia(id uint64) bool {
	return hExists(util.key(), id)
}

//GetInsigniaById 获取某个勋章的获得数量
func (util *InsigniaUtil) GetInsigniaById(id uint32) uint64 {
	value := hGet(util.key(), id)
	i, e := strconv.Atoi(value)
	if e != nil {
		return 0
	}
	return uint64(i)
}

// keyFlag 标记key
func (util *InsigniaUtil) keyFlag() string {
	return fmt.Sprintf("%s:%d", "PlayerInsigniaFlag", util.uid)
}

// SetInsigniaFlag 设置勋章是否有红点标记 flag 1:新获得的 0:旧的
func (util *InsigniaUtil) SetInsigniaFlag(id, flag uint32) {
	hSet(util.keyFlag(), id, flag)
}

//GetInsigniaFlag 获取勋章是否有红点标记
func (util *InsigniaUtil) GetInsigniaFlag() map[uint32]uint32 {
	info := hGetAll(util.keyFlag())
	ret := map[uint32]uint32{}
	for k, v := range info {
		id, e := strconv.Atoi(k)
		if e != nil {
			continue
		}
		num, e := strconv.Atoi(v)
		if e != nil {
			continue
		}
		ret[uint32(id)] = uint32(num)
	}
	return ret
}
