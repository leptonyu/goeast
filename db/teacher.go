package db

import (
	"errors"
	"fmt"
	"github.com/leptonyu/goeast/wechat"
	"strings"
)

var (
	maria = teacher{
		Id:      "maria",
		Name:    "Maria MAO",
		Chinese: "毛瑞",
		Phone:   "(86)18192201219",
		Email:   "mariamao@goeast.cn",
		Skype:   "maoruimaria",
	}
	emily = teacher{
		Id:      "emily",
		Name:    "Emily WANG",
		Chinese: "王蓉",
		Phone:   "(86)18016005118",
		Email:   "emilywang@goeast.cn",
		Skype:   "rongni_123",
	}
	jane = teacher{
		Id:      "jane",
		Name:    "Jane LUO",
		Chinese: "罗琼",
		Phone:   "(86)13916723393",
		Email:   "janeluo@goeast.cn",
		Skype:   "jane.qiongluo",
	}
	contact   = "contact"
	contactus = contact + "-us"
)

type teacher struct {
	Id      string
	Name    string
	Chinese string
	Phone   string
	Email   string
	Skype   string
}

func Teacher() func(*DBConfig, wechat.ResponseWriter, *wechat.Request) error {
	return func(c *DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
		a := strings.TrimSpace(strings.ToLower(r.Content + r.EventKey))
		switch a {
		case maria.Id:
			return t(w, maria)
		case emily.Id:
			return t(w, emily)
		case jane.Id:
			return t(w, jane)
		case contact:
			return cc(w)
		case contactus:
			return cc(w)
		default:
			return errors.New("Not found!")
		}
	}
}

func t(w wechat.ResponseWriter, tt teacher) error {
	w.ReplyText(fmt.Sprintf(`%v (%v)
Teacher & Consultant
GoEast Language Center
Telephone: 
%v
Skype:
  %v
Email: 
%v`, tt.Name, tt.Chinese, tt.Phone, tt.Skype, tt.Email))
	return nil
}

func cc(w wechat.ResponseWriter) error {
	w.ReplyText(`GoEast Language Centers
No 194-196 Zhengmin Road
Yangpu District, Shanghai
China
Telephone: 
  86-21-31326611  
Email: 
coursecenter@goeast.cn`)
	return nil
}
