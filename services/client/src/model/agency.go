package model

import "iter"

type Agency interface {
	// Iterator that yields bets from the agency's input file.
	LoadBets() iter.Seq2[Bet, error]

	// StoreWinner stores the winning bets in the agency.
	StoreWinner([]Bet) error

	/// Returns the ID of the agency.
	GetId() int
}
