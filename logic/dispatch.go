package logic

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/leptonyu/wechat"
	"github.com/leptonyu/wechat/db"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	MaxArticles = 3
	Url         = "http://www.goeastmandarin.com"
	Events      = "/events"
	Blog        = "/blog"
)

var BaseTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local).Add(8 * time.Hour)

// this struct is used for caching the GoEast site.
// Then we can speed up the responds of WeChat requests.
// There will be some goroutines used for update the cache in period time.
type Msg struct {
	Name       string    // Key of msg, list at the const in this package.
	Content    string    // Content of msg, this content is formated as json.
	ExpireTime time.Time // create time of the content.
}

func queryMsg(key string, x *db.MongoStorage) (*Msg, error) {
	msg := &Msg{}
	err := x.Query(func(m *mgo.Database) error {
		err := m.C("web").Find(bson.M{"name": key}).One(&msg)
		if err != nil || time.Since(msg.ExpireTime).Seconds() >= 0 {
			res, err := http.Get(Url + key + "?format=json-pretty")
			if err != nil {
				return err
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			msg.Name = key
			msg.Content = string(body)
			msg.ExpireTime = time.Now().Add(30 * time.Minute)
			_, err = m.C("web").Upsert(bson.M{"name": key}, msg)
		}
		return err
	})
	return msg, err
}
func Dispatch(x *db.MongoStorage) error {
	wc, err := x.GetWeChat()
	if err != nil {
		return err
	}
	wc.RegisterHandler(textDispatcher(x), wechat.MsgTypeText, wechat.MsgTypeEventClick)
	wc.RegisterHandler(func(w wechat.RespondWriter, r *wechat.Request) error {
		return help(x, w, "Help")
	}, wechat.MsgTypeEventSubscribe)

	wc.RegisterHandler(func(w wechat.RespondWriter, r *wechat.Request) error {
		return help(x, w, "Help")
	}, wechat.MsgTypeEventScan)
	return nil
}

type TextFunc func(*db.MongoStorage, wechat.RespondWriter, string) error

type TextRoute struct {
	regx   *regexp.Regexp
	handle TextFunc
}

type TextRoutes []*TextRoute

func (t *TextRoutes) Register(f TextFunc, pattern ...string) {
	for _, p := range pattern {
		r, err := regexp.Compile(p)
		if err != nil {
			panic(err)
		}
		*t = append(*t, &TextRoute{
			regx:   r,
			handle: f,
		})
	}
}

func textDispatcher(config *db.MongoStorage) wechat.HandleFunc {
	drs := &TextRoutes{}
	drs.Register(contact, `^\s*(contact(\-us)?|联系方式)\s*$`)
	drs.Register(teacher, `^\s*(maria|emily|jane)\s*$`)
	drs.Register(help, `^\s*(help|h|帮助)\s*$`)
	drs.Register(event, `^\s*(events?|事件|活动)\s*$`)
	drs.Register(blog, `^\s*(blogs?|日志|博客)\s*$`)
	drs.Register(next, `^\s*(next|下一个)\s*$`)
	drs.Register(week, `^\s*(week|本周|这周)\s*$`)
	drs.Register(home, `^\s*(home|主页|首页)\s*$`)
	return func(w wechat.RespondWriter, r *wechat.Request) error {
		txt := strings.TrimSpace(r.Content + r.EventKey)
		sig := strings.ToLower(txt)
		for _, v := range *drs {
			if v.regx.MatchString(sig) {
				err := v.handle(config, w, txt)
				if err == nil {
					return nil
				}
			}
		}
		w.ReplyText("You said: " + txt + `
Reply Help to get helps.`)
		return nil
	}
}
func home(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyNews([]wechat.Article{
		wechat.Article{
			Title:       "GoEast Language Centers",
			PicUrl:      "http://static.squarespace.com/static/52141adee4b0476aaa6af594/t/52e83ec9e4b0dfc630da5953/1390952403577/XJL_8779-small.jpg",
			Url:         Url,
			Description: `GoEast Language centers is one of Shanghai's premiere Mandarin language schools. We offer tutoring services, online classes, and on-site classes at our beautiful Shanghai campus. Our mission is to teach Mandarin and Chinese culture in a fun, personal, and effective environment. At GoEast we focus on helping you achieve your learning goals with as much personal attention as possible and no self help gimmicks! Learning language is best done with others, and we hope you'll join us at GoEast! `,
		},
	})
	return nil
}
func help(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyText(`Welcome to <a href='` + Url +
		`'>GoEast Language Centers</a>, use the following keywords to get further information:` +
		"\nHome\n  Get HomePage" +
		"\nEvent\n  Get Events" +
		"\nBlog\n  Get Blogs" +
		"\nNext\n  Get Next Event" +
		"\nWeek\n  Get Nexti 7ds Event" +
		"\nContact\n  Get Contact Info")
	return nil
}

type dua []time.Duration

func (d dua) Less(i, j int) bool {
	return d[i] < d[j]
}

func (d dua) Len() int {
	return len(d)
}
func (d dua) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func getEventArticle(event *json.Json) (wechat.Article, time.Duration) {
	s, e := time.Duration(event.Get("startDate").MustInt64()), time.Duration(event.Get("endDate").MustInt64())
	ts, te := BaseTime.Add(s*time.Millisecond), BaseTime.Add(e*time.Millisecond)
	desp := ts.Format(`Monday, Jan _2, 2006
  03:04pm`) + " - " + te.Format(`03:04pm`)
	return wechat.Article{
		Url:         Url + event.Get("fullUrl").MustString(),
		Title:       event.Get("title").MustString(),
		PicUrl:      event.Get("assetUrl").MustString(),
		Description: desp,
	}, s
}

//Handle event request.
func handleEvent(c *db.MongoStorage,
	w wechat.RespondWriter,
	txt string,
	check func(time.Time) bool,
	errstr string, max int) error {
	deferr := errors.New(errstr)
	msg, err := queryMsg(Events, c)
	if err != nil {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	mapa := map[time.Duration]wechat.Article{}
	past := j.Get("upcoming")
	for i, _ := range past.MustArray() {
		a, f := getEventArticle(past.GetIndex(i))
		if check(BaseTime.Add(f * time.Millisecond)) {
			mapa[f] = a
		}
	}
	if len(mapa) == 0 {
		log.Println(err)
		w.ReplyText(deferr.Error())
		return nil
	}
	var keys dua
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	res := []wechat.Article{}
	if len(keys) > max {
		keys = keys[:max]
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	w.ReplyNews(res)
	return nil
}

//Upcoming event, only list 3 events.
func event(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return true
	}, `Oops! No event recently!`, MaxArticles)
}

//Next event.
func next(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return true
	}, `Oops! No next event!`, 1)
}

//This week's events.
func week(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	return handleEvent(c, w, txt, func(t time.Time) bool {
		return time.Since(t).Hours() >= -7*24
	}, `Oops! No event in next 7 days!`, MaxArticles)
}

//Recent Blogs.
func blog(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	deferr := errors.New(`Oops! No blog recently!`)
	msg, err := queryMsg(Blog, c)
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		w.ReplyText(deferr.Error())
		return nil
	}
	mapa := map[time.Duration]wechat.Article{}
	past := j.Get("items")
	for i, _ := range past.MustArray() {
		event := past.GetIndex(i)
		a := wechat.Article{
			Url:         Url + event.Get("fullUrl").MustString(),
			Title:       event.Get("title").MustString(),
			PicUrl:      event.Get("assetUrl").MustString(),
			Description: event.Get("excerpt").MustString(),
		}
		t := time.Duration(event.Get("addedOn").MustInt64())
		mapa[t] = a
	}
	if len(mapa) == 0 {
		w.ReplyText(deferr.Error())
		return nil
	}
	var keys dua
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	res := []wechat.Article{}
	if len(keys) > MaxArticles {
		keys = keys[len(keys)-MaxArticles:]
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

var (
	maria = teacherinfo{
		Id:      "maria",
		Name:    "Maria MAO",
		Chinese: "毛瑞",
		Phone:   "(86)18192201219",
		Email:   "mariamao@goeast.cn",
		Skype:   "maoruimaria",
	}
	emily = teacherinfo{
		Id:      "emily",
		Name:    "Emily WANG",
		Chinese: "王蓉",
		Phone:   "(86)18016005118",
		Email:   "emilywang@goeast.cn",
		Skype:   "rongni_123",
	}
	jane = teacherinfo{
		Id:      "jane",
		Name:    "Jane LUO",
		Chinese: "罗琼",
		Phone:   "(86)13916723393",
		Email:   "janeluo@goeast.cn",
		Skype:   "jane.qiongluo",
	}
)

type teacherinfo struct {
	Id      string
	Name    string
	Chinese string
	Phone   string
	Email   string
	Skype   string
}

func teacher(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	switch strings.ToLower(txt) {
	case maria.Id:
		return t(w, maria)
	case emily.Id:
		return t(w, emily)
	case jane.Id:
		return t(w, jane)
	default:
		return errors.New("Not found!")
	}
}
func t(w wechat.RespondWriter, tt teacherinfo) error {
	w.ReplyText(fmt.Sprintf(`%v (%v)
Teacher & Consultant
GoEast Language Center

Telephone: 
  %v

Skype:
  %v

Email: 
%v`, tt.Name, tt.Chinese, tt.Phone, tt.Skype, tt.Email))
	return nil
}

func contact(c *db.MongoStorage, w wechat.RespondWriter, txt string) error {
	w.ReplyText(`GoEast Language Center

No 194-196 Zhengmin Road, Yangpu District, Shanghai, China

Telephone: 
  86-21-31326611  

  Email: 
  coursecenter@goeast.cn`)
	return nil
}
