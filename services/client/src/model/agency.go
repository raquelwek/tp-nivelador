package model

type Agency interface {
	// GetBets returns a list of bets from the agency.
	GetBets() ([]Bet, error)

	// StoreWinner stores the winning bets in the agency.
	StoreWinner([]Bet) error
}
