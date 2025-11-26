package linkpreview

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/chromedp/chromedp"
	"go.chimbori.app/butterfly/conf"
)

var ErrMissingSelector = errors.New("selector not found")

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
	js := fmt.Sprintf(`(function() {
		var el = document.querySelector(%s);
		if (el) {
			el.style.visibility = '';
			el.style.display = 'block';
			return true;
		}
		return false;
	})()`, strconv.Quote(selector))

	var foundSelector bool
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(js, &foundSelector),
	); err != nil {
		return nil, err
	}
	if !foundSelector {
		return nil, ErrMissingSelector
	}

	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.ScreenshotScale(selector, 2.0, &buf, chromedp.NodeVisible),
	); err != nil {
		return nil, err
	}

	return buf, nil
}

// takeScreenshotWithTemplate renders a provided HTML template with the given title and description,
// and then takes a screenshot of the result. The template is expected to contain <div>s with IDs “title” and “description”.
func takeScreenshotWithTemplate(ctx context.Context, url, templateContent, title, description string) ([]byte, error) {
	var cancel context.CancelFunc
	if conf.Config.Debug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	dataUrl := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(templateContent))
	selector := "#link-preview"

	js := fmt.Sprintf(`
		document.getElementById('title').innerText = %s;
		document.getElementById('description').innerText = %s;
		var el = document.querySelector('%s');
		if (el) {
			el.style.visibility = '';
			el.style.display = 'block';
		}
	`, strconv.Quote(title), strconv.Quote(description), selector)

	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(dataUrl),
		chromedp.Evaluate(js, nil),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.ScreenshotScale(selector, 2.0, &buf, chromedp.NodeVisible),
	); err != nil {
		return nil, err
	}

	return buf, nil
}

// fetchTitleAndDescription retrieves the title and description from a web page.
// OpenGraph tags are preferred (og:title, og:description), but document title is used as a fallback.
func fetchTitleAndDescription(ctx context.Context, url string) (title, description string, err error) {
	var cancel context.CancelFunc
	if conf.Config.Debug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	var ogTitle, ogDesc, docTitle string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Text("title", &docTitle, chromedp.ByQuery),
		chromedp.AttributeValue(`meta[property="og:title"]`, "content", &ogTitle, nil, chromedp.ByQuery),
		chromedp.AttributeValue(`meta[property="og:description"]`, "content", &ogDesc, nil, chromedp.ByQuery),
	)
	if err != nil {
		return "", "", err
	}

	if ogTitle != "" {
		title = ogTitle
	} else {
		title = docTitle
	}
	description = ogDesc
	return title, description, nil
}
