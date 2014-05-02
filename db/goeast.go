package db

import (
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

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
	//url
	url = "http://www.goeastmandarin.com"
)

type Msg struct {
	Name    string
	Content string
}

func (c *DBConfig) Update(key string) (r *Msg, err error) {
	res, err := http.Get(url + key + "?format=json-pretty")
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
	_, err = c.Query(func(database *mgo.Database) (interface{}, error) {
		return database.C("web").Upsert(bson.M{"name": key}, r)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
