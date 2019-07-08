package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"github.com/sapariduo/surfer/engine"
)

func main() {
	var frontendArgs URLValues

	flag.Var(&frontendArgs, "frontend", "Frontend URLs")
	flag.Parse()

	if len(frontendArgs) == 0 {
		frontendArgs = append(frontendArgs, mustParseURL("syslog+udp://:5141"))
		frontendArgs = append(frontendArgs, mustParseURL("syslog+tcp://:5514"))
		// frontendArgs = append(frontendArgs, mustParseURL("api+http://:8181/api/"))
		// frontendArgs = append(frontendArgs, mustParseURL("ui+http://:8282/"))
	}

	e, err := engine.NewEngine(frontendArgs)
	if err != nil {
		log.Fatalf("Unable to create engine: %s", err)
	}

	if err = e.Start(); err != nil {
		log.Fatalf("Unable to start engine: %s", err)
	}
	defer e.Close()

	e.Wait()

}

type URLValues []*url.URL

func (s *URLValues) String() string {
	return fmt.Sprintf("%+v", *s)
}

func (s *URLValues) Set(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	*s = append(*s, parsed)
	return nil
}

func mustParseURL(value string) *url.URL {
	parsed, err := url.Parse(value)
	if err != nil {
		log.Fatalf("Unable to parse URL: %s", err)
	}
	return parsed
}
