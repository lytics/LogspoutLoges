/*
Package logspoutloges is a log writing module for the github.com/gliderlabs/logspout daemon.

logspoutloges reads Docker logs and transforms them into JSON formats which Elasticsearch can consume and the frontend Kibana is used to searching without the need for running Logstash! Custom Elasticsearch mappings need to be applied so the appropriate fields are indexed.

Logs are written to ES indexes in the same format that Logstash would, so Kibana is able to instantaneously provide a search interface.

Logspout addressing examples:

  "logspoutloges://10.240.0.1"

To specify writing to multiple ES hosts:

  "logspoutloges://10.240.0.1+10.240.0.2+10.240.0.3"

To specify a custom client port:

  "logspoutloges://10.240.0.1:9201"
  "logspoutloges://10.240.0.1+10.240.0.2+10.240.0.3:9201"
*/
package logspoutloges
