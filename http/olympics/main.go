package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type MedalTally struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type AthleteProfile struct {
	Name           string                `json:"athlete"`
	PrimaryCountry string                `json:"country"`
	TotalMedals    MedalTally            `json:"medals"`
	YearlyMedals   map[string]MedalTally `json:"medals_by_year"`
}

type OlympicRecord struct {
	AthleteName string `json:"athlete"`
	CountryName string `json:"country"`
	EventYear   int    `json:"year"`
	SportName   string `json:"sport"`
	GoldWon     int    `json:"gold"`
	SilverWon   int    `json:"silver"`
	BronzeWon   int    `json:"bronze"`
}

type CountryYearStats struct {
	Country string `json:"country"`
	MedalTally
}

type OlympicsDataService struct {
	athletes           map[string]*AthleteProfile
	countryStatsByYear map[int]map[string]MedalTally
	sportAthleteStats  map[string]map[string]*AthleteProfile
	knownSports        map[string]struct{}
	knownYears         map[int]struct{}
}

func loadAndProcessData(filePath string) (*OlympicsDataService, error) {
	dataBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file '%s': %w", filePath, err)
	}

	var records []OlympicRecord
	if err := json.Unmarshal(dataBytes, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	service := &OlympicsDataService{
		athletes:           make(map[string]*AthleteProfile),
		countryStatsByYear: make(map[int]map[string]MedalTally),
		sportAthleteStats:  make(map[string]map[string]*AthleteProfile),
		knownSports:        make(map[string]struct{}),
		knownYears:         make(map[int]struct{}),
	}

	for _, rec := range records {
		profile, exists := service.athletes[rec.AthleteName]
		if !exists {
			profile = &AthleteProfile{
				Name:           rec.AthleteName,
				PrimaryCountry: rec.CountryName,
				TotalMedals:    MedalTally{},
				YearlyMedals:   make(map[string]MedalTally),
			}
			service.athletes[rec.AthleteName] = profile
		}

		profile.TotalMedals.Gold += rec.GoldWon
		profile.TotalMedals.Silver += rec.SilverWon
		profile.TotalMedals.Bronze += rec.BronzeWon
		profile.TotalMedals.Total += rec.GoldWon + rec.SilverWon + rec.BronzeWon

		yearStr := strconv.Itoa(rec.EventYear)
		yearly := profile.YearlyMedals[yearStr]
		yearly.Gold += rec.GoldWon
		yearly.Silver += rec.SilverWon
		yearly.Bronze += rec.BronzeWon
		yearly.Total += rec.GoldWon + rec.SilverWon + rec.BronzeWon
		profile.YearlyMedals[yearStr] = yearly

		if _, yearExists := service.countryStatsByYear[rec.EventYear]; !yearExists {
			service.countryStatsByYear[rec.EventYear] = make(map[string]MedalTally)
		}
		countryStats := service.countryStatsByYear[rec.EventYear][rec.CountryName]
		countryStats.Gold += rec.GoldWon
		countryStats.Silver += rec.SilverWon
		countryStats.Bronze += rec.BronzeWon
		countryStats.Total += rec.GoldWon + rec.SilverWon + rec.BronzeWon
		service.countryStatsByYear[rec.EventYear][rec.CountryName] = countryStats

		sportKey := strings.ToLower(rec.SportName)
		if _, sportExists := service.sportAthleteStats[sportKey]; !sportExists {
			service.sportAthleteStats[sportKey] = make(map[string]*AthleteProfile)
		}
		sportAthleteProfile, athleteInSportExists := service.sportAthleteStats[sportKey][rec.AthleteName]
		if !athleteInSportExists {
			sportAthleteProfile = &AthleteProfile{
				Name:           rec.AthleteName,
				PrimaryCountry: service.athletes[rec.AthleteName].PrimaryCountry,
				TotalMedals:    MedalTally{},
				YearlyMedals:   make(map[string]MedalTally),
			}
			service.sportAthleteStats[sportKey][rec.AthleteName] = sportAthleteProfile
		}

		sportAthleteProfile.TotalMedals.Gold += rec.GoldWon
		sportAthleteProfile.TotalMedals.Silver += rec.SilverWon
		sportAthleteProfile.TotalMedals.Bronze += rec.BronzeWon
		sportAthleteProfile.TotalMedals.Total += rec.GoldWon + rec.SilverWon + rec.BronzeWon

		sportYearStr := strconv.Itoa(rec.EventYear)
		sportYearly := sportAthleteProfile.YearlyMedals[sportYearStr]
		sportYearly.Gold += rec.GoldWon
		sportYearly.Silver += rec.SilverWon
		sportYearly.Bronze += rec.BronzeWon
		sportYearly.Total += rec.GoldWon + rec.SilverWon + rec.BronzeWon
		sportAthleteProfile.YearlyMedals[sportYearStr] = sportYearly

		service.knownSports[sportKey] = struct{}{}
		service.knownYears[rec.EventYear] = struct{}{}
	}

	log.Printf("Loaded and processed data for %d athletes, %d years, %d sports.",
		len(service.athletes), len(service.knownYears), len(service.knownSports))

	return service, nil
}

type OlympicsServer struct {
	data *OlympicsDataService
}

func NewOlympicsServer(data *OlympicsDataService) *OlympicsServer {
	return &OlympicsServer{data: data}
}

func parseLimit(r *http.Request, defaultLimit int) (int, error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultLimit, nil
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return 0, fmt.Errorf("invalid limit parameter: %s", limitStr)
	}
	return limit, nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("Responding with error %d: %s", code, message)
	http.Error(w, message, code)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *OlympicsServer) handleAthleteInfo(w http.ResponseWriter, r *http.Request) {
	athleteName := r.URL.Query().Get("name")
	if athleteName == "" {
		respondWithError(w, http.StatusBadRequest, "name parameter is required")
		return
	}

	profile, found := s.data.athletes[athleteName]
	if !found {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("athlete %s not found", athleteName))
		return
	}

	respondWithJSON(w, http.StatusOK, profile)
}

func (s *OlympicsServer) handleTopAthletesInSport(w http.ResponseWriter, r *http.Request) {
	sportName := r.URL.Query().Get("sport")
	if sportName == "" {
		respondWithError(w, http.StatusBadRequest, "sport parameter is required")
		return
	}
	sportKey := strings.ToLower(sportName)

	limit, err := parseLimit(r, 3)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	athletesInSport, sportExists := s.data.sportAthleteStats[sportKey]
	if !sportExists {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("sport '%s' not found", sportName))
		return
	}

	rankedAthletes := make([]*AthleteProfile, 0, len(athletesInSport))
	for _, profile := range athletesInSport {
		rankedAthletes = append(rankedAthletes, profile)
	}

	sort.Slice(rankedAthletes, func(i, j int) bool {
		medalsI := rankedAthletes[i].TotalMedals
		medalsJ := rankedAthletes[j].TotalMedals
		if medalsI.Gold != medalsJ.Gold {
			return medalsI.Gold > medalsJ.Gold
		}
		if medalsI.Silver != medalsJ.Silver {
			return medalsI.Silver > medalsJ.Silver
		}
		if medalsI.Bronze != medalsJ.Bronze {
			return medalsI.Bronze > medalsJ.Bronze
		}
		return rankedAthletes[i].Name < rankedAthletes[j].Name
	})

	if limit > len(rankedAthletes) {
		limit = len(rankedAthletes)
	}
	topAthletes := rankedAthletes[:limit]

	respondWithJSON(w, http.StatusOK, topAthletes)
}

func (s *OlympicsServer) handleTopCountriesInYear(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		respondWithError(w, http.StatusBadRequest, "year parameter is required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid year parameter")
		return
	}

	limit, err := parseLimit(r, 3)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	countryStats, yearExists := s.data.countryStatsByYear[year]
	if !yearExists {
		if _, trulyKnown := s.data.knownYears[year]; !trulyKnown {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("year %d not found", year))
			return
		}
		respondWithJSON(w, http.StatusOK, []CountryYearStats{})
		return
	}

	rankedCountries := make([]CountryYearStats, 0, len(countryStats))
	for countryName, tally := range countryStats {
		rankedCountries = append(rankedCountries, CountryYearStats{
			Country:    countryName,
			MedalTally: tally,
		})
	}

	sort.Slice(rankedCountries, func(i, j int) bool {
		if rankedCountries[i].Gold != rankedCountries[j].Gold {
			return rankedCountries[i].Gold > rankedCountries[j].Gold
		}
		if rankedCountries[i].Silver != rankedCountries[j].Silver {
			return rankedCountries[i].Silver > rankedCountries[j].Silver
		}
		if rankedCountries[i].Bronze != rankedCountries[j].Bronze {
			return rankedCountries[i].Bronze > rankedCountries[j].Bronze
		}
		return rankedCountries[i].Country < rankedCountries[j].Country
	})

	if limit > len(rankedCountries) {
		limit = len(rankedCountries)
	}
	topCountries := rankedCountries[:limit]

	respondWithJSON(w, http.StatusOK, topCountries)
}

func main() {
	listenPort := flag.Int("port", 8080, "Port number for the HTTP server")
	dataFile := flag.String("data", "", "Path to the Olympic winners JSON data file")
	flag.Parse()

	if *dataFile == "" {
		log.Fatal("Error: -data flag is required")
	}
	if *listenPort <= 0 || *listenPort > 65535 {
		log.Fatalf("Error: Invalid port number %d", *listenPort)
	}

	olympicsData, err := loadAndProcessData(*dataFile)
	if err != nil {
		log.Fatalf("Error initializing data service: %v", err)
	}

	server := NewOlympicsServer(olympicsData)

	mux := http.NewServeMux()
	mux.HandleFunc("/athlete-info", server.handleAthleteInfo)
	mux.HandleFunc("/top-athletes-in-sport", server.handleTopAthletesInSport)
	mux.HandleFunc("/top-countries-in-year", server.handleTopCountriesInYear)

	addr := fmt.Sprintf(":%d", *listenPort)
	log.Printf("Starting Olympics API server on %s", addr)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
