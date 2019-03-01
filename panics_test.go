package panix

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

//TestInitSlack tests inititating slack
func TestInitSlack(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "case enabled",
			enabled: false,
		},
		{
			name:    "case not enabled",
			enabled: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &SlackConfig{
				Enabled:     test.enabled,
				EnabledEnvs: []string{"dev", "staging", "prod"},
				Channel:     "channel",
			}
			InitSlack("prod", config)

			assert.NotNil(t, cfg)
			assert.NotEmpty(t, host)
			assert.NotEmpty(t, startedAt)
			assert.NotEmpty(t, environment)
			assert.Equal(t, isLogPanicEnabled, test.enabled)
		})
	}
}

//TestBadDeployment tests panic on deloyment recovery
func TestBadDeployment(t *testing.T) {
	t.Skip()
	tests := []struct {
		name    string
		enabled bool
		url     string
		status  int
		err     error
	}{
		{
			name:    "log is not enabled",
			enabled: false,
		},
		{
			name:    "log is enabled, post success",
			enabled: true,
			url:     "https://abc.com/",
			status:  200,
		},
		{
			name:    "log is enabled, post returns not found",
			enabled: true,
			url:     "https://abc.com/",
			status:  404,
		},
		{
			name:    "log is enabled, post returns not found",
			enabled: true,
			url:     "https://abc.com/",
			status:  500,
			err:     errors.New("test failed to post"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer BadDeployment()

			if test.url != "" {
				gock.New(test.url).Reply(test.status).
					SetError(test.err)
			}

			cfg.WebHookURL = test.url
			isLogPanicEnabled = test.enabled

			isMock = true
		})
	}
}

//TestBadOperation tests panic on operation recovery
func TestBadOperation(t *testing.T) {
	t.Skip()
	tests := []struct {
		name     string
		enabled  bool
		contents map[string]string
		url      string
	}{
		{
			name:    "log is not enabled",
			enabled: false,
		},
		{
			name:    "log is enabled, empty contents",
			enabled: true,
			url:     "https://bcd.com/",
		},
		{
			name:    "log is enabled, non empty contents",
			enabled: true,
			contents: map[string]string{
				"key": "value",
			},
			url: "https://cde.com/",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer BadOperation("title", test.contents)

			if test.url != "" {
				gock.New(test.url).Reply(200)
			}

			cfg.WebHookURL = test.url
			isLogPanicEnabled = test.enabled

			isMock = true
		})
	}
}

//TestGetSlackTitle tests getting slack title
func TestGetSlackTitle(t *testing.T) {
	tests := []struct {
		name    string
		request *http.Request
	}{
		{
			name: "nil request",
		},
		{
			name:    "non nil request",
			request: httptest.NewRequest(http.MethodGet, "https://abc.com/", nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetSlackTitle(test.request)

			assert.NotEmpty(t, result)
		})
	}
}

func TestSetEnablePanicEnabled(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "case normal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnablePanic()
		})
	}
}

func TestStringifyCauseStackTrace(t *testing.T) {
	type args struct {
		recoverResult interface{}
		stackTraceB   []byte
	}
	tests := []struct {
		name           string
		args           args
		wantCause      string
		wantStackTrace string
	}{
		{
			name: "case checking cause",
			args: args{
				recoverResult: "foo",
			},
			wantCause: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCause, _ := stringifyCauseStackTrace(tt.args.recoverResult, tt.args.stackTraceB)

			assert.Equal(t, tt.wantCause, gotCause, "StringifyCauseStackTrace")
		})
	}
}

func TestPostToSlack(t *testing.T) {
	scfg := &SlackConfig{
		Enabled:     false,
		Channel:     "chanchan",
		EnabledEnvs: []string{"dev", "staging", "prod"},
	}
	InitSlack("prod", scfg)
	type args struct {
		title      string
		cause      string
		stackTrace string
		contents   map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case mock",
			args: args{
				title:      "foo",
				cause:      "bar",
				stackTrace: "foobar",
				contents: map[string]string{
					"foo": "bar",
					"num": "42",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postToSlack(tt.args.title, tt.args.cause, tt.args.stackTrace, tt.args.contents)
		})
	}
}

func TestToBold(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				text: "foobar",
			},
			want: "*foobar*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toBold(tt.args.text)

			assert.Equal(t, tt.want, got, "ToBold")
		})
	}
}

func Test_toCodeSnippet(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				text: "foobar",
			},
			want: "```foobar```",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCodeSnippet(tt.args.text)

			assert.Equal(t, tt.want, got, "ToCodeSnippet")
		})
	}
}

func Test_toCodeHighlight(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				text: "foobar",
			},
			want: "`foobar`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCodeHighlight(tt.args.text)

			assert.Equal(t, tt.want, got, "ToCodeHighlight")
		})
	}
}
