package model

import (
	"log"
	"strconv"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

var (
	//CardNo 最大卡号
	CardNo int
)

type maxCardNo struct {
	MaxNo string `gorm:"column:mx"`
}

//InitCardNo 初始化卡号, 获取当前库中最大卡号
func InitCardNo(db *gorm.DB) {
	m := maxCardNo{}
	db1 := db.Table("members").Select("cardno as mx").Where("char_length(cardno)=(select max(char_length(cardno)) from members)").Order("mx desc").First(&m)
	if db1.RecordNotFound() {
		log.Printf("table members is empty!")
		return
	}
	if db1.Error != nil {
		log.Printf("Get max cardNo error %s!", db1.Error)
		return
	}
	//fmt.Println(m)

	CardNo, _ = strconv.Atoi(m.MaxNo)
	log.Printf("Init card No. %s", m.MaxNo)
}

//CheckCardNo 查询当前卡号
func CheckCardNo() int {
	return CardNo
}

//GetNewCard 申请一个卡号
func GetNewCard() string {
	CardNo++
	return strconv.Itoa(CardNo)
}

//UpdateNewCard 更新内部卡号计数
func UpdateNewCard(cardno int) bool {
	if cardno > CardNo {
		CardNo = cardno
		return true
	}
	return false
}
