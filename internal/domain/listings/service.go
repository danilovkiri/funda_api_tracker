package listings

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	fundaAPIQueryInterval = time.Millisecond * 500
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

func (s *Service) DeleteListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	err := s.repository.DeleteListingsByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to delete listings")
		return fmt.Errorf("failed to delete listings: %w", err)
	}

	return nil
}

func (s *Service) DeleteListingsByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error {
	err := s.repository.DeleteListingsByUserIDAndURLsTx(ctx, tx, userID, URLs)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to delete listings")
		return fmt.Errorf("failed to delete listings: %w", err)
	}

	return nil
}

func (s *Service) GetListingsByUserID(ctx context.Context, userID string, showOnlyNew bool) (Listings, error) {
	listings, err := s.repository.GetListingsByUserID(ctx, userID, showOnlyNew)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get listings")
		return nil, fmt.Errorf("failed to get listings: %w", err)
	}

	return listings, nil
}

func (s *Service) GetListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (Listings, error) {
	listings, err := s.repository.GetListingsByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get listings")
		return nil, fmt.Errorf("failed to get listings: %w", err)
	}

	return listings, nil
}

func (s *Service) GetCurrentlyListedListings(ctx context.Context, searchQuery string) (Listings, error) {
	parsedURL, err := url.Parse(searchQuery)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to parse search query")
		return nil, fmt.Errorf("failed to parse search query: %w", err)
	}

	var (
		pageNumber     = defaultStartPageNumber
		htmlContent    []byte
		doc            *goquery.Document
		emptyPageFound bool
		listingItems   = make([]ListingItem, 0, defaultCapacity)
		queryParams    = parsedURL.Query()
	)

	for {
		// set pagination and retrieve HTML content
		queryParams.Set("search_result", strconv.Itoa(pageNumber))
		parsedURL.RawQuery = queryParams.Encode()
		s.log.Debug().Str("url", parsedURL.String()).Int("page", pageNumber).Msg("fetching new listings for")

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

		// find json object with results
		emptyPageFound = true
		doc.Find(`script[type="application/ld+json"][data-hid="result-list-metadata"]`).Each(func(i int, selection *goquery.Selection) {
			jsonText := selection.Text()
			var listingSearchList ListingSearchList
			err = json.Unmarshal([]byte(jsonText), &listingSearchList)
			if err != nil {
				s.log.Warn().Err(err).Msg("failed to parse listings search list")
				return
			}
			listingItems = append(listingItems, listingSearchList.ItemListElement...)
			emptyPageFound = false
		})

		// break the cycle if no listing items were found previously
		if emptyPageFound {
			s.log.Warn().Str("url", parsedURL.String()).Int("page", pageNumber).Msg("stopping pagination iteration")
			break
		}

		// increment pagination
		pageNumber++

		// sleep
		time.Sleep(fundaAPIQueryInterval)
	}

	// retrieve detailed listing data in parallel
	resultsCh := make(chan *Listing, len(listingItems)) // buffer to prevent blocking
	g, ctx := errgroup.WithContext(ctx)
	for idx := range listingItems {
		time.Sleep(fundaAPIQueryInterval)
		g.Go(func() error {
			listing, gErr := s.GetListing(ctx, listingItems[idx].URL)
			if gErr != nil {
				s.log.Error().Err(gErr).Msg("failed to get listing while retrieving detailed data")
				return gErr
			}
			resultsCh <- listing
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		s.log.Error().Err(err).Msg("failed to fetch new listings in parallel")
		return nil, fmt.Errorf("failed to fetch new listings in parallel: %w", err)
	}
	close(resultsCh)

	listings := make(Listings, 0, len(listingItems))
	for l := range resultsCh {
		listings = append(listings, *l)
	}

	return listings, nil
}

func (s *Service) GetListing(ctx context.Context, URL string) (*Listing, error) {
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

func (s *Service) UpdateAndCompareListings(ctx context.Context, userID, searchQuery string) (addedListings, removedListings, leftoverListings Listings, err error) {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return nil, nil, nil, fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	currentlyListedListings, err := s.GetCurrentlyListedListings(ctx, searchQuery)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get currently listed listings")
		return nil, nil, nil, fmt.Errorf("failed to get currently listed listings: %w", err)
	}

	currentlyStoredListings, err := s.repository.GetListingsByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get currently stored listings")
		return nil, nil, nil, fmt.Errorf("failed to get currently stored listings: %w", err)
	}

	removedListings, leftoverListings = currentlyStoredListings.CompareAndGetRemovedListings(currentlyListedListings)
	addedListings = currentlyListedListings.CompareAndGetAddedListings(currentlyStoredListings)
	addedListings.SetUserID(userID)

	if err = s.repository.DeleteListingsByUserIDAndURLsTx(ctx, tx, userID, removedListings.URLs()); err != nil {
		s.log.Error().Err(err).Msg("failed to delete removed listings")
		return nil, nil, nil, fmt.Errorf("failed to delete removed listings: %w", err)
	}

	if err = s.repository.InsertListingsTx(ctx, tx, addedListings); err != nil {
		s.log.Error().Err(err).Msg("failed to add new listings")
		return nil, nil, nil, fmt.Errorf("failed to add new listings: %w", err)
	}

	if err = s.repository.UpdateListingsTx(ctx, tx, leftoverListings); err != nil {
		s.log.Error().Err(err).Msg("failed to update remaining listings")
		return nil, nil, nil, fmt.Errorf("failed to update remaining listings: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return nil, nil, nil, fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return addedListings, removedListings, leftoverListings, nil
}
