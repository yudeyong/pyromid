package main
import (
  "net/http"
  "fmt"
  "strconv"
  "log"
  "errors"
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

func newMember( phone string, cardno string, reference string, level string) *model.Member{
  ref := model.NewMember()
  var reference_id string

  if (len(reference)>0){
    err, _ := ref.FindByInfo(App.DB, reference)
    if err==nil{
        reference_id = ref.ID
    }
  }
  if err:=ref.AddNewMember(App.DB, phone, cardno, reference_id, level);err!=nil{
    log.Println(err, phone, cardno)
    return nil;
  }
  return ref
}

func (c *Controller) Consume(w http.ResponseWriter, r *http.Request){
  r.ParseForm()  //解析参数，默认是不会解析的
//  map :=
  var phone,cardno,reference string
  arr := r.Form["phone"]
  if (len(arr)>0){
    phone = arr[0]
  }
  //fmt.Println(phone)
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
  //fmt.Println(err,code)
  if code==model.ResNotFound{//新用户
    arr = r.Form["reference"]
    if (len(arr)>0){
      reference = arr[0]
    }
    m = newMember(phone, cardno, reference, "")
    if (m==nil){
      err = errors.New("用户创建失败"+phone)
    }else
    {
      err = nil
    }
  }
  if err!=nil{//其他错误
    fmt.Fprintf(w, err.Error())
    return;
  }
  fmt.Println("uid:",m.ID)

  fmt.Fprintf(w ,strconv.Itoa(code),m.String())
}

func (aa *Controller)HandleFunc(){
  fmt.Printf("b/n")
}
