package api

import (
	"time"
)

type LogEntry struct {
	Timestamp   time.Time  `json:"timestamp,omitempty"`
	Hostname    string     `json:"hostname,omitempty"`
	Application string     `json:"application,omitempty"`
	Message     string     `json:"message,omitempty"`
	MsgContent  LogMessage `json:"msg_content"`
}

type LogMessage struct {
	EventID    int64  `json:"eventid,omitempty"`
	SrcIP      string `json:"ipaddr,omitempty"`
	IPLocation string `json:"iplocation,omitempty"`
	MAC        string `json:"macaddr,omitempty"`
	URL        string `json:"url,omitempty"`
	Time       int64  `json:"time,omitempty"`
}

type QueryRequest struct {
	FromTimestamp time.Time
	ToTimestamp   time.Time
	Hostname      string
	Application   string
	Message       string
	Limit         int
	Offset        int
}

type QueryStatResponse struct {
	Stat  map[string]map[string]uint64 `json:",omitempty"`
	Error string                       `json:",omitempty"`
}

type QueryListResponse struct {
	Entries []*LogEntry `json:",omitempty"`
	Error   string      `json:",omitempty"`
}

type InsertRequest struct {
	Entry   *LogEntry
	Entries []*LogEntry
}

type InsertResponse struct {
	Error string `json:",omitempty"`
}
