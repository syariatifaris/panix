# Panix - Post Panic to Slack

## About
Panicx is an library written in Go to caputure panic event, and post it to your own slack channel. 

## How to use
To use panicx you can create a slack incoming webhook:
```
https://get.slack.help/hc/en-us/articles/115005265063-Incoming-WebHooks-for-Slack
```
And integrate it using this library. Panix is agnostic to any http handler library

### sample: 
```
package main

import (
	"github.com/gorilla/mux"
	"github.com/syariatifaris/panicx"
	"log"
	"net/http"
)

func init() {
	panix.InitSlack("staging", &panix.SlackConfig{
		Channel:     "channel-panic",
		WebHookURL:  "https://hooks.slack.com/services/xxxxxxxxxxxxxxxx",
		Enabled:     true,
		EnabledEnvs: []string{"staging", "production"},
	})
	log.Println("panic initiated")
}

func main() {
	log.Println("staring server on 0.0.0.0:9091")
	router := mux.NewRouter()
	router.HandleFunc("/test", handlePanic(testFunc)).Methods(http.MethodGet)
	if err := http.ListenAndServe("0.0.0.0:9091", router); err != nil {
		log.Fatalln("serve error", err.Error())
	}
}

func testFunc(w http.ResponseWriter, r *http.Request) {
	trigger()
	w.Write([]byte("hello world"))
}

func trigger() {
	var ds []int64
	log.Println(ds[1])
}

func handlePanic(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ttl, data, _ := panix.GetSlackTitleAndContent(r)
		defer panix.BadOperation(ttl, data)
		f(w, r)
	}
}
```
You will obtain message to slack like this one:
![Panic on Slack](img/sample.png?raw=true "Panic Message")
