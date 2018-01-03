package nsca

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/raff/tls-ext"
	"github.com/raff/tls-psk"
)

// ServerInfo contains the configuration information for an NSCA server
type ServerInfo struct {
	// Host is the IP address or host name of the NSCA server. Leave empty for localhost.
	Host string
	// Port is the IP port number (no default)
	Port string
	// EncryptionMethod specifies the message encryption to use on NSCA messages. It defaults to ENCRYPT_NONE.
	EncryptionMethod int
	// Password is used in encryption.
	Password string
	// Timeout is the connect/read/write network timeout
	Timeout time.Duration
}

// Message is the contents of an NSCA message
type message struct {
	// State is one of {STATE_OK, STATE_WARNING, STATE_CRITICAL, STATE_UNKNOWN}
	state string
	// Host is the host name to set for the NSCA-ng message
	host string
	// Service is the service name to set for the NSCA-ng message
	service string
	// Message is the "plugin output" of the NSCA-ng message
	message string
}

// NSCAServer can be used as a lower-level alternative to RunEndpoint. It is NOT safe
// to use an instance across mutiple threads.
type NSCAServer struct {
	Url        string
	conn       net.Conn
	serializer serializers.Serializer
	serverInfo ServerInfo
}

var sampleConfig = `
  Url = "icinga.dev.thermeon.com:5668"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  ##data_format = "influx"
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
func getIdentity() string {
	return "system-checker"
}

func getKey(id string) ([]byte, error) {
	return []byte("0123456789"), nil
}

// create the appropriate TLS configuration
// note that we specifiy a single cipher suite of type TLS_PSK_*
// also note that the server requires a certificate, even if not used here

var (
	config = &tls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_256_CBC_SHA},
		Certificates: []tls.Certificate{tls.Certificate{}},
		Extra: psk.PSKConfig{
			GetKey:      getKey,
			GetIdentity: getIdentity,
		},
	}
)

func generateSessionId() string {
	randbytes := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, randbytes); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(randbytes)
}

// Connect to an NSCA server.
func (n *NSCAServer) Connect() error {
	conn, err := tls.Dial("tcp", n.Url, config)
	if err != nil {
		return err
	}
	n.Close()
	n.conn = conn
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

	sessionId := generateSessionId()

	//commands which nsca-ng server accepts
	command := "MOIN 1 " + sessionId
	fmt.Fprintf(n.conn, command+"\n")

	var batch []byte
	s, err := serializers.NewNscaSerializer()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		buf, err := s.Serialize(metric)
		if err != nil {
			fmt.Printf("E! Error serializing some metrics to nsca: %s", err.Error())
		}
		batch = append(batch, buf...)
	}
	text := "PUSH " + strconv.Itoa(len(batch))
	fmt.Fprintf(n.conn, text+"\n")

	if _, err = n.conn.Write(batch); err != nil {
		fmt.Println("E! NSCA Error: " + err.Error())
	}
	return err
}

func init() {
	outputs.Add("nsca", func() telegraf.Output {
		return &NSCAServer{}
	})
}
