package loki

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap/zapcore"
)

type Client interface {
	Enqueue(LogRecord)
}

type LogRecord struct {
	Time       time.Time
	Level      string
	Path       string
	Method     string
	Status     int
	DurationMs int64
	RequestID  string
	UserID     string
	TraceID    string
	Message    string
	Error      string
	URL        string
	Query      map[string][]string
	Headers    map[string][]string
	Body       string
	BodyTrunc  bool
	RemoteIP   string
	UserAgent  string
}

type clientImpl struct {
	cfg   *config.LokiConfig
	log   logger.Logger
	queue chan LogRecord
	httpc *http.Client
}

func InitLokiClient(cfg *config.LokiConfig, log logger.Logger) Client {
	if cfg == nil || cfg.URL == "" {
		log.Warn("loki client disabled: missing config")
		return &noop{}
	}
	c := &clientImpl{
		cfg:   cfg,
		log:   log,
		queue: make(chan LogRecord, cfg.MaxQueue),
		httpc: &http.Client{Timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond},
	}
	go c.run()
	return c
}

// Enqueue adds a log record to the queue, drops if full
func (c *clientImpl) Enqueue(r LogRecord) {
	select {
	case c.queue <- r:
	default:
		c.log.Warn("loki queue full, drop log",
			zapcore.Field{Key: "path", String: r.Path},
			zapcore.Field{Key: "request_id", String: r.RequestID},
		)
	}
}

// worker
func (c *clientImpl) run() {
	batch := make([]LogRecord, 0, c.cfg.BatchSize)
	flush := time.NewTicker(time.Duration(c.cfg.FlushMs) * time.Millisecond)
	defer flush.Stop()

	for {
		select {
		case r := <-c.queue:
			batch = append(batch, r)
			if len(batch) >= c.cfg.BatchSize {
				c.flush(batch)
				batch = batch[:0]
			}
		case <-flush.C:
			if len(batch) > 0 {
				c.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// send logs to loki with retries
func (c *clientImpl) flush(records []LogRecord) {
	streams := make(map[string][][2]string) // key = labels json
	for _, r := range records {
		lbl := map[string]string{
			"service": c.cfg.Service,
			"env":     c.cfg.Env,
			"level":   r.Level,
			"path":    r.Path,
			"method":  r.Method,
		}
		ljson, _ := json.Marshal(lbl)
		line, _ := json.Marshal(r)
		streams[string(ljson)] = append(streams[string(ljson)], [2]string{
			timeToNs(r.Time), string(line),
		})
	}

	payload := struct {
		Streams []struct {
			Stream map[string]string `json:"stream"`
			Values [][2]string       `json:"values"`
		} `json:"streams"`
	}{}

	for lblJSON, vals := range streams {
		lbl := map[string]string{}
		json.Unmarshal([]byte(lblJSON), &lbl)
		payload.Streams = append(payload.Streams, struct {
			Stream map[string]string `json:"stream"`
			Values [][2]string       `json:"values"`
		}{Stream: lbl, Values: vals})
	}

	// retry logic
	body, _ := json.Marshal(payload)
	for i := 0; i < c.cfg.Retry; i++ {
		req, _ := http.NewRequest("POST", c.cfg.URL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpc.Do(req)
		if err == nil && resp.StatusCode/100 == 2 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
	}

	c.log.Warn("loki flush failed after retries",
		zapcore.Field{Key: "records", Integer: int64(len(records))},
	)
}

// convert time to nanoseconds string for loki
func timeToNs(t time.Time) string {
	return strconv.FormatInt(t.UnixNano(), 10)
}

type noop struct{}

func (n *noop) Enqueue(LogRecord) {}
