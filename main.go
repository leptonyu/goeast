package main

import (
	"flag"
	"fmt"
	"github.com/leptonyu/goeast/util"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	help := flag.Bool("help", false, "Help")
	flag.Parse()
	if *help {
		fmt.Println("This is help")
		os.Exit(0)
	}
	util.StartWeb(*port)
}
