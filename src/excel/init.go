package excel

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/fsnotify/fsnotify"
)

type LoadFunc func()

var loadFuncs = map[string]LoadFunc{
	"LoadingTips":       LoadLoadingTips,
	"Errorcode":         LoadErrorcode,
	"Mail":              LoadMail,
	"Item":              LoadItem,
	"Mapitem":           LoadMapitem,
	"Mapitemrate":       LoadMapitemrate,
	"Logic":             LoadLogic,
	"Carrier":           LoadCarrier,
	"Bombrule":          LoadBombrule,
	"Meterfix":          LoadMeterfix,
	"ChallengeTask":     LoadChallengeTask,
	"Challenge":         LoadChallenge,
	"SeasonGrade":       LoadSeasonGrade,
	"Award":             LoadAward,
	"Season":            LoadSeason,
	"Matchmode":         LoadMatchmode,
	"Rating":            LoadRating,
	"Role":              LoadRole,
	"Festival":          LoadFestival,
	"Energy":            LoadEnergy,
	"BindingAwards":     LoadBindingAwards,
	"System2":           LoadSystem2,
	"Activity":          LoadActivity,
	"SpecialReplace":    LoadSpecialReplace,
	"SpecialLevel":      LoadSpecialLevel,
	"SpecialTask":       LoadSpecialTask,
	"Adkey":             LoadAdkey,
	"Promotion":         LoadPromotion,
	"Exchangestore":     LoadExchangestore,
	"Activeness":        LoadActiveness,
	"Task":              LoadTask,
	"Gun":               LoadGun,
	"ReversionWard":     LoadReversionWard,
	"NewyearCheckin":    LoadNewyearCheckin,
	"NewyearTargetPool": LoadNewyearTargetPool,
	"SkillSystem":       LoadSkillSystem,
	"SkillEffect":       LoadSkillEffect,
	"Skill":             LoadSkill,
	"Phonenumber":       LoadPhonenumber,
	"ComradeTask":       LoadComradeTask,
	"Accomplishment":    LoadAccomplishment,
	"Achievement":       LoadAchievement,
	"AchievementLevel":  LoadAchievementLevel,
	"QuickChat":         LoadQuickChat,
	"Quickcommand":      LoadQuickcommand,
	"Detail":            LoadDetail,
	"Keywords":          LoadKeywords,
	"Box":               LoadBox,
	"Tankmode":          LoadTankmode,
	"Skillmode":         LoadSkillmode,
	"Funmode":           LoadFunmode,
	"Redandblue":        LoadRedandblue,
	"Dragonreward":      LoadDragonreward,
	"Bravereward":       LoadBravereward,
	"Luandoureward":     LoadLuandoureward,
	"Characterstate":    LoadCharacterstate,
	"Cheater_report":    LoadCheater_report,
	"Skybox":            LoadSkybox,
	"Maps":              LoadMaps,
	"Maprule":           LoadMaprule,
	"Boxreward":         LoadBoxreward,
	"Store":             LoadStore,
	"Goods":             LoadGoods,
	"Odds":              LoadOdds,
	"Seasontik":         LoadSeasontik,
	"Selling":           LoadSelling,
	"Newmapaward":       LoadNewmapaward,
	"Resultcoin":        LoadResultcoin,
	"Changename":        LoadChangename,
	"Name":              LoadName,
	"ChickenCheckin":    LoadChickenCheckin,
	"Backreward":        LoadBackreward,
	"Callbackreward":    LoadCallbackreward,
	"Callback":          LoadCallback,
	"Match":             LoadMatch,
	"Medal":             LoadMedal,
	"Addition":          LoadAddition,
	"Explaobing":        LoadExplaobing,
	"Explevel":          LoadExplevel,
	"Expmode":           LoadExpmode,
	"Expkill":           LoadExpkill,
	"Exprank":           LoadExprank,
	"Militaryrank":      LoadMilitaryrank,
	"Paysystem2":        LoadPaysystem2,
	"Paysystem":         LoadPaysystem,
	"Pay":               LoadPay,
	"WorldCupBattle":    LoadWorldCupBattle,
	"WorldCupChampion":  LoadWorldCupChampion,
	"BallStarList":      LoadBallStarList,
	"BallStarRound":     LoadBallStarRound,
	"System":            LoadSystem,
	"Ai_spawn":          LoadAi_spawn,
	"Aisys":             LoadAisys,
	"Ai":                LoadAi,
	"ThreedayCheckin":   LoadThreedayCheckin,
}

func init() {
	for _, f := range loadFuncs {
		f()
	}

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Error(err)
			return
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case ev := <-watcher.Events:
					log.Infof("Watch %s Op %s", ev.Name, ev.Op)

					if strings.Contains(ev.Name, "reload.json") {
						if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
							log.Info("Start Reload!")

							if err := Reload("../res/excel/reload.json"); err != nil {
								log.Error("Reload failed ", err)
							} else {
								log.Info("Reload excel success")
							}
						}
					}
				case err := <-watcher.Errors:
					log.Error(err)
				}
			}
		}()

		if err = watcher.Add("../res/excel/"); err != nil {
			log.Error(err)
		} else {
			log.Info("Watch excel path")
		}

		<-done
	}()

}

func Reload(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	config := make(map[string]bool)
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	for k, v := range config {
		if v {
			f, ok := loadFuncs[k]
			if ok {
				log.Info("Reload ", k)
				f()
			} else {
				log.Error("Reload ", k, " failed, can't find the load function")
			}
		}
	}

	return nil
}

func ReloadAll() {
	for _, f := range loadFuncs {
		f()
	}
}
