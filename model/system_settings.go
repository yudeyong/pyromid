package model

import (
	"database/sql"
	_ "errors"
	"fmt"
	"time"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

// 渠道充值记录结构
type SystemSettings struct {
	ID                      uint64  				`gorm:"column:id"`
	Code                    string					`gorm:"column:code"`
	Value                   string  				`gorm:"column:value"`
	Remark                  sql.NullString  `gorm:"column:remark"`
	Updtime                 time.Time 			`gorm:"column:updtime"`
}

// String
func (o *SystemSettings) String() string {
	return fmt.Sprintf("id: %d code: %s value: %s remark: %s upd: %s\n",
		o.ID, o.Code, o.Value, o.Remark, o.Updtime,
	)
}

// NewSystemSettings
func NewSystemSettings() *SystemSettings {
	return &SystemSettings{}
}

// FindByID 根据ID获取充值记录
func (o *SystemSettings) FindByCode(db *gorm.DB, code string) (*SystemSettings, error) {

	if db.First(&o,"code=?", code).RecordNotFound() {
		return nil, sql.ErrNoRows
	}
	return o, nil
}
