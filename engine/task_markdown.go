package engine

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

//go:embed js/markdown.js
var markdownJS string

type MarkdownResult struct {
	Content string `json:"content"`
}

func performMarkdownTask(ctx context.Context, task *Task, logger *zerolog.Logger) (MarkdownResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var markdown string

	includeImages := task.BoolParam("include-images", false)

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		logger.Debug().Msgf("markdown err: %v", err)

		return MarkdownResult{}, err
	}

	// Wait for HTML content to stabilize before running markdown script
	stabilizationErr := waitForHTMLStabilization(ctx, logger)
	if stabilizationErr != nil {
		logger.Warn().Err(stabilizationErr).Msg("error during HTML stabilization check, continuing anyway")
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(markdownJS, includeImages), &markdown),
	)
	if err != nil {
		logger.Debug().Msgf("markdown err: %v", err)

		return MarkdownResult{}, err
	}

	return MarkdownResult{Content: markdown}, nil
}
