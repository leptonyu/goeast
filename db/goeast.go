package db

import (
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
