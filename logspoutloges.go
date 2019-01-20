package logspoutloges

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/gliderlabs/logspout/router"
	elastigo "github.com/mattbaird/elastigo/lib"
	log "github.com/sirupsen/logrus"
)

var elastigoConn *elastigo.Conn

func init() {
	router.AdapterFactories.Register(NewLogesAdapter, "logspoutloges")
	log.SetLevel(log.DebugLevel)
}

// LogesAdapter is an adapter that streams TCP JSON to Elasticsearch
type LogesAdapter struct {
	conn    *elastigo.Conn
	route   *router.Route
	indexer *elastigo.BulkIndexer
}

// NewLogesAdapter creates a LogesAdapter with TCP Elastigo BulkIndexer as the default transport.
func NewLogesAdapter(route *router.Route) (router.LogAdapter, error) {
	log.Debugf("new LogesAdapter for route: %+v", route)

	elastigoConn = elastigo.NewConn()
	// The old standard for host was including :9200
	host := strings.Replace(route.Address, ":9200", "", -1)
	hosts := strings.Split(host, "+")
	log.Debugf("esHost variable: %s", hosts)

	elastigoConn.SetHosts(hosts)
	indexer := elastigoConn.NewBulkIndexerErrors(50, 120)
	indexer.Sender = func(buf *bytes.Buffer) error {
		l := buf.Len()
		err := indexer.Send(buf)
		if err != nil {
			log.Warnf("failed to send to Elasticsearch: %v", err)
			return err
		}
		log.Debugf("sent %d of %d bytes to Elasticsearch", l-buf.Len(), l)
		return nil
	}
	// TODO: How/when is the indexer closed?
	indexer.Start()

	l := &LogesAdapter{
		route:   route,
		conn:    elastigoConn,
		indexer: indexer,
	}
	log.Debugf("created adapter: %+v", l)
	return l, nil
}

// Stream implements the router.LogAdapter interface.
func (a *LogesAdapter) Stream(logstream chan *router.Message) {
	log.Debugf("started streaming for adapter: %+v", a)
	lid := 0
	for m := range logstream {
		log.Debugf("new message: %+v", m)
		lid++

		var mess string
		var fields map[string]interface{}
		d := []byte(m.Data)

		if json.Valid(d) {
			var msgField message
			if err := json.Unmarshal(d, &msgField); err != nil {
				log.Warnf("failed to unmarshal data %v: %v", m.Data, err)
			}
			mess = msgField.Message

			var fields map[string]interface{}
			if err := json.Unmarshal(d, &fields); err != nil {
				log.Warnf("failed to unmarshal data %v: %v", m.Data, err)
			} else {
				delete(fields, "message")
			}
		} else {
			mess = m.Data
		}

		msg := LogesMessage{
			Source:    m.Container.Config.Hostname,
			Type:      "logs",
			Timestamp: time.Now(),
			Message:   mess,
			Fields:    fields,
			Name:      m.Container.Name,
			ID:        m.Container.ID,
			Image:     m.Container.Config.Image,
			Hostname:  m.Container.Config.Hostname,
			LID:       lid,
		}

		js, err := json.Marshal(msg)
		if err != nil {
			log.Warnf("failed to marshal message: %v", err)
			continue
		}

		idx := "logstash-" + m.Time.Format("2006.01.02")
		if err := a.indexer.Index(idx, "logs", "", "", "30d", &m.Time, js); err != nil {
			log.Warnf("failed to index message: %v", err)
		}
		log.Debugf("indexed message: %+v", msg)
	}
	log.Debugf("done streaming for adapter: %+v", a)
}

// LogesMessage encapsulates the log data for Elasticsearch
type LogesMessage struct {
	Source      string                 `json:"@source"`
	Type        string                 `json:"@type"`
	Timestamp   time.Time              `json:"@timestamp"`
	Message     string                 `json:"@message"`
	Tags        []string               `json:"@tags,omitempty"`
	IndexFields map[string]interface{} `json:"@idx,omitempty"`
	Fields      map[string]interface{} `json:"@fields"`
	Name        string                 `json:"docker_name"`
	ID          string                 `json:"docker_id"`
	Image       string                 `json:"docker_image"`
	Hostname    string                 `json:"docker_hostname"`
	LID         int                    `json:"logspoutloges_id"`
}

type message struct {
	Message string `json:"message"`
}
