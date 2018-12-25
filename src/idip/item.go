package idip

/*

 IDIP 道具相关

*/

// ItemInfo 道具信息
type ItemInfo struct {
	ItemID  uint32 `json:"ItemId"`  // 道具ID
	ItemNum uint16 `json:"ItemNum"` // 道具数量
}
