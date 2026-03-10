package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SearXNGClient struct {
	client           *http.Client
	baseURL          string
	resultsPerSearch int
}

func NewSearXNGClient(baseurl string, resultsPerSearch int) *SearXNGClient {
	return &SearXNGClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:          baseurl,
		resultsPerSearch: resultsPerSearch,
	}
}

func (s *SearXNGClient) Search(ctx context.Context, query string) (string, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")

	fullUrl := s.baseURL + "/search?" + params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fullUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpRes, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to fetch response: %w", err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", httpRes.StatusCode)
	}
	body := io.LimitReader(httpRes.Body, 1<<20)

	var response struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Results) == 0 {
		return "", fmt.Errorf("no results found for query: %s", query)
	}

	results := response.Results

	if len(results) > s.resultsPerSearch {
		results = results[:s.resultsPerSearch]
	}

	var sb strings.Builder

	for i, res := range results {
		fmt.Fprintf(&sb, "%d. %s\n%s\n%s\n\n", i+1, res.Title, res.URL, res.Content)
	}

	return sb.String(), nil
}
