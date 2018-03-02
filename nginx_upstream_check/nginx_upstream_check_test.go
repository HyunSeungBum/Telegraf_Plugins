package nginx_upstream_check

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

const sampleStatusResponse = `
{
	"servers": {
		"total": 2,
		"generation": 1,
		"server": [{
				"index": 0,
				"upstream": "upstreamcluster",
				"name": "1.2.3.4:8180",
				"status": "down",
				"rise": 0,
				"fall": 1471,
				"type": "http",
				"port": 0
			},
			{
				"index": 1,
				"upstream": "upstreamcluster",
				"name": "1.2.3.4:8280",
				"status": "down",
				"rise": 0,
				"fall": 1471,
				"type": "http",
				"port": 0
			}
		]
	}
}
`

func TestNginxPlusGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/status" {
			rsp = sampleStatusResponse
			w.Header()["Content-Type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &NginxUpstream{
		Urls: []string{fmt.Sprintf("%s/status?format=json", ts.URL)},
	}

	var acc testutil.Accumulator

	err_nginx := n.Gather(&acc)

	require.NoError(t, err_nginx)

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(addr.Host)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}

	acc.AssertContainsTaggedFields(
		t,
		"nginx_upstream_check_peer",
		map[string]interface{}{
			"index":                  "0",
			"upstream_name" 		  "upstreamcluster"
			"upstream_server":        "1.2.3.4:8180",
			"status":                 "down",
			"rise":                   int64(0),
			"fall":               	  int64(3630),
			"type":          		  "http",
			"port":          		  int64(0)),
		},
		map[string]string{
			"server":           host,
			"port":             port,
			"total":         	2,
			"generation":       "1",
		})
}
