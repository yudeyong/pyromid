package model
import (
	"database/sql"
	"errors"
	"fmt"
	"time"
  "regexp"
	gorm "gopkg.in/jinzhu/gorm.v1"

	"github.com/twinj/uuid"
)

type Member struct {
	ID                      string  				`gorm:"column:id"`
	CardNo                  sql.NullString	`gorm:"column:cardno"`
	Phone                   sql.NullString  `gorm:"column:phone"`
  Level                   sql.NullString  `gorm:"column:level"`
	CreateTime              time.Time 			`gorm:"column:createtime"`
	Reference								sql.NullString  `gorm:"column:reference_id"`
}
const (
    regular = "^(13[0-9]|14[57]|15[0-35-9]|18[07-9])\\d{8}$"
)

func ValidatePhone(mobileNum string) bool {
  reg := regexp.MustCompile(regular)
  return reg.MatchString(mobileNum)
}

func NewMember() *Member{
  return &Member{}
}

const (
  ResInvalid =      403
  ResPhoneInvalid = 4031
  ResNotFound =     404
  ResFound  =       200
  ResWrongSql =     500
)
//error code:
// ResInvalid  phone,cardno 不能都为空
// ResPhoneInvalid phone 无效
// ResNotFound  找不到记录
// ResFound  成功找到
func (m *Member) FindByPhoneOrCardno(db *gorm.DB, phone string, cardno string) (error, int){
  if len(phone)==0 && len(cardno)==0 {
    return errors.New( "请输入手机号或卡号"), ResInvalid
  }
  if len(phone)==0 {
    err, code := m.FindByPhone(db, phone)
    if (err!=nil){
      return err, code
    }
    return nil, ResFound
  }
  return m.FindByCardno(db, cardno)
}

func (m *Member) FindByPhone(db *gorm.DB, phone string) (error, int){
  if !ValidatePhone(phone){
    return errors.New("无效电话号码"), ResPhoneInvalid
  }
  db1 := db.First(&m,"phone=?", phone)
	if db1.RecordNotFound() {
		return sql.ErrNoRows, ResNotFound
	}
  if db1.Error!=nil{
    return db1.Error,ResWrongSql
  }
  return nil, ResFound
}

func (m *Member) FindByCardno(db *gorm.DB, cardno string) (error, int){
  db1 := db.First(&m,"cardno=?", cardno)
	if db1.RecordNotFound() {
		return sql.ErrNoRows, ResNotFound
	}
  if db1.Error!=nil{
    return db1.Error,ResWrongSql
  }
  return nil, ResFound
}

func (m *Member) String() string{
	p,_ := m.Phone.Value()
	c,_ := m.CardNo.Value()
	r,_ := m.Reference.Value()
  return fmt.Sprintf("id=%s,p=%s,c=%s,ref=%s",m.ID,p,c,r)
}

//reference 满足 phone No.按phone算, 否则按卡号查询
func (m *Member) FindByInfo(db *gorm.DB, reference string) (error, int){
  if len(reference)==0 {
    return errors.New( "无引荐人卡号,或手机号"), ResInvalid
  }
  if !ValidatePhone(reference){
    err, code := m.FindByPhone(db, reference)
    if (err!=nil){
      return err, code
    }
    return nil, ResFound
  }
  return m.FindByCardno(db, reference)

}
func (m *Member)AddNewMember(db *gorm.DB, phone string, cardno string, level string) (bool){
	m.FillNewMember(phone, cardno, level)
	db.Create(m)
	return db.NewRecord(m)
}
func (m *Member)FillNewMember(phone string, cardno string, level string) (error, *Member){
	m.ID = uuid.NewV4().String()
	m.Phone.Scan(phone)
	m.CardNo.Scan( cardno )
	if len(level)!=0 {
		m.Level.String,m.Level.Valid = level,true
	}
	m.CreateTime = time.Now()
	return nil,m
}

func (o *Member) Save(db *gorm.DB) error {
	return db.Save(o).Error
}
