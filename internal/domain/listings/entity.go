package listings

import (
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Listing struct {
	UUID        string    `json:"UUID"`
	UserID      string    `json:"userId"`
	Context     any       `json:"@context"`
	Type        []string  `json:"@type"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Address     Address   `json:"address"`
	Offers      Offers    `json:"offers"`
	Image       string    `json:"image"`
	Photo       []Photo   `json:"photo"`
	IsNew       bool      `json:"isNew"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Offers struct {
	Type          string  `json:"@type"`
	PriceCurrency string  `json:"priceCurrency"`
	Price         float64 `json:"price"`
}

type Address struct {
	Type            string `json:"@type"`
	StreetAddress   string `json:"streetAddress"`
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
}

type Photo struct {
	Type       string `json:"@type"`
	ContentURL string `json:"contentUrl"`
}

type Listings []Listing

func (l *Listings) MapByURL() map[string]Listing {
	if l == nil || len(*l) == 0 {
		return nil
	}

	result := make(map[string]Listing)
	for idx := range *l {
		result[(*l)[idx].URL] = (*l)[idx]
	}
	return result
}

func (l *Listings) CompareAndGetRemovedListings(newListings Listings) (removedListings, leftoverListings Listings) {
	if l == nil || len(*l) == 0 {
		return removedListings, leftoverListings
	}
	removedListings = make(Listings, 0, len(*l))
	leftoverListings = make(Listings, 0, len(*l))
	currentMap := l.MapByURL()
	newMap := newListings.MapByURL()
	for url := range currentMap {
		if _, ok := newMap[url]; !ok {
			removedListings = append(removedListings, currentMap[url])
		} else {
			leftoverListings = append(leftoverListings, currentMap[url])
		}
	}
	return removedListings, leftoverListings
}

func (l *Listings) FilterByRegionsAndCities(regions, cities []string) Listings {
	if l == nil || len(*l) == 0 {
		return nil
	}
	filteredListings := make(Listings, 0, len(*l))

	for idx := range *l {
		(*l)[idx].Address.AddressRegion = strings.TrimPrefix((*l)[idx].Address.AddressRegion, "'")
		(*l)[idx].Address.AddressLocality = strings.TrimPrefix((*l)[idx].Address.AddressLocality, "'")
		if len(regions) == 0 && len(cities) == 0 {
			filteredListings = append(filteredListings, (*l)[idx])
			continue
		}

		if len(regions) != 0 && len(cities) == 0 {
			if slices.Contains(regions, strings.ToLower((*l)[idx].Address.AddressRegion)) {
				filteredListings = append(filteredListings, (*l)[idx])
				continue
			}
		}

		if len(regions) == 0 && len(cities) != 0 {
			if slices.Contains(cities, strings.ToLower((*l)[idx].Address.AddressLocality)) {
				filteredListings = append(filteredListings, (*l)[idx])
				continue
			}
		}

		if len(regions) != 0 && len(cities) != 0 {
			if slices.Contains(regions, strings.ToLower((*l)[idx].Address.AddressRegion)) || slices.Contains(cities, strings.ToLower((*l)[idx].Address.AddressLocality)) {
				filteredListings = append(filteredListings, (*l)[idx])
				continue
			}
		}
	}
	return filteredListings
}

func (l *Listings) CompareAndGetAddedListings(currentListings Listings) Listings {
	if l == nil || len(*l) == 0 {
		return nil
	}
	addedListings := make(Listings, 0, len(*l))
	currentMap := currentListings.MapByURL()
	newMap := l.MapByURL()
	for url := range newMap {
		if _, ok := currentMap[url]; !ok {
			addedListings = append(addedListings, newMap[url])
		}
	}
	return addedListings
}

func (l *Listings) Sort() {
	if l == nil || len(*l) == 0 {
		return
	}
	sort.SliceStable(*l, func(i, j int) bool {
		switch {
		case (*l)[i].CreatedAt.After((*l)[j].CreatedAt):
			return true
		case (*l)[i].CreatedAt.Before((*l)[j].CreatedAt):
			return false
		default:
			return (*l)[i].Offers.Price > (*l)[j].Offers.Price
		}
	})
}

func (l *Listings) URLs() []string {
	if l == nil || len(*l) == 0 {
		return nil
	}
	urls := make([]string, 0, len(*l))
	for idx := range *l {
		urls = append(urls, (*l)[idx].URL)
	}
	return urls
}

func (l *Listings) SetUserID(userID string) {
	if l == nil || len(*l) == 0 {
		return
	}
	for idx := range *l {
		(*l)[idx].UserID = userID
	}
}

func (l *Listings) GenerateUUIDs() {
	if l == nil || len(*l) == 0 {
		return
	}
	for idx := range *l {
		(*l)[idx].UUID = uuid.NewString()
	}
}

type ListingItem struct {
	Type     string `json:"@type"`
	Position uint   `json:"position"`
	URL      string `json:"url"`
}

type ListingSearchList struct {
	Context         any           `json:"@context"`
	Type            []string      `json:"@type"`
	Name            string        `json:"name"`
	URL             string        `json:"url"`
	ItemListElement []ListingItem `json:"itemListElement"`
}
