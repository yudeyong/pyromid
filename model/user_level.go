package model
import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	_"strconv"
	gorm "gopkg.in/jinzhu/gorm.v1"

	"github.com/shopspring/decimal"

)
var(
	LevelRatios []decimal.Decimal //配置分成比例
)
type user_level struct{
	ID                  int  				`gorm:"column:id"`
	SonID     		      string					`gorm:"column:sonnode_id"`
	AncestorID     			sql.NullString  `gorm:"column:ancestornode_id"`
  RoyaltyRatio        decimal.Decimal `gorm:"column:royaltyratio"`
	Generations         int		  				`gorm:"column:generations"`
	UpdTime		          time.Time 			`gorm:"column:updtime"`
}

func (u *user_level)CreateLevels( db * gorm.DB, member *Member )error{
	l := len(LevelRatios)
	length := l
	if (l<=0){
		return errors.New("返利配置错误")
	}

	ancestors := make([]string, l, l)
	ancestors[0] = member.ID
	var old string
	if !member.Reference.Valid{
		l = 1
	}else
	{
		old = member.Reference.String
		ancestors[1] = old

		var ul []user_level
		db1 := db.Order("generations").Find(&ul, "generations>0 and sonnode_id=?",old)
		if (db1.Error==nil){
			var lu int
			if (ul==nil){
				lu = 0
			}else
			{
				lu = len(ul)
			}

			length -= lu +2//2(自己,第一个reference)
			if length>0{//祖先代 数不足
				l -= length
			}
			fmt.Println("ancnetor count:",l, length, lu)
			for i:=2; i<l; i+=1{
				ancestors[i] = ul[i-2].AncestorID.String
			}
		}else
		{
			log.Printf("user level get ancestors sql error:%s", db1.Error)
		}
	}
	fmt.Println("valid ancnetor:",l)
	for i:=0; i<l; i+= 1{
		fmt.Println("add ul:",u.AddNewUserLevel(db, member.ID, ancestors[i],i))
	}

	return nil
}

func (u *user_level)AddNewUserLevel(db *gorm.DB, son string, ancestor string, generations int) (bool){
	u.FillNewUserLevel(son, ancestor, generations)
	db.Create(u)
	return db.NewRecord(u)
}

func (u *user_level) FillNewUserLevel(son string, ancestor string, generations int){
	u.ID = 0//自增, 清除
	u.SonID = son
	u.AncestorID.Scan( ancestor)
	u.RoyaltyRatio = LevelRatios[generations]
	u.Generations = generations
	u.UpdTime = time.Now()
}
