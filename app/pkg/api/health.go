package api

import (
	"event-data-pipeline/pkg"
	"time"
)

var (
	startTime time.Time
)

type Info struct {
	Timestamp string `json:"timestamp,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
	Version   string `json:"version,omitempty"`
}

func init() {
	startTime = time.Now() // Start monitoring uptime of the service
}

// uptime returns the duration since service start time
func uptime() time.Duration {
	return time.Since(startTime)
}

func NewInfo() *Info {
	return &Info{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    uptime().String(),
		Version:   pkg.GetVersion(),
	}

}
