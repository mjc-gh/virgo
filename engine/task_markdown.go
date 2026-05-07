package engine

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type MarkdownResult struct {
	Content string
}

func performMarkdownTask(ctx context.Context, task *Task, logger *zerolog.Logger) (MarkdownResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
	)
	if err != nil {
		logger.Debug().Msgf("markdown err: %v", err)

		return MarkdownResult{}, err
	}

	return MarkdownResult{""}, nil
}
