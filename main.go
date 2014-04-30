package main

import (
	"flag"
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

func main() {
	port := flag.Int("port", 8080, "Web service port")
	help := flag.Bool("help", false, "Help")
	flag.Parse()
	if *help {
		fmt.Println("This is help")
		os.Exit(0)
	}
	startWeb(*port)
}
func startWeb(port int) {
	m := martini.Classic()
	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		t := template.New("404.tpl")
		t, _ = t.ParseFiles("templates/404.html")
		t.Execute(w, m)
	})
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.New("index.tpl")
		t, _ = t.ParseFiles("templates/index.html")
		t.Execute(w, m)
	})
	//Create new configuration
	c, err := data.NewConfig()
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
	//Create api route
	m.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		wx.ServeHTTP(w, r)
	})
	m.Post("/api", func(w http.ResponseWriter, r *http.Request) {
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
