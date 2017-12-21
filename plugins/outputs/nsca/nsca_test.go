package nsca

import (
    "testing"

    "github.com/influxdata/telegraf/testutil"
    "github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    url := "localhost:8080"

    e := &NSCAServer{
        Url: url,
    }
    // Verify that we can connect to nsca-ng server
    err := e.Connect()
    require.NoError(t, err)

    // Verify that we can successfully write data to nsca-ng server
    err = e.Write(testutil.MockMetrics())
    require.NoError(t, err)

}
