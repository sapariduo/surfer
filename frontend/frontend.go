package frontend

import (
	"fmt"
	"net/url"

	"github.com/sapariduo/surfer/spi"
)

//NewFrontend wrap initiation for frontend server for syslog
func NewFrontend(e spi.LogEngine, frontendURL *url.URL) (spi.LogFrontend, error) {
	switch frontendURL.Scheme {
	case "syslog+tcp", "syslog+udp":
		return newSyslogServerFrontend(e, frontendURL)
		// case "api+http":
		// 	return newAPIFrontend(e, frontendURL)
		// case "ui+http":
		// 	return newUIFrontend(e, frontendURL)
	}
	return nil, fmt.Errorf("Invalid frontend %s", frontendURL.Scheme)
}
