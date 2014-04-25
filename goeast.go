package goeast

import (
	"fmt"
	"github.com/leptonyu/goeast/data"
	"github.com/leptonyu/goeast/handler"
	"github.com/wizjin/weixin"
	"net/http"
	"strings"
)

func Init(basic map[string]string) {
	http.HandleFunc("/", hello)
	wechat := weixin.New(basic["token"], basic["appid"], basic["appsecret"])
	m := handler.NewTextMap(basic)
	d := data.NewData(basic)
	wechat.HandleFunc(weixin.MsgTypeText, func(w weixin.ResponseWriter, r *weixin.Request) {
		txt := r.Content // 获取用户发送的消息
		sig := strings.ToLower(txt)
		v, ok := m[sig]
		if ok {
			v(w, d)
		} else {
			w.ReplyText(txt)
		}
	})
	wechat.HandleFunc(weixin.MsgTypeEventSubscribe, Subscribe)
	http.Handle("/api", wechat)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Goeast WeChat Service is On!")
}

func Subscribe(w weixin.ResponseWriter, r *weixin.Request) {
	w.ReplyText("Welcome to goeast wechat!") // 有新人关注，返回欢迎消息
}
