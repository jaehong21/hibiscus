package main

import (
	"flag"
	"log"

	"github.com/jaehong21/hibiscus/config"
	tviewroute53 "github.com/jaehong21/hibiscus/tviewapp/route53"
)

func main() {
	profile := flag.String("profile", "default", "AWS profile to use")
	flag.Parse()

	config.Initialize()
	config.SetAwsProfile(*profile)

	app := tviewroute53.NewApp()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
