package sessions

import (
	"strings"
	"time"
)

type Session struct {
	UserID                   string
	ChatID                   int64
	UpdateIntervalSeconds    int
	IsActive                 bool
	RegionsRaw               string
	Regions                  []string
	CitiesRaw                string
	Cities                   []string
	LastSyncedAt             time.Time
	SyncCountSinceLastChange int
	DNDActive                bool
	DNDStart                 int
	DNDEnd                   int
}

func (s *Session) ParseRawRegionsAndCities() {
	s.Regions = []string{}
	if s.RegionsRaw != "" {
		uniqueRegions := uniqueStrings(strings.Split(s.RegionsRaw, ","))
		for idx := range uniqueRegions {
			s.Regions = append(s.Regions, uniqueRegions[idx])
		}
		s.RegionsRaw = strings.Join(s.Regions, ",")
	}

	s.Cities = []string{}
	if s.CitiesRaw != "" {
		uniqueCities := uniqueStrings(strings.Split(s.CitiesRaw, ","))
		for idx := range uniqueCities {
			s.Cities = append(s.Cities, uniqueCities[idx])
		}
		s.CitiesRaw = strings.Join(s.Cities, ",")
	}
}

func (s *Session) IsWithinDND() bool {
	if !s.DNDActive {
		return false
	}

	nowTs := time.Now().UTC()
	minutes := nowTs.Hour()*60 + nowTs.Minute()

	if s.DNDStart < s.DNDEnd {
		return minutes >= s.DNDStart && minutes < s.DNDEnd
	}
	return minutes >= s.DNDStart || minutes < s.DNDEnd
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

func uniqueStrings(input []string) []string {
	encountered := make(map[string]bool)
	var result []string

	for _, s := range input {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !encountered[s] {
			encountered[s] = true
			result = append(result, s)
		}
	}

	return result
}
