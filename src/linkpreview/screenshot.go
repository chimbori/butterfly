package linkpreview

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chimbori.dev/butterfly/conf"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

var ErrMissingSelector = errors.New("selector not found")

// takeScreenshot captures a high-resolution PNG screenshot of a specific element on a web page.
// It navigates to the provided URL, ensures the element specified by the CSS selector is visible,
// and takes a screenshot at 2x scale.
func takeScreenshot(ctx context.Context, url, selector string) (png []byte, err error) {
	var cancel context.CancelFunc
	if conf.Config.Debug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithErrorf(log.Printf))
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
		chromedp.ScreenshotScale(selector, 2.0, &buf),
	); err != nil {
		return nil, err
	}

	return buf, nil
}

// takeScreenshotWithTemplate renders a provided HTML template with the given title and description,
// and then takes a screenshot of the result. The template is parsed as a Golang template, with fields
// `{{.Title}}`, `{{.Description}}`, and `{{.Url}}`.
func takeScreenshotWithTemplate(ctx context.Context, templateContent, url, title, description string) ([]byte, error) {
	var cancel context.CancelFunc
	if conf.Config.Debug {
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithErrorf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(ctx)
	}
	defer cancel()

	tmpl, err := template.New("linkpreview").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var tmplBuf bytes.Buffer
	if err := tmpl.Execute(&tmplBuf, struct {
		Title       string
		Description string
		Url         string
	}{
		Title:       title,
		Description: description,
		Url:         url,
	}); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	var screenshotBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.EmulateViewport(960, 960),
		chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString(tmplBuf.Bytes())),
		chromedp.WaitVisible("#link-preview", chromedp.ByQuery),
		chromedp.Sleep(time.Second), // Allow fonts to finish downloading.
		chromedp.ScreenshotScale("#link-preview", 2.0, &screenshotBuf),
	); err != nil {
		return nil, err
	}

	return screenshotBuf, nil
}

// fetchTitleAndDescription retrieves the title and description from a web page.
// OpenGraph tags are preferred (og:title, og:description), but document title is used as a fallback.
func fetchTitleAndDescription(ctx context.Context, url string) (title, description string, err error) {
	var doc *html.Node

	// Handle data URIs directly
	if strings.HasPrefix(url, "data:text/html,") {
		htmlContent := strings.TrimPrefix(url, "data:text/html,")
		doc, err = html.Parse(strings.NewReader(htmlContent))
		if err != nil {
			return "", "", err
		}
	} else {
		// Handle HTTP(S) URLs
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Butterfly/1.0; +https://butterfly.chimbori.com/)")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		doc, err = html.Parse(resp.Body)
		if err != nil {
			return "", "", err
		}
	}

	var ogTitle, ogDesc, docTitle string
	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && docTitle == "" {
				if n.FirstChild != nil {
					docTitle = n.FirstChild.Data
				}
			} else if n.Data == "meta" {
				var property, content string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "property":
						property = attr.Val
					case "content":
						content = attr.Val
					}
				}
				if property == "og:title" && ogTitle == "" {
					ogTitle = content
				} else if property == "og:description" && ogDesc == "" {
					ogDesc = content
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}
	parse(doc)

	if ogTitle != "" {
		title = ogTitle
	} else {
		title = docTitle
	}
	description = ogDesc
	return strings.TrimSpace(title), strings.TrimSpace(description), nil
}
