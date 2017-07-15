package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	gorm "gopkg.in/jinzhu/gorm.v1"

	"github.com/twinj/uuid"
)

type Member struct {
	ID         string         `gorm:"column:id"`
	CardNo     sql.NullString `gorm:"column:cardno"`
	Phone      sql.NullString `gorm:"column:phone"`
	Level      sql.NullString `gorm:"column:level"`
	CreateTime time.Time      `gorm:"column:createtime"`
	Reference  sql.NullString `gorm:"column:reference_id"`
	Name       sql.NullString `gorm:"column:name"`
}

const (
	regular = "^(13[0-9]|14[57]|15[0-35-9]|18[07-9])\\d{8}$"
)

func ValidatePhone(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func NewMember() *Member {
	return &Member{}
}

const (
	ResOK               = "200"
	ResInvalid          = "412"
	ResPhoneInvalid     = "4121"
	ResNotFound         = "404"
	ResFound            = "200"
	ResWrongSql         = "500"
	ResFail             = "500"
	ResFailCreateMember = "501"
)

//error code:
// ResInvalid  phone,cardno 不能都为空
// ResPhoneInvalid phone 无效
// ResNotFound  找不到记录
// ResFound  成功找到
func (m *Member) FindByPhoneOrCardno(db *gorm.DB, phone string, cardno string) (error, string) {
	if len(phone) == 0 && len(cardno) == 0 {
		return errors.New("请输入手机号或卡号"), ResInvalid
	}
	if len(phone) != 0 {
		err, code := m.FindByPhone(db, phone)
		if err != nil {
			return err, code
		}
		return nil, ResFound
	}
	return m.FindByCardno(db, cardno)
}

func (m *Member) FindByPhone(db *gorm.DB, phone string) (error, string) {
	if !ValidatePhone(phone) {
		return errors.New("无效电话号码"), ResPhoneInvalid
	}
	db1 := db.First(&m, "phone=?", phone)
	//fmt.Println("FindByPhone",phone, db1.Error)
	if db1.RecordNotFound() {
		return sql.ErrNoRows, ResNotFound
	}
	if db1.Error != nil {
		return db1.Error, ResWrongSql
	}
	return nil, ResFound
}

func (m *Member) FindByCardno(db *gorm.DB, cardno string) (error, string) {
	db1 := db.First(&m, "cardno=?", cardno)
	if db1.RecordNotFound() {
		return sql.ErrNoRows, ResNotFound
	}
	if db1.Error != nil {
		return db1.Error, ResWrongSql
	}
	return nil, ResFound
}

func (m *Member) String() string {
	p, _ := m.Phone.Value()
	c, _ := m.CardNo.Value()
	r, _ := m.Reference.Value()
	return fmt.Sprintf("id=%s,p=%s,c=%s,ref=%s,time=%s", m.ID, p, c, r, m.CreateTime)
}

//reference 满足 phone No.按phone算, 否则按卡号查询
func (m *Member) FindByInfo(db *gorm.DB, reference string) (error, string) {
	if len(reference) == 0 {
		return errors.New("无引荐人卡号,或手机号"), ResInvalid
	}
	if ValidatePhone(reference) {
		err, code := m.FindByPhone(db, reference)
		fmt.Println("ref:", reference, err, code)
		if err != nil {
			return err, code
		}
		return nil, ResFound
	}
	return m.FindByCardno(db, reference)

}
func AddNewMember(db *gorm.DB, phone string, cardno string, reference string, level string, name string) *Member {
	ref := NewMember()
	var referenceID string

	if len(reference) > 0 {
		err, _ := ref.FindByInfo(db, reference)
		if err == nil {
			referenceID = ref.ID
		}
	}
	if err := ref.CreateMember(db, phone, cardno, referenceID, level, name); err != nil {
		log.Println(err, phone, cardno)
		return nil
	}
	return ref
}

func (m *Member) CreateMember(db *gorm.DB, phone string, cardno string, reference string, level string, name string) error {
	m.FillNewMember(phone, cardno, reference, level, name)
	db.Create(m)
	if db.NewRecord(m) {
		return errors.New("用户创建失败")
	} else {
		u := &user_level{}
		go u.CreateLevels(db, m) //异步创建族谱, 快速返回用户创建请求
		return nil
	}
}
func (m *Member) FillNewMember(phone string, cardno string, reference string, level string, name string) (error, *Member) {
	m.ID = uuid.NewV4().String()
	m.Phone.Scan(phone)
	if len(cardno) != 0 {
		no, err := strconv.Atoi(cardno)
		if err == nil {
			UpdateNewCard(no)
		}
	} else {
		cardno = GetNewCard()
	}
	m.CardNo.Scan(cardno)

	if len(reference) > 0 {
		m.Reference.Scan(reference)
	}
	if len(level) != 0 {
		m.Level.String, m.Level.Valid = level, true
	}
	if len(name) > 0 {
		m.Name.Scan(name)
	}
	m.CreateTime = time.Now()
	return nil, m
}
