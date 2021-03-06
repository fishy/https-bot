package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/randbp"

	"github.com/fishy/https-bot/internal/check"
	"github.com/fishy/https-bot/internal/hnapi"
)

const (
	defaultHNTimeout      = time.Second
	defaultHNReplyTimeout = time.Second * 10
	defaultHNInterval     = time.Minute
	defaultHNWorkers      = 1
)

func hnMain(ctx context.Context, wg *sync.WaitGroup, cfg config) {
	defer wg.Done()

	if cfg.HN.Timeout <= 0 {
		cfg.HN.Timeout = defaultHNTimeout
	}
	if cfg.HN.ReplyTimeout <= 0 {
		cfg.HN.ReplyTimeout = defaultHNReplyTimeout
	}
	if cfg.HN.Workers <= 0 {
		cfg.HN.Workers = defaultHNWorkers
	}
	session := func(ctx context.Context) *hnapi.Session {
		ctx, cancel := context.WithTimeout(ctx, cfg.HN.Timeout)
		defer cancel()
		session, err := hnapi.NewSession(ctx, cfg.HN.Username, cfg.HN.Password)
		if err != nil {
			log.Fatalw("failed to login to hn", "err", err)
		}
		return session
	}(ctx)

	c := make(chan int64)
	for i := 0; i < cfg.HN.Workers; i++ {
		go hnWorker(ctx, wg, session, cfg, c)
	}

	if cfg.HN.Interval <= 0 {
		cfg.HN.Interval = defaultHNInterval
	}
	ticker := time.NewTicker(cfg.HN.Interval)
	defer ticker.Stop()

	log.Infow(
		"starting hn main ticker...",
		"interval", cfg.HN.Interval.String(),
	)

	lastMax := int64(-1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func(ctx context.Context) {
				var items int
				start := time.Now()
				defer func() {
					took := time.Now().Sub(start)
					logger := log.Debugw
					if took > cfg.HN.Interval/2 || randbp.ShouldSampleWithRate(cfg.HN.TickLogSampleRate) {
						logger = log.Infow
					}
					logger(
						"hn tick done",
						"items", items,
						"took", took.String(),
						"interval", cfg.HN.Interval.String(),
					)
				}()

				max, err := func(ctx context.Context) (int64, error) {
					ctx, cancel := context.WithTimeout(ctx, cfg.HN.Timeout)
					defer cancel()
					return hnapi.MaxItem(ctx)
				}(ctx)
				if err != nil {
					log.Errorw("Failed to get hn max item id", "err", err)
					return
				}
				defer func() {
					lastMax = max
				}()

				if lastMax < 0 {
					return
				}
				for i := max - 1; i >= lastMax; i-- {
					c <- i
					items++
				}
			}(ctx)
		}
	}
}

type result struct {
	oldURL, newURL string
	similarity     float64
}

func hnWorker(ctx context.Context, wg *sync.WaitGroup, session *hnapi.Session, cfg config, c <-chan int64) {
	defer wg.Done()

	self := strings.ToLower(cfg.HN.Username)
	for {
		select {
		case <-ctx.Done():
			return
		case i := <-c:
			func(i int64) {
				start := time.Now()
				defer func() {
					took := time.Now().Sub(start)
					log.Debugw(
						"hnWorker done",
						"i", i,
						"took", took.String(),
					)
				}()

				item, err := func(ctx context.Context) (*hnapi.Item, error) {
					ctx, cancel := context.WithTimeout(ctx, cfg.HN.Timeout)
					defer cancel()
					return hnapi.GetItem(ctx, i)
				}(ctx)
				if err != nil {
					log.Errorw("Failed to get hn item", "err", err, "id", i)
					return
				}
				if item.Deleted || item.Dead || strings.ToLower(item.By) == self {
					return
				}
				var results []*result
				for _, url := range item.URLs() {
					r := func(ctx context.Context, url string) *result {
						ctx, cancel := context.WithTimeout(ctx, cfg.HN.Timeout)
						defer cancel()
						newURL, sim, err := check.Check(ctx, url, cfg.Limit, nil)
						if err != nil {
							if !errors.Is(err, check.ErrNotHTTP) {
								log.Infow("Check failed", "err", err, "url", url)
							}
							return nil
						}
						if sim < *cfg.Threshold {
							return nil
						}
						return &result{
							oldURL:     url,
							newURL:     newURL,
							similarity: sim,
						}
					}(ctx, url)
					if r != nil {
						results = append(results, r)
					}
				}
				if len(results) == 0 {
					return
				}
				func(ctx context.Context, msg string) {
					ctx, cancel := context.WithTimeout(ctx, cfg.HN.ReplyTimeout)
					defer cancel()
					start := time.Now()
					err := session.Reply(ctx, item.ID, msg)
					took := time.Now().Sub(start)
					if err != nil {
						log.Errorw(
							"Failed to send reply",
							"err", err,
							"took", took.String(),
							"id", item.ID,
							"msg", msg,
						)
					} else {
						log.Infow(
							"Successfully replied",
							"took", took.String(),
							"parent", fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID),
						)
					}
				}(ctx, hnMessage(results))
			}(i)
		}
	}
}

func hnMessage(results []*result) string {
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(
			"%s is the HTTPS version of %s you used that works more securely and with %.2f%% similarity on their contents.\n\n",
			r.newURL,
			r.oldURL,
			r.similarity*100,
		))
	}
	sb.WriteString(
		`(I'm a bot, see https://github.com/fishy/https-bot for source code and FAQ)`,
	)
	return sb.String()
}
