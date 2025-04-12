package geo

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
)

const ()

type City struct {
	Name       string
	Population int
}

type Cities []City

type CityData struct {
	Data map[string]Cities
}

func NewCityData() *CityData {
	cityData := new(CityData)
	err := cityData.load()
	if err != nil {
		panic(err)
	}
	return cityData
}

func (cd *CityData) GetRegions() []string {
	regions := make([]string, 0, len(cd.Data))
	for region := range cd.Data {
		regions = append(regions, region)
	}
	return regions
}

func (cd *CityData) GetCitiesByRegion(region string) Cities {
	cities, _ := cd.Data[region]
	return cities
}

func (c *Cities) GetTop5ByPopulation() Cities {
	if c == nil || len(*c) == 0 {
		return nil
	}
	sort.Slice(*c, func(i, j int) bool {
		return (*c)[i].Population > (*c)[j].Population
	})

	result := make(Cities, 0, 5)
	for idx := range *c {
		if len(result) < cap(result) {
			result = append(result, (*c)[idx])
		}
	}

	return result
}

func (cd *CityData) load() error {
	cd.Data = make(map[string]Cities)

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "nl_world_maps.csv")
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}

		name := row[0]
		admin := row[1]
		pop, _ := strconv.Atoi(row[2])
		cd.Data[admin] = append(cd.Data[admin], City{Name: name, Population: pop})
	}

	return nil
}
