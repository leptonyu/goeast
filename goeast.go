package goeast

import (
	wc "github.com/wizjin/weixin"
	"net/http"
)

func Init(basic map[string]string) {
	wechat := wc.New(basic["token"], basic["appid"], basic["appsecret"])
	wechat.HandleFunc(wc.MsgTypeText, Echo)
	http.Handle("/api", wechat)
}

func Echo(w wc.ResponseWriter, r *wc.Request) {
	txt := r.Content // 获取用户发送的消息
	w.ReplyText(txt) // 回复一条文本消息
	w.PostText(txt)  // 发送一条文本消息
}
