// testclient project testclient.go

package main

import (
	"dq/app"
	//"dq/db"
	"dq/log"
	//"dq/vec2d"

	//	_ "net/http/pprof"
	//	"dq/wordsfilter"
	//"fmt"
	//"time"
	//	"os"
	//"dq/utils"
)

//func test(t interface{}) int32 {
//	a := t.(int32)
//	return a

//}

func main() {
	//新功能
	//utils.PayQuest()
	//	format := "15:04:05"
	//	a, _ := time.Parse(format, "11:00:00")
	//	b, _ := time.Parse(format, "16:00:00")
	//	c, _ := time.Parse(format, "16:01:10")
	//	fmt.Println("11time: ", a.After(b))
	//	fmt.Println("22time: ", b.After(c))
	//	fmt.Println("33time: ", c.After(a))
	//	fmt.Println("44time: ", a.After(c))
	//	fmt.Println("55time: ", b.After(a))
	//	fmt.Println("66time: ", c.After(b))
	//主线
	//log.Info("test:%d", test(30))
	//
	//	ApplicationDir, _ := os.Getwd()

	//	confPath := fmt.Sprintf("%s/bin/conf/words_filter.txt", ApplicationDir)

	//	_, err := wordsfilter.WF.GenerateWithFile(confPath)
	//	if err != nil {
	//		log.Info("path:%s", err.Error())
	//	} else {

	//		log.Info("test:%s", wordsfilter.WF.DoReplace("毛主席 12332"))
	//	}

	// 性能分析
	//	go func() {
	//		http.ListenAndServe(":8282", nil)
	//	}()

	//	v1 := vec2d.Vec2{X: 0, Y: 1}
	//	v2 := vec2d.Vec2{X: -1, Y: 0}
	//	v3 := vec2d.Vec2{X: 1, Y: 1}
	//	a1 := v1.Angle()
	//	a2 := v2.Angle()
	//	//a3 := v3.Angle()
	//	v3.Rotate(a2 - a1)
	//	log.Info("v3:%v", v3)

	//	v4 := vec2d.Vec2{X: 0, Y: 1}
	//	v5 := vec2d.Vec2{X: 0, Y: -1}

	//	log.Info("angle1:%f", v1.Angle())
	//	log.Info("angle2:%f", v2.Angle())
	//	log.Info("angle3:%f", v3.Angle())
	//	log.Info("angle4:%f", v4.Angle())
	//	log.Info("angle5:%f", v5.Angle())

	//	t1 := vec2d.Angle(vec2d.Vec2{0, 1}, vec2d.Vec2{1, 0})
	//	log.Info("angle:%f", t1)
	app := new(app.DefaultApp)

	app.Run()
	log.Info("dq over!")
	//	log.Info("111!")
	//	core := cyward.CreateWardCore()
	//	var test []*cyward.Body

	//	for i := 0; i < 10; i++ {
	//		for j := 0; j < 10; j++ {
	//			pos := vec2d.Vec2{float64(100 + i*20), float64(100 + j*15)}
	//			r := vec2d.Vec2{float64(3 + i/3), float64(3 + j/3)}
	//			t := core.CreateBody(pos, r, 30.0)
	//			t.SetTag(i*10 + j)
	//			test = append(test, t)
	//		}

	//	}
	//	var points []vec2d.Vec2
	//	points = append(points, vec2d.Vec2{-10, 0})
	//	points = append(points, vec2d.Vec2{0, 10})
	//	points = append(points, vec2d.Vec2{10, 0})
	//	points = append(points, vec2d.Vec2{0, -10})
	//	//	points[0] = vec2d.Vec2{-10, 0}
	//	//	points[1] = vec2d.Vec2{0, 10}
	//	//	points[2] = vec2d.Vec2{10, 0}
	//	//	points[3] = vec2d.Vec2{0, -10}
	//	core.CreateBodyPolygon(vec2d.Vec2{400, 200}, points, 30.0)

	//	t1 := time.Now().UnixNano()

	//	test[0].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[1].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[2].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[3].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[4].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[5].SetTarget(vec2d.Vec2{float64(500), float64(300)})

	//	t2 := time.Now().UnixNano()
	//	log.Info("main time:%d", (t2-t1)/1e6)

}
