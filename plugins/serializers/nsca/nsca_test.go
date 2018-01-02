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
		"cpu": "cpu-total",
	}
	fields := map[string]interface{}{
		"usage_user":   float64(8.37),
		"usage_system": float64(2.95),
		"usage_guest":  float64(0),
		"host":         "localhost",
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := NscaSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	split := strings.Split(mS[0], ";")
	assert.Equal(t, "localhost", split[1])
	assert.Equal(t, "CPU Load", split[2])
	assert.Equal(t, "2", split[3])
	assert.Equal(t, "load average: 11.320000", split[4])
	assert.NoError(t, err)
}
