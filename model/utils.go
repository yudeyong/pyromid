package model

import (
	"database/sql"

	"github.com/shopspring/decimal"
	gorm "gopkg.in/jinzhu/gorm.v1"
)

const (
	//ResOK 返回码
	ResOK = "200"
	//ResDup 数据重复
	ResDup = "201"
	//ResMore 进一步操作
	ResMore = "300"
	//ResMore1 另一种进一步操作
	ResMore1 = "301"
	//ResInvalid 无效参数
	ResInvalid = "412"
	//ResPhoneInvalid 无效手机号
	ResPhoneInvalid = "4121"
	//ResNotFound 没有对应记录
	ResNotFound = "404"
	//ResFound 成功找到
	ResFound = "200"
	//ResWrongSQL 查询语句错误
	ResWrongSQL = "500"
	//ResFail 异常
	ResFail = "500"
	//ResFailCreateMember 创建用户异常
	ResFailCreateMember = "501"

	//到账期限, T+n n=AvailableDays
	AvailableDays = 0
)

var (
	zero = decimal.New(0, 0) //常量
)

//Init 初始化 分级分成比例
func Init(db *gorm.DB, ratios *([]decimal.Decimal)) {
	InitLevelRatios(ratios)
	InitCardNo(db)

}

//NullStringEquals NullString与string比较
func NullStringEquals(s sql.NullString, str string) bool {
	if !s.Valid {
		return false
	}
	return str == s.String
}
