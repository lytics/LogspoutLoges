package logspoutloges

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lytics/logspout/router"
)

var (
	rawMsg       = `{"level":"error","message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`
	rawMsg2      = `{"level":"error","file":"foo.go","line":20,"message":"Number of values doesn't match number of fields:  fields:5 vals:6","severity":"ERROR","time":"2017-05-10T20:19:55Z"}`
	logspoutLog1 = `{"log":"time=\"2017-05-11T21:11:18Z\" level=info msg=\"es writing: 389 bytes\" \n","stream":"stderr","time":"2017-05-11T21:11:18.626504554Z"}`

	fileLineLog = `{"log":"{\"file\":\"/home/gaben/go/src/github.com/blackmesa/g/main.go\",\"level\":\"info\",\"line\":183,\"message\":\"send mean: 0s, stddev: 0s, 99%: 0s, 99.9%: 0s, 99.99%: 0s, max: 0s, 1min rate: 0, 15min rate: 0; merge count: 0, 1min rate: 0, 15min rate: 0; process count: 0, 1min rate: 0, 15min rate: 0\",\"severity\":\"INFO\",\"time\":\"2017-05-11T21:09:05Z\"}\n","stream":"stdout","time":"2017-05-11T21:09:05.8211386Z"}`
)

type TestIndexer struct {
	msgBuffer []*LogesMessage
}

func NewTestIndexer() *TestIndexer {

	msgBuffer := make([]*LogesMessage, 0)
	return &TestIndexer{
		msgBuffer: msgBuffer,
	}
}

func (ti *TestIndexer) Index(index string, _type string, id, parent, ttl string, date *time.Time, data interface{}) error {
	m := &LogesMessage{}
	err := json.Unmarshal(data.([]byte), m)
	if err != nil {
		return err
	}
	ti.msgBuffer = append(ti.msgBuffer, m)
	return nil
}

func newTestAdapter(route *router.Route) (router.LogAdapter, error) {
	ti := NewTestIndexer()
	return &LogesAdapter{
		route:   route,
		conn:    nil,
		indexer: ti,
	}, nil
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
	if l.Fields["file"] != "/home/gaben/go/src/github.com/blackmesa/g/main.go" {
		t.Errorf("error unmarshaling file: %v", err)
	}
	t.Logf("message parsed: %q", l.Message)
}
