package panix

import (
	"bytes"
	"fmt"
	"log"
	"net/http/httputil"
	"os"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime/debug"
)

var (
	isMock bool
)

//SlackConfig structure
type SlackConfig struct {
	Enabled     bool
	Channel     string
	WebHookURL  string
	EnabledEnvs []string
}

var (
	cfg *SlackConfig

	host              string
	startedAt         string
	environment       string
	isLogPanicEnabled bool
)

//InitSlack inits slack with configuration
//arg cfg:	configuration to initiate slack
func InitSlack(env string, slackCfg *SlackConfig) error {
	var err error
	cfg = slackCfg
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	IP, err := ExternalIP()
	if err != nil {
		return err
	}
	host = fmt.Sprintf("%s(%s)", hostname, IP)
	startedAt = time.Now().Format(time.RFC1123)
	environment = env
	setEnablePanic()
	isMock = false
	return nil
}

//setEnablePanic sets enabling log based on configuration
func setEnablePanic() {
	if cfg != nil {
		if cfg.Enabled {
			if InArrayStr(environment, cfg.EnabledEnvs) {
				isLogPanicEnabled = true
				return
			}
		}
	}
	isLogPanicEnabled = false
}

//stringifyCauseStackTrace formats panic cause and stack trace
//arg
//	recoverResult:	return value of recover()
//	stackTrace:		raw stack trace
//returns
//	cause:		formatted panic cause
//	stackTrace:	formatted panic stack trace
func stringifyCauseStackTrace(recoverResult interface{}, stackTraceB []byte) (cause, stackTrace string) {
	stackTrace = string(debug.Stack())
	cause = fmt.Sprint(recoverResult)
	return
}

//BadDeployment synchronously restore panic
//use this operation if you want to capture bad deployment / panic in main
func BadDeployment() {
	if x := recover(); x != nil || isMock {
		cause, stackTrace := stringifyCauseStackTrace(x, debug.Stack())
		if !isLogPanicEnabled {
			return
		}
		title := GetSlackTitle(nil)
		postToSlack(title, cause, stackTrace, nil)
	}
}

//BadOperation asynchronously restore the panic
//use this operation if you want to capture bad operation
//arg
//	title:		slack title
//	contents:	contents attachment for tracing
func BadOperation(title string, contents map[string]string) {
	if x := recover(); x != nil || isMock {
		cause, stackTrace := stringifyCauseStackTrace(x, debug.Stack())
		if !isLogPanicEnabled {
			return
		}
		go postToSlack(title, cause, stackTrace, contents)
	}
}

//GetSlackTitle will generate slack title based on http request,
//if http request is nil then the title will only contain basic information
//arg req:	http request where slack title will be generated
//returns slack title
func GetSlackTitle(req *http.Request) string {
	t := time.Now()
	tStr := t.Format(time.RFC1123)
	boldEnv := toBold(fmt.Sprintf("[%s]", environment))

	title := fmt.Sprintf("%s | %s", boldEnv, toCodeHighlight(tStr))
	//append information from arguments
	if req != nil {
		host := req.Host
		title = fmt.Sprintf("%s | %s | %s", boldEnv, toCodeHighlight(tStr), toCodeHighlight(host))
	}
	return title
}

//postToSlack posts the panic capture stack trace to slack
//arg
//	title:			slack title
//	stackTrace:		panic stack trace
//	contents:		contents attachment for tracing
func postToSlack(title, cause, stackTrace string, contents map[string]string) error {
	lenContents := len(contents)
	attachments := make([]map[string]interface{}, lenContents+1)
	if lenContents > 0 {
		for k, v := range contents {
			attachments = append(attachments,
				map[string]interface{}{
					"text":      toCodeSnippet(v),
					"color":     "#e50606",
					"title":     k,
					"short":     true,
					"mrkdwn_in": []string{"text"},
				})
		}
	}

	attachments = append(attachments,
		map[string]interface{}{
			"text":      toCodeSnippet(stackTrace),
			"color":     "#e50606",
			"title":     "Stack Trace",
			"short":     true,
			"mrkdwn_in": []string{"text"},
			//add fields for panic cause
			"fields": []map[string]interface{}{
				map[string]interface{}{
					"title": "Panic Cause",
					"value": cause,
					"short": true,
				},
				map[string]interface{}{
					"title": "Host & Start Time",
					"value": host + "\n" + startedAt,
					"short": true,
				},
			},
		})
	payload := map[string]interface{}{
		"text": title,
		//Enable slack to parse mention @<someone>
		"link_names":  1,
		"attachments": attachments,
	}
	if cfg.Channel != "" {
		payload["channel"] = cfg.Channel
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Println("[panics] marshal err", err, title, cause, stackTrace)
		return err
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post(cfg.WebHookURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[panics] error on capturing error : %s %s %s %s\n", err, title, cause, stackTrace)
			return err
		}
		log.Printf("[panics] error on capturing error : %s %s %s %s\n", string(b), title, cause, stackTrace)
	}
	return nil
}

func GetSlackTitleAndContent(r *http.Request) (string, map[string]string, error) {
	request, err := httputil.DumpRequest(r, true)
	if err != nil {
		return "", nil, err
	}
	return GetSlackTitle(r), map[string]string{"Request": string(request)}, nil
}

//toBold convert normal text to slack bold text
func toBold(text string) string {
	return fmt.Sprintf("*%s*", text)
}

//toCodeSnippet convert text to slack code snippet
func toCodeSnippet(text string) string {
	return fmt.Sprintf("```%s```", text)
}

//toCodeHighlight convert text to slack code highlight
func toCodeHighlight(text string) string {
	return fmt.Sprintf("`%s`", text)
}
