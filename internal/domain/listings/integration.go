package listings

import "context"

type FundaAPIClient interface {
	GetHTMLContent(ctx context.Context, URL string) ([]byte, error)
}
