package util

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"time"
)

type dispatchRoute struct {
	regx    *regexp.Regexp
	handler RouteFunc
}

type RouteFunc func(*db.DBConfig, wechat.ResponseWriter, *wechat.Request) error

type routes struct {
	rs []*dispatchRoute
}

// Register Dispatch handler
func DispatchFunc(config *db.DBConfig, wc *wechat.WeChat) {
	//Text request
	drs := &routes{}
	drs.register(`^(admin|adm|管理员)(:(help|account))?$`, db.Admin())
	drs.register(`^\s*(contact-us|maria|emily|jane)\s*$`, db.Teacher())
	drs.register(`^\s*(help|h|帮助)\s*$`, help)
	drs.register(`^\s*(events?|事件|活动)\s*$`, event)
	drs.register(`^\s*(blogs?|日志|博客)\s*$`, blog)
	drs.register(`^\s*(home|主页|首页)\s*$`, home)
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		sig := strings.ToLower(txt)
		for _, v := range drs.rs {
			if v.regx.MatchString(sig) {
				err := v.handler(config, w, r)
				if err == nil {
					return nil
				}
			}
		}
		w.ReplyText(txt)
		return nil
	})
	wc.HandleFunc(wechat.MsgTypeEventSubscribe, func(w wechat.ResponseWriter, r *wechat.Request) error {
		return help(config, w, r)
	})
	wc.HandleFunc(wechat.MsgTypeEventClick, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.EventKey
		sig := strings.ToLower(txt)
		for _, v := range drs.rs {
			if v.regx.MatchString(sig) {
				err := v.handler(config, w, r)
				if err == nil {
					return nil
				}
			}
		}
		w.ReplyText("Event key has no respond " + txt)
		return nil
	})
}

func home(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyNews([]wechat.Article{
		wechat.Article{
			Title:       "GoEast Language Centers",
			PicUrl:      "http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52e83ec9e4b0dfc630da5953/1390952403577/XJL_8779-small.jpg",
			Url:         db.Url,
			Description: `GoEast Language centers is one of Shanghai's premiere Mandarin language schools. We offer tutoring services, online classes, and on-site classes at our beautiful Shanghai campus. Our mission is to teach Mandarin and Chinese culture in a fun, personal, and effective environment. At GoEast we focus on helping you achieve your learning goals with as much personal attention as possible and no self help gimmicks! Learning language is best done with others, and we hope you'll join us at GoEast! `,
		},
	})
	return nil
}

func help(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	adm := ``
	ad, err := c.GetAdmin(r.FromUserName)
	if err == nil {
		adm = fmt.Sprintf(`
Hello %v, you can use following command:
adm:help
	Get Admin Helps
adm:test
	This is test
`, ad.Username)
	}
	w.ReplyText(`Welcome to <a href='` + db.Url + `'>GoEast Language Centers</a>, use the following keywords to get information:
Home
	Get Homepage
Help
	Get Helps
Event
	Get Events
Blog
	Get Blogs` + adm)
	return nil
}

func event(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	msg, err := c.QueryMsg(db.Events)
	if err != nil {
		w.ReplyText("Sorry, No event found!")
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		w.ReplyText(err.Error())
		return nil
	}
	mapa := map[int]wechat.Article{}
	past := j.Get("upcoming")
	for i, _ := range past.MustArray() {
		event := past.GetIndex(i)
		a := wechat.Article{
			Url:         db.Url + event.Get("fullUrl").MustString(),
			Title:       event.Get("title").MustString(),
			PicUrl:      event.Get("assetUrl").MustString(),
			Description: event.Get("excerpt").MustString(),
		}
		t := event.Get("startDate").MustInt()
		mapa[t] = a
	}
	if len(mapa) == 0 {
		return errors.New("no new events")
	}
	var keys []int
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	res := []wechat.Article{}
	if len(keys) > db.MaxArticles {
		keys = keys[len(keys)-db.MaxArticles:]
	}
	n := len(keys)
	for i := 0; i < n/2; i++ {
		keys[i], keys[n-1-i] = keys[n-1-i], keys[i]
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	w.ReplyNews(res)
	return nil
}

func blog(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyText("This is blog!")
	return nil
}

func (rs *routes) register(pattern string, handler RouteFunc) error {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rs.rs = append(rs.rs, &dispatchRoute{regx: r, handler: handler})
	return nil
}

func DeamonTask(config *db.DBConfig) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	shutdown := false
	go func() {
		for !shutdown {
			_, err := config.WC.UpdateAccessToken()
			if err != nil {
				log.Fatalln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	go func() {
		<-c
		shutdown = true
	}()
}
