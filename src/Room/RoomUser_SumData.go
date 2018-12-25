package main

import (
	"common"
	"datadef"
	"db"
	"excel"
	"math"
	"protoMsg"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/iserver"
	"zeus/linmath"
)

const (
	DEAD_NOT         = 0  //掉线
	DEAD_GUNBATTLE   = 1  //枪战死亡
	DEAD_FRIENDKILL  = 2  //友军击杀
	DEAD_BOMBKILL    = 3  //炸弹击杀
	DEAD_BOMBZONE    = 4  //轰炸区炸死
	DEAD_CARKILL     = 5  //载具撞死
	DEAD_SAFEZONEOUT = 6  //安全区外死亡
	DEAD_HANDKILL    = 7  //拳头死亡
	DEAD_FALLKILL    = 8  //坠落死亡
	DEAD_OFFLINE     = 9  //吃鸡未死亡
	DEAD_QUIT        = 10 //主动退出
	DEAD_TANKKILL    = 11 //坦克炮弹炸死
)

// SumData 统计本局数据
type SumData struct {
	user       *RoomUser
	careerdata *datadef.CareerBase

	userName         string
	rank             uint32
	battletype       uint32
	recvitemusenum   uint32
	shotnum          uint32  // 开枪次数
	revivenum        uint32  // 复活次数
	killdistance     float32 //最远击杀距离
	killstmnum       uint32  //最大连杀次数
	doubleKillNum    uint32  //两连杀次数
	tripleKillNum    uint32  //三连杀次数
	recoverHp        uint32  //恢复血量
	carusernum       uint32  //载具使用数量
	speednum         uint32  //加速次数
	attacknum        uint32  //助攻次数
	killRating       float32 //击杀评分
	winRating        float32 //生存评分
	rating           float32 //本局总评分
	begingametime    int64
	endgametime      int64
	bandageNum       uint32  //绷带使用次数
	medicalBoxNum    uint32  //医疗箱使用次数
	painkillerNum    uint32  //急救包使用次数
	carDestoryNum    uint32  //载具摧毁数量
	deadType         uint32  //死亡类型
	coin             uint32  //本局coin
	runDistance      float32 //本局跑动距离(m)
	carDistance      float32 //驾驶载具形式距离(m)
	swimDistance     float32 //本局游泳距离(m)
	gunID            uint64  //最远击杀武器枪支id，无为0
	sightID          uint32  //最远击杀武器瞄准镜id，无为0
	silienceID       uint32  //最远击杀武器消音器id
	magazineID       uint32  //最远击杀武器弹匣类id，无为0
	stockID          uint32  //最远击杀武器枪托id，无为0
	handleID         uint32  //最远击杀武器握把id，无为0
	openIDByKill     string  //被击杀人openid
	gunIDByKill      uint64  //被击杀武器枪支id，无为0
	sightIDByKill    uint32  //被击杀武器瞄准镜id，无为0
	silienceIDByKill uint32  //被击杀武器消音器id
	magazineIDByKill uint32  //被击杀武器弹匣类id，无为0
	stockIDByKill    uint32  //被击杀武器枪托id，无为0
	handleIDByKill   uint32  //被击杀武器握把id，无为0
	deadIsHead       uint32  //是否爆头
	watchType        uint32  //观战类型 0结算自动退出 1主动退出
	watchStartTime   int64   //观战开始时间
	watchEndTime     int64   //观战结束时间
	rescueNum        uint32  //救援次数
	rescueComradeNum uint32  //救援战友次数
	fistKillNum      uint32  //拳头击杀数
	grenadeKillNum   uint32  //手雷击杀数
	carKillNum       uint32  //载具击杀数
	sniperGunKillNum uint32  //狙击枪击杀数
	rpgKillNum       uint32  //rpg击杀数
	pistolKillNum    uint32  //手枪击杀数
	tankKillNum      uint32  //坦克击杀数量
	tankUseTime      int64   //使用坦克时间
	signalGunUseNum  uint32  //信号枪使用次数
	lastUpTankTime   int64
	killDistanceNum  uint32          //最远击杀距离数量
	fallDamage       uint32          //坠落伤害
	getDropBoxs      map[uint64]bool //拾取的空投箱集合

	diffkillRating float32 //击杀评分
	diffwinRating  float32 //生存评分

	landtime        int64 //落地时间
	lasttime        int64 //上次击杀时间
	tmpkillstmnum   uint32
	attackmap       map[uint64]int64 //助攻map统计
	tmprunDistance  linmath.Vector3
	tmpcarDistance  linmath.Vector3
	tmpswimDistance linmath.Vector3

	isSettble      bool //判断是否结算（分为正常退出触发和杀死后台进程)
	isGround       bool //玩家是否落地
	isParachuteDie bool //是否跳伞时被击败
	isLandDie      bool //是否着地10秒内被击败
	isGilleyWin    bool //是否穿吉利服获胜
	isOwnDrink20   bool //是否背包内饮料数量达到20瓶

	braveCoin uint32 //本次比赛获得的勇气值
	season    int    //当前赛季id
}

// NewSumData 获取统计数据管理器
func NewSumData(user *RoomUser) *SumData {
	sumData := &SumData{
		user:       user,
		careerdata: &datadef.CareerBase{},

		begingametime: time.Now().Unix(),
		endgametime:   time.Now().Unix(),

		userName:         "",
		rank:             0,
		battletype:       0,
		recvitemusenum:   0,
		shotnum:          0,
		revivenum:        0,
		killdistance:     0,
		killstmnum:       0, //最大连杀次数,
		recoverHp:        0,
		carusernum:       0,
		speednum:         0,
		attacknum:        0,
		attackmap:        make(map[uint64]int64),
		killRating:       0,
		winRating:        0,
		rating:           0,
		bandageNum:       0,
		medicalBoxNum:    0,
		painkillerNum:    0,
		carDestoryNum:    0,
		deadType:         0,
		coin:             0,
		runDistance:      0,
		carDistance:      0,
		gunID:            0,
		sightID:          0,
		silienceID:       0,
		magazineID:       0,
		stockID:          0,
		handleID:         0,
		openIDByKill:     "0",
		gunIDByKill:      0,
		sightIDByKill:    0,
		silienceIDByKill: 0,
		magazineIDByKill: 0,
		stockIDByKill:    0,
		handleIDByKill:   0,
		deadIsHead:       0,
		watchType:        0,
		watchStartTime:   0,
		watchEndTime:     0,
		rescueNum:        0,
		fistKillNum:      0,
		grenadeKillNum:   0,
		carKillNum:       0,
		sniperGunKillNum: 0,
		diffkillRating:   0,
		diffwinRating:    0,
		killDistanceNum:  0,

		lasttime:        0, //上次击杀时间
		tmpkillstmnum:   0,
		tmprunDistance:  linmath.Vector3_Invalid(),
		tmpcarDistance:  linmath.Vector3_Invalid(),
		tmpswimDistance: linmath.Vector3_Invalid(),
		getDropBoxs:     make(map[uint64]bool),

		isSettble: true,
		isGround:  false,

		season: common.GetSeason(),
	}

	sumData.InitUserName()
	sumData.initCareerData()

	return sumData
}

// InitUserName 初始化玩家昵称
func (sumData *SumData) InitUserName() {
	sumData.userName = sumData.user.GetName()
	if sumData.userName == "" {
		sumData.userName = "NoUserName"
	} else {
		r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
		sumData.userName = r.Replace(sumData.userName)
	}
}

func (sumData *SumData) initCareerData() {
	util := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, sumData.user.GetDBID(), common.GetSeason())
	if err := util.GetRoundData(sumData.careerdata); err != nil {
		sumData.user.Error("GetRoundData err: ", err)
	}
}

// IncrRecvitemUseNum 治疗道具使用次数自增
func (sumData *SumData) IncrRecvitemUseNum() {
	sumData.recvitemusenum++
}

// IncrShotNum 开枪次数自增
func (sumData *SumData) IncrShotNum() {
	sumData.shotnum++
}

// IncrReviveNum 复活次数自增
func (sumData *SumData) IncrReviveNum() {
	sumData.revivenum++
}

// IncrRescueNum 救援次数自增
func (sumData *SumData) IncrRescueNum() {
	sumData.rescueNum++
}

// IncrRescueNum 救援战友次数自增
func (sumData *SumData) IncrRescueComradeNum() {
	sumData.rescueComradeNum++
}

// KillDistance 击杀距离更新，击杀武器统计
func (sumData *SumData) KillDistance(defendid uint64) {
	space := sumData.user.GetSpace().(*Scene)

	//最高击杀距离统计
	defender, ok := space.GetEntity(defendid).(iDefender)
	if ok {
		killdistance := common.Distance(sumData.user.GetPos(), defender.GetPos())
		if sumData.killdistance < killdistance {
			sumData.killdistance = killdistance

			sumData.gunID, sumData.sightID, sumData.silienceID, sumData.magazineID, sumData.stockID, sumData.handleID = sumData.user.getUseGunInfo()
		}
		if killdistance >= float32(common.GetTBSystemValue(common.System_HonourKillDistance)) {
			sumData.killDistanceNum++
		}
		sumData.user.Info("Max kill distance: ", sumData.killdistance)
	}
}

// MaxChainKill 统计最大连杀次数
func (sumData *SumData) MaxChainKill() {

	if sumData.tmpkillstmnum == 0 {
		sumData.tmpkillstmnum = 1
		sumData.lasttime = time.Now().Unix()
	} else if sumData.tmpkillstmnum >= 1 {
		chainbytime, ok := excel.GetSystem(38)
		if !ok {
			sumData.user.Warn("Get system config failed")
			return
		}

		if time.Now().Unix()-sumData.lasttime <= int64(chainbytime.Value) {
			sumData.tmpkillstmnum++

			if sumData.killstmnum < sumData.tmpkillstmnum {
				sumData.killstmnum = sumData.tmpkillstmnum
			}

			if sumData.tmpkillstmnum == 2 {
				sumData.doubleKillNum++
			}

			if sumData.tmpkillstmnum == 3 {
				sumData.tripleKillNum++
			}
		} else {
			sumData.tmpkillstmnum = 1
			sumData.lasttime = time.Now().Unix()
		}
	}
}

// AddRecoverHp 恢复血量增加
func (sumData *SumData) AddRecoverHp(recoverhp uint32) {
	sumData.recoverHp += recoverhp
}

// IncrCarUserNum 载具使用数量自增
func (sumData *SumData) IncrCarUserNum() {
	sumData.carusernum++
}

// IncrSpeedNum 加速次数自增
func (sumData *SumData) IncrSpeedNum() {
	sumData.speednum++
}

// DisposeAttackNum 助攻次数统计
func (sumData *SumData) DisposeAttackNum(attackid uint64) {
	for i, j := range sumData.attackmap {
		if j+10 < time.Now().Unix() {
			continue
		}

		space := sumData.user.GetSpace().(*Scene)
		if space == nil {
			return
		}

		if !space.IsInOneTeam(i, attackid) {
			continue
		}
		attacker, ok := space.GetEntity(i).(*RoomUser)
		if ok {
			attacker.sumData.attacknum++
		}
	}
}

// DisposeAttackNumMap 助攻次数map填充
func (sumData *SumData) DisposeAttackNumMap(defendid uint64) {
	space := sumData.user.GetSpace().(*Scene)
	tmpdefender, ok := space.GetEntity(defendid).(*RoomUser)
	if ok {
		tmpdefender.sumData.attackmap[sumData.user.GetID()] = time.Now().Unix()
	}
}

// IncrBandageNum 绷带自增
func (sumData *SumData) IncrBandageNum() {
	sumData.bandageNum++
}

// IncrMedicalBoxNum 医疗箱自增
func (sumData *SumData) IncrMedicalBoxNum() {
	sumData.medicalBoxNum++
}

// IncrPainkillerNum 急救包自增
func (sumData *SumData) IncrPainkillerNum() {
	sumData.painkillerNum++
}

// IncrCarDestoryNum 载具摧毁自增
func (sumData *SumData) IncrCarDestoryNum() {
	sumData.carDestoryNum++
}

// DeadType 死亡类型判断,攻击者武器统计
func (sumData *SumData) DeadType(injuredType uint32, defendid uint64) {
	space := sumData.user.GetSpace().(*Scene)
	if space == nil {
		return
	}
	if space.firstDie == 0 {
		space.firstDie = sumData.user.GetDBID()
		sumData.user.Debug("First Die ")
	}
	if defendid != 0 {
		attacker, ok := space.GetEntity(defendid).(iAttacker)
		if ok {
			sumData.openIDByKill, _ = dbservice.Account(attacker.GetDBID()).GetUsername()
			user, ok := space.GetEntityByDBID("Player", attacker.GetDBID()).(*RoomUser)
			if ok {
				sumData.gunIDByKill, sumData.sightIDByKill, sumData.silienceIDByKill, sumData.magazineIDByKill, sumData.stockIDByKill, sumData.handleIDByKill = user.getUseGunInfo()
			}
		}
	}
	if injuredType == 6 {
		sumData.deadIsHead = 1
	}

	if space.IsInOneTeam(sumData.user.GetID(), defendid) {
		sumData.deadType = DEAD_FRIENDKILL
		return
	}

	if injuredType == 0 || injuredType == 6 {
		sumData.deadType = DEAD_GUNBATTLE
		return
	}
	if sumData.gunIDByKill == 0 && injuredType == 1 {
		sumData.deadType = DEAD_HANDKILL
		return
	}
	if injuredType == 2 || injuredType == 5 {
		sumData.deadType = DEAD_SAFEZONEOUT
		return
	}
	if injuredType == 3 {
		sumData.deadType = DEAD_BOMBZONE
		return
	}
	if injuredType == 4 || injuredType == 9 || injuredType == 10 {
		sumData.deadType = DEAD_CARKILL
		return
	}
	if injuredType == 7 {
		sumData.deadType = DEAD_FALLKILL
		return
	}
	if injuredType == 8 {
		sumData.deadType = DEAD_BOMBKILL
		return
	}
	if injuredType == 12 {
		sumData.deadType = DEAD_TANKKILL
		return
	}

	sumData.deadType = DEAD_OFFLINE
}

// KillInfo 玩家击杀成功时 数据记录
func (sumData *SumData) KillInfo(injuredType uint32) {
	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	if injuredType == gunAttack || injuredType == headShotAttack || injuredType == fistAttack {
		weapon := sumData.user.GetInUseWeapon()
		if weapon == 0 { //拳头击杀
			sumData.fistKillNum++
		} else if excelData, ok := excel.GetItem(uint64(weapon)); ok {
			if excelData.Type == 2 && excelData.Subtype == 1 { //狙击枪击杀
				sumData.sniperGunKillNum++
				sumData.user.Debug("sniperGunKillNum ", sumData.sniperGunKillNum)
			}
			if excelData.Type == ItemTypeBazooka {
				sumData.rpgKillNum++
			}
			if excelData.Type == ItemTypeWeaonPistol {
				sumData.pistolKillNum++
			}
		}
	}
	if injuredType == carhit {
		sumData.carKillNum++
	}
	if injuredType == throwdam {
		sumData.grenadeKillNum++
	}
	if injuredType == tankShell {
		sumData.tankKillNum++
	}
	//记录是否本局的首杀
	if scene.firstKill == 0 {
		scene.firstKill = sumData.user.GetDBID()
	}
}

// GetCoin 每局coin计算
func (sumData *SumData) GetCoin(rank uint32, killnum uint32) uint32 {

	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return 0
	}

	var ret uint32

	switch scene.GetMatchMode() {
	//普通模式，勇者模式
	case common.MatchModeNormal, common.MatchModeBrave:
		{
			var singleCoin, twoCoin, fourCoin uint32
			if scene.mapdata.Map_pxs_ID == 2 {
				base, ok := excel.GetNewmapaward(uint64(rank))
				if !ok {
					return ret
				}
				singleCoin, twoCoin, fourCoin = uint32(base.Single), uint32(base.Two), uint32(base.Four)
			} else {
				base, ok := excel.GetResultcoin(uint64(rank))
				if !ok {
					return ret
				}
				singleCoin, twoCoin, fourCoin = uint32(base.Single), uint32(base.Two), uint32(base.Four)
			}

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					ret += twoCoin
				} else if scene.teamMgr.teamType == 1 {
					ret += fourCoin
				}
			} else {
				ret += singleCoin
			}

			ret += killnum * uint32(common.GetTBSystemValue(common.System_KillGetCoin))
		}
	//快速模式
	case common.MatchModeScuffle, common.MatchModeEndless:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetFunRewards(rank, 2, common.Item_Coin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetFunRewards(rank, 4, common.Item_Coin)
				}
			} else {
				rewards = common.GetFunRewards(rank, 1, common.Item_Coin)
			}

			if rewards == nil {
				sumData.user.Error("GetFunRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_Coin]
		}
	//精英模式
	case common.MatchModeElite:
		{
			if scene.teamMgr.isTeam {
				return ret
			}

			rewards := common.GetDragonRewards(rank, common.Item_Coin)
			if rewards == nil {
				sumData.user.Error("GetDragonRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_Coin]
		}
	//红蓝对决模式
	case common.MatchModeVersus:
		{
			if scene.teamMgr.isTeam {
				return ret
			}

			rewards := common.GetRedAndBlueRewards(rank, common.Item_Coin)
			if rewards == nil {
				sumData.user.Error("GetRedAndBlueRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_Coin]
		}
	//娱乐模式
	case common.MatchModeArcade:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetFunmode(rank, 2, common.Item_Coin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetFunmode(rank, 4, common.Item_Coin)
				}
			} else {
				rewards = common.GetFunmode(rank, 1, common.Item_Coin)
			}

			if rewards == nil {
				sumData.user.Error("GetFunmode failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_Coin]
		}
	//坦克大战
	case common.MatchModeTankWar:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetTankWarmode(rank, 2, common.Item_Coin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetTankWarmode(rank, 4, common.Item_Coin)
				}
			} else {
				rewards = common.GetTankWarmode(rank, 1, common.Item_Coin)
			}

			if rewards == nil {
				sumData.user.Error("GetTankWarmode failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_Coin]
		}
	}

	sumData.user.Info("GetCoin before addition, num: ", ret, " rank: ", rank, " killnum: ", killnum)

	//战友组队金币加成
	if scene.teamMgr.isComradeGame(sumData.user) {
		add, ok := excel.GetAddition(uint64(common.AdditionBonus_ComradeCoin))
		if ok {
			ret += uint32(float32(ret) * float32(add.Value/100))
		}
	}

	vRate := common.VeteranCoinRate(sumData.user.calVeteranNum())
	iRate := float32(sumData.user.getMaxItemBonus(ItemTypeCoinBonus)) / 100.0
	return uint32(float32(ret) * (1 + vRate) * (1 + iRate))
}

// GetBraveCoin 获取本场比赛获得的勇气值
func (sumData *SumData) GetBraveCoin(rank uint32, killnum uint32) uint32 {
	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return 0
	}

	var ret uint32

	switch scene.GetMatchMode() {
	//勇者战场
	case common.MatchModeBrave:
		{
			base, ok := excel.GetBravereward(uint64(rank))
			if !ok {
				sumData.user.Error("Notf Found Reward ", rank)
				return ret
			}

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					ret += uint32(base.Two)
				} else if scene.teamMgr.teamType == 1 {
					ret += uint32(base.Four)
				}
			} else {
				ret += uint32(base.Single)
			}

			ret += killnum * 10
		}
		//快速模式
	case common.MatchModeScuffle, common.MatchModeEndless:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetFunRewards(rank, 2, common.Item_BraveCoin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetFunRewards(rank, 4, common.Item_BraveCoin)
				}
			} else {
				rewards = common.GetFunRewards(rank, 1, common.Item_BraveCoin)
			}

			if rewards == nil {
				sumData.user.Error("GetFunRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_BraveCoin]
		}
		//精英模式
	case common.MatchModeElite:
		{
			if scene.teamMgr.isTeam {
				return ret
			}

			rewards := common.GetDragonRewards(rank, common.Item_BraveCoin)
			if rewards == nil {
				sumData.user.Error("GetDragonRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_BraveCoin]
		}
		//红蓝对决模式
	case common.MatchModeVersus:
		{
			if scene.teamMgr.isTeam {
				return ret
			}

			rewards := common.GetRedAndBlueRewards(rank, common.Item_BraveCoin)
			if rewards == nil {
				sumData.user.Error("GetRedAndBlueRewards failed, rank: ", rank)
				return ret
			}

			ret += rewards[common.Item_BraveCoin]
		}
	//娱乐模式
	case common.MatchModeArcade:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetFunmode(rank, 2, common.Item_BraveCoin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetFunmode(rank, 4, common.Item_BraveCoin)
				}
			} else {
				rewards = common.GetFunmode(rank, 1, common.Item_BraveCoin)
			}

			if rewards == nil {
				return ret
			}

			ret += rewards[common.Item_BraveCoin]
		}
	//坦克大战
	case common.MatchModeTankWar:
		{
			var rewards map[uint32]uint32

			if scene.teamMgr.isTeam {
				if scene.teamMgr.teamType == 0 {
					rewards = common.GetTankWarmode(rank, 2, common.Item_BraveCoin)
				} else if scene.teamMgr.teamType == 1 {
					rewards = common.GetTankWarmode(rank, 4, common.Item_BraveCoin)
				}
			} else {
				rewards = common.GetTankWarmode(rank, 1, common.Item_BraveCoin)
			}

			if rewards == nil {
				return ret
			}

			ret += rewards[common.Item_BraveCoin]
		}
	}

	sumData.user.Info("GetBraveCoin before addition, num: ", ret, " rank: ", rank, " killnum: ", killnum)

	iRate := float32(sumData.user.getMaxItemBonus(ItemTypeBraveBonus)) / 100.0
	return uint32(float32(ret) * (1 + iRate))
}

// RunDistance 跑动距离更新
func (sumData *SumData) RunDistance() {
	state := sumData.user.GetBaseState()
	if state == RoomPlayerBaseState_Watch || state == RoomPlayerBaseState_Dead {
		return
	}

	if !sumData.tmprunDistance.IsEqual(linmath.Vector3_Invalid()) {
		runDistance := common.Distance(sumData.user.GetPos(), sumData.tmprunDistance)
		sumData.runDistance += runDistance
		sumData.tmprunDistance = sumData.user.GetPos()
	}
}

// CarDistance 载具使用距离更新
func (sumData *SumData) CarDistance() {
	state := sumData.user.GetBaseState()
	if state == RoomPlayerBaseState_Watch || state == RoomPlayerBaseState_Dead {
		return
	}

	if !sumData.tmpcarDistance.IsEqual(linmath.Vector3_Invalid()) {
		carDistance := common.Distance(sumData.user.GetPos(), sumData.tmpcarDistance)
		sumData.carDistance += carDistance
		sumData.tmpcarDistance = sumData.user.GetPos()
	}
}

// SwimDistance 游泳距离更新
func (sumData *SumData) SwimDistance() {
	state := sumData.user.GetBaseState()
	if state == RoomPlayerBaseState_Watch || state == RoomPlayerBaseState_Dead {
		return
	}

	if !sumData.tmpswimDistance.IsEqual(linmath.Vector3_Invalid()) {
		swimDistance := common.Distance(sumData.user.GetPos(), sumData.tmpswimDistance)
		sumData.swimDistance += swimDistance
		sumData.tmpswimDistance = sumData.user.GetPos()
	}
}

// singleKillWinRating 本场rating分差值计算
func (sumData *SumData) singleKillWinRating() {
	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	ratingMap := excel.GetRatingMap()

	myRating := scene.scores[sumData.user.GetDBID()]
	averageRating := scene.averageRating
	if sumData.season != common.GetSeason() {
		myRating = float64(ratingMap[4001].Value)
		averageRating = float64(ratingMap[4001].Value)
	}

	var s, k1, k2, k3, rp, p float32
	wr, r := 0, myRating
	rf, rs := averageRating, myRating
	c1, c2, c3, c4, c5, e, num, f, b1, b2, ratingFloor, ratingUpper := ratingMap[4006].Value, ratingMap[4007].Value, ratingMap[4008].Value, ratingMap[4009].Value, ratingMap[4010].Value, ratingMap[4016].Value, ratingMap[4017].Value, ratingMap[4018].Value, ratingMap[4004].Value, ratingMap[4005].Value, ratingMap[4027].Value, ratingMap[4028].Value

	if rf < float64(ratingFloor) || rf > float64(ratingUpper) {
		rf = float64(ratingMap[4001].Value)
	}
	if rs < float64(ratingFloor) || rs > float64(ratingUpper) {
		rs = float64(ratingMap[4001].Value)
		r = rs
	}

	var battleType string
	var totalPlayerNum uint32
	switch sumData.battletype {
	case 0:
		battleType = common.PlayerCareerSoloData
		k3 = ratingMap[4011].Value
		totalPlayerNum = scene.maxUserSum

		r = myRating
	case 1:
		battleType = common.PlayerCareerDuoData
		k3 = ratingMap[4012].Value
		totalPlayerNum = scene.teamMgr.TeamsSum()

		if team, ok := scene.teamMgr.teams[sumData.user.GetUserTeamID()]; ok {
			var teamRating, i float64 = 0, 0
			for _, memid := range team {
				info := scene.teamMgr.GetSpaceMemberInfo(memid)
				if info == nil {
					continue
				}
				i++
				teamRating += float64(info.rating)
			}

			if i != 0 {
				r = teamRating / i
			}
		}
	case 2:
		battleType = common.PlayerCareerSquadData
		k3 = ratingMap[4013].Value
		totalPlayerNum = scene.teamMgr.TeamsSum()

		if team, ok := scene.teamMgr.teams[sumData.user.GetUserTeamID()]; ok {
			var teamRating, i float64 = 0, 0
			for _, memid := range team {
				info := scene.teamMgr.GetSpaceMemberInfo(memid)
				if info == nil {
					continue
				}
				i++
				teamRating += float64(info.rating)
			}

			if i != 0 {
				r = teamRating / i
			}
		}
	default:
		sumData.user.Error("Unkonwn battle typ: ", sumData.battletype)
		return
	}
	if totalPlayerNum == 0 {
		return
	}

	s = c1 * float32(math.Pow((float64(c2)*math.Log10((rf+rs)/rs+float64(c3))/rs), math.Log10(float64(c4)*rf/rs+float64(c5)*rs/rf)))
	rp = (float32(totalPlayerNum) + 1 - float32(sumData.rank)) / float32(totalPlayerNum)
	p = 1 / (1 + float32(math.Pow(float64(b1), (rf-r)/float64(b2))))

	data := &datadef.CareerBase{}
	dbUtil := db.PlayerCareerDataUtil(battleType, sumData.user.GetDBID(), common.GetSeason())
	if err := dbUtil.GetRoundData(data); err != nil {
		sumData.user.Error("GetRoundData err: ", err)
		return
	}

	if rp > p {
		data.WinStreak++
		data.FailStreak = 0
		if data.WinStreak > uint32(num) {
			wr = 1
		}
	} else if rp < p {
		data.WinStreak = 0
		data.FailStreak++
		if data.FailStreak > uint32(num) {
			wr = 1
		}
	} else {
		data.WinStreak = 0
		data.FailStreak = 0
	}

	if err := dbUtil.SetRoundData(data); err != nil {
		sumData.user.Error("SetRoundData err: ", err)
		return
	}

	k1 = 1 + e*float32(wr)
	k2 = 1 + f/float32(data.TotalBattleNum+1)
	sumData.user.Debug("c1:", c1, " c2:", c2, " c3:", c3, " c4:", c4, " c5:", c5, " rf:", rf, " rs:", rs, " e:", e, " wr:", wr, " f:", f, " tr:", data.TotalBattleNum+1, " k3:", k3, " b1:", b1, " b2:", b2, " r:", r)
	sumData.user.Debug("s:", s, " k1:", k1, " k2:", k2, " k3:", k3, " rp:", rp, " p;", p, " totalPlayerNum:", totalPlayerNum, " rank:", sumData.rank)
	sumData.diffwinRating = s * k1 * k2 * k3 * (rp - p)

	// diffkillRating 计算
	var sgb, rp0, p0 float32
	m, g1, g2, maxK, maxD, h1, h2, h3 := ratingMap[4026].Value, ratingMap[4019].Value, ratingMap[4020].Value, ratingMap[4022].Value, ratingMap[4021].Value, ratingMap[4023].Value, ratingMap[4024].Value, ratingMap[4025].Value

	if float32(sumData.user.kill) < maxK {
		maxK = float32(sumData.user.kill)
	}
	if float32(sumData.user.effectharm) < maxD {
		maxD = float32(sumData.user.effectharm)
	}

	sgb = s / m
	rp0 = maxK/g1 + maxD/g2
	p0 = h1*float32(math.Log10(float64(100*p))/math.Log10(float64(h3))) + h2
	sumData.user.Debug("s:", s, " m:", m, " maxk:", maxK, " maxD:", maxD, " g1:", g1, " g2:", g2, " h1:", h1, " p:", p, " h3:", h3, " h2:", h2)
	sumData.user.Debug("sgb:", sgb, " rp0:", rp0, " p0:", p0)
	sumData.diffkillRating = sgb * (rp0 - p0)
}

// killwinRating rating分统计
func (sumData *SumData) totalKillWinRating() {
	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	ratingMap := excel.GetRatingMap()
	ow, a1, a2, k, n, ratingFloor, ratingUpper := ratingMap[4001].Value, ratingMap[4002].Value, ratingMap[4003].Value, ratingMap[4014].Value, ratingMap[4015].Value, ratingMap[4027].Value, ratingMap[4028].Value

	if sumData.season != common.GetSeason() {
		sumData.careerdata = &datadef.CareerBase{}
	}

	variation := sumData.diffwinRating*a1 + sumData.diffkillRating*a2
	if math.IsNaN(float64(variation)) {
		variation = 0
	}
	if math.IsNaN(float64(sumData.careerdata.SoloRating)) {
		sumData.careerdata.SoloRating = ow
	}
	if math.IsNaN(float64(sumData.careerdata.DuoRating)) {
		sumData.careerdata.DuoRating = ow
	}
	if math.IsNaN(float64(sumData.careerdata.SquadRating)) {
		sumData.careerdata.SquadRating = ow
	}

	switch sumData.battletype {
	case 0:
		totalPlayerNum := scene.maxUserSum
		if totalPlayerNum == 0 {
			return
		}

		if variation < 0 {
			if sumData.careerdata.SoloRating < ratingFloor || sumData.careerdata.SoloRating > ratingUpper {
				if ow >= k {
					sumData.careerdata.SoloRating = ow + variation
				} else {
					sumData.careerdata.SoloRating = ow + n*(float32(totalPlayerNum)+1-float32(sumData.rank))/float32(totalPlayerNum)
				}
			} else if sumData.careerdata.SoloRating < k {
				sumData.careerdata.SoloRating += n * (float32(totalPlayerNum) + 1 - float32(sumData.rank)) / float32(totalPlayerNum)
			} else {
				sumData.careerdata.SoloRating += variation
			}
		} else {
			if sumData.careerdata.SoloRating < ratingFloor || sumData.careerdata.SoloRating > ratingUpper {
				sumData.careerdata.SoloRating = ow + variation
			} else {
				sumData.careerdata.SoloRating += variation
			}
		}
	case 1:
		totalPlayerNum := scene.teamMgr.TeamsSum()
		if totalPlayerNum == 0 {
			return
		}

		if variation < 0 {
			if sumData.careerdata.DuoRating < ratingFloor || sumData.careerdata.DuoRating > ratingUpper {
				if ow >= k {
					sumData.careerdata.DuoRating = ow + variation
				} else {
					sumData.careerdata.DuoRating = ow + n*(float32(totalPlayerNum)+1-float32(sumData.rank))/float32(totalPlayerNum)
				}
			} else if sumData.careerdata.DuoRating < k {
				sumData.careerdata.DuoRating += n * (float32(totalPlayerNum) + 1 - float32(sumData.rank)) / float32(totalPlayerNum)
			} else {
				sumData.careerdata.DuoRating += variation
			}
		} else {
			if sumData.careerdata.DuoRating < ratingFloor || sumData.careerdata.DuoRating > ratingUpper {
				sumData.careerdata.DuoRating = ow + variation
			} else {
				sumData.careerdata.DuoRating += variation
			}
		}
	case 2:
		totalPlayerNum := scene.teamMgr.TeamsSum()
		if totalPlayerNum == 0 {
			return
		}

		if variation < 0 {
			if sumData.careerdata.SquadRating < ratingFloor || sumData.careerdata.SquadRating > ratingUpper {
				if ow >= k {
					sumData.careerdata.SquadRating = ow + variation
				} else {
					sumData.careerdata.SquadRating = ow + n*(float32(totalPlayerNum)+1-float32(sumData.rank))/float32(totalPlayerNum)
				}
			} else if sumData.careerdata.SquadRating < k {
				sumData.careerdata.SquadRating += n * (float32(totalPlayerNum) + 1 - float32(sumData.rank)) / float32(totalPlayerNum)
			} else {
				sumData.careerdata.SquadRating += variation
			}
		} else {
			if sumData.careerdata.SquadRating < ratingFloor || sumData.careerdata.SquadRating > ratingUpper {
				sumData.careerdata.SquadRating = ow + variation
			} else {
				sumData.careerdata.SquadRating += variation
			}
		}
	default:
		sumData.user.Error("Unkonwn battle typ:", sumData.battletype)
		return
	}
}

// rankCount 本场rank统计
func (sumData *SumData) rankCal() {
	if sumData.shotnum == 0 {
		//sumData.user.Info("没有开过枪")
		sumData.shotnum = 1
	}

	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	if scene.teamMgr.isTeam {
		if sumData.user.bVictory {
			sumData.rank = scene.teamMgr.SurplusTeamSum()
		} else {
			sumData.rank = scene.teamMgr.SurplusTeamSum() + 1
		}

		if scene.teamMgr.teamType == 0 {
			sumData.battletype = 1
		} else if scene.teamMgr.teamType == 1 {
			sumData.battletype = 2
		} else {
			sumData.user.Warn("Unknown battle type: ", sumData.battletype)
		}
	} else {
		if sumData.user.bVictory {
			sumData.rank = uint32(scene.getMemSum())
		} else {
			sumData.rank = uint32(scene.getMemSum()) + 1
		}
		sumData.battletype = 0
	}
}

// totalScore 综合评分计算
func (sumData *SumData) totalScore(min float32, mid float32, max float32) float32 {
	var tmp float32
	rating := excel.GetRatingMap()
	if min > mid {
		tmp = min
		min = mid
		mid = tmp
	}
	if mid > max {
		tmp = mid
		mid = max
		max = tmp
	}
	if min > mid {
		tmp = min
		min = mid
		mid = tmp
	}
	return rating[2501].Value*max + rating[2502].Value*mid + rating[2503].Value*min
}

//AchievementData 刷新成就数据
func (sumData *SumData) AchievementData() {
	scene := sumData.user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	freshData := map[uint64]float64{}
	freshData[common.AchievementBattleNum] = 1
	if sumData.rank == 1 {
		if sumData.battletype == 0 {
			freshData[common.AchievementSoloTopOne] = 1
		}
		if sumData.battletype == 1 {
			freshData[common.AchievementDuoTopOne] = 1
		}
		if sumData.battletype == 2 {
			freshData[common.AchievementSquadTopOne] = 1
		}
	}
	if sumData.rank <= 10 {
		freshData[common.AchievementTopTen] = 1
	}
	if sumData.user.effectharm > 0 {
		freshData[common.AchievementDamageTotal] = float64(sumData.user.effectharm)
	}
	if sumData.user.kill > 0 {
		freshData[common.AchievementKillTotal] = float64(sumData.user.kill)
	}
	if sumData.runDistance > 0 {
		freshData[common.AchievementRunDistance] = float64(sumData.runDistance)
	}
	if sumData.fistKillNum > 0 {
		freshData[common.AchievementFistKill] = float64(sumData.fistKillNum)
	}
	if sumData.grenadeKillNum > 0 {
		freshData[common.AchievementGrenadeKill] = float64(sumData.grenadeKillNum)
	}
	if sumData.carKillNum > 0 {
		freshData[common.AchievementCarKill] = float64(sumData.carKillNum)
	}
	if sumData.sniperGunKillNum > 0 {
		freshData[common.AchievementSniperKill] = float64(sumData.sniperGunKillNum)
	}
	if scene.firstDie == sumData.user.GetDBID() {
		freshData[common.AchievementFirstDie] = 1
	}
	if scene.firstKill == sumData.user.GetDBID() {
		freshData[common.AchievementFirstKill] = 1
	}
	util := db.PlayerAchievementUtil(sumData.user.GetDBID())
	result := map[uint64]float64{}
	for k, v := range freshData {
		result[k] = util.AddAchievementData(uint32(k), v)
		if k == common.AchievementRunDistance {
			result[k] = result[k] / 1000
		}
	}
	msg := sumData.user.FreshAchievement(sumData.user.GetDBID(), result)
	if msg != nil {
		sumData.user.RPC(iserver.ServerTypeClient, "AchievementAddNotify", msg)
	}
}

//FreshAchievement 刷新成就进度
func (user *RoomUser) FreshAchievement(uid uint64, freshData map[uint64]float64) *protoMsg.AchievmentInfo {
	if len(freshData) == 0 {
		return nil
	}
	util := db.PlayerAchievementUtil(uid)
	achieveProcess := util.GetAchieveInfo()
	var needSave []*db.AchieveInfo
	excelData := excel.GetAccomplishmentMap()
	level, exp := util.GetLevelInfo()
	for _, data := range excelData {
		newValue, ok := freshData[data.Condition1]
		if !ok {
			continue
		}

		oneAchieve := achieveProcess[data.Id]
		if oneAchieve == nil {
			oneAchieve = &db.AchieveInfo{
				Id: data.Id,
			}
		}
		oneProcess := len(oneAchieve.Stamp)
		var fresh bool
		for k, num := range data.Amount {
			if oneProcess >= k+1 {
				continue
			}
			if newValue >= float64(num) {
				oneAchieve.Stamp = append(oneAchieve.Stamp, uint64(time.Now().Unix()))
				oneAchieve.Flag = 1 // 1:代表是新获得的成就, 0:旧成就
				fresh = true
				exp += data.Experience[k]
				levelData, ok := excel.GetAchievementLevel(uint64(level))
				if ok {
					for exp >= uint32(levelData.Experience) {
						needExp := levelData.Experience
						levelData, ok = excel.GetAchievementLevel(uint64(level + 1))
						if !ok {
							break
						}
						exp -= uint32(needExp)
						level++
					}
				}
				user.tlogAchievementFlow(uint32(oneAchieve.Id), level, uint32(exp))
			}
		}
		if fresh {
			needSave = append(needSave, oneAchieve)
		}
	}
	if len(needSave) > 0 {
		util.AddAchieve(needSave)
		util.SetExp(exp)
		util.SetLevel(level)

		msg := &protoMsg.Uint32Array{}
		for _, v := range needSave {
			msg.List = append(msg.List, uint32(v.Id))
		}
		user.RPC(common.ServerTypeLobby, "UpdateAchievement", msg)
	}
	msg := &protoMsg.AchievmentInfo{}
	msg.Level = level
	msg.Exp = exp
	for _, v := range needSave {
		msg.List = append(msg.List, &protoMsg.Achievement{
			Id:    uint32(v.Id),
			Stamp: v.Stamp,
			Flag:  v.Flag,
		})
	}
	for k, v := range freshData {
		msg.Process = append(msg.Process, &protoMsg.AchievementProcess{
			Id:  uint32(k),
			Num: float32(v),
		})
	}
	return msg
}

//HonorInfo 统计本局游戏的勋章数据
func (sumData *SumData) HonorInfo() {
	util := db.PlayerInsigniaUtil(sumData.user.GetDBID())
	add := map[uint64]uint8{}
	flag := map[uint32]uint32{}
	insigniaInfo := util.GetInsignia()
	for _, data := range excel.GetMedalMap() {
		complete := false
		switch data.Id {
		case common.HonorKill:
			complete = sumData.user.kill >= uint32(data.Condition)
		case common.HonorHeadShot:
			complete = sumData.user.headshotnum >= uint32(data.Condition)
		case common.HonorRecover:
			complete = sumData.recoverHp >= uint32(data.Condition)
		case common.HonorHarm:
			complete = sumData.user.effectharm >= uint32(data.Condition)
		case common.HonorMultiKill:
			complete = sumData.killstmnum >= uint32(data.Condition)
		case common.HonorWin:
			complete = sumData.rank == 1
		case common.HonorKillDistance:
			complete = sumData.killDistanceNum >= uint32(data.Condition)
		case common.HonorRank:
			complete = sumData.rank <= uint32(data.Condition)
		case common.HonorRescue:
			complete = sumData.rescueNum >= uint32(data.Condition)
		default:
			complete = false
		}
		if complete {
			insigniaInfo[uint32(data.Id)]++
			//util.AddInsignia(data.Id)
			add[data.Id] = 1

			if insigniaInfo[uint32(data.Id)] == uint32(data.Lv2_amount) {
				util.SetInsigniaFlag(uint32(data.Id), 1) // 设置新获得的勋章红点标记
				flag[uint32(data.Id)] = 1
			}
		}
	}
	if len(add) > 0 {
		util.SetInsignia(insigniaInfo)
		msg := &protoMsg.InsigniaInfo{}
		for id, num := range insigniaInfo {
			if _, ok := add[uint64(id)]; !ok {
				continue
			}
			msg.Info = append(msg.Info, &protoMsg.Insignia{
				Id:    id,
				Count: num,
				Flag:  flag[id],
			})
			sumData.user.tlogInsigniaFlow(id, num)
		}
		sumData.user.RPC(iserver.ServerTypeClient, "InsigniaAddNotify", msg)
	}
}
