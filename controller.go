package main
import (
  "net/http"
  "fmt"
  _ "strings"
)

type StatResponse struct{
  msg string
  code string
}

func (r *StatResponse)string() string{
  return fmt.Sprintf("%s:%s", r.msg, r.code)
}
func (c *Controller) Consume(w http.ResponseWriter, r *http.Request){
  r.ParseForm()  //解析参数，默认是不会解析的
//  map :=
  phone := r.Form["phone"]
  cardno := r.Form["cardno"]
/*  for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
  */
  var res StatResponse
  if phone==nil && cardno==nil {
    res.msg = "请输入手机号或卡号"
  }
  fmt.Fprintf(w, res.string()) //这个写入到w的是输出到客户端的
}

func (aa *Controller)HandleFunc(){
  fmt.Printf("b/n")
}
