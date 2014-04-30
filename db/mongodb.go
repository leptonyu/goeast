/*
This package is used to connect to the database.
*/
package db

import (
	"labix.org/v2/mgo"
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
