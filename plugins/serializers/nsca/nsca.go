package nsca

import (
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

const (
	// CPU load is in ok state
	STATE_OK = 6

	// CPU load is in warning state
	STATE_WARNING = 8
)

type NscaSerializer struct {
}

func (s *NscaSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	metricsName := metric.Name()
	var byteResp []byte
	switch metricsName {
	case "cpu":
		byteResp = getCPULoad(metric)
	}
	return byteResp, nil
}

func getCPULoad(metric telegraf.Metric) []byte {
	tags := metric.Tags()
	var message []byte
	if tags[metric.Name()] == "cpu-total" {

		field := metric.Fields()

		user := field["usage_user"]
		system := field["usage_system"]
		guest := field["usage_guest"]
		host := field["host"]

		cpuTotal := user.(float64) + system.(float64) + guest.(float64)

		var cpuStatus string
		if cpuTotal <= float64(STATE_OK) {
			//status=OK
			cpuStatus = "0"
		} else if cpuTotal > float64(STATE_OK) && cpuTotal < float64(STATE_WARNING) {
			//status=WARNING
			cpuStatus = "1"
		} else {
			//status=CRITICAL
			cpuStatus = "2"
		}
		service := "CPU Load"
		loadMsg := "load average: " + strconv.FormatFloat(cpuTotal, 'f', 6, 64)
		fmt.Println("message", loadMsg)
		message = buildMessage(cpuStatus, service, loadMsg, host.(string))

	}
	return message
}
func buildMessage(status, service, loadMsg, hostname string) []byte {

	tim := time.Now().Unix()
	//hostname = "passive.dev.thermeon.com;"
	loadMessage := "[" + strconv.Itoa(int(tim)) + "]" + " PROCESS_SERVICE_CHECK_RESULT;" +
		hostname + ";" + service + ";" + status + ";" + loadMsg + ";\n"
	return []byte(loadMessage)
}
