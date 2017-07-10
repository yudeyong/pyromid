package conf
import (
  "fmt"
  gorm "gopkg.in/jinzhu/gorm.v1"
  "strconv"
  "strings"
  "log"
  _ "errors"

  "../model"

  "github.com/shopspring/decimal"
)

const (
	Levels    = "levels"  // 层数
  LevelRatioString = "level%dratio"  //各级分成比例
  LevelRatioSql = "level%ratio"
)

func InitLevels(db *gorm.DB,	levelRatios *([]decimal.Decimal) )(error){
  var err error;
  ss :=model.NewSystemSettings();

  _,err = ss.FindByCode(db ,Levels)
  if err!=nil {
    return err;
  }

  //var v 配置数量
  v, err := strconv.Atoi(ss.Value);
  if err!=nil {
    return err;
  }

  var sss []model.SystemSettings
  db.Where("code like ?",LevelRatioSql).Order("id").Find(&sss)
  //实际数量
  size := len(sss);
  if size!=v+1 {
    log.Printf("Config %s warning: (SystemSettings)%d!=(db)%d", Levels, v, size-1)
  }
  *levelRatios = make([]decimal.Decimal,size,size)//空着第一个,从1开始编号
  var i, start int
  start = strings.Index( LevelRatioSql, "%")
  //ASSERT 只支持1位数的分级层数>=10的层数不支持

  for i,*ss = range sss {

    if i,err = strconv.Atoi( string([]rune(ss.Code)[start:start+1]));err!=nil {
      return err;
    }

    if (*levelRatios)[i],err = decimal.NewFromString(ss.Value); err!=nil {
      return err;
    }
  }
  fmt.Printf("%s\n",levelRatios);
  return nil
}
