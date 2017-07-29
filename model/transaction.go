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

//Transaction 用户关系表
type Transaction struct {
	ID              string          `gorm:"column:id"`
	OrderID         sql.NullString  `gorm:"column:order_id"`
	SourceID        string          `gorm:"column:source_id"`
	TargetID        string          `gorm:"column:target_id"`
	Amount          decimal.Decimal `gorm:"column:amount"`
	TransactionTime time.Time       `gorm:"column:transactiontime"`
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
