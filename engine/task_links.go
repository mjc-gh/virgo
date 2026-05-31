package engine

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog"
)

//go:embed js/links.js
var linksJS string

type LinksResult struct {
	Content string `json:"content"`
}

type linkData struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

func performLinksTask(ctx context.Context, task *Task, logger *zerolog.Logger) (LinksResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	linksJSONStr, err := extractLinksJSON(ctx, task, logger)
	if err != nil {
		return LinksResult{}, err
	}

	links, err := parseLinksJSON(linksJSONStr, logger)
	if err != nil {
		return LinksResult{}, err
	}

	logger.Debug().Msgf("parsed %d links", len(links))

	search := task.StringParam("search", "")
	threshold := task.IntParam("threshold", 3)
	content := filterAndFormatLinks(links, search, threshold)

	return LinksResult{Content: content}, nil
}

func extractLinksJSON(ctx context.Context, task *Task, logger *zerolog.Logger) (string, error) {
	var linksJSONStr string

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		logger.Debug().Msgf("links err: %v", err)

		return "", err
	}

	// Wait for HTML content to stabilize before running markdown script
	stabilizationErr := waitForHTMLStabilization(ctx, logger)
	if stabilizationErr != nil {
		logger.Warn().Err(stabilizationErr).Msg("error during HTML stabilization check, continuing anyway")
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(linksJS, &linksJSONStr),
	)
	if err != nil {
		logger.Debug().Msgf("links err: %v", err)

		return "", err
	}

	logger.Debug().Msgf("links raw result: %q (len=%d)", linksJSONStr, len(linksJSONStr))

	return linksJSONStr, nil
}

func parseLinksJSON(linksJSONStr string, logger *zerolog.Logger) ([]linkData, error) {
	var links []linkData
	err := json.Unmarshal([]byte(linksJSONStr), &links)
	if err != nil {
		// Maybe the result is wrapped in quotes by chromedp
		var wrapped string
		errWrapped := json.Unmarshal([]byte(linksJSONStr), &wrapped)
		if errWrapped == nil {
			logger.Debug().Msgf("unwrapped string: %q", wrapped)
			err = json.Unmarshal([]byte(wrapped), &links)
			if err != nil {
				logger.Debug().Msgf("links unmarshal error after unwrapping: %v", err)

				return nil, fmt.Errorf("unmarshal links: %w", err)
			}
		} else {
			logger.Debug().Msgf("links unmarshal error: %v", err)

			return nil, fmt.Errorf("unmarshal links: %w", err)
		}
	}

	return links, nil
}

func filterAndFormatLinks(links []linkData, search string, threshold int) string {
	var results []string
	for _, link := range links {
		// Skip links that resolve to an empty location hash
		if link.Href == "#" {
			continue
		}

		// If no search term, include all links; otherwise apply fuzzy matching
		if search == "" {
			results = append(results, fmt.Sprintf("- [%s](%s)", link.Text, link.Href))
		} else {
			rank := fuzzy.RankMatchNormalizedFold(search, link.Text)
			if rank >= 0 && rank <= threshold {
				results = append(results, fmt.Sprintf("- [%s](%s)", link.Text, link.Href))
			}
		}
	}

	return strings.Join(results, "\n")
}
