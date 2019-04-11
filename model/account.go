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
// valid: true 有效期内; false 未到有效期的,不含过期的
func GetAmountByMember(db *gorm.DB, mid string, valid bool) (decimal.Decimal, error) {
	a := AccountPoint{}
	var sql string
	if valid {
		sql = ">="
	} else {
		sql = "<"
	}
	db1 := db.Table("accounts").Select("member_id,sum(amount) as sumamount").Where("current_date "+sql+" startdate and ((current_date<=expiredate) or (expiredate is null)) and amount>0 and member_id=?", mid).Group("member_id").First(&a)
	if db1.RecordNotFound() {
		return zero, nil
	}
	if db1.Error != nil {
		return zero, db1.Error
	}

	//fmt.Println("amount ", a.MemberID, a.Amount)
	return a.Amount, nil
}

//Cashout 消费金额
// 	m member
//  amountStr	金额字符串形式, 分为单位,例, 120 1块2毛
//  order:订单id
//	return point, code, message
func Cashout(db *gorm.DB, m *Member, amountStr string, orderID string) (string, string, string) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return "", ResInvalid, err.Error()
	}
	var validcode int
	validcode, err = vaildOrderID(db, m.ID, orderID)
	if validcode > 0 {
		var code, msg string
		if validcode == 1 {
			code = ResDup
			msg = "order No. exist"
		} else {
			code = ResFail
			msg = err.Error()
		}
		return "", code, msg
	}
	var point decimal.Decimal
	var t *Transaction
	var consumedPoints []Account
	amount, point, t, consumedPoints = getConsumeAccount(db, m.ID, amount, orderID)
	if t == nil {
		return "", ResWrongSQL, "积分消耗错误"
	}
	if amount.IsPositive() {
		return "", ResInvalid, "余额不足"
	}
	ts := make([]Transaction, 0) //兼容saveConsume, 空数据集
	as := make([]Account, 0)     //兼容saveConsume, 空数据集
	if false == saveConsume(db, ts, as, t, consumedPoints) {
		return "", ResWrongSQL, "保存错误"
	}
	return point.String(), ResOK, "OK"
}

//Consume 消费金额
// 	m member
//  amountStr	金额字符串形式, 分为单位,例, 120 1块2毛
//  usePoint	是否使用账户金额
//  order:订单id
func Consume(db *gorm.DB, m *Member, amountStr string, usePoint string, orderID string) (*ConsumeResult, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, err
	}
	var validcode int
	validcode, err = vaildOrderID(db, m.ID, orderID)
	if validcode > 0 {
		if err == nil { //assert(result=1)
			err = errors.New("order No. exist")
		}
		return nil, err
	}
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

	//fmt.Println("consume:", point, amount, isuse, len(consumedPoints))
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
	a := "0"
	if len(accounts) > 0 {
		a = accounts[0].Amount.Round(0).String()
	}

	result := &ConsumeResult{point.Round(0).String(), amount.Round(0).String(), a, totalAmount(accounts).Round(0).String()}
	return result, nil
}

func totalAmount(as []Account) (total decimal.Decimal) {
	total = zero
	for _, a := range as {
		total = total.Add(a.Amount)
	}
	return
}

//getConsumeAccount 获取消费账户对应记录列表,产生交易记录,计算消耗金额
//	返回值依次: 剩余需支付消费金额,抵用金额,交易记录对象,储值更新对象列表
//	t=nil until all exceptions unhappened
//	check t == nil
func getConsumeAccount(db *gorm.DB, mID string, amount decimal.Decimal, orderID string) (remindAmount decimal.Decimal, point decimal.Decimal, t *Transaction, result []Account) {
	LIMITATION := 4 //常量
	offset := 0
	now := time.Now()
	var as []Account
	//var result []Account
	remindAmount = amount
	for remindAmount.GreaterThan(zero) {
		//order by 优先有效期,次优先小amount
		db1 := db.Order("expiredate, amount").Offset(offset).Limit(LIMITATION).Find(&as, "current_date >= startdate and ((current_date<=expiredate) or (expiredate is null)) and amount>0 and member_id=?", mID)
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
			// fmt.Println("consume:", remindAmount, a.Amount)
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
	t.fillTransaction(orderID, mID, mID, point.Neg())
	//fmt.Println(len(result), remindAmount, point, amount, result[len(result)-1].Amount, t)
	return remindAmount, point, t, result
}

func saveConsume(db *gorm.DB, transactions []Transaction, accounts []Account, t *Transaction, consumedPoints []Account) bool {
	tx := db.Begin() //开启事务

	// fmt.Println("account:", len(accounts))
	for _, a := range accounts {
		if err := a.saveNew(tx); err != nil {
			tx.Rollback()
			goboot.Log.Error("consume save error, rollback.")
			return false
		}
	}
	// fmt.Println("points:", len(consumedPoints))
	for _, a := range consumedPoints {
		db1 := tx.Save(a)
		if db1.Error != nil {
			tx.Rollback()
			goboot.Log.Error(db1.Error)
			return false
		}
	}
	// fmt.Println("trans:", len(transactions))
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
	if len(ts) <= 0 {
		return arr
	}
	//mid := ts[0].SourceID
	now := time.Now()
	tday := now.AddDate(0, 0, AvailableDays)
	for i, t := range ts {
		arr[i].ID = uuid.NewV4().String()
		arr[i].MemberID = t.TargetID
		arr[i].Amount = t.Amount
		//arr[i].ExpireDate = &tomorrow
		//arr[i].ExpireDate leave null
		arr[i].StartDate = tday
		arr[i].GetDate = now
		arr[i].GetAmount = t.Amount
		arr[i].UpdTime = now
	}
	return arr
}
