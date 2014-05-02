/*
This package is used to connect to the database.
*/
package db

import (
	"encoding/json"
	"github.com/leptonyu/goeast/wechat"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

//MongoDB configuration
type DBConfig struct {
	DBName string //Mongodb database name
	DBUrl  string //Mangodb connect url: mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]]
}

//
func NewDBConfigWithUser(dbname, host, username, password string) *DBConfig {
	return &DBConfig{
		DBName: dbname,
		DBUrl:  "mongodb://" + username + ":" + password + "@" + host,
	}
}

func NewDBConfig(dbname string) *DBConfig {
	return &DBConfig{
		DBName: dbname,
		DBUrl:  "mongodb://localhost",
	}
}

// One query on one connection.
func (c *DBConfig) Query(f func(*mgo.Database) (interface{}, error)) (data interface{}, err error) {
	session, err := mgo.Dial(c.DBUrl)
	if err != nil {
		return
	}
	defer session.Close()
	return f(session.DB(c.DBName))
}

type property struct {
	_id   bson.ObjectId
	Name  string
	Value string
}

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
