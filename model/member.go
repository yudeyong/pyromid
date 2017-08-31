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

	"github.com/e2u/goboot"
	"github.com/twinj/uuid"
)

//Member 数据库对象
type Member struct {
	ID         string         `gorm:"column:id"`
	CardNo     sql.NullString `gorm:"column:cardno"`
	Phone      sql.NullString `gorm:"column:phone"`
	Level      sql.NullString `gorm:"column:level"`
	CreateTime time.Time      `gorm:"column:createtime"`
	Reference  sql.NullString `gorm:"column:reference_id"`
	Name       sql.NullString `gorm:"column:name"`
}

//MemberOutput json输出对象
type MemberOutput struct {
	MemberID string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	CardNo   string `json:"cardNo"`
	RefID    string `json:"refID"`
	Time     string `json:"createTime"`
	Level    string `json:"level"`
}

const (
	regular = "^(13[0-9]|14[57]|15[0-35-9]|18[07-9])\\d{8}$"
)

//MapMembers2Output 转换数据库 member数组对象输出json
func MapMembers2Output(members []Member) []MemberOutput {
	mo := make([]MemberOutput, len(members))
	for i, m := range members {
		mo[i] = *m.Map2Output()
	}
	return mo
}

//Map2Output 转换数据库 member对象输出json
func (m *Member) Map2Output() *MemberOutput {
	mo := &MemberOutput{}
	mo.MemberID = m.ID
	mo.Name = m.Name.String
	mo.Phone = m.Phone.String
	mo.CardNo = m.CardNo.String
	mo.RefID = m.Reference.String
	mo.Time = m.CreateTime.Format("2006-01-02 15:04")
	mo.Level = m.Level.String
	return mo
}

//BindMemberReference 绑定推荐会员
func BindMemberReference(db *gorm.DB, mid string, ref string) error {
	m := NewMember()
	err := m.FindByID(db, mid)
	if err != nil {
		return err
	}
	if true == m.Reference.Valid {
		return errors.New("用户已有推荐用户")
	}
	var isAncestor bool
	//被推荐用户不能是推荐用户的'祖先'
	isAncestor, err = checkAncestor(db, ref, mid)
	if err != nil {
		return err
	}
	if isAncestor {
		return errors.New("不能循环推荐")
	}
	r := NewMember()
	if err = r.FindByID(db, ref); err != nil {
		return err
	}
	//fmt.Println(m, r)
	m.Reference.Scan(r.ID)
	tx := db.Begin() //开启事务
	db1 := tx.Save(m)
	if db1.Error != nil {
		tx.Rollback()
		goboot.Log.Error(db1.Error)
		return db1.Error
	}
	ul, err := CreateLevels(tx, m, 1)
	if ul == nil {
		tx.Rollback()
		goboot.Log.Error(err)
		return err
	}
	tx.Commit()
	return nil
}

//ValidatePhone 校验手机号格式
func ValidatePhone(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

//NewMember 空Member
func NewMember() *Member {
	return &Member{}
}

//FindByPhoneOrCardno 按手机号查找, 找不到时,按卡号查找
// error code:
//	ResInvalid  phone,cardno 不能都为空
//	ResPhoneInvalid phone 无效
//	ResNotFound  找不到记录
//	ResFound  成功找到
//	ResFail		异常
func (m *Member) FindByPhoneOrCardno(db *gorm.DB, phone string, cardno string) (string, error) {
	if len(phone) == 0 && len(cardno) == 0 {
		return ResInvalid, errors.New("请输入手机号或卡号")
	}
	if len(phone) != 0 {
		code, err := m.FindByPhone(db, phone)
		//fmt.Println("find phone", code, err, phone, m)
		if err != nil {
			return code, err
		}
		return ResFound, nil
	}
	return m.FindByCardno(db, cardno)
}

//FindByID 按id查找
func (m *Member) FindByID(db *gorm.DB, id string) error {
	db1 := db.Where("id=?", id).Find(&m)
	if db1.RecordNotFound() {
		return sql.ErrNoRows
	}
	return db1.Error
}

//FindByPhone 按电话查找
func (m *Member) FindByPhone(db *gorm.DB, phone string) (string, error) {
	if !ValidatePhone(phone) {
		return ResPhoneInvalid, errors.New("无效电话号码")
	}
	db1 := db.First(&m, "phone=?", phone)
	//fmt.Println("FindByPhone",phone, db1.Error)
	if db1.RecordNotFound() {
		return ResNotFound, sql.ErrNoRows
	}
	if db1.Error != nil {
		return ResWrongSQL, db1.Error
	}
	return ResFound, nil
}

//FindByCardno 按卡号查找
func (m *Member) FindByCardno(db *gorm.DB, cardno string) (string, error) {
	db1 := db.First(&m, "cardno=?", cardno)
	if db1.RecordNotFound() {
		return ResNotFound, sql.ErrNoRows
	}
	if db1.Error != nil {
		return ResWrongSQL, db1.Error
	}
	return ResFound, nil
}

//FindMemberLikeName 按姓名模糊查找
func FindMemberLikeName(db *gorm.DB, name string) ([]Member, error) {
	var ms []Member
	db1 := db.Model(&Member{}).Where("name like ?", "%"+name+"%").Find(&ms)
	//fmt.Println(ms)
	if db1.RecordNotFound() {
		return nil, sql.ErrNoRows
	}
	if db1.Error != nil {
		return nil, db1.Error
	}
	return ms, nil

}
func (m *Member) String() string {
	p, _ := m.Phone.Value()
	c, _ := m.CardNo.Value()
	r, _ := m.Reference.Value()
	return fmt.Sprintf("id=%s,p=%s,c=%s,ref=%s,time=%s", m.ID, p, c, r, m.CreateTime)
}

//FindByInfo reference 满足 phone No.按phone算, 否则按卡号查询
func (m *Member) FindByInfo(db *gorm.DB, reference string) (string, error) {
	if len(reference) == 0 {
		return ResInvalid, errors.New("无引荐人卡号,或手机号")
	}
	if ValidatePhone(reference) {
		code, err := m.FindByPhone(db, reference)
		//fmt.Println("ref:", reference, err, code)
		if err != nil {
			return code, err
		}
		return ResFound, nil
	}
	return m.FindByCardno(db, reference)

}

//AddNewMember 查找推荐用户,添加新用户
//	name, phone, cardno, refname, refphone, refcardno, refID, level string
//	return
//		*Member	:	推荐用户,
//		Member[]: 推荐用户列表,
//		code		:	ResInvalid/ResNotFound/ResMore1/ResFailCreateMember/ResOK,
//		err message
func AddNewMember(db *gorm.DB, name, phone, cardno, refname, refphone, refcardno, refID, level string) (*Member, []Member, string, string) {
	//fmt.Println(name, ";", phone, ";", cardno, ";", refname, ";", refphone, ";", refcardno, ";", refID, ";", level)
	if len(phone) == 0 && len(cardno) == 0 {
		return nil, nil, ResInvalid, "请提供会员数据"
	}
	ref := NewMember()
	if len(refID) == 0 {
		if len(refphone) != 0 || len(refcardno) != 0 || len(refname) != 0 {
			members, _, _ := SearchMembersByInfo(db, refphone, refcardno, refname)
			//fmt.Println("ref search rsult:", members)
			if members == nil {
				// errstr = "介绍用户没找到"
				// code = model.ResNotFound
				return nil, nil, ResNotFound, "介绍用户没找到"
			}
			if len(members) > 1 {
				return nil, members, ResMore1, "请选择引荐用户"
			}
			refID = members[0].ID
		}
	}
	if err := ref.createMember(db, phone, cardno, refID, level, name); err != nil {
		log.Println(err, phone, cardno)
		return nil, nil, ResFailCreateMember, "用户创建失败," + phone + err.Error()
	}
	return ref, nil, ResOK, ""
}

//createMember 简单创建用户
func (m *Member) createMember(db *gorm.DB, phone string, cardno string, reference string, level string, name string) error {
	m.fillNewMember(phone, cardno, reference, level, name)
	db.Create(m)
	if db.NewRecord(m) {
		return errors.New("用户创建失败")
	}

	go CreateLevels(db, m, 0) //异步创建族谱, 快速返回用户创建请求
	return nil
}

//fillNewMember 填充新member对象
func (m *Member) fillNewMember(phone string, cardno string, reference string, level string, name string) (*Member, error) {
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
	return m, nil
}

//SearchMembersByInfo 搜索用户,根据
//	phone,cardno or name
//	返回 member list, msg code, msg
//		code:
//			ResInvalid:	参数不足
//			ResFail		:	异常
//			ResMore		:	结果超过1个
//			ResNotFound:未找到
//			ResFound	:	成功找到1个
func SearchMembersByInfo(db *gorm.DB, phone string, cardno string, name string) ([]Member, string, string) {
	var m *Member
	var err error
	m = NewMember()
	var members []Member
	if len(phone) == 0 && len(cardno) == 0 {
		if len(name) == 0 {
			return nil, ResInvalid, ""
		} // else {
		members, err = FindMemberLikeName(db, name)
		if err != nil {
			return nil, ResFail, err.Error()
		}
		if len(members) > 1 {
			return members, ResMore, ""
		}
		if len(members) < 1 {
			return nil, ResNotFound, ""
		}
		return members, ResFound, ""
		//}
	}
	var code string
	code, err = m.FindByPhoneOrCardno(db, phone, cardno)
	if err != nil {
		return nil, code, err.Error()
	}
	return []Member{*m}, code, ""
}

//SearchMembers 搜索member根据
//	id, phone, cardno or name
//	id 优先
//	返回 member list, msg code, msg
//		code 定义,详见 SearchMembersByInfo
func SearchMembers(db *gorm.DB, id string, phone string, cardno string, name string) ([]Member, string, string) {
	if len(id) == 0 {
		// err = m.FindByID(app.App.DB, id)
		// if err != nil {
		// 	fmt.Fprintf(w, errMsg.messageString(model.ResFail, err.Error()))
		// 	return
		// }
		//} else {
		return SearchMembersByInfo(db, phone, cardno, name)
	}

	m := NewMember()
	err := m.FindByID(db, id)
	if err != nil {
		return nil, ResFail, err.Error()
	}
	return []Member{*m}, ResFound, ""
}

//UpdateMember 更新member
//	cardno 忽略
func UpdateMember(db *gorm.DB, id, phone, cardno, name string) error {
	//	db.Table("users").Where("id IN (?)", []int{10, 11}).
	//Updates(map[string]interface{}{"name": "hello", "age": 18})
	db1 := db.Table("members").Where("id=?", id).Update(map[string]interface{}{"name": name, "phone": phone})
	if db1.Error != nil {
		return db1.Error
	}
	return nil
}
