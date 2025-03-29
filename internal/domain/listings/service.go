package listings

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"net/url"
	"strconv"
	"strings"
)

type Service struct {
	repository     Repository
	fundaAPIClient FundaAPIClient
	log            *zerolog.Logger
}

func NewService(
	repository Repository,
	fundaAPIClient FundaAPIClient,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repository:     repository,
		fundaAPIClient: fundaAPIClient,
		log:            log,
	}
}

func (s *Service) Reset(ctx context.Context) error {
	s.log.Debug().Msg("running listings.Reset")

	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	if err = s.repository.TruncateSearchQueryTable(ctx, tx); err != nil {
		s.log.Error().Err(err).Msg("failed to truncate search query table")
		return fmt.Errorf("failed to truncate search query table: %w", err)
	}

	if err = s.repository.TruncateListingsTable(ctx, tx); err != nil {
		s.log.Error().Err(err).Msg("failed to truncate listings table")
		return fmt.Errorf("failed to truncate listings table: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) ResetAndUpdate(ctx context.Context, URL string) error {
	s.log.Debug().Msg("running listings.ResetAndUpdate")

	if err := s.Reset(ctx); err != nil {
		s.log.Error().Err(err).Msg("failed to reset the database")
		return fmt.Errorf("failed to reset the database: %w", err)
	}

	if err := s.repository.CreateSearchQuery(ctx, URL); err != nil {
		s.log.Error().Err(err).Msg("failed to set a new search query")
		return fmt.Errorf("failed to set a new search query: %w", err)
	}

	return nil
}

func (s *Service) UpdateAndCompareListings(ctx context.Context) (addedListings, removedListings Listings, err error) {
	s.log.Debug().Msg("running listings.UpdateAndCompareListings")

	currentListings, err := s.repository.GetAllCurrentListings(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get current listings")
		return nil, nil, fmt.Errorf("failed to get current listings: %w", err)
	}

	newListings, err := s.GetNewListings(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get new listings")
		return nil, nil, fmt.Errorf("failed to get new listings: %w", err)
	}

	if err = s.repository.MUpdateListings(ctx, newListings); err != nil {
		s.log.Error().Err(err).Msg("failed to update listings in DB")
		return nil, nil, fmt.Errorf("failed to update listings in DB: %w", err)
	}

	removedListings = currentListings.CompareAndGetRemovedListings(newListings)
	addedListings = newListings.CompareAndGetAddedListings(currentListings)
	return addedListings, removedListings, nil
}

func (s *Service) GetNewListings(ctx context.Context) (Listings, error) {
	s.log.Debug().Msg("running listings.GetNewListings")

	currentURL, err := s.repository.GetCurrentSearchQuery(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get current search query")
		return nil, fmt.Errorf("failed to get current search query: %w", err)
	}

	parsedURL, err := url.Parse(currentURL)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to parse current search query")
		return nil, fmt.Errorf("failed to parse current search query: %w", err)
	}

	var (
		pageNumber   = defaultStartPageNumber
		htmlContent  []byte
		doc          *goquery.Document
		endOfSearch  bool
		listingItems = make([]ListingItem, 0, defaultCapacity)
		queryParams  = parsedURL.Query()
	)

	for {
		// set pagination and retrieve HTML content
		queryParams.Set("search_result", strconv.Itoa(pageNumber))
		parsedURL.RawQuery = queryParams.Encode()
		htmlContent, err = s.fundaAPIClient.GetHTMLContent(ctx, parsedURL.String())
		if err != nil {
			s.log.Error().Err(err).Msg("failed to load HTML content while getting listing items")
			return nil, fmt.Errorf("failed to load HTML content while getting listing items: %w", err)
		}

		// transform to goquery.Document
		reader := bytes.NewReader(htmlContent)
		doc, err = goquery.NewDocumentFromReader(reader)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to parse HTML content while getting listing items")
			return nil, fmt.Errorf("failed to parse HTML content while getting listing items: %w", err)
		}

		// trigger end-of-search by full-text search, since PAI does not provide 404 for such cases
		doc.Find("body *").Each(func(i int, selection *goquery.Selection) {
			text := strings.TrimSpace(selection.Text())
			if strings.Contains(text, pageNotFoundText) {
				endOfSearch = true
			}
		})
		if endOfSearch {
			break
		}

		// find json object with results
		doc.Find(`script[type="application/ld+json"][data-hid="result-list-metadata"]`).Each(func(i int, selection *goquery.Selection) {
			jsonText := selection.Text()
			var listingSearchList ListingSearchList
			err = json.Unmarshal([]byte(jsonText), &listingSearchList)
			if err != nil {
				s.log.Warn().Err(err).Msg("failed to parse listings search list")
				return
			}
			listingItems = append(listingItems, listingSearchList.ItemListElement...)
		})

		// increment pagination
		pageNumber++
	}

	listings := make(Listings, 0, len(listingItems))
	for idx := range listingItems {
		var listing *Listing
		listing, err = s.GetListing(ctx, listingItems[idx].URL)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to get listing while retrieving detailed data")
			continue
		}
		listings = append(listings, *listing)
	}

	return listings, nil
}

func (s *Service) GetListing(ctx context.Context, URL string) (*Listing, error) {
	s.log.Debug().Msg("running listings.GetListing")

	var (
		err         error
		htmlContent []byte
		doc         *goquery.Document
	)

	htmlContent, err = s.fundaAPIClient.GetHTMLContent(ctx, URL)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to load HTML content while getting detailed listing")
		return nil, fmt.Errorf("failed to load HTML content while getting detailed listing: %w", err)
	}

	// transform to goquery.Document
	reader := bytes.NewReader(htmlContent)
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to parse HTML content while getting detailed listing")
		return nil, fmt.Errorf("failed to parse HTML content while getting detailed listing: %w", err)
	}

	// find json object with results
	var listing Listing
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, selection *goquery.Selection) {
		jsonText := selection.Text()
		err = json.Unmarshal([]byte(jsonText), &listing)
		if err != nil {
			s.log.Warn().Err(err).Msg("failed to parse detailed listing")
			return
		}
	})

	return &listing, nil
}
