
-- ----------------------------
-- Table structure for playerdaydata
-- ----------------------------

CREATE TABLE `playerdaydata` (
  `DayID` bigint(30) unsigned NOT NULL DEFAULT '0' COMMENT '本局ID',
  `Season` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '赛季',
  `UID` bigint(30) unsigned NOT NULL DEFAULT '0' COMMENT '玩家UID',
  `NowTime` int(20) unsigned NOT NULL DEFAULT '0' COMMENT '该天最后一场时间',
  `StartTime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '记录那一天的数据',
  `Model` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '游戏模式',
  `DayFirstNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总吃鸡数',
  `DayTopTenNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总前10名',
  `Rating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '本模式的rating分',
  `WinRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT 'winrating分',
  `KillRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT 'killrating',
  `DayEffectHarm` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总有效伤害',
  `DayShotNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总开枪次数',
  `DaySurviveTime` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总生存时间',
  `DayDistance` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '一天总行进距离',
  `DayAttackNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天总助攻次数',
  `DayRecoverNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '一天治疗道具使用次数',
  `DayReviveNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '每天复活次数',
  `DayHeadShotNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '每天爆头次数',
  `DayBattleNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '战斗场次',
  `DayKillNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '击杀数量',
  `TotalRank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总排名',
  `Rank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '该模式的排名',
  `UserName` varchar(50) NOT NULL DEFAULT '' COMMENT '玩家名字',
  UNIQUE KEY `DayID` (`UID`,`DayID`,`Model`) USING BTREE,
  KEY `ix_UID` (`UID`),
  KEY `ix_DayID` (`DayID`),
  KEY `ix_UserName` (`UserName`),
  KEY `ix_Model` (`Model`),
  KEY `ix_StartTime` (`StartTime`),
  KEY `ix_NowTime` (`NowTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for playerrounddata
-- ----------------------------

CREATE TABLE `playerrounddata` (
  `GameID` bigint(30) unsigned NOT NULL DEFAULT '0' COMMENT '本局ID',
  `UID` bigint(30) unsigned NOT NULL DEFAULT '0' COMMENT '玩家UID',
  `Season` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '赛季',
  `Rank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '排名',
  `KillNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '击杀数量',
  `HeadShotNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '爆头数量',
  `EffectHarm` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '有效伤害值',
  `RecoverUseNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '治疗道具使用次数',
  `ShotNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '开枪次数',
  `ReviveNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '复活次数',
  `KillDistance` float(10,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '最远击杀距离',
  `KillStmNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '最大连杀次数',
  `FinallHp` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '最终血量',
  `RecoverHp` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '本局总恢复血量',
  `RunDistance` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '行进距离',
  `CarUseNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '载具使用使用数量',
  `CarDestoryNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '载具摧毁数量',
  `AttackNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '助攻次数',
  `SpeedNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '加速次数',
  `Coin` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '金币数',
  `StartUnix` int(20) unsigned NOT NULL DEFAULT '0' COMMENT '开始时间戳',
  `EndUnix` int(20) unsigned NOT NULL DEFAULT '0' COMMENT '结束时间戳',
  `StartTime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '战斗开始时间',
  `EndTime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '结束时间',
  `BattleType` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '战斗类型（单、双、四）',
  `KillRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '本局击杀评分',
  `WinRatig` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '本局胜利评分',
  `SoloKillRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前单人击杀rating',
  `SoloWinRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前单人胜利rating',
  `SoloRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前单人rating分',
  `SoloRank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '全服单人模式排名',
  `DuoKillRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前双人击杀rating',
  `DuoWinRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前双人胜利rating',
  `DuoRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前双人rating',
  `DuoRank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '全服双人模式排名',
  `SquadKillRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前四人击杀rating',
  `SquadWinRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前四人胜利rating',
  `SquadRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '当前四人rating分',
  `SquadRank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '全服四人模式排名',
  `TotalRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '综合评分',
  `TotalRank` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '全服综合分总排名',
  `TopRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '全服最高评分',
  `TotalBattleNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总场次',
  `TotalFirstNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总吃鸡场次',
  `TotalTopTenNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '前10总场次',
  `TotalHeadShot` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总爆头数',
  `TotalKillNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总击杀量',
  `TotalShotNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总开枪次数',
  `TotalEffectHarm` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总有效伤害',
  `TotalSurviveTime` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总生存时间',
  `TotalDistance` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '总行进距离',
  `TotalRecvItemUseNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总治疗道具使用次数',
  `SingleMaxKill` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总场次中最高击杀',
  `SingleMaxHeadShot` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '总场次中最大爆头数',
  `DeadType` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '死亡类型',
  `UserName` varchar(50) NOT NULL DEFAULT '' COMMENT '玩家名字',
  `SkyBox` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '本局天空盒',
  `PlatID` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '平台ID',
  `LoginChannel` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '登录渠道',
  UNIQUE KEY `GameID` (`UID`,`GameID`) USING BTREE,
  KEY `ix_UID` (`UID`),
  KEY `ix_GameID` (`GameID`),
  KEY `ix_UserName` (`UserName`),
  KEY `ix_StartTime` (`StartTime`),
  KEY `ix_StartUnix` (`StartUnix`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for playersearchnum
-- ----------------------------

CREATE TABLE `playersearchnum` (
  `UID` bigint(30) unsigned NOT NULL DEFAULT '0' COMMENT '玩家UID',
  `UserName` varchar(50) NOT NULL DEFAULT '' COMMENT '玩家名字',
  `SearchNum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '搜索次数',
  `TotalRating` float(30,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '总rating',
  PRIMARY KEY (`UID`),
  KEY `ix_UserName` (`UserName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
