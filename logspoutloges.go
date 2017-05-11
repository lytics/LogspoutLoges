package logspoutloges

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gliderlabs/logspout/router"
	"github.com/mattbaird/elastigo/lib"
)

var elastigoConn *elastigo.Conn

func init() {
	router.AdapterFactories.Register(NewLogesAdapter, "logspoutloges")
	elastigoConn = elastigo.NewConn()
}

// LogesAdapter is an adapter that streams TCP JSON to Elasticsearch
type LogesAdapter struct {
	conn    *elastigo.Conn
	route   *router.Route
	indexer *elastigo.BulkIndexer
}

// NewLogesAdapter creates a LogesAdapter with TCP Elastigo BulkIndexer as the default transport.
// eg URI: `logspoutloges://10.240.0.1+10.240.0.2+10.240.0.3`
func NewLogesAdapter(route *router.Route) (router.LogAdapter, error) {
	hosts := parseEsAddr(route.Address)
	log.Debugf("ES Hosts: %s", hosts)
	elastigoConn.SetHosts(hosts)

	indexer := elastigoConn.NewBulkIndexerErrors(50, 120)
	indexer.Sender = func(buf *bytes.Buffer) error {
		log.Infof("es writing: %d bytes", buf.Len())
		return indexer.Send(buf)
	}
	indexer.Start()

	return &LogesAdapter{
		route:   route,
		conn:    elastigoConn,
		indexer: indexer,
	}, nil
}

func parseEsAddr(addr string) []string {
	esHosts := strings.Replace(addr, ":9200", "", -1)
	return strings.Split(esHosts, "+")
}

func parseFields(rawMsg []byte) (*LogFields, error) {
	lf := new(LogFields)
	err := json.Unmarshal(rawMsg, lf)
	if err != nil {
		return nil, err
	}
	return lf, nil
}

// Stream implements the router.LogAdapter interface.
func (a *LogesAdapter) Stream(logstream chan *router.Message) {
	lid := 0
	for m := range logstream {
		lid++
		// Un-escape the newline characters so logs look nice
		var msgVal string
		msgVal = EncodeNewlines(m.Data)

		fieldMap := make(map[string]interface{})
		if fields, err := parseFields([]byte(m.Data)); err == nil && fields.Message != "" {
			msgVal = fields.Message
			fieldMap["level"] = fields.Level
			fieldMap["severity"] = fields.Severity
			fieldMap["line"] = fields.Line
			fieldMap["file"] = fields.File
			fieldMap["rawtime"] = fields.RawTime
		}
		fieldMap["host"] = m.Container.Config.Hostname
		fieldMap["image"] = m.Container.Config.Image

		msg := LogesMessage{
			Source:    m.Container.Config.Hostname,
			Type:      "logspout",
			Fields:    fieldMap,
			Timestamp: time.Now(),
			Message:   msgVal,
		}
		js, err := json.Marshal(msg)
		if err != nil {
			log.Errorf("loges marshal error: %v", err)
			continue
		}

		idx := "logstash-" + m.Time.Format("2006.01.02")
		//Index(index string, _type string, id,         ttl string, date *time.Time, data interface{}, refresh bool)
		if err := a.indexer.Index(idx, "logspout", "", "", "90d", &m.Time, js); err != nil {
			log.Errorf("Index(ing) error: %v\n", err)
		}
	}
}

// LogesMessage Encapsulates the log data for Elasticsearch
type LogesMessage struct {
	Source      string                 `json:"@source"`
	Type        string                 `json:"@type"`
	Timestamp   time.Time              `json:"@timestamp"`
	Message     string                 `json:"@message"`
	Tags        []string               `json:"@tags,omitempty"`
	IndexFields map[string]interface{} `json:"@idx,omitempty"`
	Fields      map[string]interface{} `json:"@fields,omitempty"`
}

type LogFields struct {
	Level    string `json:"level"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	RawTime  string `json:"time"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}
