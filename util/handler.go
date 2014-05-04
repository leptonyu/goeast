package util

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/leptonyu/goeast/db"
	"github.com/leptonyu/goeast/wechat"
	"log"
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
	drs.register(`^\s*(help|h|帮助)\s*$`, help)
	drs.register(`^\s*(events?|事件|活动)\s*$`, event)
	drs.register(`^\s*(blogs?|日志|博客)\s*$`, blog)
	drs.register(`^\s*(home|主页|首页)\s*$`, home)
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		sig := strings.ToLower(txt)
		log.Println(r)
		for _, v := range drs.rs {
			if v.regx.MatchString(sig) {
				return v.handler(config, w, r)
			}
		}
		w.ReplyText(txt)
		return nil
	})
}

func home(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyText(`<a href='` + db.Url + `'>GoEast Language Centers</a>`)
	return nil
}

func help(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyText(`Welcome to <a href='` + db.Url + `'>GoEast Language Centers</a>, use the following keywords to get information:
Home
	Get Homepage
Help
	Get Helps
Event
	Get Events
Blog
	Get Blogs`)
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
	go func() {
		for {
			_, err := config.WC.UpdateAccessToken()
			if err != nil {
				log.Fatalln(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	f := func(wait time.Duration, keys ...string) {
		if len(keys) > 0 {
			m := map[string]db.Msg{}
			for _, v := range keys {
				r, err := config.QueryMsg(v)
				if err == nil {
					m[v] = *r
				}
			}
			for {
				time.Sleep(wait)
				for _, v := range keys {
					msg, ok := m[v]
					if ok && time.Since(msg.CreateTime.Add(wait)).Seconds() >= 0 {
						fmt.Println(v)
						config.UpdateMsg(v)
					}
				}
			}
		}
	}
	go f(30*time.Minute, db.Blog, db.Events)
	go f(24*time.Hour,
		db.Home,
		db.Campus,
		db.Contact,
		db.Galleries,
		db.One2one,
		db.Online,
		db.Onsite,
		db.Teachers,
		db.Testimonials)
}
