package db

import (
	"fmt"
	"math"
	"zeus/dbservice"

	"github.com/garyburd/redigo/redis"
)

// 一局游戏的临时信息, 保存在Cache中

const (
	sceneTempPrefix = "SceneTempInfo"
	fieldMapID      = "mapid"
	fieldSkyBox     = "skybox"
	fieldWatcherNum = "watchernum"
	fieldMatchMode  = "matchMode"
)

type sceneTempUtil struct {
	id uint64
}

// SceneTempUtil 获取工具类
func SceneTempUtil(spaceID uint64) *sceneTempUtil {
	return &sceneTempUtil{
		id: spaceID,
	}
}

// remove 清理场景的临时数据
func (util *sceneTempUtil) Remove() {
	c := dbservice.GetServerRedis()
	defer c.Close()
	c.Do("DEL", util.key())
}

// SetInfo 保存一局比赛的临时信息
func (util *sceneTempUtil) SetInfo(mapid uint32, skybox, matchmode uint32, watchNum int) error {
	c := dbservice.GetServerRedis()
	defer c.Close()

	args := redis.Args{}
	args = args.Add(util.key())
	args = args.Add(fieldMapID, mapid)
	args = args.Add(fieldSkyBox, skybox)
	args = args.Add(fieldWatcherNum, watchNum)
	args = args.Add(fieldMatchMode, matchmode)

	_, err := c.Do("HMSET", args...)
	return err
}

func (util *sceneTempUtil) GetInfo() (uint32, uint32, uint32, uint32, error) {
	c := dbservice.GetServerRedis()
	defer c.Close()

	info := struct {
		Mapid     uint32 `redis:"mapid"`
		Skybox    uint32 `redis:"skybox"`
		Num       uint32 `redis:"watchernum"`
		MatchMode uint32 `redis:"matchMode"`
	}{}
	values, err := redis.Values(c.Do("HGETALL", util.key()))
	if err != nil {
		return math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, err
	}

	err = redis.ScanStruct(values, &info)
	if err != nil {
		return math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, err
	}

	return info.Mapid, info.Skybox, info.Num, info.MatchMode, nil
}

// SetMapID 保存MapID
func (util *sceneTempUtil) SetMapID(mapid uint32) error {
	return dbservice.CacheHSET(util.key(), fieldMapID, mapid)
}

// GetMapID 获取MapID
func (util *sceneTempUtil) GetMapID() (uint32, error) {
	mapid, err := redis.Uint64(dbservice.CacheHGET(util.key(), fieldMapID))
	if err != nil {
		return math.MaxUint32, err
	}

	return uint32(mapid), nil
}

// SetSkyBox 保存SkyBox
func (util *sceneTempUtil) SetSkyBox(skybox uint32) error {
	return dbservice.CacheHSET(util.key(), fieldSkyBox, skybox)
}

// GetSkyBox 获取SkyBox
func (util *sceneTempUtil) GetSkyBox() (uint32, error) {
	skybox, err := redis.Uint64(dbservice.CacheHGET(util.key(), fieldSkyBox))
	if err != nil {
		return math.MaxUint32, err
	}

	return uint32(skybox), nil
}

// SetWatcherNum 保存当前观战人数
func (util *sceneTempUtil) SetWatcherNum(num uint32) error {
	return dbservice.CacheHSET(util.key(), fieldWatcherNum, num)
}

// GetWatcherNum 获取当前观战人数
func (util *sceneTempUtil) GetWatcherNum() (uint32, error) {
	num, err := redis.Uint64(dbservice.CacheHGET(util.key(), fieldWatcherNum))
	if err != nil {
		return 0, err
	}

	return uint32(num), nil
}

// GetSingleUserWatchNum 单人游戏的观战人数
func (util *sceneTempUtil) GetSingleUserWatchNum(uid uint64) uint32 {
	c := dbservice.GetServerRedis()
	defer c.Close()
	d, e := redis.Uint64(c.Do("HGET", util.key(), uid))
	if e != nil {
		return 0
	}
	return uint32(d)
}

// SetSingleUserWatchNum 单人游戏的观战人数
func (util *sceneTempUtil) IncrbySingleUserWatchNum(uid uint64, num int) uint32 {
	c := dbservice.GetServerRedis()
	defer c.Close()
	d, e := redis.Uint64(c.Do("HINCRBY", util.key(), uid, num))
	if e != nil {
		return 0
	}
	return uint32(d)
}

func (util *sceneTempUtil) key() string {
	return fmt.Sprintf("%s:%d", sceneTempPrefix, util.id)
}
