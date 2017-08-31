package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	gorm "gopkg.in/jinzhu/gorm.v1"

	"github.com/shopspring/decimal"
)

var (
	//levelRatios 配置分成比例
	levelRatios []decimal.Decimal
	totalRatio  decimal.Decimal
)

//UserLevel 用户关系表
type UserLevel struct {
	ID           int             `gorm:"column:id"`
	SonID        string          `gorm:"column:sonnode_id"`
	AncestorID   string          `gorm:"column:ancestornode_id"`
	RoyaltyRatio decimal.Decimal `gorm:"column:royaltyratio"`
	Generations  int             `gorm:"column:generations"`
	UpdTime      time.Time       `gorm:"column:updtime"`
}

//ReferenceRelationship 用户关系视图
type ReferenceRelationship struct {
	//相关推荐人id
	ID           string          `gorm:"column:id"`
	CardNo       sql.NullString  `gorm:"column:cardno"`
	Phone        sql.NullString  `gorm:"column:phone"`
	Level        sql.NullString  `gorm:"column:level"`
	CreateTime   time.Time       `gorm:"column:createtime"`
	Reference    sql.NullString  `gorm:"column:reference_id"`
	Name         sql.NullString  `gorm:"column:name"`
	RoyaltyRatio decimal.Decimal `gorm:"column:royaltyratio"`
	Generations  int             `gorm:"column:generations"`
}

//ReferenceOutput 用户关系视图输出json
type ReferenceOutput struct {
	//相关推荐人id
	ID           string          `json:"id"`
	CardNo       string          `json:"cardNo"`
	Phone        string          `json:"phone"`
	Level        string          `json:"level"`
	CreateTime   time.Time       `json:"createTime"`
	Reference    string          `json:"refID"`
	Name         string          `json:"name"`
	RoyaltyRatio decimal.Decimal `json:"royaltyratio"`
	Generations  int             `json:"generations"`
}

// 是否mid祖先中包含 ref
// return nil,nil 不包含
func checkAncestor(db *gorm.DB, mid string, ref string) (*UserLevel, error) {
	ul := UserLevel{}
	db1 := db.Where("sonnode_id=? and ancestornode_id=?", mid, ref).First(&ul)
	if db1.RecordNotFound() {
		return nil, nil
	}
	if db1.Error != nil {
		return nil, db1.Error
	}
	return &ul, nil
}

//InitLevelRatios 初始化分成比例
func InitLevelRatios(ratios *([]decimal.Decimal)) error {
	if len(*ratios) <= 0 {
		return errors.New("无返利配置")
	}
	levelRatios = make([]decimal.Decimal, len(*ratios))
	totalRatio = decimal.New(0, 4)
	for i, d := range *ratios {
		levelRatios[i] = d
		totalRatio.Add(d)
	}
	return nil
}

//CreateLevels 创建level记录
func CreateLevels(db *gorm.DB, member *Member) (*UserLevel, error) {
	u := &UserLevel{}
	l := len(levelRatios)
	length := l
	if l <= 0 {
		return nil, errors.New("返利配置错误")
	}

	ancestors := make([]string, l, l)
	ancestors[0] = member.ID
	var old string
	if !member.Reference.Valid {
		l = 1
	} else {
		old = member.Reference.String
		ancestors[1] = old

		var ul []UserLevel
		db1 := db.Order("generations").Limit(l-2).Find(&ul, "generations>0 and sonnode_id=?", old)
		if db1.Error == nil {
			var lu int
			if ul == nil {
				lu = 0
			} else {
				lu = len(ul)
			}

			length -= lu + 2 //2(自己,第一个reference)
			if length > 0 {  //祖先代 数不足
				l -= length
			}
			//fmt.Println("ancnetor count:",l, length, lu)
			for i := 2; i < l; i++ {
				ancestors[i] = ul[i-2].AncestorID
			}
		} else {
			log.Printf("user level get ancestors sql error:%s", db1.Error)
		}
	}
	fmt.Println("valid ancnetor:", l)
	for i := 0; i < l; i++ {
		//fmt.Println("add ul:",
		u.AddNewUserLevel(db, member.ID, ancestors[i], i)
	}

	return u, nil
}

//AddNewUserLevel 用输入字段创建user level记录
func (u *UserLevel) AddNewUserLevel(db *gorm.DB, son string, ancestor string, generations int) bool {
	u.fillNewUserLevel(son, ancestor, generations)
	db.Create(u)
	return db.NewRecord(u)
}

//FindReferenceByID 按reference_id查找继承关系
func FindReferenceByID(db *gorm.DB, id string) ([]ReferenceRelationship, error) {
	var refs []ReferenceRelationship
	db1 := db.Table("members").Joins("JOIN user_levels on user_levels.ancestornode_id=members.id").Select("members.*,royaltyratio,generations").Where("sonnode_id=?", id).Find(&refs)
	if db1.RecordNotFound() {
		return nil, sql.ErrNoRows
	}
	if db1.Error != nil {
		return nil, db1.Error
	}
	return refs, nil
}

//MapReference2Output 数据视图转换输出json对象
func MapReference2Output(rs []ReferenceRelationship) []ReferenceOutput {
	ros := make([]ReferenceOutput, len(rs))
	for i, rr := range rs {
		ros[i].ID = rr.ID
		ros[i].CardNo = rr.CardNo.String
		ros[i].CreateTime = rr.CreateTime
		ros[i].Generations = rr.Generations
		ros[i].Level = rr.Level.String
		ros[i].Name = rr.Name.String
		ros[i].Phone = rr.Phone.String
		ros[i].Reference = rr.Reference.String
		ros[i].RoyaltyRatio = rr.RoyaltyRatio
	}
	return ros
}

//FillNewUserLevel 用输入字段创建 user level 对象
func (u *UserLevel) fillNewUserLevel(son string, ancestor string, generations int) {
	u.ID = 0 //自增, 清除
	u.SonID = son
	u.AncestorID = ancestor
	u.RoyaltyRatio = levelRatios[generations]
	u.Generations = generations
	u.UpdTime = time.Now()
}

//GetLevelsByMember 获取用户
func getLevelsByMember(db *gorm.DB, mid string) ([]UserLevel, error) {
	var ul []UserLevel
	db1 := db.Order("generations").Limit(len(levelRatios)).Find(&ul, "sonnode_id=?", mid)
	if db1.Error != nil {
		fmt.Println(db1.Error)
	} else { //校验返回结果 有序, 连续
		for i, u := range ul {
			if u.Generations != i {
				return nil, errors.New("user level 不连续. ")
			}
		}
	}

	return ul, db1.Error
}
