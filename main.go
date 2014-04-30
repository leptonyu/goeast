package main

import (
	"flag"
	"github.com/leptonyu/goeast/util"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	dbname := flag.String("dbname", "goeast", "MongoDB name")
	help := flag.Bool("help", false, "Help")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	util.StartWeb(*port, *dbname)
}
