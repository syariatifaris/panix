package main

import (
	"github.com/gorilla/mux"
	"github.com/syariatifaris/panicx"
	"log"
	"net/http"
)

func init() {
	panix.InitSlack("staging", &panix.SlackConfig{
		Channel:     "core-panic",
		WebHookURL:  "https://hooks.slack.com/services/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",
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
