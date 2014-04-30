package util

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/leptonyu/goeast/data"
	"github.com/leptonyu/goeast/handler"
	"github.com/wizjin/weixin"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func Template(key string, m interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New(key)
		t, _ = t.ParseFiles("templates/" + key)
		t.Execute(w, m)
	}
}

func StartWeb(port int, dbname string, api string) {
	m := martini.Classic()
	m.NotFound(Template("404.tpl", m))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	m.Get("/", Template("index.tpl", m))

	//Create new configuration
	c, err := data.NewConfig(dbname, api)
	if port == 8080 {
		c.Interval(30*time.Minute, data.Blog, data.Events)
		c.Interval(24*time.Hour,
			data.Home,
			data.Campus,
			data.Contact,
			data.Galleries,
			data.One2one,
			data.Online,
			data.Onsite,
			data.Teachers,
			data.Testimonials)
	}
	if err != nil {
		panic(err)
	}
	b := c.Basic
	//Create text request handler
	mr := handler.NewTextMap()
	wx := weixin.New(b.Token, b.Appid, b.Secret)
	wx.HandleFunc(weixin.MsgTypeText, func(w weixin.ResponseWriter, r *weixin.Request) {
		txt := r.Content
		sig := strings.ToLower(txt)
		go c.Save(r)
		v, ok := mr[sig]
		if ok {
			v(w, c)
		} else {
			w.ReplyText(txt)
		}
	})
	wx.HandleFunc(weixin.MsgTypeEventSubscribe, func(w weixin.ResponseWriter, r *weixin.Request) {
	})
	//Create api route
	m.Get("/"+c.Api, func(w http.ResponseWriter, r *http.Request) {
		wx.ServeHTTP(w, r)
	})
	m.Post("/"+c.Api, func(w http.ResponseWriter, r *http.Request) {
		wx.ServeHTTP(w, r)
	})
	err = http.ListenAndServe(":"+strconv.Itoa(port), m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("stop!")
	c.Session.Close()
	os.Exit(0)
}
