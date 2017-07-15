package model

import (
	_ "errors"
	"fmt"
	_ "strconv"
	_ "time"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

type Account struct {
	MemberID string `gorm:"column:member_id"`
	Amount   int    `gorm:"column:sumamount"`
}

func GetAmountByMember(db *gorm.DB, m *Member) (error, int) {
	a := Account{}
	db1 := db.Table("accounts").Select("member_id,sum(amount) as sumamount").Where("current_date between startdate and expiredate and member_id=?", m.ID).Group("member_id").First(&a)
	if db1.RecordNotFound() {
		return nil, 0
	}
	if db1.Error != nil {
		return db1.Error, 0
	}

	fmt.Println("amount ", a.MemberID, a.Amount)
	return nil, a.Amount
}
