package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type MatchService struct {
	matchRepo      repository.MatchRepository
	sportRepo      repository.SportRepository
	teamRepo       repository.TeamRepository
	tournamentRepo repository.TournamentRepository
	firebaseDb     FirebaseClient
}

func NewMatchService(matchRepo repository.MatchRepository, sportRepo repository.SportRepository, teamRepo repository.TeamRepository, tournamentRepo repository.TournamentRepository, firebaseDb FirebaseClient) *MatchService {
	return &MatchService{
		matchRepo:      matchRepo,
		sportRepo:      sportRepo,
		teamRepo:       teamRepo,
		tournamentRepo: tournamentRepo,
		firebaseDb:     firebaseDb,
	}
}

func (s *MatchService) SyncToFirebase(ctx context.Context, sportID int) error {
	if s.firebaseDb == nil {
		return nil
	}
	sport, err := s.sportRepo.GetByID(ctx, sportID)
	if err != nil {
		return err
	}

	tournament, err := s.tournamentRepo.GetByID(ctx, sport.TournamentID)
	if err != nil {
		return err
	}

	matches, err := s.matchRepo.GetBySportID(ctx, sportID)
	if err != nil {
		return err
	}

	teams, err := s.teamRepo.GetAll(ctx, 0)
	if err != nil {
		return err
	}

	teamMap := make(map[int]model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	firebaseMatches := make(map[string]FirebaseMatch)
	for _, curr := range matches {
		var homeName, awayName string
		if curr.HomeID != nil {
			if t, ok := teamMap[*curr.HomeID]; ok {
				homeName = t.Name
			}
		}
		if curr.AwayID != nil {
			if t, ok := teamMap[*curr.AwayID]; ok {
				awayName = t.Name
			}
		}

		var homeScore, awayScore string
		if curr.HomeScore != nil {
			homeScore = *curr.HomeScore
		}
		if curr.AwayScore != nil {
			awayScore = *curr.AwayScore
		}

		canEdit := curr.HomeID == nil || curr.AwayID == nil
		nextMatchId := ""
		if curr.RoundID > 1 {
			nextMatchId = strconv.Itoa(curr.RoundID / 10)
		}
		fbMatch := FirebaseMatch{
			MatchId:     curr.ID,
			Name:        curr.Round + " - Match ",
			NextMatchId: nextMatchId,
			Participants: []*FirebaseParticipant{
				nil,
				{
					Name:         homeName,
					ResultText:   homeScore,
					IsWinner:     curr.WinnerID != nil && curr.HomeID != nil && *curr.WinnerID == *curr.HomeID,
					CanEditTeams: canEdit,
					TeamsID:      curr.HomeID,
				},
				{
					Name:         awayName,
					ResultText:   awayScore,
					IsWinner:     curr.WinnerID != nil && curr.AwayID != nil && *curr.WinnerID == *curr.AwayID,
					CanEditTeams: canEdit,
					TeamsID:      curr.AwayID,
				},
			},
			StartTime:           curr.StartDate.Format("15:04"),
			State:               string(curr.State),
			TournamentRoundText: curr.Round,
		}
		if curr.ImageUrl != nil {
			fbMatch.ImageUrl = *curr.ImageUrl
		}
		switch curr.Round {
		case "Final":
			fbMatch.Name = "Final"
		case "Perebutan juara 3":
			fbMatch.Name = "Perebutan juara 3"
		}

		firebaseMatches[strconv.Itoa(curr.RoundID)] = fbMatch
	}

	path := fmt.Sprintf("%s/sports/%s/matches", tournament.Slug, sport.Slug)
	ref := s.firebaseDb.NewRef(path)
	return ref.Set(ctx, firebaseMatches)
}

func (s *MatchService) GetAll(w http.ResponseWriter, r *http.Request) {
	var matches []model.Match
	var err error

	sportIDStr := r.URL.Query().Get("sportId")
	if sportIDStr != "" {
		sportID, errConv := strconv.Atoi(sportIDStr)
		if errConv != nil {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid sportId")
			return
		}
		matches, err = s.matchRepo.GetBySportID(r.Context(), sportID)
	} else {
		matches, err = s.matchRepo.GetAll(r.Context())
	}

	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, matches, "", "Matches retrieved")
}

func (s *MatchService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	match, err := s.matchRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, match, "", "Match retrieved")
}

func (s *MatchService) Create(w http.ResponseWriter, r *http.Request) {
	var match model.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	if err := s.matchRepo.Create(r.Context(), match); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), match.SportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Match created")
}

func (s *MatchService) Update(w http.ResponseWriter, r *http.Request) {
	const maxRequestSize = 3 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			helper.WriteResponse(w, http.StatusRequestEntityTooLarge, false, nil, helper.ErrEntityTooLarge, "Request body too large (max 3MB)")
			return
		}
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Error parsing form")
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	existingMatch, err := s.matchRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	// State transition logic
	if existingMatch.State == model.DONE {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrMatchInvalidStateTransition, "Cannot update match that is already DONE")
		return
	}

	newState := model.MatchState(r.FormValue("state"))

	if existingMatch.State == model.SOON {
		// Update home/away ID if provided
		if r.FormValue("home_id") != "" {
			homeID, err := strconv.Atoi(r.FormValue("home_id"))
			if err == nil {
				existingMatch.HomeID = &homeID
			}
		}
		if r.FormValue("away_id") != "" {
			awayID, err := strconv.Atoi(r.FormValue("away_id"))
			if err == nil {
				existingMatch.AwayID = &awayID
			}
		}
		// Update start_time if provided
		if r.FormValue("start_time") != "" {
			startTime, err := time.Parse(time.RFC3339, r.FormValue("start_time"))
			if err != nil {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid start_time format (use RFC3339, e.g., 2024-04-16T15:30:00Z)")
				return
			}
			existingMatch.StartDate = startTime
		}

		if newState == model.LIVE {
			file, handler, err := r.FormFile("image")
			if err != nil {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrMatchImageRequired, "Image is required when starting match")
				return
			}
			defer file.Close()

			if !helper.IsImage(handler) {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid image format")
				return
			}

			if !helper.ValidateImageSize(handler, 2*1024*1024) {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Image size must be less than 2MB")
				return
			}

			var fileReader io.Reader = file
			fileName := fmt.Sprintf("match-%d-%d%s", existingMatch.ID, time.Now().UnixNano(), filepath.Ext(handler.Filename))

			if helper.IsHEIC(handler) {
				jpegData, err := helper.ConvertHEICToJPEG(file)
				if err != nil {
					helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, "Error converting HEIC to JPEG")
					return
				}
				fileReader = bytes.NewReader(jpegData)
				fileName = fmt.Sprintf("match-%d-%d.jpg", existingMatch.ID, time.Now().UnixNano())
			}

			matchDir := filepath.Join(".", "storage", "match")
			helper.CheckDirectory(matchDir)

			if err := helper.UploadFile(fileReader, matchDir, fileName); err != nil {
				helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
				return
			}

			imageUrl := fmt.Sprintf("/storage/match/%s", fileName)
			existingMatch.ImageUrl = &imageUrl
			existingMatch.State = model.LIVE
		}
	} else if existingMatch.State == model.LIVE {
		if newState != "" && newState != model.LIVE && newState != model.DONE {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrMatchInvalidStateTransition, "Can only update status from LIVE to DONE or update scores")
			return
		}

		if newState == model.DONE {
			winnerIDStr := r.FormValue("winner_id")
			if winnerIDStr == "" {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrMatchWinnerRequired, "Winner ID is required when finishing match")
				return
			}
			winnerID, err := strconv.Atoi(winnerIDStr)
			if err != nil {
				helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid winner_id")
				return
			}
			existingMatch.WinnerID = &winnerID
			existingMatch.State = model.DONE
		}

		homeScore := r.FormValue("home_score")
		awayScore := r.FormValue("away_score")
		if homeScore != "" {
			existingMatch.HomeScore = &homeScore
		}
		if awayScore != "" {
			existingMatch.AwayScore = &awayScore
		}
	}

	// Update the match
	if err := s.matchRepo.Update(r.Context(), existingMatch); err != nil {
		if existingMatch.State == model.LIVE && existingMatch.ImageUrl != nil {
			// If we just uploaded a new image, delete it on failure
			helper.DeleteFile(filepath.Join(".", *existingMatch.ImageUrl))
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	// Logic to update next round based on winner_id
	if existingMatch.State == model.DONE && existingMatch.NextRoundID != nil && existingMatch.WinnerID != nil {
		nextMatch, err := s.matchRepo.GetByID(r.Context(), *existingMatch.NextRoundID)
		if err == nil {
			if existingMatch.RoundID%10 == 1 {
				nextMatch.HomeID = existingMatch.WinnerID
			} else {
				nextMatch.AwayID = existingMatch.WinnerID
			}
			s.matchRepo.Update(r.Context(), nextMatch)
			go func() {
				s.SyncToFirebase(context.Background(), nextMatch.SportID)
			}()
		}

		// Update loser match for semifinals
		var loserID *int
		if existingMatch.HomeID != nil && *existingMatch.WinnerID == *existingMatch.HomeID {
			loserID = existingMatch.AwayID
		} else if existingMatch.AwayID != nil && *existingMatch.WinnerID == *existingMatch.AwayID {
			loserID = existingMatch.HomeID
		}

		if loserID != nil {
			allMatches, err := s.matchRepo.GetBySportID(r.Context(), existingMatch.SportID)
			if err == nil {
				for _, m := range allMatches {
					if m.RoundID == 2 { // 3rd place match
						if existingMatch.RoundID%10 == 1 {
							m.HomeID = loserID
						} else {
							m.AwayID = loserID
						}
						s.matchRepo.Update(r.Context(), m)
						go func() {
							s.SyncToFirebase(context.Background(), m.SportID)
						}()
						break
					}
				}
			}
		}
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), existingMatch.SportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Match updated")
}

func (s *MatchService) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	match, err := s.matchRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if err := s.matchRepo.Delete(r.Context(), id); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), match.SportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Match deleted")
}

func (s *MatchService) DeleteBySportID(w http.ResponseWriter, r *http.Request) {
	sportID, err := strconv.Atoi(chi.URLParam(r, "sportId"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid sportId")
		return
	}

	if err := s.matchRepo.DeleteBySportID(r.Context(), sportID); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), sportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Matches deleted")
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
	TeamsID      *int   `json:"teams_id,omitempty"`
}

type FirebaseMatch struct {
	MatchId             int                    `json:"matchId,omitempty"`
	Name                string                 `json:"name"`
	NextMatchId         string                 `json:"nextMatchId"`
	NextLooserMatchId   string                 `json:"nextLooserMatchId,omitempty"`
	Participants        []*FirebaseParticipant `json:"participants"`
	StartTime           string                 `json:"startTime"`
	State               string                 `json:"state"`
	TournamentRoundText string                 `json:"tournamentRoundText"`
	ImageUrl            string                 `json:"imageUrl,omitempty"`
}

func (s *MatchService) GetTeamHistoryImages(w http.ResponseWriter, r *http.Request) {
	matchID, err := strconv.Atoi(chi.URLParam(r, "matchId"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid matchId")
		return
	}

	teamID, err := strconv.Atoi(chi.URLParam(r, "teamId"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid teamId")
		return
	}

	currentMatch, err := s.matchRepo.GetByID(r.Context(), matchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, "Match not found")
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	matches, err := s.matchRepo.GetBySportID(r.Context(), currentMatch.SportID)
	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	type HistoryImage struct {
		MatchID  int    `json:"match_id"`
		Round    string `json:"round"`
		ImageUrl string `json:"image_url"`
	}

	var history []HistoryImage
	currentRoundIDStr := strconv.Itoa(currentMatch.RoundID)

	for _, m := range matches {
		// History must have an image
		if m.ImageUrl == nil || *m.ImageUrl == "" {
			continue
		}

		// Must involve the team
		if (m.HomeID != nil && *m.HomeID == teamID) || (m.AwayID != nil && *m.AwayID == teamID) {
			// Special handling for 3rd place match (RoundID = 2)
			if currentMatch.RoundID == 2 {
				// Include all other matches involving the team
				if m.RoundID != 2 {
					history = append(history, HistoryImage{
						MatchID:  m.ID,
						Round:    m.Round,
						ImageUrl: *m.ImageUrl,
					})
				}
			} else {
				// Standard bracket logic: history has longer RoundID and current is prefix
				// Example: current 11 (semifinal), history 111 (quarterfinal)
				mRoundIDStr := strconv.Itoa(m.RoundID)
				if len(mRoundIDStr) > len(currentRoundIDStr) && strings.HasPrefix(mRoundIDStr, currentRoundIDStr) {
					history = append(history, HistoryImage{
						MatchID:  m.ID,
						Round:    m.Round,
						ImageUrl: *m.ImageUrl,
					})
				}
			}
		}
	}

	helper.WriteResponse(w, http.StatusOK, true, history, "", "Team history images retrieved")
}

func (s *MatchService) GetTeamHistoryImagesByTeam(w http.ResponseWriter, r *http.Request) {
	sportID, err := strconv.Atoi(chi.URLParam(r, "sportId"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid sportId")
		return
	}

	teamID, err := strconv.Atoi(chi.URLParam(r, "teamId"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid teamId")
		return
	}

	matches, err := s.matchRepo.GetBySportID(r.Context(), sportID)
	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	type HistoryImage struct {
		MatchID  int    `json:"match_id"`
		Round    string `json:"round"`
		ImageUrl string `json:"image_url"`
	}

	var history []HistoryImage
	for _, m := range matches {
		if m.ImageUrl == nil || *m.ImageUrl == "" {
			continue
		}
		if (m.HomeID != nil && *m.HomeID == teamID) || (m.AwayID != nil && *m.AwayID == teamID) {
			history = append(history, HistoryImage{
				MatchID:  m.ID,
				Round:    m.Round,
				ImageUrl: *m.ImageUrl,
			})
		}
	}

	helper.WriteResponse(w, http.StatusOK, true, history, "", "Team history images retrieved")
}

func (s *MatchService) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateMatchesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	sport, err := s.sportRepo.GetByID(r.Context(), req.SportID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrMatchSportNotFound, "Sport not found")
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	tournament, err := s.tournamentRepo.GetByID(r.Context(), sport.TournamentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrMatchTournamentNotFound, "Tournament not found")
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
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
		roundName := getRoundName(curr.Depth)

		// Create match struct
		m := model.Match{
			SportID:     req.SportID,
			RoundID:     roundID,
			NextRoundID: nil, // Will be set after getting DB IDs
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
		if err := s.matchRepo.Create(r.Context(), m); err != nil {
			// Log error but continue? Or fail?
			// Better to fail, but rolling back might be hard without transaction.
			// Ideally matchRepo should support BulkCreate or Transaction.
			// For now, log.
			fmt.Println("Error creating match in Postgres:", err)
		}
	}

	// Query created matches to get their IDs and update Firebase data
	createdMatches, err := s.matchRepo.GetBySportID(r.Context(), req.SportID)
	if err == nil {
		// Build map of RoundID to Match ID
		roundIDToMatchID := make(map[int]int)
		for _, m := range createdMatches {
			roundIDToMatchID[m.RoundID] = m.ID
		}

		// Update NextRoundID for each match with correct Match IDs
		for _, m := range createdMatches {
			if m.RoundID > 1 { // Not the final match (RoundID 1)
				parentRoundID := m.RoundID / 10
				if parentMatchID, ok := roundIDToMatchID[parentRoundID]; ok {
					m.NextRoundID = &parentMatchID
					s.matchRepo.Update(r.Context(), m)
				}
			}
		}

		// Update Firebase matches with IDs and NextMatchIds
		for roundIDStr, fbMatch := range firebaseMatches {
			roundID, _ := strconv.Atoi(roundIDStr)
			if matchID, ok := roundIDToMatchID[roundID]; ok {
				fbMatch.MatchId = matchID

				// Update NextMatchId to actual Match ID
				if fbMatch.NextMatchId != "" {
					nextRoundID, _ := strconv.Atoi(fbMatch.NextMatchId)
					if nextMatchID, ok := roundIDToMatchID[nextRoundID]; ok {
						fbMatch.NextMatchId = strconv.Itoa(nextMatchID)
					}
				}

				firebaseMatches[roundIDStr] = fbMatch
			}
		}
	}

	// Save to Firebase
	if s.firebaseDb != nil {
		path := fmt.Sprintf("%s/sports/%s/matches", tournament.Slug, sport.Slug)
		ref := s.firebaseDb.NewRef(path)
		if err := ref.Set(r.Context(), firebaseMatches); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrMatchFirebaseError, "Error saving to Firebase: "+err.Error())
			return
		}
	}

	helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Matches generated successfully")
}
