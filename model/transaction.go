package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/twinj/uuid"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

const (
	defaultPageSize = 500
)

//Transaction 用户关系表
type Transaction struct {
	ID              string          `gorm:"column:id"`
	OrderID         sql.NullString  `gorm:"column:order_id"`
	SourceID        string          `gorm:"column:source_id"`
	TargetID        string          `gorm:"column:target_id"`
	Amount          decimal.Decimal `gorm:"column:amount"`
	TransactionTime time.Time       `gorm:"column:transactiontime"`
}

//HistoryTransaction 历史记录视图
type HistoryTransaction struct {
	ID              string          `gorm:"column:id"`
	OrderID         sql.NullString  `gorm:"column:order_id"`
	MemberID        string          `gorm:"column:member_id"`
	MemberName      string          `gorm:"column:mname"`
	MemberPhone     string          `gorm:"column:phone"`
	RelationID      string          `gorm:"column:relation_id"`
	RelationName    string          `gorm:"column:rname"`
	Amount          decimal.Decimal `gorm:"column:amount"`
	TransactionTime time.Time       `gorm:"column:transactiontime" json:"time"`
}

//TransactionHistoryByID 获取交易记录 根据member.id
//	greaterOrLess : ">"获取记录,"<"消费记录
func TransactionHistoryByID(db *gorm.DB, mid string, pageSize int, offset int, greatOrLess string) ([]HistoryTransaction, error) {
	var history []HistoryTransaction
	var db1 *gorm.DB
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	db1 = db.Order("transactiontime").Limit(pageSize).Offset(offset).Table("transactions t")
	db1 = db1.Joins("JOIN members m1 ON source_id=m1.id").Joins("JOIN members m2 ON target_id=m2.id")
	db1 = db1.Select("t.id id,order_id,m1.id member_id,m1.name mname,m1.phone phone,m2.id relation_id,m2.name rname,amount,transactiontime")
	db1 = db1.Where("amount"+greatOrLess+"0 and target_id=?", mid)
	db1 = db1.Find(&history)
	//db1 = db.Limit(pageSize).Offset(offset).Find(&history, "target_id=?", mid)
	if db1.RecordNotFound() {
		return history, nil
	}
	if db1.Error != nil {
		return nil, db1.Error
	}
	return history, nil
}
func (t *Transaction) fillTransaction(orderID string, sourceID string, targetID string, amount decimal.Decimal) {
	t.ID = uuid.NewV4().String()
	if len(orderID) > 0 {
		t.OrderID.Scan(orderID)
	}
	t.SourceID = sourceID
	t.TargetID = targetID
	t.Amount = amount
	t.TransactionTime = time.Now()
}

//CreateTransactionsByLevels 根据用户祖先返回产生返利关系,交易记录
func createTransactionsByLevels(db *gorm.DB, ul []UserLevel, amount decimal.Decimal, orderID string) []Transaction {
	ts := make([]Transaction, len(ul))
	id := ul[0].SonID
	//now := time.Now()
	for i := range ts {
		d1 := amount.Mul(levelRatios[i])
		d1.Round(4)
		//fmt.Println("createTransactionsByLevels", d1, levelRatios[i])
		ts[i].fillTransaction(orderID, id, ul[i].AncestorID, d1)
	}
	//4位精度造成精度差的概率极低,
	//如果做到严格准确, 可以把最后一个返利金额, 如下操作
	//[len-1:].amount = amount * sum(levelRatios) - sum([:len-1].amount)
	return ts
}

func (t *Transaction) saveNew(db *gorm.DB) error {
	db.Create(t)
	if db.NewRecord(t) {
		return errors.New("交易创建失败")
	}
	fmt.Println("done:", t)
	return nil
}
