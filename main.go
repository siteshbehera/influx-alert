package main

/*
Generate slack token: https://api.slack.com/web

Environment Variables:
  * SLACK_API_TOKEN
  * SLACK_ROOM
  * HIPCHAT_API_TOKEN
  * HIPCHAT_ROOM_ID
  * HIPCHAT_SERVER (optional)

*/

import (
	"fmt"
	"github.com/bluele/slack"
	flag "github.com/ogier/pflag"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type Trigger struct {
	Operator string
	Value    int64
}

type Notifier struct {
	Name string
}

type Alert struct {
	Name         string
	Type         string
	Function     string
	Limit        int
	Timeshift    string
	Query        string
	Interval     float64
	Trigger      Trigger
	NotifiersRaw []string   `yaml:"notifiers"`
	Notifiers    []Notifier `yaml:"-"`
}

var err error
var slack_api *slack.Slack
var slack_channel *slack.Channel
var hipchat_api *hipchat.Client

func main() {
	var file *string = flag.StringP("config", "c", "", "Config file to use")
	flag.Parse()

	setupInflux()

	alerts := []Alert{}

	data, _ := ioutil.ReadFile(*file)
	err := yaml.Unmarshal(data, &alerts)
	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%+v\n", alerts)
	}

	setupSlack()
	setupHipchat()
	done := make(chan bool)
	for _, alert := range alerts {
		go func(alert Alert) {
			for {
				alert.Run()
				time.Sleep(time.Duration(alert.Interval) * time.Second)
			}
		}(alert)
	}
	<-done // wait
}
