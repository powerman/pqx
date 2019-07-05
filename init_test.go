package pqx

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/powerman/gotest/testinit"
)

func TestMain(m *testing.M) { testinit.Main(m) }

var (
	testTimeFactor = floatGetenv("GO_TEST_TIME_FACTOR", 1.0)
	testSecond     = time.Duration(float64(time.Second) * testTimeFactor)
)

func floatGetenv(name string, def float64) float64 {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return def
	}
	return v
}
