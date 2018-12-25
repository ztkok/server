package db

import (
	"fmt"
	"zeus/dbservice"

	"github.com/garyburd/redigo/redis"
)

type playerRankUtil struct {
	typ    string
	season int
}

func PlayerRankUtil(typ string, season int) *playerRankUtil {
	return &playerRankUtil{
		typ:    typ,
		season: season,
	}
}

// GetPlayerRank 获取rank
func (u *playerRankUtil) GetPlayerRank(uid uint64) (uint32, error) {
	reply, err := u.getValue(uid)
	if err != nil || reply == nil {
		return 0, err
	}

	v, err := redis.Uint64(reply, nil)
	if err != nil {
		return 0, err
	}
	return uint32(v) + 1, nil
}

// GetPlayerScore 获取分数
func (u *playerRankUtil) GetPlayerScore(uid uint64) (uint32, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	reply, err := c.Do("ZSCORE", u.key(), uid)
	if err != nil || reply == nil {
		return 0, err
	}

	v, err := redis.Float64(reply, nil)
	if err != nil {
		return 0, err
	}

	return uint32(v), nil
}

// GetTopScore 获取最高分数
func (u *playerRankUtil) GetTopScore() ([]interface{}, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	return redis.Values(c.Do("ZREVRANGE", u.key(), 0, 0, "WITHSCORES"))
}

// GetTopNumScore 获取最高num个分数
func (u *playerRankUtil) GetTopNumScore(num int) ([]interface{}, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	return redis.Values(c.Do("ZREVRANGE", u.key(), 0, num-1, "WITHSCORES"))
}

// GetScoreByRank 获取指定名次的玩家的分数
func (u *playerRankUtil) GetScoreByRank(rank uint64) ([]interface{}, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	return redis.Values(c.Do("ZREVRANGE", u.key(), rank-1, rank-1, "WITHSCORES"))
}

// GetScoreByRankRange 获取指定排名范围内的玩家分数
func (u *playerRankUtil) GetScoreByRankRange(beg, end uint64) ([]interface{}, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	return redis.Values(c.Do("ZREVRANGE", u.key(), beg-1, end-1, "WITHSCORES"))
}

// RemRankByUid 删除指定玩家的排名
func (u *playerRankUtil) RemRankByUid(uid uint64) error {
	c := dbservice.GetRankRedis()
	defer c.Close()

	_, err := c.Do("ZREM", u.key(), uid)
	return err
}

// PipeRanks 通过管道获取单个rank
func (u *playerRankUtil) PipeRanks(score float32, uid uint64) (uint32, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	c.Send("ZADD", u.key(), score, uid)
	c.Send("ZREVRANK", u.key(), uid)
	c.Flush()
	c.Receive()
	reply, err := c.Receive()
	if err != nil || reply == nil {
		return 0, err
	}

	v, err := redis.Uint64(reply, nil)
	if err != nil {
		return 0, err
	}
	return uint32(v) + 1, nil
}

// PipeRanks 通过管道获取单个rank
func (u *playerRankUtil) PipeRanksUint64(score float64, uid uint64) (uint32, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()

	c.Send("ZADD", u.key(), score, uid)
	c.Send("ZREVRANK", u.key(), uid)
	c.Flush()
	c.Receive()
	reply, err := c.Receive()
	if err != nil || reply == nil {
		return 0, err
	}

	v, err := redis.Uint64(reply, nil)
	if err != nil {
		return 0, err
	}
	return uint32(v) + 1, nil
}

// SetPlayerRank 设置rank属性
func (u *playerRankUtil) SetPlayerRank(score float32, uid uint64) error {
	c := dbservice.GetRankRedis()
	defer c.Close()

	_, err := c.Do("ZADD", u.key(), score, uid)
	return err
}

func (u *playerRankUtil) GetTotalNum() (uint64, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()
	return redis.Uint64(c.Do("ZCARD", u.key()))
}

func (u *playerRankUtil) getValue(uid uint64) (interface{}, error) {
	c := dbservice.GetRankRedis()
	defer c.Close()
	return c.Do("ZREVRANK", u.key(), uid)
}

func (u *playerRankUtil) setValue(score float32, uid uint64) error {
	c := dbservice.GetRankRedis()
	defer c.Close()
	_, err := c.Do("ZADD", u.key(), score, uid)
	return err
}

func (u *playerRankUtil) key() string {
	return fmt.Sprintf("%s:%d", u.typ, u.season)
}
