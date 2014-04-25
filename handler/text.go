package handler

import (
	"github.com/leptonyu/goeast/data"
	"github.com/wizjin/weixin"
)

func NewTextMap(basic map[string]string) map[string]func(weixin.ResponseWriter, data.Data) {
	m := map[string]func(weixin.ResponseWriter, data.Data){}
	m["web"] = web
	m["help"] = help
	return m
}

func web(w weixin.ResponseWriter, d data.Data) {
	w.ReplyText("Welcome to " + d.Url())
}

func help(w weixin.ResponseWriter, d data.Data) {
	w.ReplyText("Usage: help | web | event")
}
