package logspoutloges

import (
	"net/url"
	"testing"
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

func TestParsePort(t *testing.T) {
	eg := "logspoutloges://10.240.0.1"
	p, err := parsePort(eg)
	if err != PortNotParsed {
		t.Errorf("error returned should be PortNotParsed; returned: %v", err)
	}
	if p != "9200" {
		t.Errorf("if not specified, port returned should default to 9200: returned %d", p)
	}
}
func TestParseIPsPortDefault(t *testing.T) {
	eg := "logspoutloges://10.240.0.1+10.240.0.42"
	p, err := parsePort(eg)
	if err != PortNotParsed {
		t.Errorf("error returned should be PortNotParsed; returned: %v", err)
	}
	if p != "9200" {
		t.Errorf("if not specified, port returned should default to 9200: returned %d", p)
	}
}

func TestParsePortSpecified(t *testing.T) {
	eg := "logspoutloges://10.240.0.1:9201"
	p, err := parsePort(eg)
	if err != nil {
		t.Errorf("no error should be returned: %v", err)
	}
	if p != "9201" {
		t.Errorf("port returned should be 9201: returned %d", p)
	}
}

func TestParseIPsPortSpecified(t *testing.T) {
	eg := "logspoutloges://10.240.0.1+10.240.0.12:9201"
	p, err := parsePort(eg)
	if err != nil {
		t.Errorf("no error should be returned: %v", err)
	}
	if p != "9201" {
		t.Errorf("port returned should be 9201: returned %d", p)
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
