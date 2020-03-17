package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"tesou.io/platform/foot-parent/foot-api/common/base"
	"tesou.io/platform/foot-parent/foot-core/common/base/service/mysql"
	"tesou.io/platform/foot-parent/foot-core/module/analy/service"
)

func main() {

	//关闭SQL输出
	fmt.Println(strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	mysql.ShowSQL(false)
	base.Log.Info("---------------------------------------------------------------")
	base.Log.Info("---------------C1模型--------------")
	base.Log.Info("---------------------------------------------------------------")
	c1 := new(service.C2Service)
	c1.MaxLetBall = 1
	//c1.Analy(true)
	base.Log.Info("---------------------------------------------------------------")
	base.Log.Info("---------------E1模型--------------")
	base.Log.Info("---------------------------------------------------------------")
	e1 := new(service.C1Service)
	e1.MaxLetBall = 1
	e1.PrintOddData = true
	//e1.Analy(true)
	//关闭SQL输出
	mysql.ShowSQL(true)
}

