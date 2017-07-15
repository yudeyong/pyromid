package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	_ "strings"

	"./model"
)

//GetPara 获取key对应参数值, 不存在返回""
func GetPara(r *http.Request, key string) string {
	arr := r.Form[key]
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}

//Consume 消耗积分
//  id      : memberid
//  amount  : 消费积分
//  usePoint: 是否使用余额,缺省否
func (c *Controller) Consume(w http.ResponseWriter, r *http.Request) {

}

type userResp struct {
	RespCode string `json:"respCode"`
	RespMsg  string `json:"respMsg"`
	MemberID string `json:"id"`
	Amount   string `json:"amount"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

func (r *userResp) String() string {
	jb, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(jb)
}
func (r *userResp) SetMessage(code string, message string) {
	r.RespCode = code
	r.RespMsg = message
}
func (r *userResp) CopyMemberInfo(m *model.Member) {
	r.MemberID = m.ID
	r.Name = m.Name.String
	r.Phone = m.Phone.String
}
func (r *userResp) Message(code string, message string) string {
	r.SetMessage(code, message)
	return r.String()
}

//CheckUser 检查用户
//  phone     : 用户手机号
//  cardno    : 用户卡号,与手机号,至少一个非空. 不存在时, 创建新用户, 及其推荐返利关系树
//  reference : 推荐人,识别为11位手机号,按手机号,否则按卡号查询; 老用户无效
//  Name      : 用户名,老用户无效
func (c *Controller) CheckUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	//  map :=
	resp := userResp{Amount: "0"}
	phone := GetPara(r, "phone")
	//fmt.Println(phone)
	cardno := GetPara(r, "cardno")
	/*  for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	*/
	m := model.NewMember()
	err, code := m.FindByPhoneOrCardno(App.DB, phone, cardno)
	//fmt.Println(err,code)
	switch code {
	case model.ResNotFound: //新用户
		name := GetPara(r, "name")
		reference := GetPara(r, "reference")
		m = model.AddNewMember(App.DB, phone, cardno, reference, "", name)
		if m == nil {
			err = errors.New("用户创建失败" + phone)
			code = model.ResFailCreateMember
		} else {
			err = nil
			code = model.ResOK
			resp.CopyMemberInfo(m)
		}
	case model.ResFound: //老用户
		var i int
		err, i = model.GetAmountByMember(App.DB, m)
		if err != nil {
			code = model.ResFail
		} else {
			resp.Amount = strconv.Itoa(i)
			resp.CopyMemberInfo(m)
		}
	}
	if err != nil { //其他错误
		fmt.Fprintf(w, resp.Message(code, err.Error()))
		return
	}
	resp.RespCode = code
	fmt.Fprintf(w, resp.String())
}
