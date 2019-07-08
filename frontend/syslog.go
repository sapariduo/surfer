package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sapariduo/surfer/api"
	"github.com/sapariduo/surfer/spi"
	"github.com/sapariduo/surfer/utils"

	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
	"gopkg.in/natefinch/lumberjack.v2"
)

type syslogServerFrontend struct {
	e spi.LogEngine
	// b      spi.LogBackend
	logsQ  syslog.LogPartsChannel
	stopQ  chan *sync.Cond
	format format.Format
	server *syslog.Server
}

func writer() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/tmp/log/surfer.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    // days
		Compress:   false, // disabled by default
	})
}

func newSyslogServerFrontend(e spi.LogEngine, frontendURL *url.URL) (*syslogServerFrontend, error) {
	if frontendURL.Host == "" {
		return nil, fmt.Errorf("Empty host in frontend URL '%s'", frontendURL)
	}

	syslogFormat, err := utils.GetSyslogFormatQueryParam(frontendURL, "format", syslog.RFC3164)
	if err != nil {
		return nil, err
	}

	queueSize, err := utils.GetIntQueryParam(frontendURL, "queueSize", 512)
	if err != nil {
		return nil, err
	}

	timeout, err := utils.GetDurationQueryParam(frontendURL, "timeout", 0*time.Second)
	if err != nil {
		return nil, err
	}

	f := syslogServerFrontend{}
	f.e = e

	logsQ := make(syslog.LogPartsChannel, queueSize)
	f.logsQ = logsQ

	stopQ := make(chan *sync.Cond, 1)
	f.stopQ = stopQ

	f.format = syslogFormat

	server := syslog.NewServer()
	server.SetFormat(syslogFormat)
	server.SetTimeout(int64(timeout.Seconds() * 1000))
	server.SetHandler(syslog.NewChannelHandler(logsQ))
	switch strings.ToLower(frontendURL.Scheme) {
	case "syslog+tcp":
		err = server.ListenTCP(frontendURL.Host)
	case "syslog+udp":
		err = server.ListenUDP(frontendURL.Host)
	}
	if err != nil {
		return nil, err
	}
	f.server = server

	return &f, nil
}

func (f *syslogServerFrontend) Start() error {

	// _, b := f.e.GetBackend()
	// f.b = b

	err := f.server.Boot()
	if err != nil {
		return err
	}

	go f.run()

	return nil
}

func (f *syslogServerFrontend) Close() error {

	cond := sync.NewCond(&sync.Mutex{})
	cond.L.Lock()
	f.stopQ <- cond
	cond.Wait()
	cond.L.Unlock()

	return f.server.Kill()
}

func (f *syslogServerFrontend) run() {
	writer()
	for {
		select {
		case logParts := <-f.logsQ:
			// log.Println(logParts)
			src := f.toLogEntry(logParts)
			f.parseContent(src.Message)
			// f.b.Insert(&api.InsertRequest{Entry: f.toLogEntry(logParts)})
		case cond := <-f.stopQ:
			cond.Broadcast()
			return
		}
	}
}

func (f *syslogServerFrontend) toLogEntry(logParts format.LogParts) *api.LogEntry {
	e := api.LogEntry{}
	switch f.format {
	case syslog.RFC3164:
		if val, ok := logParts["timestamp"].(time.Time); ok {
			e.Timestamp = val
		} else {
			e.Timestamp = time.Now()
		}
		if val, ok := logParts["hostname"].(string); ok {
			e.Hostname = val
		}
		if val, ok := logParts["tag"].(string); ok {
			e.Application = val
		}
		if val, ok := logParts["content"].(string); ok {
			e.Message = val
			if cnt, err := f.parseContent(e.Message); err == nil {
				// fmt.Printf("%s\n", cnt)
				e.MsgContent = valueMapper(cnt, e.Timestamp)
			}
		}
	case syslog.RFC5424:
		if val, ok := logParts["timestamp"].(time.Time); ok {
			e.Timestamp = val
		} else {
			e.Timestamp = time.Now()
		}
		if val, ok := logParts["hostname"].(string); ok {
			e.Hostname = val
		}
		if val, ok := logParts["app_name"].(string); ok {
			e.Application = val
		}
		if val, ok := logParts["message"].(string); ok {
			e.Message = val
			if cnt, err := f.parseContent(e.Message); err == nil {
				// fmt.Printf("%s\n", cnt)
				e.MsgContent = valueMapper(cnt, e.Timestamp)
			}

		}
	}
	echo, err := json.Marshal(&e)

	if err != nil {
		log.Panicf("Cannot be parsed to json")
	}
	log.Println(string(echo))

	return &e
}

func (f *syslogServerFrontend) parseContent(content string) (map[string]interface{}, error) {
	kv, err := kvparser(content)
	if err != nil {
		return nil, err
	}
	return kv, nil
}

func kvparser(input string) (kv map[string]interface{}, err error) {
	// fmt.Println("---source----", input)
	ent := strings.Split(input, ",")
	// fmt.Printf("--initiation---lenghth: %d, %s\n", len(ent), ent)
	data := make(map[string]interface{})
	for e := range ent {
		set := strings.Split(ent[e], ": ")
		if len(set) == 2 {
			if strings.Index(set[0], "error") == -1 {
				k := strings.TrimPrefix(set[0], " ")
				v := strings.TrimPrefix(set[1], " ")
				data[k] = v
			}
		}
	}
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	return data, nil
}

func valueMapper(dataParts map[string]interface{}, ts time.Time) api.LogMessage {
	c := api.LogMessage{}
	c.IPLocation = "PAQUES"
	c.EventID = ts.UnixNano()
	c.Time = ts.UnixNano() / 1000000000
	if val, ok := dataParts["SrcIP"].(string); ok {
		c.SrcIP = val
	}
	if val, ok := dataParts["MAC"].(string); ok {
		c.MAC = val
	}
	if val, ok := dataParts["URL"].(string); ok {
		c.URL = val
	}
	return c
}
