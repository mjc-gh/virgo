package engine

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type PlaintextResult struct {
	Content string
}

func performPlaintextTask(ctx context.Context, task *Task, logger *zerolog.Logger) (PlaintextResult, error) {
	var plaintext string

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		chromedp.Text("main", &plaintext, chromedp.ByQuery),
	)
	if err != nil {
		logger.Debug().Msgf("plaintext err: %v", err)

		return PlaintextResult{}, err
	}

	return PlaintextResult{plaintext}, nil
}
