//Copyright 2014 leptonyu. All rights reserved. Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
/*
This package is used to connect to the database.
*/
package db

import (
	"encoding/json"
	"github.com/leptonyu/goeast/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

//MongoDB configuration, contains a wechat object.
type DBConfig struct {
	Prefix string         //Mongodb database prefix name
	DBName string         //Mongodb database name
	DBUrl  string         //Mangodb connect url: mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]]
	WC     *wechat.WeChat //WeChat SDK object.
	admins []*User
}

//MongoDB configuration with specific username and password.
func NewDBConfigWithUser(dbname,
	host,
	username,
	password string) (*DBConfig, error) {
	c := &DBConfig{
		Prefix: "wechat_",
		DBName: dbname,
		DBUrl:  "mongodb://" + username + ":" + password + "@" + host,
	}
	return c.init()
}

//MongoDB configuration with default username and password to connect to the localhost DB.
func NewDBConfig(dbname string) (*DBConfig, error) {
	c := &DBConfig{
		Prefix: "wechat_",
		DBName: dbname,
		DBUrl:  "mongodb://localhost",
	}
	return c.init()
}

// Init the DBConfig, which creates the WeChat SDK object.
func (c *DBConfig) init() (*DBConfig, error) {
	_, err := c.CreateWeChat()
	if err == nil {
		c.admins = c.FindAdmins()
	}
	return c, err
}

// Query Database function, all single operations of MongoDB should implement this function.
type QueryFunc func(*mgo.Database) (interface{}, error)

// One operation on MongoDB, this is final method used for operate the database.
func (c *DBConfig) Query(f QueryFunc) (data interface{}, err error) {
	session, err := mgo.Dial(c.DBUrl)
	if err != nil {
		return
	}
	defer session.Close()
	return f(session.DB(c.Prefix + c.DBName))
}

type property struct {
	_id   bson.ObjectId
	Name  string
	Value string
}

// Read wechat.AccessToken from MongoDB.
// If not found, it will return error.
func (c *DBConfig) Read() (*wechat.AccessToken, error) {
	x, err := c.Query(func(dbm *mgo.Database) (data interface{}, err error) {
		c := dbm.C("wechat")
		v := &property{}
		err = c.Find(bson.M{"name": "accesstoken"}).One(v)
		if err != nil {
			return
		}
		data = &wechat.AccessToken{}
		err = json.Unmarshal([]byte(v.Value), data)
		return
	})
	if err != nil {
		return nil, err
	}
	return x.(*wechat.AccessToken), nil
}

func (c *DBConfig) Save(msg wechat.Request) error {
	c.Log(&msg)
	return nil
}

// Write wechat.AccessToken into MongoDB.
func (c *DBConfig) Write(at *wechat.AccessToken) error {
	bs, err := json.Marshal(at)
	if err != nil {
		return err
	}
	_, err = c.Query(func(dbm *mgo.Database) (data interface{}, err error) {
		c := dbm.C("wechat")
		v := &property{}
		v.Name = "accesstoken"
		v.Value = string(bs)
		data, err = c.Upsert(bson.M{"name": "accesstoken"}, v)
		return
	})
	return err

}

type storeWeChat struct {
	Name   string
	Appid  string
	Secret string
	Token  string
}

//Create WeChat Object from DBConfig.
func (c *DBConfig) CreateWeChat() (*wechat.WeChat, error) {
	xx := storeWeChat{}
	_, err := c.Query(func(database *mgo.Database) (interface{}, error) {
		err := database.C("wechat").Find(bson.M{"name": "wechat"}).One(&xx)
		return xx, err
	})
	if err != nil {
		return nil, err
	}
	wc, err := wechat.New(xx.Appid, xx.Secret, xx.Token, c)
	if err != nil {
		return nil, err
	}
	c.WC = wc
	return wc, err
}

// Initialize the WeChat SDK, write the basic information into MongoDB.
// Such as appid, appsecret, apitoken, etc.
// After invoking this method, WeChat SDK object can work.
func (c *DBConfig) Init(appid, secret, token string) error {
	xx := storeWeChat{}
	xx.Name = "wechat"
	xx.Appid = appid
	xx.Secret = secret
	xx.Token = token
	_, err := c.Query(func(database *mgo.Database) (interface{}, error) {
		return database.C("wechat").Upsert(bson.M{"name": "wechat"}, xx)
	})
	if err == nil {
		log.Println("Set appid as " + appid)
		log.Println("Set secret as " + secret)
		log.Println("Set token as " + token)
	}
	return err
}
