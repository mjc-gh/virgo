package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog"
)

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

	search := task.StringParam("search", "")
	threshold := task.IntParam("threshold", 3)

	// Use a simpler approach - just return an array, let chromedp handle JSON marshalling
	var linksJSONStr string

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(extractLinksJS, &linksJSONStr),
	)
	if err != nil {
		logger.Debug().Msgf("links err: %v", err)

		return LinksResult{}, err
	}

	// Debug logging
	logger.Debug().Msgf("links raw result: %q (len=%d)", linksJSONStr, len(linksJSONStr))

	// The result from Evaluate should be the JSON string
	var links []linkData
	err = json.Unmarshal([]byte(linksJSONStr), &links)
	if err != nil {
		// Maybe the result is wrapped in quotes by chromedp
		var wrapped string
		errWrapped := json.Unmarshal([]byte(linksJSONStr), &wrapped)
		if errWrapped == nil {
			logger.Debug().Msgf("unwrapped string: %q", wrapped)
			err = json.Unmarshal([]byte(wrapped), &links)
			if err != nil {
				logger.Debug().Msgf("links unmarshal error after unwrapping: %v", err)

				return LinksResult{}, fmt.Errorf("unmarshal links: %w", err)
			}
		} else {
			logger.Debug().Msgf("links unmarshal error: %v", err)

			return LinksResult{}, fmt.Errorf("unmarshal links: %w", err)
		}
	}

	logger.Debug().Msgf("parsed %d links", len(links))

	var results []string
	for _, link := range links {
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

	return LinksResult{Content: strings.Join(results, "\n")}, nil
}

const extractLinksJS = `
(function() {
  const links = [];
  const anchorElements = document.querySelectorAll('a');
  
  for (const anchor of anchorElements) {
    const href = anchor.getAttribute('href');
    const text = anchor.textContent.trim();
    
    // Skip anchors without href or text
    if (href && text) {
      links.push({ text, href });
    }
  }
  
  return JSON.stringify(links);
})()
`
