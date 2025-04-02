package sessions

import (
	"strings"
	"time"
)

type Session struct {
	UserID                string
	ChatID                int64
	UpdateIntervalSeconds int
	IsActive              bool
	RegionsRaw            string
	Regions               []string
	CitiesRaw             string
	Cities                []string
	LastSyncedAt          time.Time
}

func (s *Session) ParseRawRegionsAndCities() {
	if s.RegionsRaw != "" {
		s.Regions = strings.Split(s.RegionsRaw, ",")
		for idx := range s.Regions {
			s.Regions[idx] = strings.TrimSpace(s.Regions[idx])
		}
	} else {
		s.Regions = []string{}
	}
	if s.CitiesRaw != "" {
		s.Cities = strings.Split(s.CitiesRaw, ",")
		for idx := range s.Cities {
			s.Cities[idx] = strings.TrimSpace(s.Cities[idx])
		}
	} else {
		s.Cities = []string{}
	}
}

type Sessions []Session

func (s *Sessions) SelectForSync() Sessions {
	if s == nil || len(*s) == 0 {
		return nil
	}

	result := make(Sessions, 0, len(*s))
	for idx := range *s {
		interval := time.Duration((*s)[idx].UpdateIntervalSeconds) * time.Second
		lastSyncedAt := (*s)[idx].LastSyncedAt
		if lastSyncedAt.Add(interval).Before(time.Now()) {
			result = append(result, (*s)[idx])
		}
	}

	return result
}
