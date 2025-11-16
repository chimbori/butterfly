package main

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

func takeScreenshot(ctx context.Context, url, selector string) (png []byte, err error) {
	ctx, cancel := chromedp.NewContext(
		ctx,
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	if selector == "" {
		return nil, fmt.Errorf("missing selector")
	}

	// Un-hide the selected element before attempting a screenshot.
	js := "var el=document.querySelector('" + selector + "');el.style.visibility='';el.style.display='block';"

	var res string
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(js, &res),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.ScreenshotScale(selector, 2.0, &buf, chromedp.NodeVisible),
	); err != nil {
		return nil, err
	}

	return buf, nil
}
