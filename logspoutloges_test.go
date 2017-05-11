package logspoutloges

import (
	"net/url"
	"testing"
	"time"

	"github.com/lytics/logspout/router"
)

var (
	rawMsg       = `{"level":"error","message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`
	rawMsg2      = `{"level":"error","file":"foo.go","line":20,"message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`
	logspoutLog1 = `{"log":"time=\"2017-05-11T21:11:18Z\" level=info msg=\"es writing: 389 bytes\" \n","stream":"stderr","time":"2017-05-11T21:11:18.626504554Z"}`
	fileLineLog  = `{"log":"{\"file\":\"/home/gaben/go/src/github.com/blackmesa/g/main.go\",\"level\":\"info\",\"line\":183,\"message\":\"send mean: 0s, stddev: 0s, 99%: 0s, 99.9%: 0s, 99.99%: 0s, max: 0s, 1min rate: 0, 15min rate: 0; merge count: 0, 1min rate: 0, 15min rate: 0; process count: 0, 1min rate: 0, 15min rate: 0\",\"severity\":\"INFO\",\"time\":\"2017-05-11T21:09:05Z\"}\n","stream":"stdout","time":"2017-05-11T21:09:05.8211386Z"}`
)

func TestESAddrs(t *testing.T) {
	eg := "logspoutloges://10.240.0.1+10.240.0.2+10.240.0.3:9200"
	egUrl, err := url.Parse(eg)
	if err != nil {
		t.Errorf("error parsing %s: %v", eg, err)
	}

	esHosts := parseEsAddr(egUrl.Host)
	if len(esHosts) != 3 {
		t.Errorf("error parsing multi-host route URI: %#v", esHosts)
	}
}

func TestESAddr(t *testing.T) {
	eg := "logspoutloges://10.240.0.1"
	egUrl, err := url.Parse(eg)
	if err != nil {
		t.Errorf("error parsing %s: %v", eg, err)
	}

	esHosts := parseEsAddr(egUrl.Host)
	if len(esHosts) != 1 {
		t.Errorf("error parsing single-host route URI: %#v", esHosts)
	}
}

func TestLogFieldUnmarshal(t *testing.T) {
	lf, err := parseFields([]byte(rawMsg))
	if err != nil {
		t.Errorf("error parsing fields: %v", err)
	}
	if lf == nil {
		t.Fatal("logfields are nil")
	}
	t.Logf("LogFields: %#v", *lf)
	if lf.Level != "error" {
		t.Errorf("'level' field not expected 'error' value: %v", lf.Level)
	}
}

func TestLogFieldUnmarshalFileLine(t *testing.T) {
	lf, err := parseFields([]byte(rawMsg2))
	if err != nil {
		t.Errorf("error parsing fields: %v", err)
	}
	if lf == nil {
		t.Fatal("logfields are nil")
	}
	t.Logf("LogFields: %#v", *lf)
	if lf.Level != "error" {
		t.Errorf("'level' field not expected 'error' value: %v", lf.Level)
	}
	if lf.Line != 20 {
		t.Errorf("error parsing line number: %v not equal to 20", lf.Line)
	}
	if lf.File != "foo.go" {
		t.Errorf("error parsing file name: %v", lf.File)
	}
}

func TestParseRawLog(t *testing.T) {
	rl, err := parseRawLog([]byte(fileLineLog))
	if err != nil {
		t.Errorf("error parsing raw log message: %v\n%#v", err, rl)
	}
	if rl.Stream != "stdout" {
		t.Errorf("error decoding simple string field")
	}
	t.Logf("%#v", rl.LogFields)
	if rl.LogFields.Line != 183 {
		t.Errorf("line number not parsed from log message properly: %#v", rl.LogFields)
	}
}

func TestProcessMessage(t *testing.T) {
	m := &router.Message{
		Container: nil,
		Source:    "testing",
		Data:      fileLineLog,
		Time:      time.Now(),
	}

	l, err := processMessage(m)
	if err != nil {
		t.Errorf("error processing message %#v\n%v", m, err)
	}
	if l.Fields["host"] != "???" {
		t.Errorf("empty Container field should have resulted in question marks")
	}
	if l.Fields["line"] != 183 {
		t.Errorf("line number not parsed from log message properly: %#v", l.Fields)
	}
}
