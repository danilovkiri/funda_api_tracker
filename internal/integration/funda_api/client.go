package funda_api

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"net/http"
)

const (
	name = "Funda API client"
)

type FundaAPIClient struct { //nolint:golint
	cfg    *Config
	client *resty.Client
	log    *zerolog.Logger
}

func NewFundaAPIClient(
	cfg *Config,
	log *zerolog.Logger,
) *FundaAPIClient {
	log.Info().Msg(fmt.Sprintf("initializing %s", name))
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	return &FundaAPIClient{
		cfg:    cfg,
		client: client,
		log:    log,
	}
}

func (c *FundaAPIClient) GetHTMLContent(ctx context.Context, URL string) ([]byte, error) {
	resp, err := c.client.R().SetContext(ctx).SetHeader("referer", URL).SetHeaders(provideHeaders()).Get(URL)
	if err != nil {
		c.log.Error().Err(err).Str("client", name).Msg("failed to execute request")
		return nil, fmt.Errorf("failed to execute request in %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		c.log.Warn().Str("client", name).Msg(fmt.Sprintf("got response code %d", resp.StatusCode()))
		return nil, fmt.Errorf("got response code %d from %s", resp.StatusCode(), name)
	}

	return resp.Body(), nil
}

func provideHeaders() map[string]string {
	return map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":           "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"cache-control":             "max-age=0",
		"dnt":                       "\"656086674\"",
		"priority":                  "u=0, i",
		"sec-ch-ua":                 "\"Chromium\";v=\"134\", \"Not:A-Brand\";v=\"24\", \"Google Chrome\";v=\"134\"",
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "\"macOS\"",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	}
}
