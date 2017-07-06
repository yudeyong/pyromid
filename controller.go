package main
import (
  "net/http"
  "fmt"
  "strconv"
  _"errors"
  _ "strings"
  "./model"
)

type StatResponse struct{
  msg string
  code string
}

func (r *StatResponse)String() string{
  return fmt.Sprintf("%s:%s", r.msg, r.code)
}

func (c *Controller) Consume(w http.ResponseWriter, r *http.Request){
  r.ParseForm()  //解析参数，默认是不会解析的
//  map :=
  var phone,cardno string
  arr := r.Form["phone"]
  if (len(arr)>0){
    phone = arr[0]
  }
  fmt.Println(phone)
  arr = r.Form["cardno"]
  if (len(arr)>0){
    cardno = arr[0]
  }
/*  for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
  */
  m := model.NewMember()
  err,code := m.FindByPhoneOrCardno(App.DB, phone, cardno)
  if code==model.ResNotFound{//新用户
    ref := model.NewMember()
    var reference string
    arr = r.Form["reference"]
    if (len(arr)>0){
      reference = arr[0]
    }
    err, code = ref.FindByInfo(App.DB, reference)
    if err==nil{
        m.Reference.Scan(ref.ID)
    }
    fmt.Fprintf(w, "新用户创建")
    if m.AddNewMember(App.DB, phone, cardno, ""){
      fmt.Fprintf(w, "新用户创建失败")
      //return;
    }
  }else  if err!=nil{//其他错误
    fmt.Fprintf(w, err.Error())
    return;
  }else//老用户
  {//found, add consume amount
    fmt.Println("old:",m.ID)
  }

  fmt.Fprintf(w ,strconv.Itoa(code),m.String()) //这个写入到w的是输出到客户端的
}

func (aa *Controller)HandleFunc(){
  fmt.Printf("b/n")
}
