package nginx_upstream_check

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NginxUpstream struct {
	Urls []string

	client *http.Client

	ResponseTimeout internal.Duration
}

var sampleConfig = `
  # An array of Nginx status of upstream URI to gather stats.
  urls = ["http://localhost/status?format=json"]

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
`

func (n *NginxUpstream) SampleConfig() string {
	return sampleConfig
}

func (n *NginxUpstream) Description() string {
	return "Read Nginx's status of upstream information (nginx_upstream_check_module)"
}

func (n *NginxUpstream) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval

	if n.client == nil {
		client, err := n.createHttpClient()
		if err != nil {
			return err
		}
		n.client = client
	}

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(n.gatherUrl(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (n *NginxUpstream) createHttpClient() (*http.Client, error) {

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (n *NginxUpstream) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	resp, err := n.client.Get(addr.String())

	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	switch contentType {
	case "application/json":
		return gatherStatusUrl(bufio.NewReader(resp.Body), getTags(addr), acc)
	default:
		return fmt.Errorf("%s returned unexpected content type %s", addr.String(), contentType)
	}
}

type Status struct {
	Servers HealthCheckNodeArray `json:servers`
}

type HealthCheckNodeArray struct {
	Total      int64             `json:"total"`
	Generation *int              `json:"generation"`
	Server     []HealthCheckNode `json:"server"`
}

type HealthCheckNode struct {
	Index    *int   `json:"index"`
	Upstream string `json:"upstream"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Rise     int64  `json:"rise"`
	Fall     int64  `json:"fall"`
	Type     string `json:"type"`
	Port     int64  `json:"port"`
}

func gatherStatusUrl(r *bufio.Reader, tags map[string]string, acc telegraf.Accumulator) error {
	dec := json.NewDecoder(r)
	status := &Status{}
	if err := dec.Decode(status); err != nil {
		return fmt.Errorf("Error while decoding JSON response")
	}
	status.Gather(tags, acc)
	return nil
}

func (s *Status) Gather(tags map[string]string, acc telegraf.Accumulator) {
	s.gatherUpstreamMetrics(tags, acc)
}

func (s *Status) gatherUpstreamMetrics(tags map[string]string, acc telegraf.Accumulator) {

	upstreamFields := map[string]interface{}{
		"total":      s.Servers.Total,
		"generation": *s.Servers.Generation,
	}

	acc.AddFields(
		"nginx_upstream_check",
		upstreamFields,
		tags,
	)

	for _, peer := range s.Servers.Server {

		peerFields := map[string]interface{}{
			"status": peer.Status,
			"rise":   peer.Rise,
			"fall":   peer.Fall,
			"type":   peer.Type,
			"port":   peer.Port,
		}

		peerTags := map[string]string{}

		peerTags["upstream_name"] = peer.Upstream
		peerTags["upstream_server"] = peer.Name

		if peer.Index != nil {
			peerTags["index"] = strconv.Itoa(*peer.Index)
		}
		acc.AddFields("nginx_upstream_check_peer", peerFields, peerTags)
	}
}

func getTags(addr *url.URL) map[string]string {
	h := addr.Host
	host, port, err := net.SplitHostPort(h)
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
	return map[string]string{"server": host, "port": port}
}

func init() {
	inputs.Add("nginx_upstream_check", func() telegraf.Input {
		return &NginxUpstream{}
	})
}
