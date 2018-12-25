package db

import (
	"encoding/json"
	"excel"
	"fmt"
	"protoMsg"
	"time"

	log "github.com/cihub/seelog"
)

const (
	friendPrefix     = "Friend"
	applyPrefix      = "FriendApply"
	platFrientPrefix = "platFrient"
	blackListPrefix  = "BlackList"
)

const (
	chatKey    = "ChatInfo" //聊天key
	maxChatNum = 100        //保存的最大聊天数量
)

// 好友信息
type FriendInfo struct {
	ID   uint64 // 好友id
	Name string // 好友名称
	Time int64  // 时间戳
}

// 申请信息
type ApplyInfo struct {
	ID        uint64 // 申请者id
	Name      string // 申请者名称
	ApplyTime int64  // 申请时间
}

// 聊天信息
type ChatInfo struct {
	ID      uint64 // 好友id
	Time    int64  // 时间戳
	Content string // 聊天内容
}

// friendUtil 好友工具
type FriendUtil struct {
	uid uint64
}

// GetFriendUtil 获取好友工具
func GetFriendUtil(uid uint64) *FriendUtil {
	return &FriendUtil{
		uid: uid,
	}
}

// friendkey 获取好友key值
func (f *FriendUtil) friendkey() string {
	return fmt.Sprintf("%s:%d", friendPrefix, f.uid)
}

// applykey 获取申请key值
func (f *FriendUtil) applykey() string {
	return fmt.Sprintf("%s:%d", applyPrefix, f.uid)
}

// platkey 获取平台好友key值
func (f *FriendUtil) platkey() string {
	return fmt.Sprintf("%s:%d", platFrientPrefix, f.uid)
}

// blackkey 获取黑名单key值
func (f *FriendUtil) blackkey() string {
	return fmt.Sprintf("%s:%d", blackListPrefix, f.uid)
}

// chatkey 获取聊天key值
func (f *FriendUtil) chatkey() string {
	return chatKey
}

// AddFriend 添加好友
func (f *FriendUtil) AddFriend(info FriendInfo) bool {

	log.Info("AddFriend", info)

	if hExists(f.friendkey(), info.ID) {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Info("AddFriend err ", err)
		return false
	}

	// 添加好友数据
	hSet(f.friendkey(), info.ID, string(d))

	return true
}

// DelFriend 删除好友
func (f *FriendUtil) DelFriend(id uint64) bool {
	if !hExists(f.friendkey(), id) {
		return false
	}

	hDEL(f.friendkey(), id)

	return true
}

// GetFriendList 获取好友列表
func (f *FriendUtil) GetFriendList() []*FriendInfo {

	friendSet := hGetAll(f.friendkey())

	ret := make([]*FriendInfo, 0)
	for _, friendStream := range friendSet {

		info := &FriendInfo{}
		if err := json.Unmarshal([]byte(friendStream), &info); err != nil {
			log.Warn("GetFriendList Failed to Unmarshal ", err)
			return nil
		}

		ret = append(ret, info)
	}

	return ret

}

// IsFriendByID 根据id判断是否为好友
func (f *FriendUtil) IsFriendByID(id uint64) bool {
	return hExists(f.friendkey(), id)
}

// IsReachLimit 是否达到好友上线
func (f *FriendUtil) IsReachLimit() bool {
	limit, success := excel.GetSystem(33)
	if !success {
		log.Warn("好友上限配置出错！")
		return false
	}
	return hLen(f.friendkey()) >= limit.Value
}

// AddApplyInfo 添加申请信息
func (f *FriendUtil) AddApplyInfo(info ApplyInfo) bool {

	log.Info("AddApplyInfo", info)

	if hExists(f.applykey(), info.ID) {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Info("AddApplyInfo err ", err)
		return false
	}

	// 添加好友数据
	hSet(f.applykey(), info.ID, string(d))

	return true
}

// DelApply 删除申请请求
func (f *FriendUtil) DelApply(id uint64) bool {
	if !hExists(f.applykey(), id) {
		return false
	}

	hDEL(f.applykey(), id)

	return true
}

// GetSigleApplyReq 获取申请请求
func (f *FriendUtil) GetSigleApplyReq(id uint64) *FriendInfo {
	if !hExists(f.applykey(), id) {
		return nil
	}

	hGet(f.applykey(), id)

	v := hGet(f.applykey(), id)
	var d *FriendInfo
	if err := json.Unmarshal([]byte(v), &d); err != nil {
		log.Warn("GetSigleApplyReq Failed to Unmarshal ", err)
		return nil
	}
	return d

}

// GetApplyList 获取申请列表
func (f *FriendUtil) GetApplyList() []*ApplyInfo {

	applySet := hGetAll(f.applykey())

	ret := make([]*ApplyInfo, 0)
	curtime := time.Now().Unix()
	for _, applyStream := range applySet {

		info := &ApplyInfo{}
		if err := json.Unmarshal([]byte(applyStream), &info); err != nil {
			log.Warn("GetApplyList Failed to Unmarshal ", err)
			return nil
		}

		dueTime := int64(0)
		system, success := excel.GetSystem(31)
		if success {
			dueTime = int64(system.Value)
		}
		if info.ApplyTime+dueTime <= curtime {
			f.DelApply(info.ID) // 删除过期申请表
		} else {
			ret = append(ret, info)
		}
	}

	return ret
}

// InApplyListByID 根据id判断是否在申请列表中
func (f *FriendUtil) InApplyListByID(id uint64) bool {

	if !hExists(f.applykey(), id) {
		return false
	}

	hGet(f.applykey(), id)

	v := hGet(f.applykey(), id)
	var d *ApplyInfo
	if err := json.Unmarshal([]byte(v), &d); err != nil {
		log.Warn("InApplyListByID Failed to Unmarshal ", err)
		return false
	}

	dueTime := int64(0)
	system, success := excel.GetSystem(31)
	if success {
		dueTime = int64(system.Value)
	}
	if d.ApplyTime+dueTime < time.Now().Unix() {
		f.DelApply(id)
		return false
	}

	return true
}

// IsReachApplyLimit 是否达到好友申请上限
func (f *FriendUtil) IsReachApplyLimit() bool {
	limit, success := excel.GetSystem(34)
	if !success {
		log.Warn("好友申请上限配置出错！")
		return false
	}
	return hLen(f.applykey()) >= limit.Value
}

// UpdatePlatFrientInfo 更新平台好友信息
func (f *FriendUtil) UpdatePlatFrientInfo(info []uint64) {

	d, e := json.Marshal(info)
	if e != nil {
		log.Warn("UpdatePlatFrientInfo error e = ", e)
		return
	}

	hSet(f.platkey(), f.uid, string(d))

	// if err := dbservice.CacheHSET(); err != nil {
	// 	log.Error(err)
	// 	return
	// }

}

// GetPlatFrientInfo 获取平台好友信息
func (f *FriendUtil) GetPlatFrientInfo() []uint64 {

	ret := make([]uint64, 0)

	if !hExists(f.platkey(), f.uid) {
		//log.Error("获取平台好友信息失败 ", f.uid)
		return ret
	}

	v := hGet(f.platkey(), f.uid)

	if err := json.Unmarshal([]byte(v), &ret); err != nil {
		log.Warn("获取平台好友信息失败 Unmarshal err ", err)
		return nil
	}
	return ret

}

// AddBlacker 拉黑
func (f *FriendUtil) AddBlacker(info FriendInfo) bool {
	if f.IsBlacker(info.ID) {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Info("Marshal err: ", err)
		return false
	}

	hSet(f.blackkey(), info.ID, string(d))

	return true
}

// DelBlacker 删除黑名单
func (f *FriendUtil) DelBlacker(id uint64) bool {
	if !f.IsBlacker(id) {
		return false
	}

	hDEL(f.blackkey(), id)

	return true
}

// IsBlacker 是否被拉黑
func (f *FriendUtil) IsBlacker(id uint64) bool {
	return hExists(f.blackkey(), id)
}

// GetBlackList 获取黑名单
func (f *FriendUtil) GetBlackList() []*FriendInfo {
	friendSet := hGetAll(f.blackkey())

	ret := make([]*FriendInfo, 0)
	for _, friendStream := range friendSet {

		info := &FriendInfo{}
		if err := json.Unmarshal([]byte(friendStream), &info); err != nil {
			log.Warn("Unmarshal err: ", err)
			return nil
		}

		ret = append(ret, info)
	}

	return ret
}

// AddChat 添加聊天记录
func (f *FriendUtil) AddChat(info *protoMsg.ChatInfo) bool {
	msg := f.getChatDetail()
	if len(msg.Infos) >= maxChatNum {
		msg.Infos = append(msg.Infos[:0], msg.Infos[1:]...)
	}

	msg.Infos = append(msg.Infos, info)
	f.setChatDetail(msg)

	return true
}

// DelChat 删除聊天记录
func (f *FriendUtil) DelChat(uid uint64) bool {
	msg := f.getChatDetail()

	var count int
	for _, v := range msg.Infos {
		if v.Uid == uid {
			count++
		}
	}

	for i := 0; i < count; i++ {
		for j, v := range msg.Infos {
			if v.Uid == uid {
				msg.Infos = append(msg.Infos[:j], msg.Infos[j+1:]...)
				break
			}
		}
	}

	f.setChatDetail(msg)
	return true
}

// GetChatDetail 获取聊天记录
func (f *FriendUtil) GetChatDetail() *protoMsg.UnreadChatDetail {
	return f.getChatDetail()
}

func (f *FriendUtil) getChatDetail() *protoMsg.UnreadChatDetail {
	data := hGet(f.chatkey(), f.uid)
	msg := &protoMsg.UnreadChatDetail{}

	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		log.Error("Unmarshal err: ", err)
	}

	return msg
}

func (f *FriendUtil) setChatDetail(msg *protoMsg.UnreadChatDetail) {
	d, err := json.Marshal(msg)
	if err != nil {
		log.Error("Marshal err: ", err)
		return
	}

	hSet(f.chatkey(), f.uid, string(d))
	return
}
