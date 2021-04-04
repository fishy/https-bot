package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/runtimebp"
	yaml "gopkg.in/yaml.v2"
)

var (
	configPath = flag.String(
		"config",
		"/data/config.yaml",
		"path to the config file",
	)
)

const (
	defaultThreshold = 0.95
	defaultLimit     = 1024 * 10
)

type config struct {
	Threshold *float64 `yaml:"similarity_threshold"`
	Limit     int64    `yaml:"read_limit"`

	HN struct {
		Username     string        `yaml:"username"`
		Password     string        `yaml:"password"`
		Timeout      time.Duration `yaml:"timeout"`
		ReplyTimeout time.Duration `yaml:"reply_timeout"`
		Interval     time.Duration `yaml:"interval"`
		Workers      int           `yaml:"workers"`
	} `yaml:"hn"`
}

func main() {
	flag.Parse()
	log.InitLogger(log.InfoLevel)

	cfg := parseConfig(*configPath)
	if cfg.Threshold == nil {
		t := defaultThreshold
		cfg.Threshold = &t
	}
	if cfg.Limit <= 0 {
		cfg.Limit = defaultLimit
	}

	go func() {
		// for pprof
		http.ListenAndServe("localhost:6060", nil)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go runtimebp.HandleShutdown(
		context.Background(),
		func(signal os.Signal) {
			defer wg.Done()
			log.Infow("shutting down...", "signal", signal)
			cancel()
		},
	)

	wg.Add(1)
	go hnMain(ctx, &wg, cfg)

	wg.Wait()
}

func parseConfig(path string) config {
	var cfg config
	f, err := os.Open(path)
	if err != nil {
		log.Fatalw("Cannot open config file", "err", err, "path", path)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	decoder.SetStrict(true)
	if err := decoder.Decode(&cfg); err != nil {
		log.Fatalw("Cannot parse config file", "err", err, "path", path)
	}
	return cfg
}
