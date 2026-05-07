package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

func StartLocal(ctx context.Context, headfull bool) (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("block-new-web-contents", true),
	}

	if !headfull {
		opts = append(opts, chromedp.Headless)
	}

	return chromedp.NewExecAllocator(ctx, opts...)
}

func StartRemote(ctx context.Context, url string) (context.Context, context.CancelFunc) {
	return chromedp.NewRemoteAllocator(ctx, url)
}
