package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/twinj/uuid"

	"github.com/e2u/goboot"

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
	UpdTime    time.Time       `gorm:"column:updtime"`
}

//ConsumeResult 消费接口返回结果
type ConsumeResult struct {
	//PointUsed 消耗抵用金额
	PointUsed string
	//PayAmount 剩余需支付
	PayAmount string
	//SelfGainPoints 为自己累积奖励金额
	SelfGainPoints string
	//GainPoints 总累积奖励金额
	GainPoints string
}

//NewAccount 空Account
func NewAccount() *AccountPoint {
	return &AccountPoint{}
}

//GetAmountByMember 获取有效账户余额 by member //todo
func GetAmountByMember(db *gorm.DB, mid string) (decimal.Decimal, error) {
	a := AccountPoint{}
	zero := decimal.New(0, 0)
	db1 := db.Table("accounts").Select("member_id,sum(amount) as sumamount").Where("current_date >= startdate and ((current_date<=expiredate) or (expiredate is null)) and amount>0 and member_id=?", mid).Group("member_id").First(&a)
	if db1.RecordNotFound() {
		return zero, nil
	}
	if db1.Error != nil {
		return zero, db1.Error
	}

	//fmt.Println("amount ", a.MemberID, a.Amount)
	return a.Amount, nil
}

//Consume 消费金额
//	m member
//	amountStr	金额字符串形式, 分为单位,例, 120 1块2毛
//	usePoint	是否使用账户金额
//	order			订单id
func Consume(db *gorm.DB, m *Member, amountStr string, usePoint string, orderID string) (*ConsumeResult, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, err
	}
	//amount = amount.Mul(decimal.New(1, -2)) //分为单位,转换成元
	var isuse bool
	if len(usePoint) > 0 {
		//if err, isuse=false
		isuse, _ = strconv.ParseBool(usePoint)
	} //else isuse=false
	var point decimal.Decimal
	var t *Transaction
	var consumedPoints []Account
	if isuse {
		amount, point, t, consumedPoints = getConsumeAccount(db, m.ID, amount, orderID)
		if t == nil {
			return nil, errors.New("积分消耗错误")
		}
	}

	fmt.Println("consume:", point, amount, isuse, len(consumedPoints))
	//reward := []decimal.Decimal
	ul, _ := getLevelsByMember(db, m.ID)
	length := len(ul)
	if length <= 0 {
		return nil, errors.New("用户关系错误")
	}
	ul = ul[:length]
	//fmt.Println(ul)
	transactions := createTransactionsByLevels(db, ul, amount, orderID)
	accounts := getAccountPoints(db, transactions)
	if true {
		if false == saveConsume(db, transactions, accounts, t, consumedPoints) {
			return nil, errors.New("保存错误")
		}
	}

	result := &ConsumeResult{point.Round(0).String(), amount.Round(0).String(), accounts[0].Amount.Round(0).String(), totalAmount(accounts).Round(0).String()}
	return result, nil
}

func totalAmount(as []Account) (total decimal.Decimal) {
	total = decimal.New(0, 0)
	for _, a := range as {
		total = total.Add(a.Amount)
	}
	return
}

//getConsumeAccount 获取消费账户对应记录列表,产生交易记录,计算消耗金额
//	返回值依次: 剩余需支付消费金额,抵用金额,交易记录对象,储值更新对象列表
//	t=nil until all exceptions unhappened
//	check t == nil
func getConsumeAccount(db *gorm.DB, mid string, amount decimal.Decimal, orderID string) (remindAmount decimal.Decimal, point decimal.Decimal, t *Transaction, result []Account) {
	LIMITATION := 4 //常量
	offset := 0
	zero := decimal.New(0, 0)
	now := time.Now()
	var as []Account
	//var result []Account
	remindAmount = amount
	for remindAmount.GreaterThan(zero) {
		db1 := db.Order("expiredate").Offset(offset).Limit(LIMITATION).Find(&as, "current_date >= startdate and ((current_date<=expiredate) or (expiredate is null)) and amount>0 and member_id=?", mid)
		if db1.RecordNotFound() {
			break
		}
		if db1.Error != nil {
			goboot.Log.Error(db1.Error)
			return
		}
		var i int
		var a Account
		for i, a = range as {
			//fmt.Println(remindAmount, a.Amount)
			as[i].UpdTime = now
			if remindAmount.LessThanOrEqual(a.Amount) {
				as = as[:i+1]
				as[i].Amount = a.Amount.Sub(remindAmount)
				remindAmount = zero
				break
			}
			remindAmount = remindAmount.Sub(a.Amount)
			as[i].Amount = zero
		}
		result = append(result, as...)
		if i < LIMITATION-1 {
			break
		}
		offset += LIMITATION
	}
	point = amount.Sub(remindAmount)
	t = new(Transaction)
	t.fillTransaction(orderID, mid, mid, point.Neg())
	//fmt.Println(len(result), remindAmount, point, amount, result[len(result)-1].Amount, t)
	return remindAmount, point, t, result
}

func saveConsume(db *gorm.DB, transactions []Transaction, accounts []Account, t *Transaction, consumedPoints []Account) bool {
	tx := db.Begin() //开启事务

	fmt.Println("account:", len(accounts))
	for _, a := range accounts {
		if err := a.saveNew(tx); err != nil {
			tx.Rollback()
			goboot.Log.Error("consume save error, rollback.")
			return false
		}
	}
	fmt.Println("points:", len(consumedPoints))
	for _, a := range consumedPoints {
		db1 := tx.Save(a)
		if db1.Error != nil {
			tx.Rollback()
			goboot.Log.Error(db1.Error)
			return false
		}
	}
	fmt.Println("trans:", len(transactions))
	for _, a := range transactions {
		if err := a.saveNew(tx); err != nil {
			tx.Rollback()
			goboot.Log.Error("consume save error, rollback.")
			return false
		}
	}
	if t != nil {
		//fmt.Println("tran")
		if err := t.saveNew(tx); err != nil {
			tx.Rollback()
			goboot.Log.Error("consume save error, rollback.")
			return false
		}
	}
	tx.Commit()
	//fmt.Println("consume done")
	return true

}
func (a *Account) saveNew(db *gorm.DB) error {
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
		//arr[i].ExpireDate = &tomorrow
		//arr[i].ExpireDate leave null
		arr[i].StartDate = tomorrow
		arr[i].GetDate = now
		arr[i].GetAmount = t.Amount
		arr[i].UpdTime = now
	}
	return arr
}
