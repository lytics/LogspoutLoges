package logspoutloges

import (
	"net/url"
	"testing"
)

var rawMsg = `{"level":"error","message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`
var rawMsg2 = `{"level":"error","file":"foo.go","line":20,"message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`

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
