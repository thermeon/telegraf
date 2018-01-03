package nsca

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/metric"
)

func TestSerializeCPULoad(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host": "passive.dev.thermeon.com",
	}
	fields := map[string]interface{}{
		"load1":  float64(1.62),
		"load5":  float64(1.51),
		"load15": float64(1.51),
	}
	m, err := metric.New("system", tags, fields, now)
	assert.NoError(t, err)

	s := NscaSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	split := strings.Split(mS[0], ";")
	assert.Equal(t, "passive.dev.thermeon.com", split[1])
	assert.Equal(t, "CPU Load", split[2])
	assert.Equal(t, "0", split[3])
	assert.Equal(t, "OK - load average: 1.620000,1.510000,1.510000", split[4])
	assert.NoError(t, err)
}

func TestSerializeDiskState(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"host": "passive.dev.thermeon.com",
	}
	fields := map[string]interface{}{
		"used_percent": float64(68.24),
	}
	m, err := metric.New("mem", tags, fields, now)
	assert.NoError(t, err)

	s := NscaSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	split := strings.Split(mS[0], ";")
	assert.Equal(t, "passive.dev.thermeon.com", split[1])
	assert.Equal(t, "Disk Space", split[2])
	assert.Equal(t, "0", split[3])
	assert.Equal(t, "Disk State: OK", split[4])
	assert.NoError(t, err)
}
