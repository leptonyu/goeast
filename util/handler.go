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
	drs.register(`^\s*(help|h|帮助)\s*$`, help)
	drs.register(`^\s*(events?|事件|活动)\s*$`, event)
	drs.register(`^\s*(blogs?|日志|博客)\s*$`, blog)
	drs.register(`^\s*(home|主页|首页)\s*$`, home)
	wc.HandleFunc(wechat.MsgTypeText, func(w wechat.ResponseWriter, r *wechat.Request) error {
		txt := r.Content
		sig := strings.ToLower(txt)
		go config.Log(r)
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
}

func home(c *db.DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
	w.ReplyText(`<a href='` + db.Url + `'>GoEast Language Centers</a>`)
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
	f := func(wait time.Duration, keys ...string) (*time.Timer, error) {
		if len(keys) > 0 {
			m := map[string]db.Msg{}
			for _, v := range keys {
				r, err := config.QueryMsg(v)
				if err == nil {
					m[v] = *r
				}
			}
			return time.AfterFunc(10*time.Second, func() {
				for _, v := range keys {
					msg, ok := m[v]
					if ok && time.Since(msg.CreateTime.Add(wait)).Seconds() >= 0 {
						fmt.Println(v)
						config.UpdateMsg(v)
					}
				}
			}), nil
		}
		return nil, errors.New("No keys")
	}
	t1, err1 := f(30*time.Minute, db.Blog, db.Events)
	t2, err2 := f(24*time.Hour,
		db.Home,
		db.Campus,
		db.Contact,
		db.Galleries,
		db.One2one,
		db.Online,
		db.Onsite,
		db.Teachers,
		db.Testimonials)
	go func() {
		<-c
		shutdown = true
		if err1 == nil {
			t1.Stop()
		}
		if err2 == nil {
			t2.Stop()
		}
	}()
}
