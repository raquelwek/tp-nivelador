package model

import (
	"bufio"
	"fmt"
	"iter"
	"os"
	"strconv"
	"strings"
)

type AgencyImpl struct {
	Id         int
	OutputFile string
	InputFile  string
}

func CreateAgency(id int, outputFile string, inputFile string) *AgencyImpl {
	return &AgencyImpl{
		Id:         id,
		OutputFile: outputFile,
		InputFile:  inputFile,
	}
}

func parseBetLine(line string, agencyID int) (Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 5 {
		return Bet{}, fmt.Errorf("invalid bet line: expected 5 fields, got %d", len(fields))
	}

	document, err := strconv.Atoi(fields[2])
	if err != nil {
		return Bet{}, fmt.Errorf("invalid document %q: %w", fields[2], err)
	}

	number, err := strconv.Atoi(fields[4])
	if err != nil {
		return Bet{}, fmt.Errorf("invalid number %q: %w", fields[4], err)
	}

	return Bet{
		AgencyId:  agencyID,
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  document,
		Birthdate: strings.ReplaceAll(fields[3], "-", ""),
		Number:    number,
	}, nil
}

func formatBetLine(bet Bet) string {
	birthdate := bet.Birthdate
	birthdate = birthdate[0:4] + "-" + birthdate[4:6] + "-" + birthdate[6:8]

	return strings.Join([]string{
		bet.FirstName,
		bet.LastName,
		strconv.Itoa(bet.Document),
		birthdate,
		strconv.Itoa(bet.Number),
	}, ",")
}

func (a *AgencyImpl) LoadBets() iter.Seq2[Bet, error] {
	return func(yield func(Bet, error) bool) {
		file, err := os.Open(a.InputFile)
		if err != nil {
			yield(Bet{}, err)
			return
		}
		defer file.Close() // Se cierra automáticamente al terminar de iterar

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			bet, err := parseBetLine(scanner.Text(), a.Id)
			if err != nil {
				yield(Bet{}, err)
				return
			}

			// yield envía el valor. Si el consumidor hace un "break", yield devuelve false y salimos.
			if !yield(bet, nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield(Bet{}, err)
		}
	}
}

func (a *AgencyImpl) StoreWinner(winningBets []Bet) error {
	file, err := os.OpenFile(a.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	defer file.Close()

	for _, bet := range winningBets {
		if _, err := file.WriteString(formatBetLine(bet) + "\n"); err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
	}
	return nil
}

func (a *AgencyImpl) GetId() int {
	return a.Id
}
