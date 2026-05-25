package engine

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mjc-gh/virgo/internal/browser"
	"github.com/rs/zerolog"
)

type config struct {
	concurrency int
	remoteURL   string
	headfull    bool
}

type Engine struct {
	browserCancel context.CancelFunc
	config        config
	logger        *zerolog.Logger
	results       chan Result
	tasks         chan Task
	wg            sync.WaitGroup
}

type Option func(*Engine)

func WithRemoteAllocator(host string, port int) Option {
	return func(e *Engine) {
		host := net.JoinHostPort(host, strconv.Itoa(port))
		e.config.remoteURL = fmt.Sprintf("http://%s/json/version", host)
	}
}

func WithHeadfullLocalAllocator() Option {
	return func(e *Engine) {
		e.config.headfull = true
	}
}

func WithLogger(l *zerolog.Logger) Option {
	return func(e *Engine) {
		e.logger = l
	}
}

func New(concurrency int, opts ...Option) *Engine {
	if concurrency < 1 {
		concurrency = 1
	}

	e := Engine{
		config: config{
			concurrency: concurrency,
		},
		results: make(chan Result),
		tasks:   make(chan Task, concurrency),
		wg:      sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(&e)
	}

	if e.logger == nil {
		l := zerolog.New(os.Stderr).With().Timestamp().Logger()
		e.logger = &l
	}

	return &e
}

func (e *Engine) Start(ctx context.Context) {
	if e.config.remoteURL != "" {
		ctx, e.browserCancel = browser.StartRemote(ctx, e.config.remoteURL)
	} else {
		ctx, e.browserCancel = browser.StartLocal(ctx, e.config.headfull)
	}

	for i := range e.config.concurrency {
		e.wg.Add(1)

		go func(idx int, tasks <-chan Task, results chan<- Result, done func(), logger *zerolog.Logger) {
			logger.Debug().Msgf("task worker #%d started", idx)
			defer done()
			defer logger.Debug().Msgf("task worker #%d stopped", idx)

			for task := range tasks {
				logger.Debug().Msgf("task worker #%d got task", idx)

				r := performTask(ctx, &task, logger)

				// If the task its own buffered channel, send
				// the result there. If no channel was provided,
				// we'll use the unbuffered results channel
				// provided by the engine.
				if task.resultCh != nil {
					task.resultCh <- r
				} else {
					results <- r
				}

				logger.Debug().Msgf("task worker #%d sent result", idx)
			}
		}(i+1, e.tasks, e.results, e.wg.Done, e.logger)
	}
}

func (e *Engine) Shutdown() {
	e.logger.Debug().Msg("shutdown called")
	defer e.logger.Debug().Msg("shutdown done")

	if e.browserCancel != nil {
		defer e.browserCancel()
	}

	close(e.tasks)
	e.wg.Wait()
	close(e.results)
}

func (e *Engine) Results() <-chan Result {
	return e.results
}

func (e *Engine) Add(t Task) {
	e.tasks <- t
}

// waitForHTMLStabilization polls the HTML content until it stabilizes or timeout occurs.
// It considers content stabilized when the length change between consecutive checks
// is less than 1%. It polls for up to 5 seconds before returning regardless of stability.
func waitForHTMLStabilization(ctx context.Context, logger *zerolog.Logger) error {
	const (
		pollInterval       = 200 * time.Millisecond
		maxWaitTime        = 5 * time.Second
		stabilityThreshold = 0.01 // 1%
	)

	startTime := time.Now()
	var previousLength int

	for {
		var htmlContent string
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlContent),
		)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to evaluate HTML content during stabilization check")

			return err
		}

		currentLength := len(htmlContent)

		// On first check, just record the length
		if previousLength == 0 {
			previousLength = currentLength
		} else {
			// Calculate percentage change
			change := float64(currentLength-previousLength) / float64(previousLength)
			changePercent := math.Abs(change)

			logger.Debug().Msgf("HTML stabilization check: length %d (change: %.2f%%)", currentLength, changePercent*100)

			// If content has stabilized (change < 1%), we can return early
			if changePercent < stabilityThreshold {
				logger.Debug().Msg("HTML content stabilized")

				return nil
			}

			previousLength = currentLength
		}

		// Check if we've exceeded the timeout
		if time.Since(startTime) >= maxWaitTime {
			logger.Debug().Msg("HTML stabilization timeout reached (5s)")

			return nil
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}
