package nsca

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"

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

var sampleConfig = `
  subject = "telegraf"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
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

// NSCAServer can be used as a lower-level alternative to RunEndpoint. It is NOT safe
// to use an instance across mutiple threads.
type NSCAServer struct {
	conn       net.Conn
	serializer serializers.Serializer
	serverInfo ServerInfo
}

// Connect to an NSCA server.
func (n *NSCAServer) Connect() error {
	conn, err := tls.Dial("tcp", "icinga.dev.thermeon.com:5668", config)
	if err != nil {
		fmt.Println(err)
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
	text := "MOIN 1 " + sessionId
	fmt.Fprintf(n.conn, text+"\n")
	message, _ := bufio.NewReader(n.conn).ReadString('\n')
	fmt.Print("Message from server: " + message)

	tim := time.Now().Unix()
	msg := getCPULoad()
	log := "[" + strconv.Itoa(int(tim)) + "]" + " PROCESS_SERVICE_CHECK_RESULT;passive.dev.thermeon.com;" + msg.service + ";" + msg.state + ";" + msg.message
	text = "PUSH " + strconv.Itoa(len(log)+2)
	fmt.Fprintf(n.conn, text+"\n")
	message, _ = bufio.NewReader(n.conn).ReadString('\n')
	fmt.Print("Message from server: " + message)

	fmt.Fprintf(n.conn, log+";\n")
	message, _ = bufio.NewReader(n.conn).ReadString('\n')
	fmt.Print("Message from server: " + message)
	return nil
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
// also note that the â€œserverâ€ requires a certificate, even if not used here

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

func getCPULoad() (msg message) {
	load1 := cpuLoad(1)
	load5 := cpuLoad(5)
	load15 := cpuLoad(15)

	loadF1 := average(load1)
	loadF5 := average(load5)
	loadF15 := average(load15)

	totalLoadF := loadF1 + loadF5 + loadF15

	totalLoad := float64(runtime.NumCPU()) * totalLoadF / 100
	fmt.Println("total load :: ", totalLoad)

	s1 := strconv.FormatFloat(loadF1, 'f', 6, 64)
	s2 := strconv.FormatFloat(loadF5, 'f', 6, 64)
	s3 := strconv.FormatFloat(loadF15, 'f', 6, 64)

	loadAvarage := s1 + ", " + s2 + ", " + s3

	var cpuStatus string
	if totalLoad <= float64(1.60) {
		//status=OK
		cpuStatus = "0"
	} else if totalLoad > float64(1.60) && totalLoad < float64(1.80) {
		//status=WARNING
		cpuStatus = "1"
	} else {
		//status=CRITICAL
		cpuStatus = "2"
	}
	message := "load average: " + loadAvarage
	fmt.Println("message", message)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	msg.message = message
	msg.state = cpuStatus
	msg.host = hostname
	msg.service = "CPU Load"
	return
}

func cpuLoad(duration int) []float64 {
	values, _ := cpu.Percent(time.Duration(duration)*time.Second, true)
	return values
}

func average(values []float64) float64 {
	var average float64
	var sum float64
	for i := 0; i < len(values); i++ {
		sum += values[i]
	}
	average = sum / float64(len(values))
	return average

}

func init() {
	outputs.Add("nsca", func() telegraf.Output {
		return &NSCAServer{}
	})
}
