package main

import (
	"flag"
	"net/http"
	"fmt"
	"strings"
	"time"
	_ "strconv"

	_ "net/http/pprof"

	"github.com/e2u/goboot"
	"github.com/e2u/goboot/jobs"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"./app"
	"./conf"
)


var (
	RunEnv  string
	App     *app.AppContext
	stoping bool
	ListenPort int

	retrySQSQueue string
)

type Controller struct {
}

const (

)
func initApp() error {
	App = app.NewAppContext()
	goboot.Log.Info("starting...")
	//fmt.Printf("found: %s\n", App)

	conf.InitLevels(App.DB)
	return nil
}


// 初始化函数
func init() {
	flag.StringVar(&RunEnv, "env", "dev", "app run env: [dev|prod]")
	flag.IntVar(&ListenPort, "port", 9000, "http listen port: [9000|9001]")
	flag.Parse()

	goboot.Init(RunEnv)
	goboot.OnAppStart(initApp, 10)
	goboot.Startup()
}


func srvMain(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()  //解析参数，默认是不会解析的
	fmt.Println(r.Form)  //这些信息是输出到服务器端的打印信息
	action :=string([]rune( r.URL.Path)[1:])
	fmt.Println("path", r.URL.Path, action)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") //这个写入到w的是输出到客户端的
}

func main() {
	jobs.SelfConcurrent = false // 不允许并发,只能运行完一个任务再运行下一个任务
//	go jobs.Every(time.Minute, HealthJob{})

	c := &Controller{}
	r := mux.NewRouter()
	r.HandleFunc("/consume", c.Consume)
	r.HandleFunc("/", srvMain) //设置访问的路由
	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", ListenPort),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe();
	
	defer func() {
		App.DB.Close()
	}()
}
