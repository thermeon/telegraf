package nsca

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

type NscaSerializer struct {
	CPUWarning   float64
	CPUCritical  float64
	DiskWarning  float64
	DiskCritical float64
}

func (s *NscaSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	metricsName := metric.Name()
	var byteResp []byte
	switch metricsName {
	case "system":
		byteResp = s.getCPULoad(metric)
	case "mem":
		byteResp = s.getDiskStatus(metric)
	}

	return byteResp, nil
}

func (s *NscaSerializer) getCPULoad(metric telegraf.Metric) []byte {

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
		if cpuTotal < s.CPUWarning {
			//status -> OK
			cpuStatus = "0"
			status = "OK"
		} else if cpuTotal >= s.CPUWarning && cpuTotal < s.CPUCritical {
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
func (s *NscaSerializer) getDiskStatus(metric telegraf.Metric) []byte {
	var message []byte
	tags := metric.Tags()
	field := metric.Fields()

	usedPercent := field["used_percent"]
	host := tags["host"]

	diskState := usedPercent.(float64)

	var diskStatus, status string
	if diskState < s.DiskWarning {
		//status -> OK
		diskStatus = "0"
		status = "OK"
	} else if diskState >= s.DiskWarning && diskState < s.DiskCritical {
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
