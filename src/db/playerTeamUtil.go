package db

import (
	"fmt"
	"zeus/dbservice"

	"github.com/garyburd/redigo/redis"
)

// 玩家队伍的临时数据，记录服务器id 观战人数
type playerTeamUtil struct {
	teamID uint64
}

func PlayerTeamUtil(teamID uint64) *playerTeamUtil {
	return &playerTeamUtil{teamID: teamID}
}

// IncryTeamWatchNum 设置队伍所在的匹配服务器
func (r *playerTeamUtil) IncryTeamWatchNum(num int) uint32 {
	c := dbservice.GetServerRedis()
	defer c.Close()

	d, e := redis.Uint64(c.Do("HINCRBY", r.key(), "watchNum", num))
	if e != nil {
		return 0
	}
	return uint32(d)
}

// SetMatchSrvID 设置队伍所在的匹配服务器
func (r *playerTeamUtil) GetTeamWatchNum() uint32 {
	num, e := redis.Uint64(dbservice.CacheHGET(r.key(), "watchNum"))
	if e != nil {
		return 0
	}
	return uint32(num)
}

// SetMatchSrvID 设置队伍所在的匹配服务器
func (r *playerTeamUtil) SetMatchSrvID(srvID uint64) error {
	return dbservice.CacheHSET(r.key(), "matchsrvid", srvID)
}

// GetMatchSrvID 获取队伍所在的匹配服务器
func (r *playerTeamUtil) GetMatchSrvID() (uint64, error) {
	return redis.Uint64(dbservice.CacheHGET(r.key(), "matchsrvid"))
}

// SetMatchMode 设置队伍的匹配模式
func (r *playerTeamUtil) SetMatchMode(mode uint32) error {
	return dbservice.CacheHSET(r.key(), "matchmode", mode)
}

// GetMatchMode 获取队伍的匹配模式
func (r *playerTeamUtil) GetMatchMode() (uint32, error) {
	mode, err := redis.Uint64(dbservice.CacheHGET(r.key(), "matchmode"))
	return uint32(mode), err
}

// SetTeamType 设置队伍的类型，0代表双排队伍，1代表四排队伍
func (r *playerTeamUtil) SetTeamType(typ uint8) error {
	return dbservice.CacheHSET(r.key(), "teamtype", typ)
}

// GetTeamType 获取队伍的类型，0代表双排队伍，1代表四排队伍
func (r *playerTeamUtil) GetTeamType() (uint8, error) {
	typ, err := redis.Uint64(dbservice.CacheHGET(r.key(), "teamtype"))
	return uint8(typ), err
}

// SetTeamNum 设置队伍人数
func (r *playerTeamUtil) SetTeamNum(num uint32) error {
	return dbservice.CacheHSET(r.key(), "teamnum", num)
}

// GetTeamNum 获取队伍人数
func (r *playerTeamUtil) GetTeamNum() (uint32, error) {
	num, err := redis.Uint64(dbservice.CacheHGET(r.key(), "teamnum"))
	return uint32(num), err
}

// SetTeamMap 设置队伍选定的地图
func (r *playerTeamUtil) SetTeamMap(mapid uint32) error {
	return dbservice.CacheHSET(r.key(), "teammap", mapid)
}

// GetTeamMap 获取队伍选定的地图
func (r *playerTeamUtil) GetTeamMap() (uint32, error) {
	id, err := redis.Uint64(dbservice.CacheHGET(r.key(), "teammap"))
	return uint32(id), err
}

// Remove 结束时 清理队伍
func (r *playerTeamUtil) Remove() {
	c := dbservice.GetServerRedis()
	defer c.Close()
	c.Do("DEL", r.key())
}
func (r *playerTeamUtil) key() string {
	return fmt.Sprintf("%s:%d", "PlayerTeam", r.teamID)
}
