package logspoutloges

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/logspout/router"
	log "github.com/sirupsen/logrus"
	elastic "gopkg.in/olivere/elastic.v3"
)

func init() {
	router.AdapterFactories.Register(NewLogesAdapter, "logspoutloges")
	if os.Getenv("DEBUG") != "" {
		log.Infof("Debug logging enabled.")
		log.SetLevel(log.DebugLevel)
	}
}

// LogesAdapter is an adapter that streams TCP JSON to Elasticsearch
type LogesAdapter struct {
	route *router.Route
	bp    *elastic.BulkProcessor
}

// NewLogesAdapter creates a LogesAdapter with Elasticsearch as the default transport.
func NewLogesAdapter(route *router.Route) (router.LogAdapter, error) {
	log.Debugf("new LogesAdapter for route: %+v", route)

	hosts := strings.Split(route.Address, "+")
	hosts = normalize(hosts)

	retrier := elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(time.Millisecond, 5*time.Minute))

	// FIXME: Client is never stopped.
	c, err := elastic.NewClient(
		elastic.SetURL(hosts...),
		elastic.SetErrorLog(log.StandardLogger()),
		elastic.SetInfoLog(log.StandardLogger()),
		elastic.SetRetrier(retrier),
	)
	if err != nil {
		log.Warnf("failed to create Elasticsearch client: %v", err)
		return nil, err
	}

	bp, err := c.BulkProcessor().FlushInterval(time.Minute).Do()
	if err != nil {
		log.Warnf("failed to create bulk processor: %v", err)
		return nil, err
	}

	// FIXME: Processor is never stopped.
	if err := bp.Start(); err != nil {
		log.Warnf("failed to start bulk processor: %v", err)
		return nil, err
	}

	l := &LogesAdapter{
		route: route,
		bp:    bp,
	}
	log.Debugf("created adapter: %+v", l)
	return l, nil
}

// Stream implements the router.LogAdapter interface.
func (a *LogesAdapter) Stream(logstream chan *router.Message) {
	log.Debugf("started streaming for adapter: %+v", a)
	for m := range logstream {
		var msg string
		var fields Fields

		if err := json.Unmarshal([]byte(m.Data), &fields); err == nil && fields.Message != "" {
			msg = fields.Message
			fields.Message = ""
		} else {
			// Un-escape the newline characters so logs look nice
			msg = EncodeNewlines(m.Data)
		}
		fields.Host = m.Container.Config.Hostname
		fields.Image = m.Container.Config.Image

		l := Log{
			Source:    m.Container.Config.Hostname,
			Type:      "logspout",
			Fields:    fields,
			Timestamp: time.Now(),
			Message:   msg,
		}

		idx := "logstash-" + m.Time.Format("2006.01.02")
		r := elastic.NewBulkIndexRequest().
			Doc(l).
			Index(idx).
			Type("logs")

		a.bp.Add(r)
		log.Debugf("indexed log: %+v", l)
	}
	log.Debugf("done streaming for adapter: %+v", a)
}

// Log encapsulates log data for Elasticsearch.
type Log struct {
	Source      string                 `json:"@source"`
	Type        string                 `json:"@type"`
	Timestamp   time.Time              `json:"@timestamp"`
	Message     string                 `json:"@message"`
	Tags        []string               `json:"@tags,omitempty"`
	IndexFields map[string]interface{} `json:"@idx,omitempty"`
	Fields      Fields                 `json:"@fields"`
}

// Fields encapsulates standardized log fields.
type Fields struct {
	Host     string `json:"host"`
	Image    string `json:"image"`
	Level    string `json:"level,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
	RawTime  string `json:"time,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

func normalize(urls []string) []string {
	log.Debugf("normalizing urls: %v", urls)
	out := make([]string, 0, len(urls))
	for _, s := range urls {
		url, err := url.Parse(fmt.Sprintf("http://%v", s)) // Need to add scheme for proper parsing
		if err != nil {
			log.Warnf("failed to parse %q to URL: %v", s, err)
			continue
		}
		if url.Port() == "" {
			log.Warnf("URL %q missing port, fixing for now.", url)
			url.Host = fmt.Sprintf("%v:%v", url.Hostname(), "9200")
		}
		log.Debugf("normalized URL: %v", url)
		out = append(out, url.String())
	}
	log.Debugf("normalized urls: %v", out)
	return out
}
