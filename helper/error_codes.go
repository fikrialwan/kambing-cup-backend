package helper

const (
	// General Errors (0xx)
	ErrInternalServer = "001"
	ErrBadRequest     = "002"
	ErrNotFound       = "003"
	ErrUnauthorized   = "004"
	ErrEntityTooLarge = "006"

	// Tournament Errors (1xx)
	ErrTournamentSlugTaken          = "101"
	ErrTournamentNameRequired      = "102"
	ErrTournamentTotalSurahInvalid = "103"
	ErrTournamentImageRequired     = "104"

	// Sport Errors (2xx)
	ErrSportNameTaken          = "201"
	ErrSportNameRequired      = "202"
	ErrSportTournamentNotFound = "203"

	// Match Errors (3xx)
	ErrMatchSportNotFound           = "301"
	ErrMatchTournamentNotFound      = "302"
	ErrMatchFirebaseError           = "303"
	ErrMatchInvalidStateTransition  = "304"
	ErrMatchImageRequired           = "305"
	ErrMatchWinnerRequired          = "306"

	// Team Errors (4xx)
	ErrTeamNameTaken       = "401"
	ErrTeamRequiredFields = "402"

	// User Errors (5xx)
	ErrUserAlreadyExists  = "501"
	ErrUserRequiredFields = "502"
	ErrUserInvalidRole    = "503"

	// Auth Errors (6xx)
	ErrAuthRequiredFields     = "601"
	ErrAuthInvalidCredentials = "602"
)
