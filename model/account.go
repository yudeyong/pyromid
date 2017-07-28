package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/twinj/uuid"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

//AccountPoint 用户账户可用余额
type AccountPoint struct {
	MemberID string          `gorm:"column:member_id"`
	Amount   decimal.Decimal `gorm:"column:sumamount"`
}

//Account 用户账户记录
type Account struct {
	ID         string          `gorm:"column:id"`
	MemberID   string          `gorm:"column:member_id"`
	Amount     decimal.Decimal `gorm:"column:amount"`
	ExpireDate *time.Time      `gorm:"column:expiredate"`
	StartDate  time.Time       `gorm:"column:startdate"`
	GetDate    time.Time       `gorm:"column:getdate"`
	GetAmount  decimal.Decimal `gorm:"column:getamount"`
}

//NewAccount 空Account
func NewAccount() *AccountPoint {
	return &AccountPoint{}
}

//GetAmountByMember 获取有效账户余额 by member //todo
func GetAmountByMember(db *gorm.DB, m *Member) (decimal.Decimal, error) {
	a := AccountPoint{}
	zero := decimal.New(0, 0)
	db1 := db.Table("accounts").Select("member_id,sum(amount) as sumamount").Where("current_date > startdate and ((current_date<expiredate) or (expiredate is null)) and member_id=?", m.ID).Group("member_id").First(&a)
	if db1.RecordNotFound() {
		return zero, nil
	}
	if db1.Error != nil {
		return zero, db1.Error
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
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return err
	}
	amount = amount.Mul(decimal.New(1, -2)) //分为单位,转换成元
	var isuse bool
	if len(usePoint) > 0 {
		//if err, isuse=false
		isuse, _ = strconv.ParseBool(usePoint)
	} //else isuse=false
	var point decimal.Decimal

	if isuse {
		point, err = GetAmountByMember(db, m)
		fmt.Println(amount, point)
		if err != nil {
			return err
		}
		if amount.GreaterThan(point) {
			amount = amount.Sub(point)
		} else {
			point = amount
			amount = decimal.New(0, 0)
		}
		fmt.Println(amount, point)
	}

	fmt.Println("consume:", point, amount, isuse)
	//reward := []decimal.Decimal
	ul, _ := getLevelsByMember(db, m.ID)
	length := len(ul)
	if length <= 0 {
		return errors.New("用户关系错误")
	}
	ul = ul[:length]
	//fmt.Println(ul)
	transactions := createTransactionsByLevels(db, ul, amount, orderID)
	accounts := getAccountPoints(db, transactions)
	saveConsume(db, transactions, accounts, point)
	return nil
}

func saveConsume(db *gorm.DB, transactions []Transaction, accounts []Account, point decimal.Decimal) {
	tx := db.Begin() //开启事务
	for _, t := range accounts {
		if err := t.save(tx); err != nil {
			fmt.Println(err)
		}
	}
	//fmt.Println("trans:", transactions)
	for _, t := range transactions {
		if err := t.save(tx); err != nil {
			fmt.Println(err)
		}
	}
	tx.Commit()

}
func (a *Account) save(db *gorm.DB) error {
	db.Create(a)
	if db.NewRecord(a) {
		return errors.New("交易创建失败")
	}
	fmt.Println("done:", a)
	return nil
}

func getAccountPoints(db *gorm.DB, ts []Transaction) []Account {
	arr := make([]Account, len(ts))
	mid := ts[0].SourceID
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	for i, t := range ts {
		arr[i].ID = uuid.NewV4().String()
		arr[i].MemberID = mid
		arr[i].Amount = t.Amount
		arr[i].ExpireDate = &tomorrow
		//arr[i].ExpireDate leave null
		arr[i].StartDate = tomorrow
		arr[i].GetDate = now
		arr[i].GetAmount = t.Amount
	}
	return arr
}
