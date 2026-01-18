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

func (m *MockSportRepository) Create(sport model.Sport) error {
	args := m.Called(sport)
	return args.Error(0)
}

func (m *MockSportRepository) GetAll() ([]model.Sport, error) {
	args := m.Called()
	return args.Get(0).([]model.Sport), args.Error(1)
}

func (m *MockSportRepository) GetByID(id int) (model.Sport, error) {
	args := m.Called(id)
	return args.Get(0).(model.Sport), args.Error(1)
}

func (m *MockSportRepository) Update(sport model.Sport) error {
	args := m.Called(sport)
	return args.Error(0)
}

func (m *MockSportRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockTournamentRepository
type MockTournamentRepository struct {
	mock.Mock
}

func (m *MockTournamentRepository) GetAll() ([]model.Tournament, error) {
	args := m.Called()
	return args.Get(0).([]model.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Create(tournament model.Tournament) error {
	args := m.Called(tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) Update(tournament model.Tournament) error {
	args := m.Called(tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTournamentRepository) GetBySlug(slug string) (model.Tournament, error) {
	args := m.Called(slug)
	return args.Get(0).(model.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) GetByID(id int) (model.Tournament, error) {
	args := m.Called(id)
	return args.Get(0).(model.Tournament), args.Error(1)
}

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAll() ([]model.User, error) {
	args := m.Called()
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailPassword(email string, password string) (model.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetById(id int) (model.User, error) {
	args := m.Called(id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user model.CreateUserRequest) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user model.UpdateUserRequest) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) SuperadminExists() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateSuperadmin(username, email, password string) error {
	args := m.Called(username, email, password)
	return args.Error(0)
}

func (m *MockUserRepository) GetSuperadminByUsername(username string) (model.User, error) {
	args := m.Called(username)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateSuperadminEmail(id int, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}

// MockTeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(team model.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetAll() ([]model.Team, error) {
	args := m.Called()
	return args.Get(0).([]model.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByID(id int) (model.Team, error) {
	args := m.Called(id)
	return args.Get(0).(model.Team), args.Error(1)
}

func (m *MockTeamRepository) Update(team model.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockMatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(match model.Match) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetAll() ([]model.Match, error) {
	args := m.Called()
	return args.Get(0).([]model.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByID(id int) (model.Match, error) {
	args := m.Called(id)
	return args.Get(0).(model.Match), args.Error(1)
}

func (m *MockMatchRepository) Update(match model.Match) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchRepository) Delete(id int) error {
	args := m.Called(id)
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
