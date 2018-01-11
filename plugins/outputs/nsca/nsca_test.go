package nsca

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	url := "icinga.dev.thermeon.com:5668"

	e := &NSCAServer{
		Url:          url,
		Identity:     "system-checker",
		Key:          "0123456789",
		CPUWarning:   3.6,
		CPUCritical:  4.0,
		DiskWarning:  80.0,
		DiskCritical: 90.0,
	}
	// Verify that we can connect to nsca-ng server
	err := e.Connect()
	assert.NoError(t, err, "should not cause an error")

	// Verify that we can successfully write data to nsca-ng server
	err = e.Write(testutil.MockMetrics())
	assert.NoError(t, err, "should not cause an error")

}
