package data

import (
	"errors"
	json "github.com/bitly/go-simplejson"
	"github.com/wizjin/weixin"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"sort"
	"time"
)

type Config struct {
	Name    string
	Basic   *Info
	Session *mgo.Session
	close   bool
}

const (
	//Home
	Home = `/`
	//Contact
	Contact = `/contact-us`
	//AboutUs
	Campus       = "/campus"
	Teachers     = "/our-teachers"
	Galleries    = "/galleries"
	Testimonials = "/testimonials"
	//Courses
	One2one = "/1-on-1-tutoring"
	Online  = "/online-courses"
	Onsite  = "/on-site-courses"
	//Blog
	Blog = "/blog"
	//Events
	Events = "/events"
)

type Info struct {
	Key    string `bson:"key"`
	Name   string `bson:"name"`
	Url    string `bson:"url"`
	Token  string `bson:"token"`
	Appid  string `bson:"appid"`
	Secret string `bson:"secret"`
}

func NewConfig(dbname string) (c *Config, err error) {
	sess, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		return
	}
	sess.SetSafe(&mgo.Safe{})
	coll := sess.DB(dbname).C("config")
	r := Info{}
	coll.Find(bson.M{"key": "config"}).One(&r)
	c = &Config{
		Name:    dbname,
		Basic:   &r,
		Session: sess,
		close:   false,
	}
	return
}

func (c *Config) Close() {
	c.Session.Close()
	c.close = true
}

type Msg struct {
	Name    string `bson:"name"`
	Content string `bson:cont`
}

func (c *Config) fetch(key string) (r *Msg, err error) {
	mes := c.Session.DB(c.Name).C("message")
	r = &Msg{}
	err = mes.Find(bson.M{"name": key}).One(&r)
	if err != nil {
		r, err = c.update(key)
	}
	return
}

func (c *Config) Interval(wait time.Duration, keys ...string) {
	if len(keys) > 0 {
		go func() {
			for !c.close {
				time.Sleep(wait)
				for _, v := range keys {
					c.update(v)
				}
			}
		}()
	}
}

func (c *Config) update(key string) (r *Msg, err error) {
	res, err := http.Get(c.Basic.Url + key + "?format=json-pretty")
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	r = &Msg{Name: key,
		Content: string(body)}
	mes := c.Session.DB(c.Name).C("message")
	err = mes.Insert(r)
	return
}

func (c *Config) Save(r *weixin.Request) {
	coll := c.Session.DB(c.Name).C("user")
	coll.Insert(r)
}

func (c *Config) FetchKeys(key string) (res []weixin.Article, err error) {
	switch key {
	case Events:
		return c.fetchk(key, 3, true, event)
	case Blog:
		return c.fetchk(key, 3, false, blog)
	default:
		err = errors.New("No key " + key + " registed")
		return
	}

}

func (c *Config) fetchk(key string,
	max int,
	fromSmallToLarge bool,
	handler func(map[int]weixin.Article,
		*json.Json,
		*Config) (err error)) (res []weixin.Article, err error) {
	msg, err := c.fetch(key)
	if err != nil {
		return
	}
	j, err := json.NewJson([]byte(msg.Content))
	if err != nil {
		return
	}
	mapa := map[int]weixin.Article{}
	err = handler(mapa, j, c)
	if err != nil {
		return
	}
	if len(mapa) == 0 {
		err = errors.New("no new events")
		return
	}
	var keys []int
	for k, _ := range mapa {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	res = []weixin.Article{}
	if len(keys) <= max {
	} else if fromSmallToLarge {
		keys = keys[:max]
	} else {
		keys = keys[len(keys)-max:]
	}
	if !fromSmallToLarge {
		n := len(keys)
		for i := 0; i < n/2; i++ {
			keys[i], keys[n-1-i] = keys[n-1-i], keys[i]
		}
	}
	for _, k := range keys {
		res = append(res, mapa[k])
	}
	return
}

func event(mapa map[int]weixin.Article, j *json.Json, c *Config) (err error) {
	past := j.Get("upcoming")
	for i, _ := range past.MustArray() {
		event := past.GetIndex(i)
		a := weixin.Article{
			Url:         c.Basic.Url + event.Get("fullUrl").MustString(),
			Title:       event.Get("title").MustString(),
			PicUrl:      event.Get("assetUrl").MustString(),
			Description: event.Get("excerpt").MustString(),
		}
		t := event.Get("startDate").MustInt()
		mapa[t] = a
	}
	return
}

func blog(mapa map[int]weixin.Article, j *json.Json, c *Config) (err error) {
	past := j.Get("items")
	for i, _ := range past.MustArray() {
		event := past.GetIndex(i)
		a := weixin.Article{
			Url:         c.Basic.Url + event.Get("fullUrl").MustString(),
			Title:       event.Get("title").MustString(),
			PicUrl:      event.Get("assetUrl").MustString(),
			Description: event.Get("excerpt").MustString(),
		}
		t := event.Get("addedOn").MustInt()
		mapa[t] = a
	}
	return
}
