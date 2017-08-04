package app

import (
	"fmt"

	"github.com/e2u/goboot"
	gorm "gopkg.in/jinzhu/gorm.v1"
)

var (
	//App 开发环境
	App *AppContext
)

//Postgres 数据库配置
type Postgres struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string
}

//AppContext 应用实例, 名称,数据库,缓存配置,当前环境(测试,开发,生产)
type AppContext struct {
	AppName  string
	DB       *gorm.DB
	DevMode  bool
	postgres *Postgres
}

//Init 初始化
func Init() *AppContext {
	App = newAppContext()
	return App
}

//Close 资源释放
func Close() {
	App.DB.Close()
}

func newAppContext() *AppContext {
	app := &AppContext{
		AppName: goboot.Config.MustString("app.name", "pyramid"),
		DevMode: goboot.Config.MustBool("mode.dev", true),
		postgres: &Postgres{
			Host:     goboot.Config.MustString("pg.host", "127.0.0.1"),
			Port:     goboot.Config.MustInt("pg.port", 5432),
			DBName:   goboot.Config.MustString("pg.dbname", "pyramid"),
			User:     goboot.Config.MustString("pg.user", "postgres"),
			Password: goboot.Config.MustString("pg.password", "none"),
			SSLMode:  goboot.Config.MustString("pg.sslmode", "disable"),
		},
	}

	if err := app.initPostgres(); err != nil {
		goboot.Log.Criticalf("初始化数据库连接错误: %v", err.Error())
		panic(fmt.Sprintf("初始化数据库连接错误: %v", err.Error()))
	}
	return app
}

func (a *AppContext) initPostgres() error {
	var err error
	a.DB, err = gorm.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", a.postgres.User, a.postgres.Password, a.postgres.Host, a.postgres.Port, a.postgres.DBName, a.postgres.SSLMode))
	if err != nil {
		return err
	}
	if a.DevMode {
		a.DB.LogMode(true)
	}
	if err = a.DB.Exec("set timezone = 'PRC'").Error; err != nil {
		return err
	}
	return a.DB.DB().Ping()
}
