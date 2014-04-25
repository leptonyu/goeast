package goeast

import (
<<<<<<< HEAD
	"github.com/leptonyu/goeast/data"
	"github.com/leptonyu/goeast/handler"
	"github.com/wizjin/weixin"
	"html/template"
=======
	"fmt"
	"github.com/leptonyu/goeast/data"
	"github.com/leptonyu/goeast/handler"
	"github.com/wizjin/weixin"
>>>>>>> a8044706a05206398f3a6e1a31b307f6049ba8e9
	"net/http"
	"strings"
)

type Info struct {
	URL string
}

func Init(basic map[string]string) {
<<<<<<< HEAD
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		e := Info{URL: basic["url"]}
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, e)
	})
=======
	http.HandleFunc("/", hello)
>>>>>>> a8044706a05206398f3a6e1a31b307f6049ba8e9
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

<<<<<<< HEAD
=======
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Goeast WeChat Service is On!")
}

>>>>>>> a8044706a05206398f3a6e1a31b307f6049ba8e9
func Subscribe(w weixin.ResponseWriter, r *weixin.Request) {
	w.ReplyText("Welcome to goeast wechat!") // 有新人关注，返回欢迎消息
}
