package logspoutloges

import (
	"bytes"
	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gliderlabs/logspout/router"
	"github.com/mattbaird/elastigo/lib"
)

var elastigoConn *elastigo.Conn

func init() {
	router.AdapterFactories.Register(NewLogesAdapter, "logspoutloges")
}

// LogesAdapter is an adapter that streams TCP JSON to Elasticsearch
type LogesAdapter struct {
	conn    *elastigo.Conn
	route   *router.Route
	indexer *elastigo.BulkIndexer
}

// NewLogesAdapter creates a LogesAdapter with TCP Elastigo BulkIndexer as the default transport.
func NewLogesAdapter(route *router.Route) (router.LogAdapter, error) {

	addr := route.Address

	elastigoConn = elastigo.NewConn()
	// The old standard for host was including :9200
	esHost := strings.Replace(addr, ":9200", "", -1)
	log.Infof("esHost variable: %s\n", esHost)

	elastigoConn.SetHosts([]string{esHost})
	indexer := elastigoConn.NewBulkIndexerErrors(10, 120)
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

// Stream implements the router.LogAdapter interface.
func (a *LogesAdapter) Stream(logstream chan *router.Message) {
	lid := 0
	for m := range logstream {
		lid++
		// Un-escape the newline characters so logs look nice
		m.Data = EncodeNewlines(m.Data)

		msg := LogesMessage{
			Message:  m.Data,
			Name:     m.Container.Name,
			ID:       m.Container.ID,
			Image:    m.Container.Config.Image,
			Hostname: m.Container.Config.Hostname,
			LID:      lid,
		}
		js, err := json.Marshal(msg)
		if err != nil {
			log.Println("loges:", err)
			continue
		}

		idx := "logstash-" + m.Time.Format("2006.01.02")
		//Index(index string, _type string, id,         ttl string, date *time.Time, data interface{}, refresh bool)
		if err := a.indexer.Index(idx, "golog", msg.ID, "30d", &m.Time, js, false); err != nil {
			log.Errorf("Index(ing) error: %v\n", err)
		}
	}
}

// LogESMessage Encapsulates the log data for Elasticsearch
type LogesMessage struct {
	Message  string `json:"message"`
	Name     string `json:"docker.name"`
	ID       string `json:"docker.id"`
	Image    string `json:"docker.image"`
	Hostname string `json:"docker.hostname"`
	LID      int    `json:logspout.loges.lid"`
}
