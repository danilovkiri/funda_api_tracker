package listings

import "time"

type Listing struct {
	Context     any       `json:"@context"`
	Type        []string  `json:"@type"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Address     Address   `json:"address"`
	Offers      Offers    `json:"offers"`
	Image       string    `json:"image"`
	Photo       []Photo   `json:"photo"`
	LastSeen    time.Time `json:"-"`
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

func (l *Listings) CompareAndGetRemovedListings(newListings Listings) Listings {
	if l == nil || len(*l) == 0 {
		return nil
	}
	removedListings := make(Listings, 0, len(*l))
	currentMap := l.MapByURL()
	newMap := newListings.MapByURL()
	for url := range currentMap {
		if _, ok := newMap[url]; !ok {
			removedListings = append(removedListings, currentMap[url])
		}
	}
	return removedListings
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
