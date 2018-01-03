package nsca

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

const (

	// CPU load is in warning state
	// if cpu core is 4 then warning thresold is 3.6(90%)
	STATE_WARNING = 3.6

	// CPU load is in critical state
	// if cpu core is 4 then critical thresold is 4(100%)
	STATE_CRITICAL = 4

	// Disk State is in warning if used disk equal/greated than 80%
	DISK_WARNING = 80

	// Disk State is in critical if used disk equal/greated than 90%
	DISK_CRITICAL = 90
)

type NscaSerializer struct {
}

func (s *NscaSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	metricsName := metric.Name()
	var byteResp []byte
	switch metricsName {
	case "system":
		byteResp = getCPULoad(metric)
	case "mem":
		byteResp = getDiskStatus(metric)
	}

	return byteResp, nil
}

func getCPULoad(metric telegraf.Metric) []byte {

	var message []byte
	tags := metric.Tags()
	field := metric.Fields()

	//performing metric only when it contains "load1" or we can check any of the below map
	//map[n_users: n_cpus: load1: load5: load15:]
	if _, ok := field["load1"]; ok {

		load1 := field["load1"]
		load5 := field["load5"]
		load15 := field["load15"]
		host := tags["host"]
		cpuTotal := load1.(float64)

		var cpuStatus, status string
		if cpuTotal < float64(STATE_WARNING) {
			//status -> OK
			cpuStatus = "0"
			status = "OK"
		} else if cpuTotal >= float64(STATE_WARNING) && cpuTotal < float64(STATE_CRITICAL) {
			//status -> WARNING
			cpuStatus = "1"
			status = "WARNING"
		} else {
			//status -> CRITICAL
			cpuStatus = "2"
			status = "CRITICAL"
		}

		service := "CPU Load"
		load1Str := strconv.FormatFloat(load1.(float64), 'f', 6, 64)
		load5Str := strconv.FormatFloat(load5.(float64), 'f', 6, 64)
		load15Str := strconv.FormatFloat(load15.(float64), 'f', 6, 64)
		loadMsg := status + " - load average: " + load1Str + "," + load5Str + "," + load15Str
		message = buildMessage(cpuStatus, service, loadMsg, host)
	}

	return message
}
func getDiskStatus(metric telegraf.Metric) []byte {
	var message []byte
	tags := metric.Tags()
	field := metric.Fields()

	usedPercent := field["used_percent"]
	host := tags["host"]

	diskState := usedPercent.(float64)

	var diskStatus, status string
	if diskState < float64(DISK_WARNING) {
		//status -> OK
		diskStatus = "0"
		status = "OK"
	} else if diskState >= float64(DISK_WARNING) && diskState < float64(DISK_CRITICAL) {
		//status -> WARNING
		diskStatus = "1"
		status = "WARNING"
	} else {
		//status -> CRITICAL
		diskStatus = "2"
		status = "CRITICAL"
	}
	service := "Disk Space"
	loadMsg := "Disk : " + status
	message = buildMessage(diskStatus, service, loadMsg, host)

	return message
}
func buildMessage(status, service, loadMsg, hostname string) []byte {

	tim := time.Now().Unix()
	//hostname = "passive.dev.thermeon.com"
	loadMessage := "[" + strconv.Itoa(int(tim)) + "]" + " PROCESS_SERVICE_CHECK_RESULT;" +
		hostname + ";" + service + ";" + status + ";" + loadMsg + ";\n"
	return []byte(loadMessage)
}
