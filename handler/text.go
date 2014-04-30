package handler

import (
	"fmt"
	"github.com/leptonyu/goeast/data"
	"github.com/wizjin/weixin"
)

func NewTextMap() map[string]func(weixin.ResponseWriter, *data.Config) {
	m := map[string]func(weixin.ResponseWriter, *data.Config){}
	m["help"] = help
	m["event"] = handle(data.Events, "event")
	m["events"] = handle(data.Events, "event")
	m["blog"] = handle(data.Blog, "article")
	return m
}

func help(w weixin.ResponseWriter, d *data.Config) {
	w.ReplyText(fmt.Sprintf(`Welcome to <a href="%s">%s</a>`+
		"\nReply \"Event\" to get recent events\n"+
		"Reply \"Help\" to get help", d.Basic.Url, d.Basic.Name))
}

func handle(key, name string) func(w weixin.ResponseWriter, d *data.Config) {
	return func(w weixin.ResponseWriter, d *data.Config) {
		articles, err := d.FetchKeys(key)
		if err != nil {
			w.ReplyText("Sorry, GoEast has no " + name + " recently!")
		} else {
			w.ReplyNews(articles)
		}
	}
}
