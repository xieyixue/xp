package main

import (
	"flag"
	"log"
	"proxy"
	"os"
	"fmt"
)

var (
	logFileName = flag.String("log", "/var/log/main.log", "Log file name")
)


func main() {
	logFile, logErr := os.OpenFile(*logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to find", *logFile, "cServer start Failed")
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags|log.Lshortfile)
	log.Print("proxy start listen 0.0.0.0:8082")
	server := proxy.Server{"0.0.0.0", "8082", "Server"}
	err := server.Listen()
	if err != nil {
		log.Panic(err)
	}
}
