package service

import (
	"math"
	"tesou.io/platform/foot-parent/foot-api/common/base"
	entity5 "tesou.io/platform/foot-parent/foot-api/module/analy/pojo"
	"tesou.io/platform/foot-parent/foot-api/module/match/pojo"
	entity3 "tesou.io/platform/foot-parent/foot-api/module/odds/pojo"
	"tesou.io/platform/foot-parent/foot-core/module/analy/constants"
	"tesou.io/platform/foot-parent/foot-core/module/match/service"
	"time"
)

type C3Service struct {
	AnalyService
	service.BFScoreService
	service.BFBattleService
	service.BFFutureEventService
	service.BFJinService

	//最大让球数据
	MaxLetBall float64
}

func (this *C3Service) ModelName() string {
	return "C3"
}

func (this *C3Service) Analy(analyAll bool) {
	this.AnalyService.Analy(analyAll, this)

}

func (this *C3Service) Analy_Near() {
	this.AnalyService.Analy_Near(this)
}

/**
  -1 参数错误
  -2 不符合让球数
  -3 计算分析错误
  0  新增的分析结果
  1  需要更新结果
 */
func (this *C3Service) analyStub(v *pojo.MatchLast) (int, *entity5.AnalyResult) {
	temp_data := this.Find(v.Id, this.ModelName())
	matchId := v.Id
	//声明使用变量
	var a18Bet *entity3.AsiaHis
	//亚赔
	aList := this.AsiaHisService.FindByMatchIdCompId(matchId, constants.C1_REFER_ASIA)
	if len(aList) < 1 {
		return -1, temp_data
	}
	a18Bet = aList[0]
	temp_data.LetBall = a18Bet.ELetBall
	if math.Abs(a18Bet.ELetBall) > this.MaxLetBall {
		//temp_data.Result =""
		return -2, temp_data
	}

	//限制初盘,即时盘让球在0.25以内
	sLetBall := math.Abs(a18Bet.SLetBall)
	eLetBall := math.Abs(a18Bet.ELetBall)
	if math.Abs(sLetBall-eLetBall) > 0.25 {
		//temp_data.Result =""
		//return -2, temp_data
	}

	//得出结果
	preResult := -1
	letBall := 0.00
	//------
	bfs_arr := this.BFScoreService.FindByMatchId(matchId)
	if len(bfs_arr) < 1 {
		return -1, temp_data
	}

	var temp_val float64
	var mainZongBfs *pojo.BFScore
	var mainZhuBfs *pojo.BFScore
	var guestZongBfs *pojo.BFScore
	var guestKeBfs *pojo.BFScore
	for _, e := range bfs_arr { //bfs_arr有多语言版本,条数很多
		if e.TeamId == v.MainTeamId {
			if e.Type == "总" {
				mainZongBfs = e
			}
			if e.Type == "主" {
				mainZhuBfs = e
			}
		} else if e.TeamId == v.GuestTeamId {
			if e.Type == "总" {
				guestZongBfs = e
			}
			if e.Type == "客" {
				guestKeBfs = e
			}
		}
	}
	if mainZongBfs == nil || guestZongBfs == nil || mainZhuBfs == nil || guestKeBfs == nil {
		return -1, temp_data
	}
	baseVal := 0.068
	rankDiff := 3.0
	if mainZongBfs.MatchCount >= 8 && guestZongBfs.MatchCount >= 8 {
		//排名越小越好
		temp_val = float64(mainZongBfs.Ranking - guestZongBfs.Ranking)
		if temp_val >= rankDiff {
			xishu := 1.0
			if temp_val > 11 {
				xishu = 3.0
			} else if temp_val > 5 {
				xishu = 2.0
			}
			letBall -= (temp_val / rankDiff) * baseVal * xishu
		}
		temp_val = float64(guestZongBfs.Ranking - mainZongBfs.Ranking)
		if temp_val >= rankDiff {
			xishu := 1.0
			if temp_val > 11 {
				xishu = 3.0
			} else if temp_val > 5 {
				xishu = 2.0
			}
			letBall += (temp_val / rankDiff) * baseVal * xishu
		}
		temp_val = float64(mainZhuBfs.Ranking - guestKeBfs.Ranking)
		if temp_val >= rankDiff {
			xishu := 1.0
			letBall -= (temp_val / rankDiff / 2) * baseVal * xishu
		}
		temp_val = float64(guestKeBfs.Ranking - mainZhuBfs.Ranking)
		if temp_val >= rankDiff {
			xishu := 1.0
			letBall += (temp_val / rankDiff / 2) * baseVal * xishu
		}
	} else {
		//return -1, temp_data
	}

	//------
	bffe_main := this.BFFutureEventService.FindNextBattle(matchId, v.MainTeamId)
	bffe_guest := this.BFFutureEventService.FindNextBattle(matchId, v.GuestTeamId)

	if this.IsCupMatch(bffe_main.EventLeagueId) {
		//下一场打杯赛
		//return -3, temp_data
	} else {
		//如果主队下一场打客场,战意充足
		if v.MainTeamId == bffe_main.EventGuestTeamId {
			letBall += (baseVal)
		}
	}

	if this.IsCupMatch(bffe_guest.EventLeagueId) {
		//下一场打杯赛
		//return -3, temp_data
	} else {
		//如果客队下一场打主场，战意懈怠
		if v.GuestTeamId == bffe_guest.EventMainTeamId {
			letBall += (baseVal)
		}
	}

	//1.0判断主队是否是让球方
	mainLetball := this.AnalyService.mainLetball(a18Bet)

	//2.0区间段
	var sectionBlock1, sectionBlock2 int
	tLetBall := math.Abs(letBall)
	//maxLetBall := math.Max(sLetBall, eLetBall)
	tempLetball1 := math.Abs(sLetBall - tLetBall)
	if tempLetball1 < 0.0 {
		sectionBlock1 = 1
	} else if tempLetball1 < 0.256 {
		sectionBlock1 = 2
	} else if tempLetball1 < 0.51 {
		sectionBlock1 = 3
	} else if tempLetball1 < 0.76 {
		sectionBlock1 = 4
	} else {
		sectionBlock1 = 10000
	}

	tempLetball2 := math.Abs(eLetBall - tLetBall)
	if tempLetball2 < 0.0 {
		sectionBlock2 = 1
	} else if tempLetball2 < 0.256 {
		sectionBlock2 = 2
	} else if tempLetball2 < 0.51 {
		sectionBlock2 = 3
	} else if tempLetball2 < 0.76 {
		sectionBlock2 = 4
	} else {
		sectionBlock2 = 10000
	}

	//3.0即时盘赔率大于等于初盘赔率
	endUp := eLetBall >= sLetBall
	//3.1即时盘初盘
	notZero := tLetBall >= eLetBall && tLetBall >= sLetBall

	//看两个区间是否属于同一个区间
	//if sectionBlock1 == 1 && sectionBlock2 == 1 {
	if sectionBlock1 <= 2 && sectionBlock2 <= 2 {
		if mainLetball && letBall > 0.1 && endUp && notZero {
			preResult = 3
		} else if !mainLetball && letBall < -0.1 && endUp && notZero {
			//preResult = 0
		}
	}

	if preResult < 0 {
		return -3, temp_data
	}

	if preResult == 3 && this.IsCupMatch(bffe_main.EventLeagueId) {
		//下一场打杯赛
		return -3, temp_data
	} else if preResult == 0 && this.IsCupMatch(bffe_guest.EventLeagueId) {
		//下一场打杯赛
		return -3, temp_data
	}

	base.Log.Info("所属于区间:", sectionBlock1, "-", sectionBlock2, ",对阵", v.MainTeamId+":"+v.GuestTeamId, ",计算得出让球为:", letBall, ",初盘让球:", a18Bet.SLetBall, ",即时盘让球:", a18Bet.ELetBall)
	var data *entity5.AnalyResult
	if len(temp_data.Id) > 0 {
		temp_data.PreResult = preResult
		temp_data.HitCount = temp_data.HitCount + 1
		temp_data.LetBall = a18Bet.ELetBall
		temp_data.MyLetBall = Decimal(letBall)
		data = temp_data
		//比赛结果
		data.Result = this.IsRight(v, data)
		return 1, data
	} else {
		data = new(entity5.AnalyResult)
		data.MatchId = v.Id
		data.MatchDate = v.MatchDate
		data.SLetBall = a18Bet.SLetBall
		data.LetBall = a18Bet.ELetBall
		data.AlFlag = this.ModelName()
		format := time.Now().Format("0102150405")
		data.AlSeq = format
		data.PreResult = preResult
		data.HitCount = 3
		data.MyLetBall = Decimal(letBall)
		//比赛结果
		data.Result = this.IsRight(v, data)
		return 0, data
	}
}


