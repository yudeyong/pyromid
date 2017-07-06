package model

import (
	_"fmt"
	"log"
	"strconv"

  gorm "gopkg.in/jinzhu/gorm.v1"
)

var (
	CardNo int
)

type maxCardNo struct{
	MaxNo	string `gorm:"column:mx"`
}

func InitCardNo(db *gorm.DB){
	m := maxCardNo{}
	db1 := db.Table("members").Select("cardno as mx").Where("char_length(cardno)=(select max(char_length(cardno)) from members)").Order("mx desc").First(&m)
	if db1.RecordNotFound() {
		log.Printf("table members is empty!")
		return
	}
  if db1.Error!=nil{
		log.Printf("Get max cardNo error %s!", db1.Error)
		return
  }
	//fmt.Println(m)

	CardNo,_ = strconv.Atoi(m.MaxNo)
	log.Printf("Init card No. %s", m.MaxNo)
}
func CheckCardNo() int{
	return CardNo
}
func GetNewCard() string{
	CardNo += 1
	return strconv.Itoa(CardNo);
}

func UpdateNewCard(cardno int) bool{
	if (cardno>CardNo){
		CardNo=cardno
		return true
	}
	return false
}
