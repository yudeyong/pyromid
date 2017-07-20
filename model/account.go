package model

import (
	"errors"
	"fmt"
	"strconv"
	_ "time"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

//Account 用户账户可用余额
type Account struct {
	MemberID string `gorm:"column:member_id"`
	Amount   int    `gorm:"column:sumamount"`
}

//NewAccount 空Account
func NewAccount() *Account {
	return &Account{}
}

//GetAmountByMember 获取有效账户余额 by member
func GetAmountByMember(db *gorm.DB, m *Member) (int, error) {
	a := Account{}
	db1 := db.Table("accounts").Select("member_id,sum(amount) as sumamount").Where("current_date between startdate and expiredate and member_id=?", m.ID).Group("member_id").First(&a)
	if db1.RecordNotFound() {
		return 0, nil
	}
	if db1.Error != nil {
		return 0, db1.Error
	}

	fmt.Println("amount ", a.MemberID, a.Amount)
	return a.Amount, nil
}

//Consume 消费金额
//	m member
//	amountStr	金额字符串形式, 分为单位,例, 120 1块2毛
//	usePoint	是否使用账户金额
//	order			订单id
func Consume(db *gorm.DB, m *Member, amountStr string, usePoint string, orderID string) error {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return err
	}
	var isuse bool
	if len(usePoint) > 0 {
		//if err, isuse=false
		isuse, _ = strconv.ParseBool(usePoint)
	} //else isuse=false
	var point int

	if isuse {
		point, err = GetAmountByMember(db, m)
		fmt.Println(amount, point)
		if err != nil {
			return err
		}
		if amount > point {
			amount -= point
		} else {
			point = amount
			amount = 0
		}
		fmt.Println(amount, point)
	}

	fmt.Println("consume:", point, amount, isuse)
	//reward := []decimal.Decimal
	ul, _ := GetLevelsByMember(db, m)
	length := len(ul)
	if length <= 0 {
		return errors.New("用户关系错误")
	}
	ul = ul[:length]
	fmt.Println(ul)
	transactions := CreateTransactionsByLevels(db, ul, amount, orderID)
	points := getAccountPoints(db, ul)
	fmt.Println(transactions, points)
	return nil
}

func getAccountPoints(db *gorm.DB, ul []UserLevel) []Account {
	return nil
}
