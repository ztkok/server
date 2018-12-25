package db

import (
	"encoding/json"
	"fmt"
	"protoMsg"
	"zeus/dbservice"

	"github.com/garyburd/redigo/redis"
)

// 玩家在线情况下各种临时信息的存储, 不需要持久化

const (
	playerTempPrefix = "PlayerTemp"
)

type playerTempUtil struct {
	uid uint64
}

// PlayerTempUtil 获取工具类
func PlayerTempUtil(uid uint64) *playerTempUtil {
	return &playerTempUtil{
		uid: uid,
	}
}

func (u *playerTempUtil) getStringValue(field string) string {
	value, err := redis.String(dbservice.CacheHGET(u.key(), field))
	if err != nil {
		return ""
	}
	return value
}

func (u *playerTempUtil) getUintValue(field string) uint64 {
	value, err := redis.Uint64(dbservice.CacheHGET(u.key(), field))
	if err != nil {
		return 0
	}
	return value
}

func (u *playerTempUtil) getUintSliceValue(field string) []uint64 {
	data, err := redis.String(dbservice.CacheHGET(u.key(), field))
	if err != nil {
		return nil
	}

	var list []uint64
	err = json.Unmarshal([]byte(data), &list)
	if err != nil {
		return nil
	}

	return list
}

func (u *playerTempUtil) setValue(field string, value interface{}) error {
	return dbservice.CacheHSET(u.key(), field, value)
}

func (u *playerTempUtil) setComplexValue(field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return dbservice.CacheHSET(u.key(), field, string(data))
}

// GetPlayerLocation 获取玩家所在地理位置
func (u *playerTempUtil) GetPlayerLocation() string {
	return u.getStringValue("location")
}

// SetPlayerLocation 设置玩家所在地理位置
func (u *playerTempUtil) SetPlayerLocation(location string) error {
	return u.setValue("location", location)
}

// GetPlayerTeamID 获取玩家所在队伍id
func (u *playerTempUtil) GetPlayerTeamID() uint64 {
	return u.getUintValue("teamid")
}

// SetPlayerTeamID 设置玩家所在队伍id
func (u *playerTempUtil) SetPlayerTeamID(teamID uint64) error {
	return u.setValue("teamid", teamID)
}

// GetPlayerAutoMatch 获取玩家自动补充队友的标识
func (u *playerTempUtil) GetPlayerAutoMatch() uint32 {
	return uint32(u.getUintValue("automatch"))
}

// SetPlayerAutoMatch 设置玩家自动补充队友的标识
func (u *playerTempUtil) SetPlayerAutoMatch(automatch uint32) error {
	return u.setValue("automatch", automatch)
}

// GetPlayerMapID 获取玩家所在地图id
func (u *playerTempUtil) GetPlayerMapID() uint32 {
	return uint32(u.getUintValue("mapid"))
}

// SetPlayerMapID 设置玩家所在地图id
func (u *playerTempUtil) SetPlayerMapID(mapID uint32) error {
	return u.setValue("mapid", mapID)
}

// GetPlayerMatchMode 获取玩家的匹配模式
func (u *playerTempUtil) GetPlayerMatchMode() uint32 {
	return uint32(u.getUintValue("matchmode"))
}

// SetPlayerMatchMode 设置玩家的匹配模式
func (u *playerTempUtil) SetPlayerMatchMode(mode uint32) error {
	return u.setValue("matchmode", mode)
}

// GetPlayerJumpAir 获取玩家跳过飞机
func (u *playerTempUtil) GetPlayerJumpAir() uint64 {
	return u.getUintValue("jumpair")
}

// SetPlayerJumpAir 设置是否跳过飞机
func (u *playerTempUtil) SetPlayerJumpAir(value uint64) error {
	return u.setValue("jumpair", value)
}

// GetGameState 获取玩家游戏状态
func (u *playerTempUtil) GetGameState() uint64 {
	return u.getUintValue("gamestate")
}

// SetGameState 设置玩家游戏状态
func (u *playerTempUtil) SetGameState(state uint64) error {
	return u.setValue("gamestate", state)
}

// GetEnterGameTime 获取玩家进入游戏时间
func (u *playerTempUtil) GetEnterGameTime() uint64 {
	return u.getUintValue("entertime")
}

// SetEnterGameTime 设置玩家进入游戏时间
func (u *playerTempUtil) SetEnterGameTime(timestamp uint64) error {
	return u.setValue("entertime", timestamp)
}

// GetPlayerTeamID 获取玩家所在Space id
func (u *playerTempUtil) GetPlayerSpaceID() uint64 {
	return u.getUintValue("spaceid")
}

// SetPlayerTeamID 设置玩家所在Space id
func (u *playerTempUtil) SetPlayerSpaceID(spaceID uint64) error {
	return u.setValue("spaceid", spaceID)
}

// GetBookingFriend 获取预约我的好友
func (u *playerTempUtil) GetBookingFriend() uint64 {
	return u.getUintValue("BookingFriend")
}

// SetBookingFriend 设置预约我的好友
func (u *playerTempUtil) SetBookingFriend(uid uint64) error {
	return u.setValue("BookingFriend", uid)
}

// GetTmpBookingFriends 获取向我发送预约邀请的好友
func (u *playerTempUtil) GetTmpBookingFriends() []uint64 {
	return u.getUintSliceValue("TmpBookingFriends")
}

// AddTmpBookingFriend 添加向我发送预约邀请的好友
func (u *playerTempUtil) AddTmpBookingFriend(uid uint64) error {
	list := u.GetTmpBookingFriends()
	list = append(list, uid)
	return u.setComplexValue("TmpBookingFriends", list)
}

// DelTmpBookingFriend 删除向我发送预约邀请的好友
func (u *playerTempUtil) DelTmpBookingFriend(uid uint64) error {
	list := u.GetTmpBookingFriends()
	for i, v := range list {
		if v == uid {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return u.setComplexValue("TmpBookingFriends", list)
}

// GetBookedFriends 获取我预约的好友
func (u *playerTempUtil) GetBookedFriends() []uint64 {
	return u.getUintSliceValue("BookedFriends")
}

// AddBookedFriend 添加我预约的好友
func (u *playerTempUtil) AddBookedFriend(uid uint64) error {
	list := u.GetBookedFriends()
	list = append(list, uid)
	return u.setComplexValue("BookedFriends", list)
}

// DelBookedFriend 删除我预约的好友
func (u *playerTempUtil) DelBookedFriend(uid uint64) error {
	list := u.GetBookedFriends()
	for i, v := range list {
		if v == uid {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return u.setComplexValue("BookedFriends", list)
}

// GetTmpBookedFriends 获取我发送了预约邀请的好友
func (u *playerTempUtil) GetTmpBookedFriends() []uint64 {
	return u.getUintSliceValue("TmpBookedFriends")
}

// AddTmpBookedFriend 添加我发送了预约邀请的好友
func (u *playerTempUtil) AddTmpBookedFriend(uid uint64) error {
	list := u.GetTmpBookedFriends()
	list = append(list, uid)
	return u.setComplexValue("TmpBookedFriends", list)
}

// DelTmpBookedFriend 删除我发送了预约邀请的好友
func (u *playerTempUtil) DelTmpBookedFriend(uid uint64) error {
	list := u.GetTmpBookedFriends()
	for i, v := range list {
		if v == uid {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return u.setComplexValue("TmpBookedFriends", list)
}

// GetReadyBookedFriends 获取返回大厅等待组队的好友
func (u *playerTempUtil) GetReadyBookedFriends() []uint64 {
	return u.getUintSliceValue("ReadyBookedFriends")
}

// AddReadyBookedFriend 添加返回大厅等待组队的好友
func (u *playerTempUtil) AddReadyBookedFriend(uid uint64) error {
	list := u.GetReadyBookedFriends()
	list = append(list, uid)
	return u.setComplexValue("ReadyBookedFriends", list)
}

// DelReadyBookedFriend 删除返回大厅等待组队的好友
func (u *playerTempUtil) DelReadyBookedFriend(uid uint64) error {
	list := u.GetReadyBookedFriends()
	for i, v := range list {
		if v == uid {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return u.setComplexValue("ReadyBookedFriends", list)
}

// GetTeamCustoms 获取发布的队伍定制信息
func (u *playerTempUtil) GetTeamCustoms() []*protoMsg.TeamCustom {
	data, err := redis.String(dbservice.CacheHGET(u.key(), "TeamCustoms"))
	if err != nil {
		return nil
	}

	var list []*protoMsg.TeamCustom
	err = json.Unmarshal([]byte(data), &list)
	if err != nil {
		return nil
	}

	return list
}

// AddTeamCustom 添加发布的队伍定制信息
func (u *playerTempUtil) AddTeamCustom(info *protoMsg.TeamCustom) error {
	list := u.GetTeamCustoms()
	list = append(list, info)
	return u.setComplexValue("TeamCustoms", list)
}

// ClearTeamCustoms 清除发布的队伍定制信息
func (u *playerTempUtil) ClearTeamCustoms() error {
	return dbservice.CacheHDEL(u.key(), "TeamCustoms")
}

// SetToJoinTeam 设置即将加入的队伍
func (u *playerTempUtil) SetToJoinTeam(teamID uint64) error {
	return u.setValue("ToJoinTeam", teamID)
}

// GetToJoinTeam 获取即将加入的队伍
func (u *playerTempUtil) GetToJoinTeam() uint64 {
	return u.getUintValue("ToJoinTeam")
}

// SetToInviteUser 设置即将邀请的玩家
func (u *playerTempUtil) SetToInviteUser(uid uint64) error {
	return u.setValue("ToInviteUser", uid)
}

// GetToInviteUser 获取即将加入的队伍
func (u *playerTempUtil) GetToInviteUser() uint64 {
	return u.getUintValue("ToInviteUser")
}

// DelKey 删除redis key
func (u *playerTempUtil) DelKey() error {
	return dbservice.CacheDelKey(u.key())
}

func (u *playerTempUtil) key() string {
	return fmt.Sprintf("%s:%d", playerTempPrefix, u.uid)
}
