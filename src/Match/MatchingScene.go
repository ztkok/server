package main

// MatchingScene 匹配上的房间
type MatchingScene struct {
	IWaitingScene
	matchingDegree uint32
}

// MatchingScenes 匹配上的房间列表
type MatchingScenes []*MatchingScene

// Len 实现sort接口
func (mss MatchingScenes) Len() int { return len(mss) }

// Swap 实现sort接口
func (mss MatchingScenes) Swap(i, j int) { mss[i], mss[j] = mss[j], mss[i] }

// Less 实现sort接口
func (mss MatchingScenes) Less(i, j int) bool {
	if mss[i].matchingDegree < mss[j].matchingDegree {
		return true
	} else if mss[i].matchingDegree > mss[j].matchingDegree {
		return false
	}

	// 相同的情况下比较场景创建时间, 创建早的排后面
	if mss[i].GetCreateTime() >= mss[j].GetCreateTime() {
		return true
	}

	return false
}
