package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"tesou.io/platform/foot-parent/foot-api/common/base"
	entity5 "tesou.io/platform/foot-parent/foot-api/module/analy/pojo"
	"tesou.io/platform/foot-parent/foot-api/module/match/pojo"
	entity3 "tesou.io/platform/foot-parent/foot-api/module/odds/pojo"
	"tesou.io/platform/foot-parent/foot-core/common/utils"
	"tesou.io/platform/foot-parent/foot-core/module/analy/constants"
	"tesou.io/platform/foot-parent/foot-core/module/match/service"
	"time"
)

type C1Service struct {
	AnalyService
	service.BFScoreService
	service.BFBattleService
	service.BFFutureEventService
	service.BFJinService

	//最大让球数据
	MaxLetBall float64
}

func (this *C1Service) ModelName() string {
	//alFlag := reflect.TypeOf(*this).Name()
	return "C1"
}

func (this *C1Service) Analy(analyAll bool) {
	var matchLasts []*pojo.MatchLast
	if analyAll {
		matchHis := this.MatchHisService.FindAll()
		for _, e := range matchHis {
			matchLasts = append(matchLasts, &e.MatchLast)
		}
		//matchLasts = this.MatchLastService.FindAll()
	} else {
		matchLasts = this.MatchLastService.FindNotFinished()
	}
	this.Analy_Process(matchLasts)

}

func (this *C1Service) Analy_Near() {
	matchList := this.MatchLastService.FindNear()
	this.Analy_Process(matchList)
}

func (this *C1Service) Analy_Process(matchList []*pojo.MatchLast) {
	hit_count_str := utils.GetVal(constants.SECTION_NAME, "hit_count")
	hit_count, _ := strconv.Atoi(hit_count_str)
	data_list_slice := make([]interface{}, 0)
	data_modify_list_slice := make([]interface{}, 0)
	var rightCount = 0
	var errorCount = 0

	for _, v := range matchList {
		stub, data := this.analyStub(v)
		if nil != data {
			if strings.EqualFold(data.Result, "命中") {
				rightCount++
			}
			if strings.EqualFold(data.Result, "错误") {
				errorCount++
			}
		}

		if stub == 0 || stub == 1 {
			data.TOVoid = false
			hours := v.MatchDate.Sub(time.Now()).Hours()
			if hours > 0 {
				data.THitCount = hit_count
			} else {
				data.THitCount = 1
			}
			//如其他模型存在互斥选项，设置为作废
			diff_preResult := this.FindOtherAlFlag(data.MatchId, data.AlFlag, data.PreResult)
			if diff_preResult {
				data.TOVoid = true
				//与其他模型互斥
				data.TOVoidDesc = "斥"
			}

			if stub == 0 {
				data_list_slice = append(data_list_slice, data)
			} else if stub == 1 {
				data_modify_list_slice = append(data_modify_list_slice, data)
			}
		} else {
			if stub != -2 {
				data = this.Find(v.Id, this.ModelName())
			}
			data.TOVoid = true
			if len(data.Id) > 0 {
				if data.HitCount >= hit_count {
					data.HitCount = (hit_count / 2) - 1
				} else {
					data.HitCount = 0
				}
				this.AnalyService.Modify(data)
			}
		}
	}
	base.Log.Info("------------------")
	base.Log.Info("------------------")
	base.Log.Info("------------------")
	base.Log.Info("GOOOO场次:", rightCount)
	base.Log.Info("X0000场次:", errorCount)
	base.Log.Info("------------------")

	this.AnalyService.SaveList(data_list_slice)
	this.AnalyService.ModifyList(data_modify_list_slice)
}

/**
  -1 参数错误
  -2 不符合让球数
  -3 计算分析错误
  0  新增的分析结果
  1  需要更新结果
 */
func (this *C1Service) analyStub(v *pojo.MatchLast) (int, *entity5.AnalyResult) {
	matchId := v.Id
	if matchId == "1779520"{
		fmt.Println("-------------------")
	}
	//声明使用变量
	var a18betData *entity3.AsiaHis
	//亚赔
	aList := this.AsiaHisService.FindByMatchIdCompId(matchId, constants.C1_REFER_ASIA)
	if len(aList) < 1 {
		return -1, nil
	}
	a18betData = aList[0]
	if math.Abs(a18betData.ELetBall) > this.MaxLetBall {
		temp_data := this.Find(v.Id, this.ModelName())
		temp_data.LetBall = a18betData.ELetBall
		//temp_data.Result =""
		return -2, temp_data
		//return -2, nil
	}

	//限制初盘,即时盘让球在0.25以内
	sLetBall := math.Abs(a18betData.SLetBall)
	eLetBall := math.Abs(a18betData.ELetBall)
	if math.Abs(sLetBall-eLetBall) > 0.25 {
		temp_data := this.Find(v.Id, this.ModelName())
		temp_data.LetBall = a18betData.ELetBall
		temp_data.Result =""
		return -2, temp_data
		//return -2, nil
	}

	//得出结果
	preResult := -1
	letBall := 0.00
	//------
	bfs_arr := this.BFScoreService.FindByMatchId(matchId)
	if len(bfs_arr) < 1 {
		return -1, nil
	}

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
		return -1, nil
	}
	baseVal := 0.25
	//积分排名----排名越小越好
	if mainZongBfs.MatchCount >= 8 && guestZongBfs.MatchCount >= 8 {
		temp_val := math.Abs(float64(mainZongBfs.Ranking - guestZongBfs.Ranking))
		if temp_val > 4 {
			if mainZongBfs.Ranking < guestZongBfs.Ranking {
				letBall += baseVal
			} else {
				letBall -= baseVal
			}
		}
		temp_val = math.Abs(float64(guestZongBfs.Ranking - mainZongBfs.Ranking))
		if temp_val > 4 {
			if mainZongBfs.Ranking < guestZongBfs.Ranking {
				letBall += baseVal
			} else {
				letBall -= baseVal
			}
		}
	} else {
		//return -1, nil
	}

	//两队对战
	bfb_arr := this.BFBattleService.FindNearByMatchId(matchId, 4)
	mainWin := 0
	guestWin := 0
	for _, e := range bfb_arr {
		if e.BattleMainTeamId == v.MainTeamId && e.BattleMainTeamGoals > e.BattleGuestTeamGoals {
			mainWin++
		}
		if e.BattleGuestTeamId == v.MainTeamId && e.BattleGuestTeamGoals > e.BattleMainTeamGoals {
			mainWin++
		}
		if e.BattleMainTeamId == v.GuestTeamId && e.BattleMainTeamGoals > e.BattleGuestTeamGoals {
			guestWin++
		}
		if e.BattleGuestTeamId == v.GuestTeamId && e.BattleGuestTeamGoals > e.BattleMainTeamGoals {
			guestWin++
		}
	}
	if mainWin > guestWin {
		letBall += baseVal
	}
	if guestWin > mainWin {
		letBall -= baseVal
	}
	//近期战绩
	bfj_main := this.BFJinService.FindNearByTeamName(v.MainTeamId, 4)
	bfj_guest := this.BFJinService.FindNearByTeamName(v.GuestTeamId, 4)
	bfj_mainWin := 0
	bfj_guestWin := 0
	for _, e := range bfj_main {
		if e.HomeTeam == v.MainTeamId && e.HomeScore > e.GuestScore {
			bfj_mainWin++
		}
		if e.GuestTeam == v.MainTeamId && e.GuestScore > e.HomeScore {
			bfj_mainWin++
		}
	}
	for _, e := range bfj_guest {
		if e.HomeTeam == v.GuestTeamId && e.HomeScore > e.GuestScore {
			bfj_guestWin++
		}
		if e.GuestTeam == v.GuestTeamId && e.GuestScore > e.HomeScore {
			bfj_guestWin++
		}
	}
	if bfj_mainWin > (bfj_guestWin + 1) {
		letBall += baseVal
	}
	if bfj_guestWin > (bfj_mainWin + 1) {
		letBall -= baseVal
	}

	//未来赛事
	bffe_main := this.BFFutureEventService.FindNextBattle(matchId, v.MainTeamId)
	bffe_guest := this.BFFutureEventService.FindNextBattle(matchId, v.GuestTeamId)
	if strings.Contains(bffe_main.EventLeagueId, "杯") {
		//下一场打杯赛
		return -3, nil
	} else {
		//如果主队下一场打客场,战意充足
		if v.MainTeamId == bffe_main.EventGuestTeamId {
			letBall += 0.075
		}
	}
	if strings.Contains(bffe_guest.EventLeagueId, "杯") {
		//下一场打杯赛
			return -3, nil
	} else {
		//如果客队下一场打主场，战意懈怠
		if v.GuestTeamId == bffe_guest.EventMainTeamId {
			letBall += 0.075
		}
	}

	//1.0判断主队是否是让球方
	//mainLetball := true
	//if a18betData.ELetBall > 0 {
	//	mainLetball = true
	//} else if a18betData.ELetBall < 0 {
	//	mainLetball = false
	//} else {
	//	if letBall > 0{
	//		mainLetball = false
	//	} else {
	//		mainLetball = true
	//	}
	//}
	//2.0区间段
	var sectionBlock1, sectionBlock2 int
	//maxLetBall := math.Max(sLetBall, eLetBall)
	tempLetball1 := math.Abs(a18betData.SLetBall - letBall)
	if tempLetball1 < 0.0 {
		sectionBlock1 = 1
	} else if tempLetball1 < 0.26 {
		sectionBlock1 = 2
	} else if tempLetball1 < 0.51 {
		sectionBlock1 = 3
	} else if tempLetball1 < 0.76 {
		sectionBlock1 = 4
	} else {
		sectionBlock1 = 10000
	}

	tempLetball2 := math.Abs(a18betData.ELetBall - letBall)
	if tempLetball2 < 0.0 {
		sectionBlock2 = 0
	}else if tempLetball2 < 0.1 {
		sectionBlock2 = 1
	} else if tempLetball2 < 0.26 {
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

	//看两个区间是否属于同一个区间
	if  sectionBlock2 == 0 ||  sectionBlock2 == 2 && endUp{
		if letBall > 0 && a18betData.ELetBall > 0{
			preResult = 3
		}else if letBall > 0 && a18betData.ELetBall < 0{
			preResult = 3
		}else if letBall < 0 && a18betData.ELetBall < 0{
			preResult = 0
		}else if letBall < 0 && a18betData.ELetBall > 0{
			preResult = 0
		}
	}

	if preResult < 0 {
		return -3, nil
	}

	base.Log.Info("所属于区间:", sectionBlock1, "-", sectionBlock2, ",对阵", v.MainTeamId+":"+v.GuestTeamId, ",计算得出让球为:", letBall, ",初盘让球:", a18betData.SLetBall, ",即时盘让球:", a18betData.ELetBall)
	var data *entity5.AnalyResult
	temp_data := this.Find(v.Id, this.ModelName())
	if len(temp_data.Id) > 0 {
		temp_data.PreResult = preResult
		temp_data.HitCount = temp_data.HitCount + 1
		temp_data.LetBall = a18betData.ELetBall
		temp_data.MyLetBall = Decimal(letBall)
		data = temp_data
		//比赛结果
		data.Result = this.IsRight(v, data)
		return 1, data
	} else {
		data = new(entity5.AnalyResult)
		data.MatchId = v.Id
		data.MatchDate = v.MatchDate
		data.LetBall = a18betData.ELetBall
		data.AlFlag = this.ModelName()
		format := time.Now().Format("0102150405")
		data.AlSeq = format
		data.PreResult = preResult
		data.HitCount = 3
		data.LetBall = a18betData.ELetBall
		data.MyLetBall = Decimal(letBall)
		//比赛结果
		data.Result = this.IsRight(v, data)
		return 0, data
	}
}
