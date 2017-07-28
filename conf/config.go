package conf

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	gorm "gopkg.in/jinzhu/gorm.v1"

	"../model"

	"github.com/shopspring/decimal"
)

const (
	//Levels 层数配置
	Levels = "levels"
	//LevelRatioString 各级分成比例
	LevelRatioString = "level%dratio"
	//LevelRatioSQL 各级对应sql查询语句
	LevelRatioSQL = "level%ratio"
)

//InitLevels 初始化各级回扣比例
func InitLevels(db *gorm.DB) (*([]decimal.Decimal), error) {
	var err error
	var levelRatios ([]decimal.Decimal)
	ss := model.NewSystemSettings()

	_, err = ss.FindByCode(db, Levels)
	if err != nil {
		return nil, err
	}

	//var v 配置数量
	v, err := strconv.Atoi(ss.Value)
	if err != nil {
		return nil, err
	}

	var sss []model.SystemSettings
	db.Where("code like ?", LevelRatioSQL).Order("id").Find(&sss)
	//实际数量
	size := len(sss)
	if size != v+1 {
		log.Printf("Config %s warning: (SystemSettings)%d!=(db)%d", Levels, v, size-1)
	}
	levelRatios = make([]decimal.Decimal, size, size) //空着第一个,从1开始编号
	var i, start int
	start = strings.Index(LevelRatioSQL, "%")
	//ASSERT 只支持1位数的分级层数>=10的层数不支持

	for i, *ss = range sss {

		if i, err = strconv.Atoi(string([]rune(ss.Code)[start : start+1])); err != nil {
			return nil, err
		}

		if (levelRatios)[i], err = decimal.NewFromString(ss.Value); err != nil {
			return nil, err
		}
	}
	fmt.Println("lvl ", levelRatios)
	return &levelRatios, nil
}
