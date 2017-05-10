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
