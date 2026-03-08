package search

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const ddgLiteURL = "https://lite.duckduckgo.com/lite/"
const maxResults = 5

type DuckDuckGoClient struct {
	client *http.Client
}

func NewDuckDuckGoClient() *DuckDuckGoClient {
	return &DuckDuckGoClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

func (d *DuckDuckGoClient) Search(ctx context.Context, query string) (string, error) {
	reader, err := d.fetch(ctx, query)
	if err != nil {
		return "", err
	}

	searchResults, err := d.parse(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse ddg response: %w", err)
	}

	if len(searchResults) == 0 {
		return "", fmt.Errorf("no results found for query: %s", query)
	}

	formattedResult := d.formatResults(searchResults, query)

	return formattedResult, nil
}

func (d *DuckDuckGoClient) fetch(ctx context.Context, query string) (io.Reader, error) {
	postData := url.Values{
		"q": {query},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ddgLiteURL,
		strings.NewReader(postData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://duckduckgo.com")

	res, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ddg response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", res.StatusCode)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return bytes.NewReader(responseBody), nil
}

func (d *DuckDuckGoClient) parse(body io.Reader) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	doc.Find("a.result-link").Slice(0, maxResults).Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())

		link, exists := s.Attr("href")
		if !exists {
			return
		}
		snippet := s.Closest("tr").Next().Find(".result-snippet").Text()

		results = append(results, SearchResult{
			Title:   title,
			URL:     link,
			Snippet: snippet,
		})
	})

	return results, nil
}

func (d *DuckDuckGoClient) formatResults(searchResults []SearchResult, query string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Search results for query: %s\n\n", query)

	for i, result := range searchResults {
		fmt.Fprintf(&sb, "%d. %s\n%s\n%s\n\n", i+1, result.Title, result.URL, result.Snippet)
	}
	return sb.String()
}
