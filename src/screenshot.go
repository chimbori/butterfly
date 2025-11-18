package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
	"go.chimbori.app/butterfly/conf"
)

// takeScreenshot captures a high-resolution PNG screenshot of a specific element on a web page.
// It navigates to the provided URL, ensures the element specified by the CSS selector is visible,
// and takes a screenshot at 2x scale.
func takeScreenshot(ctx context.Context, url, selector string) (png []byte, err error) {
	var cancel context.CancelFunc
	if conf.Config.Debug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
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
