package global

import (
	"flag"
	"log"
)

var (
	IsDevelopment bool
)

func LoadGlobalFlags() {
	parseFlags()
}

func parseFlags() {
	log.Println("Parsing flags")
	isDevelopment := flag.Bool("debug", false, "set the environment to development")
	flag.Parse()
	IsDevelopment = *isDevelopment
	log.Println("CMD flags parsed successfully")
}
