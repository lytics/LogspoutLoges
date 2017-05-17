package logspoutloges

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

// strip port designation since elastigo expects 9200
// split IP addresses if multiple are designated and separated by '+'
func parseEsAddr(addr string) []string {
	esHosts := strings.Replace(addr, ":9200", "", -1)
	return strings.Split(esHosts, "+")
}

// parsePort returns the port designation from the logspoutloges addressing scheme. Only necessary if a port other than 9200 needs to be specified for client connections.
// returns 9200 and signifying PortNotParsed error message if no port was found.
func parsePort(addr string) (string, error) {
	addrUrl, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	if p := addrUrl.Port(); p != "" {
		return p, nil
	}
	return "9200", PortNotParsed
}

func parseFields(rawMsg []byte) (*LogFields, error) {
	rl := new(LogFields)
	err := json.Unmarshal(rawMsg, rl)
	if err != nil {
		return nil, err
	}
	return rl, nil
}

func parseRawLog(rawMsg []byte) (*RawLog, error) {
	rm := new(RawLog)
	rm.LogFields = new(LogFields)
	err := json.Unmarshal(rawMsg, rm)
	if err != nil {
		return nil, err
	}
	unq, err := strconv.Unquote(string(rm.Log))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(unq), rm.LogFields)
	if err != nil {
		return nil, err
	}
	return rm, nil
}
