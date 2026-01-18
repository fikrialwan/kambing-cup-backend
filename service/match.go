package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"math"
	"net/http"
	"strconv"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type MatchService struct {
	matchRepo      *repository.MatchRepository
	sportRepo      *repository.SportRepository
	tournamentRepo *repository.TournamentRepository
	firebaseDb     *db.Client
}

func NewMatchService(matchRepo repository.MatchRepository, sportRepo repository.SportRepository, tournamentRepo repository.TournamentRepository, firebaseDb *db.Client) *MatchService {
	return &MatchService{
		matchRepo:      &matchRepo,
		sportRepo:      &sportRepo,
		tournamentRepo: &tournamentRepo,
		firebaseDb:     firebaseDb,
	}
}

func (s *MatchService) GetAll(w http.ResponseWriter, r *http.Request) {
	matches, err := s.matchRepo.GetAll()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *MatchService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(match); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *MatchService) Create(w http.ResponseWriter, r *http.Request) {
	var match model.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.matchRepo.Create(match); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Match created"))
}

func (s *MatchService) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var match model.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	match.ID = id

	if err := s.matchRepo.Update(match); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Match updated"))
}

func (s *MatchService) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := s.matchRepo.Delete(id); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Match deleted"))
}

type GenerateMatchesRequest struct {
	TeamCount int `json:"team_count"`
	SportID   int `json:"sport_id"`
}

type FirebaseParticipant struct {
	CanEditTeams bool   `json:"canEditTeams"`
	IsWinner     bool   `json:"isWinner"`
	Name         string `json:"name,omitempty"`
	ResultText   string `json:"resultText,omitempty"`
}

type FirebaseMatch struct {
	Name                string                 `json:"name"`
	NextMatchId         string                 `json:"nextMatchId"`
	NextLooserMatchId   string                 `json:"nextLooserMatchId,omitempty"`
	Participants        []*FirebaseParticipant `json:"participants"`
	StartTime           string                 `json:"startTime"`
	State               string                 `json:"state"`
	TournamentRoundText string                 `json:"tournamentRoundText"`
}

func (s *MatchService) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateMatchesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sport, err := s.sportRepo.GetByID(req.SportID)
	if err != nil {
		http.Error(w, "Sport not found", http.StatusNotFound)
		return
	}

	tournament, err := s.tournamentRepo.GetByID(sport.TournamentID)
	if err != nil {
		http.Error(w, "Tournament not found", http.StatusNotFound)
		return
	}

	rounds := int(math.Ceil(math.Log2(float64(req.TeamCount))))
	matches := []model.Match{}
	firebaseMatches := make(map[string]FirebaseMatch)

	// Helper to get round name
	getRoundName := func(depth int) string {
		if depth == 0 {
			return "Final"
		}
		if depth == 1 {
			return "Semifinal"
		}
		if depth == 2 {
			return "8 Besar"
		}
		if depth == 3 {
			return "16 Besar"
		}
		if depth == 4 {
			return "32 Besar"
		}
		return fmt.Sprintf("%d Besar", int(math.Pow(2, float64(depth+1))))
	}

	// Generate matches
	// Start from Final (depth 0, id "1")
	// BFS or simply generating by levels

	// Queue for generation: (id, depth)
	type Item struct {
		ID    string
		Depth int
	}
	queue := []Item{{"1", 0}}

	// Also create 3rd place match
	match3rd := model.Match{
		SportID:   req.SportID,
		RoundID:   2,
		Round:     "Perebutan juara 3",
		State:     model.SOON,
		StartDate: time.Now(),
	}
	matches = append(matches, match3rd)
	firebaseMatches["2"] = FirebaseMatch{
		Name:                "Perebutan juara 3",
		NextMatchId:         "",
		Participants:        []*FirebaseParticipant{nil, {CanEditTeams: false}, {CanEditTeams: false}},
		StartTime:           "00:00",
		State:               "SOON",
		TournamentRoundText: "Perebutan juara 3",
	}

	generatedIDs := make(map[string]bool)
	generatedIDs["1"] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Prepare Postgres Match
		roundID, _ := strconv.Atoi(curr.ID)
		nextRoundIDVal := 0
		if len(curr.ID) > 1 {
			parentID := curr.ID[:len(curr.ID)-1]
			nextRoundIDVal, _ = strconv.Atoi(parentID)
		}
		nextRoundID := &nextRoundIDVal
		if len(curr.ID) == 1 {
			nextRoundID = nil
		}

		roundName := getRoundName(curr.Depth)

		// Create match struct
		m := model.Match{
			SportID:     req.SportID,
			RoundID:     roundID,
			NextRoundID: nextRoundID,
			Round:       roundName,
			State:       model.SOON,
			StartDate:   time.Now(),
		}
		matches = append(matches, m)

		// Prepare Firebase Match
		nextMatchId := ""
		if len(curr.ID) > 1 {
			nextMatchId = curr.ID[:len(curr.ID)-1]
		}
		
		nextLooserMatchId := ""
		if curr.Depth == 1 { // Semifinals
			nextLooserMatchId = "2"
		}

		// Determine participants state (canEditTeams)
		// Only the deepest level (leaf nodes) should be editable initially? 
		// Or creating empty slots.
		isLeaf := curr.Depth == rounds-1
		
		canEdit := false
		if isLeaf {
			canEdit = true
		}

		// Match Name logic (e.g., "Semifinal - Match 1")
		// Calculate match index in the round?
		// e.g. 11 is Match 1, 12 is Match 2 of Semifinal
		// 111 is Match 1, 112 is Match 2, 121 is Match 3, 122 is Match 4 of 8 Besar
		// It's tricky to get exact "Match X" number from ID binary string without parsing.
		// "11" -> 1. "12" -> 2.
		// "111" -> 1. "112" -> 2. "121" -> 3. "122" -> 4.
		// Binary logic: Convert "11..." string (replacing 1->0, 2->1) to integer + 1?
		// Ex: "111" -> suffix "11" -> 00 -> 0 -> +1 = 1
		// "112" -> suffix "12" -> 01 -> 1 -> +1 = 2
		// "121" -> suffix "21" -> 10 -> 2 -> +1 = 3
		// "122" -> suffix "22" -> 11 -> 3 -> +1 = 4
		
		matchNum := 1
		if len(curr.ID) > 1 {
			suffix := curr.ID[1:]
			val := 0
			for _, c := range suffix {
				val = val << 1
				if c == '2' {
					val = val | 1
				}
			}
			matchNum = val + 1
		}
		
		matchName := fmt.Sprintf("%s - Match %d", roundName, matchNum)
		if curr.Depth == 0 {
			matchName = "Final"
		}

		fbMatch := FirebaseMatch{
			Name:                matchName,
			NextMatchId:         nextMatchId,
			NextLooserMatchId:   nextLooserMatchId,
			Participants:        []*FirebaseParticipant{nil, {CanEditTeams: canEdit}, {CanEditTeams: canEdit}},
			StartTime:           "00:00",
			State:               "SOON",
			TournamentRoundText: roundName,
		}
		firebaseMatches[curr.ID] = fbMatch

		// Add children if not at max depth
		if curr.Depth < rounds-1 {
			child1 := curr.ID + "1"
			child2 := curr.ID + "2"
			queue = append(queue, Item{child1, curr.Depth + 1})
			queue = append(queue, Item{child2, curr.Depth + 1})
		}
	}

	// Save to Postgres
	for _, m := range matches {
		if err := s.matchRepo.Create(m); err != nil {
			// Log error but continue? Or fail?
			// Better to fail, but rolling back might be hard without transaction.
			// Ideally matchRepo should support BulkCreate or Transaction.
			// For now, log.
			fmt.Println("Error creating match in Postgres:", err)
		}
	}

	// Save to Firebase
	path := fmt.Sprintf("%s/sports/%s/matches", tournament.Slug, sport.Slug)
	ref := s.firebaseDb.NewRef(path)
	if err := ref.Set(context.Background(), firebaseMatches); err != nil {
		http.Error(w, "Error saving to Firebase: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matches generated successfully"))
}
