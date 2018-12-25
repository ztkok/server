package main

import (
	"common"
	"excel"
	"math/rand"
	"strings"
	"time"

	log "github.com/cihub/seelog"
)

type stMapItemRate struct {
	rate  int
	items []uint32
}

type stMapItemConfig struct {
	createrate int
	num        int
	total      int
	itemrate   []stMapItemRate
}

type MapItemRate struct {
	mapitemrate map[uint64][]stMapItemConfig
}

func GetMapItemRate(tb *Scene) *MapItemRate {
	if tb.refreshItemInst == nil {
		tb.refreshItemInst = &MapItemRate{}
		tb.refreshItemInst.mapitemrate = make(map[uint64][]stMapItemConfig)

		for k, v := range excel.GetMapitemrateMap() {
			readItemList(tb.refreshItemInst, k, v.Itemlist1)
			readItemList(tb.refreshItemInst, k, v.Itemlist2)
			readItemList(tb.refreshItemInst, k, v.Itemlist3)
			readItemList(tb.refreshItemInst, k, v.Itemlist4)
			readItemList(tb.refreshItemInst, k, v.Itemlist5)
			readItemList(tb.refreshItemInst, k, v.Itemlist6)
			readItemList(tb.refreshItemInst, k, v.Itemlist7)
			readItemList(tb.refreshItemInst, k, v.Itemlist8)
			readItemList(tb.refreshItemInst, k, v.Itemlist9)
			readItemList(tb.refreshItemInst, k, v.Itemlist10)
			readItemList(tb.refreshItemInst, k, v.Itemlist11)
		}
	}

	return tb.refreshItemInst
}

func readItemList(mapitem *MapItemRate, k uint64, str string) {
	var tmp stMapItemConfig
	strlist := strings.Split(str, ";")
	if len(strlist) != 2 {
		return
	}

	ratenum := strings.Split(strlist[0], "-")
	if len(ratenum) != 2 {
		return
	}

	tmp.createrate = common.StringToInt(ratenum[0])
	tmp.num = common.StringToInt(ratenum[1])

	itemlist := strings.Split(strlist[1], ",")
	if len(itemlist) < 1 {
		return
	}

	for _, str := range itemlist {
		item := strings.Split(str, ":")
		if len(item) != 2 {
			continue
		}

		var itemrate stMapItemRate
		items := strings.Split(item[0], "-")
		rate := common.StringToInt(item[1])
		if rate != 0 {
			tmp.total += rate
			itemrate.rate = rate
			for _, idstr := range items {
				id := common.StringToUint32(idstr)
				if id != 0 {
					itemrate.items = append(itemrate.items, id)
				} else {
					// log.Warn("配置id错误", str, idstr)
				}
			}

			tmp.itemrate = append(tmp.itemrate, itemrate)
		} else {
			// log.Warn("配置错误", str)
		}
	}

	mapitem.mapitemrate[k] = append(mapitem.mapitemrate[k], tmp)
	// log.Info("测试", str, len(mapitem.mapitemrate))
}

func (sf *MapItemRate) randRuleItem(r *rand.Rand, rule uint64) []uint32 {
	ret := make([]uint32, 0)
	rand.Seed(time.Now().UnixNano())
	config, ok := sf.mapitemrate[rule]
	if !ok {
		//log.Warn("错误配置规则", rule)
		return ret
	}

	for _, v := range config {
		rate := r.Intn(10000) + 1
		if rate > v.createrate {
			continue
		}

		for index := 0; index < v.num; index++ {
			randnum := r.Intn(v.total) + 1
			var percent int

			for _, itemrate := range v.itemrate {
				percent += itemrate.rate
				if percent >= randnum {

					for _, itembaseid := range itemrate.items {
						_, ok := excel.GetItem(uint64(itembaseid))
						if !ok {
							log.Error("Base id not exist, id: ", itembaseid)
							continue
						}

						ret = append(ret, itembaseid)
					}

					break
				}
			}
		}
	}

	return ret
}
