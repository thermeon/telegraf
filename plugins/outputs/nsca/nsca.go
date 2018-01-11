package nsca

import (
	"fmt"
	"net"
	"strconv"

	"github.com/google/uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/raff/tls-ext"
	"github.com/raff/tls-psk"
)

// NSCAServer can be used as a lower-level alternative to RunEndpoint. It is NOT safe
// to use an instance across mutiple threads.
type NSCAServer struct {
	Url          string
	Identity     string
	Key          string
	conn         net.Conn
	serializer   serializers.Serializer
	CPUWarning   float64
	CPUCritical  float64
	DiskWarning  float64
	DiskCritical float64
}

var sampleConfig = `
  # These all values are server specific.
  Url = "icinga.dev.thermeon.com:5668"
  Identity = "system-checker"
  Key = "0123456789"
  CPUWarning = 3.6
  CPUCritical = 4.0
  DiskWarning = 80.0
  DiskCritical = 90.0
`

func (n *NSCAServer) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NSCAServer) SampleConfig() string {
	return sampleConfig
}

func (n *NSCAServer) Description() string {
	return "Send telegraf measurements to nsca"
}

// define GetKey and GetIdentity methods
func (n *NSCAServer) getIdentity() string {
	return n.Identity
}

func (n *NSCAServer) getKey(id string) ([]byte, error) {
	return []byte(n.Key), nil
}

// Connect to an NSCA server.
func (n *NSCAServer) Connect() error {
	config := &tls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_256_CBC_SHA},
		Certificates: []tls.Certificate{tls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      n.getKey,
			GetIdentity: n.getIdentity,
		},
	}
	conn, err := tls.Dial("tcp", n.Url, config)
	if err != nil {
		return err
	}
	n.conn = conn
	uuid := uuid.New()

	//commands which nsca-ng server accepts
	command := "MOIN 1 " + uuid.String()
	fmt.Fprintf(n.conn, command+"\n")

	s, err := serializers.NewNscaSerializer(n.CPUWarning, n.CPUCritical, n.DiskWarning, n.DiskCritical)
	if err != nil {
		return err
	}
	n.serializer = s
	return nil
}

// Close the connection and clean up.
func (n *NSCAServer) Close() error {
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	return nil
}

func (n *NSCAServer) Write(metrics []telegraf.Metric) error {
	var batch []byte
	var err error
	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			fmt.Printf("Error serializing some metrics to nsca: %s", err.Error())
			return err
		}
		batch = append(batch, buf...)
	}
	text := "PUSH " + strconv.Itoa(len(batch))
	fmt.Fprintf(n.conn, text+"\n")

	if _, err = n.conn.Write(batch); err != nil {
		fmt.Println("NSCA Error: " + err.Error())
	}
	return err
}

func init() {
	outputs.Add("nsca", func() telegraf.Output {
		return &NSCAServer{}
	})
}
