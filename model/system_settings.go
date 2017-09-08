package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	gorm "gopkg.in/jinzhu/gorm.v1"
)

//SystemSettings 渠道充值记录结构
type SystemSettings struct {
	ID      uint64         `gorm:"column:id"`
	Code    string         `gorm:"column:code"`
	Value   string         `gorm:"column:value"`
	Remark  sql.NullString `gorm:"column:remark"`
	Updtime time.Time      `gorm:"column:updtime"`
}

// String
func (o *SystemSettings) String() string {
	return fmt.Sprintf("id: %d code: %s value: %s remark: %s upd: %s\n",
		o.ID, o.Code, o.Value, o.Remark.String, o.Updtime,
	)
}

//NewSystemSettings get empty object
func NewSystemSettings() *SystemSettings {
	return &SystemSettings{}
}

// FindByCode 根据code获取充值记录
func (o *SystemSettings) FindByCode(db *gorm.DB, code string) (*SystemSettings, error) {

	if db.First(&o, "code=?", code).RecordNotFound() {
		return nil, sql.ErrNoRows
	}
	return o, nil
}

//UpdateRatios 更新费率
// return code, msg
func UpdateRatios(db *gorm.DB, r []string, sync, updAll string) (string, string) {
	var err error
	var mask []bool //优化更新,仅更新不同项
	l := len(r)
	tmp := make([]decimal.Decimal, l)
	var updateExist, updateAll bool //是否更新已有记录
	var flag = true                 //队尾标志
	var needUpdate bool             //真需要更新
	if len(sync) > 0 {
		updateExist, err = strconv.ParseBool(sync)
		if err != nil {
			return ResFail, err.Error()
		}
	}
	if len(updAll) > 0 {
		updateAll, err = strconv.ParseBool(updAll)
		if err != nil {
			return ResFail, err.Error()
		}
	}
	mask = make([]bool, len(levelRatios))
	//total := zero
	onepercent := decimal.New(1, -2)
	for i := l; i > 0; {
		i--
		tmp[i], err = decimal.NewFromString(r[i])
		if err != nil {
			return ResInvalid, err.Error()
		}

		//清除队尾为0的记录
		if flag && zero.Equal(tmp[i]) {
			l--
			continue
		} else {
			if flag {
				flag = false
			}
		}
		tmp[i] = onepercent.Mul(tmp[i])
		//total = total.Add(tmp[i])
		if i < len(mask) {
			mask[i] = !levelRatios[i].Equal(tmp[i])
			needUpdate = needUpdate || mask[i]
		}
	}
	if l < len(r) {
		tmp = tmp[:l]
		if l < len(mask) {
			mask = mask[:l]
		}
	}
	if !needUpdate && l == len(mask) {
		return ResOK, "无需更新"
	}
	err = updateRatiosDB(db, tmp)
	if err != nil {
		return ResFail, err.Error()
	}
	oldRatios := levelRatios
	InitLevelRatios(&tmp)
	if updateExist {
		//耗时操作,异步进行
		go updateAllUserLevel(db, mask, oldRatios, updateAll)
		return ResOK, "更新中"
	}
	return ResOK, "更新成功"
}

func updateRatiosDB(db *gorm.DB, ratios []decimal.Decimal) error {
	tx := db.Begin() //开启事务
	ss := make([]SystemSettings, len(ratios))
	now := time.Now()
	db1 := tx.Table("system_settings").Where("code='levels'").Update(map[string]interface{}{"value": len(ratios) - 1})
	//fmt.Println("update")
	if db1.Error != nil {
		tx.Rollback()
		return db1.Error
	}
	db1 = tx.Delete(SystemSettings{}, "code like 'level%ratio'")
	//fmt.Println("delete")
	if db1.Error != nil {
		tx.Rollback()
		return db1.Error
	}
	for i, r := range ratios {
		ss[i].Code = "level" + strconv.Itoa(i) + "ratio"
		ss[i].Value = r.String()
		ss[i].Remark.Scan("第" + strconv.Itoa(i) + "层分成比例")
		ss[i].Updtime = now
		//fmt.Println("add"+strconv.Itoa(i), " ", ss[i])
		tx.Create(&(ss[i]))
		if tx.NewRecord(&(ss[i])) {
			tx.Rollback()
			return errors.New("用户创建失败")
		}
	}
	tx.Commit()
	return nil
}
