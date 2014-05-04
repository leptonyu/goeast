package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leptonyu/goeast/wechat"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

const (
	//Base URL
	Url = `http://www.goeastmandarin.com`
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
	//url
	url = "http://www.goeastmandarin.com"
	//
	MaxArticles = 3
)

// this struct is used for caching the GoEast site.
// Then we can speed up the responds of WeChat requests.
// There will be some goroutines used for update the cache in period time.
type Msg struct {
	Name       string    // Key of msg, list at the const in this package.
	Content    string    // Content of msg, this content is formated as json.
	CreateTime time.Time // create time of the content.
}

// Query the specific Msg by key, such as
/*
	r := c.QueryMsg(db.Events)
*/
func (c *DBConfig) QueryMsg(key string) (r *Msg, err error) {
	r = &Msg{}
	_, err = c.Query(func(database *mgo.Database) (interface{}, error) {
		err := database.C("web").Find(bson.M{"name": key}).One(&r)
		return r, err
	})
	if err != nil {
		return c.UpdateMsg(key)
	}
	return r, nil
}

// Update Msg into database.
// If the Msg with key Msg.Name does not exist, then it will create a new one.
func (c *DBConfig) UpdateMsg(key string) (r *Msg, err error) {
	res, err := http.Get(url + key + "?format=json-pretty")
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	r = &Msg{
		Name:       key,
		Content:    string(body),
		CreateTime: time.Now(),
	}
	_, err = c.Query(func(database *mgo.Database) (interface{}, error) {
		return database.C("web").Upsert(bson.M{"name": key}, r)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

type requestLog struct {
	From   string
	To     string
	Create int
	Id     int64
	Type   string
	Value  string
}

func (c *DBConfig) Log(r *wechat.Request) {
	bs, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
		return
	}
	rl := &requestLog{
		From:   r.FromUserName,
		To:     r.ToUserName,
		Create: r.CreateTime,
		Id:     r.MsgId,
		Type:   r.MsgType,
		Value:  string(bs),
	}
	c.Query(func(d *mgo.Database) (interface{}, error) {
		d.C("userinfo").Upsert(bson.M{
			"from": rl.From,
			"to":   rl.To,
			"type": rl.Type,
			"id":   rl.Id,
		}, rl)
		return nil, nil
	})
}

func (c *DBConfig) QueryLog(username, typename string) ([]*wechat.Request, error) {
	if username == "" && typename == "" {
		return nil, errors.New("Pamater invalid")
	} else if username == "" {
		rs := []*wechat.Request{}
		_, err := c.Query(func(d *mgo.Database) (interface{}, error) {
			return nil,
				d.C("userinfo").Find(bson.M{"type": typename}).All(&rs)
		})
		return rs, err
	} else if typename == "" {
		rs := []*wechat.Request{}
		_, err := c.Query(func(d *mgo.Database) (interface{}, error) {
			return nil, d.C("userinfo").Find(bson.M{"from": username}).All(&rs)
		})
		return rs, err
	} else {
		rs := []*wechat.Request{}
		_, err := c.Query(func(d *mgo.Database) (interface{}, error) {
			return nil, d.C("userinfo").Find(bson.M{"from": username, "type": typename}).All(&rs)
		})
		return rs, err
	}
}

type User struct {
	Id       string
	Username string
	Admin    bool
}

func (c *DBConfig) IsAdmin(id string) bool {
	if id == "" {
		return false
	}
	for _, user := range c.admins {
		if user.Id == id {
			return true
		}
	}
	return false
}

func (c *DBConfig) GetAdmin(id string) (*User, error) {
	if id == "" {
		return nil, errors.New("Speficy the admin id")
	}
	for _, user := range c.admins {
		if user.Id == id {
			return user, nil
		}
	}
	return nil, errors.New("User with id " + id + " is not admin!")
}
func (c *DBConfig) FindAdmins() []*User {
	user := []*User{}
	_, err := c.Query(func(d *mgo.Database) (interface{}, error) {
		return nil, d.C("user").Find(bson.M{"admin": true}).All(&user)
	})
	if err != nil {
		return []*User{}
	}
	return user
}

func (c *DBConfig) Upsert(user *User) error {
	_, err := c.Query(func(d *mgo.Database) (interface{}, error) {
		return d.C("user").Upsert(bson.M{"id": user.Id}, user)
	})
	if err == nil {
		c.admins = c.FindAdmins()
	}
	return err
}

func (c *DBConfig) UpsertWithUser(id, username string, isAdmin bool) error {
	return c.Upsert(&User{
		Id:       id,
		Username: username,
		Admin:    isAdmin,
	})
}

func Admin() func(*DBConfig, wechat.ResponseWriter, *wechat.Request) error {
	return func(c *DBConfig, w wechat.ResponseWriter, r *wechat.Request) error {
		for _, v := range c.admins {
			if r.FromUserName == v.Id {
				w.ReplyText(fmt.Sprintf(`Hello %v, you are the admin.
adm:help
	Get Admin Help
adm:status
	Get Admin Status
`, v.Username))
				return nil
			}
		}
		return errors.New("You are not administrator!")
	}
}
