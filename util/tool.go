//Copyright 2014 leptonyu. All rights reserved. Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
/*
This package is used for create a web service.
	util.StartWeb(8080,config)
*/
package util

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/leptonyu/goeast/db"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

//Create html from templates files.
func Template(key string, m interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.New(key)
		t, _ = t.ParseFiles("templates/" + key)
		t.Execute(w, m)
	}
}

// Start Web Service.
func StartWeb(port int, config *db.DBConfig) {
	m := martini.Classic()
	m.NotFound(Template("404.tpl", m))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	m.Get("/", Template("index.tpl", m))
	wc := config.WC
	if port == 8080 {
		DeamonTask(config)
	}
	DispatchFunc(config, wc)
	//Create api route
	ff := wc.CreateHandlerFunc()
	m.Get("/"+config.DBName, ff)
	m.Post("/"+config.DBName, ff)
	err := http.ListenAndServe(":"+strconv.Itoa(port), m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("stop!")
	os.Exit(0)
}
