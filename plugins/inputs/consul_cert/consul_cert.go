package consul_cert

import (
	"net/http"
	"regexp"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Consul struct {
	Address    string
	Scheme     string
	Token      string
	Username   string
	Password   string
	Datacentre string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	// client used to connect to Consul agnet
	client *api.Client
}

var sampleConfig = `
  ## Most of these values defaults to the one configured on a Consul's agent level.
  ## Optional Consul server address (default: "localhost")
  # address = "localhost"
  ## Optional URI scheme for the Consul server (default: "http")
  # scheme = "http"
  ## Optional ACL token used in every request (default: "")
  # token = ""
  ## Optional username used for request HTTP Basic Authentication (default: "")
  # username = ""
  ## Optional password used for HTTP Basic Authentication (default: "")
  # password = ""
  ## Optional data centre to query the health checks from (default: "")
  # datacentre = ""
`

var expiryMatcher = regexp.MustCompile("/(.*)/expires$")

func (c *Consul) Description() string {
	return "Gather certificate validity time for consul-cert-postman domains"
}

func (c *Consul) SampleConfig() string {
	return sampleConfig
}

func (c *Consul) createAPIClient() (*api.Client, error) {
	config := api.DefaultConfig()

	if c.Address != "" {
		config.Address = c.Address
	}

	if c.Scheme != "" {
		config.Scheme = c.Scheme
	}

	if c.Datacentre != "" {
		config.Datacenter = c.Datacentre
	}

	if c.Username != "" {
		config.HttpAuth = &api.HttpBasicAuth{
			Username: c.Username,
			Password: c.Password,
		}
	}

	tlsCfg, err := internal.GetTLSConfig(
		c.SSLCert, c.SSLKey, c.SSLCA, c.InsecureSkipVerify)

	if err != nil {
		return nil, err
	}

	config.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return api.NewClient(config)
}

func (c *Consul) GatherCertCheck(acc telegraf.Accumulator, certs []*api.KVPair) {
	for _, cert := range certs {
		if m := expiryMatcher.FindStringSubmatch(cert.Key); len(m) == 2 {
			expires, _ := time.Parse(time.RFC3339, string(cert.Value))
			d := expires.Sub(time.Now())

			record := make(map[string]interface{})
			tags := make(map[string]string)

			record["check_name"] = "certificate"

			record["hours_remaining"] = int(d.Hours())

			tags["domain"] = m[1]

			acc.AddFields("consul_cert_checks", record, tags)
		}
	}
}

func (c *Consul) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		newClient, err := c.createAPIClient()

		if err != nil {
			return err
		}

		c.client = newClient
	}

	kv := c.client.KV()
	certs, _, err := kv.List("certs/", nil)

	if err != nil {
		return err
	}

	c.GatherCertCheck(acc, certs)

	return nil
}

func init() {
	inputs.Add("consul_cert", func() telegraf.Input {
		return &Consul{}
	})
}
