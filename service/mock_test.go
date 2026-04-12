package service_test

import (
	"context"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"

	"github.com/stretchr/testify/mock"
)

// MockSportRepository
type MockSportRepository struct {
	mock.Mock
}

func (m *MockSportRepository) Create(ctx context.Context, sport model.Sport) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockSportRepository) GetAll(ctx context.Context, tournamentID int) ([]model.Sport, error) {
	args := m.Called(ctx, tournamentID)
	return args.Get(0).([]model.Sport), args.Error(1)
}

func (m *MockSportRepository) GetByID(ctx context.Context, id int) (model.Sport, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Sport), args.Error(1)
}

func (m *MockSportRepository) Update(ctx context.Context, sport model.Sport) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockSportRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSportRepository) GetByNameAndTournament(ctx context.Context, name string, tournamentID int) (model.Sport, error) {
	args := m.Called(ctx, name, tournamentID)
	return args.Get(0).(model.Sport), args.Error(1)
}

// MockTournamentRepository
type MockTournamentRepository struct {
	mock.Mock
}

func (m *MockTournamentRepository) GetAll(ctx context.Context) ([]model.Tournament, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) GetActive(ctx context.Context) (model.Tournament, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Create(ctx context.Context, tournament model.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) Update(ctx context.Context, tournament model.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) DeactivateAllExcept(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTournamentRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTournamentRepository) GetBySlug(ctx context.Context, slug string) (model.Tournament, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(model.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) GetByID(ctx context.Context, id int) (model.Tournament, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Tournament), args.Error(1)
}

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailPassword(ctx context.Context, email string, password string) (model.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetById(ctx context.Context, id int) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsernameOrEmail(ctx context.Context, username, email string) (model.User, error) {
	args := m.Called(ctx, username, email)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user model.CreateUserRequest) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user model.UpdateUserRequest) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) SuperadminExists(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateSuperadmin(ctx context.Context, username, email, password string) error {
	args := m.Called(ctx, username, email, password)
	return args.Error(0)
}

func (m *MockUserRepository) GetSuperadminByUsername(ctx context.Context, username string) (model.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateSuperadminEmail(ctx context.Context, id int, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

// MockTeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team model.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) CreateBulk(ctx context.Context, teams []model.Team) error {
	args := m.Called(ctx, teams)
	return args.Error(0)
}

func (m *MockTeamRepository) GetAll(ctx context.Context) ([]model.Team, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id int) (model.Team, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Team), args.Error(1)
}

func (m *MockTeamRepository) Update(ctx context.Context, team model.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByNameAndSport(ctx context.Context, name string, sportID int) (model.Team, error) {
	args := m.Called(ctx, name, sportID)
	return args.Get(0).(model.Team), args.Error(1)
}

// MockMatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(ctx context.Context, match model.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetAll(ctx context.Context) ([]model.Match, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Match), args.Error(1)
}

func (m *MockMatchRepository) GetBySportID(ctx context.Context, sportID int) ([]model.Match, error) {
	args := m.Called(ctx, sportID)
	return args.Get(0).([]model.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id int) (model.Match, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Match), args.Error(1)
}

func (m *MockMatchRepository) Update(ctx context.Context, match model.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchRepository) DeleteBySportID(ctx context.Context, sportID int) error {
	args := m.Called(ctx, sportID)
	return args.Error(0)
}

// MockFirebaseClient
type MockFirebaseClient struct {
	mock.Mock
}

func (m *MockFirebaseClient) NewRef(path string) service.FirebaseRef {
	args := m.Called(path)
	return args.Get(0).(service.FirebaseRef)
}

// MockFirebaseRef
type MockFirebaseRef struct {
	mock.Mock
}

func (m *MockFirebaseRef) Set(ctx context.Context, v interface{}) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}
