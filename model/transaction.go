package model

import (
	"database/sql"
	_ "errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

//Transaction 用户关系表
type Transaction struct {
	ID              int                 `gorm:"column:id"`
	OrderID         sql.NullString      `gorm:"column:order_id"`
	SourceID        string              `gorm:"column:source_id"`
	TargetID        string              `gorm:"column:target_id"`
	Amount          decimal.NullDecimal `gorm:"column:amount"`
	TransactionTime time.Time           `gorm:"column:transactiontime"`
}

//CreateTransactionsByLevels 根据用户祖先返回产生返利关系,交易记录
func CreateTransactionsByLevels(db *gorm.DB, ul []UserLevel, amount int, orderID string) []Transaction {
	ts := make([]Transaction, len(ul))
	id := ul[0].SonID
	now := time.Now()
	for i := range ts {
		if len(orderID) > 0 {
			ts[i].OrderID.Scan(orderID)
		}
		ts[i].SourceID = id
		ts[i].TargetID = ul[i].AncestorID
		d := decimal.New(int64(amount), -2)
		d1 := d.Mul(LevelRatios[i])
		d.Round(4)
		fmt.Println(d, d1, LevelRatios[i])
		ts[i].Amount.Scan(d)
		ts[i].TransactionTime = now
	}
	//4位精度造成精度差的概率极低,
	//如果做到严格准确, 可以把最后一个返利金额, 如下操作
	//[len-1:].amount = amount * sum(LevelRatios) - sum([:len-1].amount)
	return ts
}
