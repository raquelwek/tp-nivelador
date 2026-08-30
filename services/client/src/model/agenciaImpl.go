package model

import (
	"encoding/csv"
	"fmt"
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

func (a *AgencyImpl) GetBets() ([]Bet, error) {
	file, err := os.Open(a.InputFile)
	if err != nil {
		return nil, fmt.Errorf("error opening input file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var bets []Bet

	for row, err := reader.Read(); err == nil; row, err = reader.Read() {
		name, lastName, document, birthdate, number := row[0], row[1], row[2], row[3], row[4]

		doc, err := strconv.Atoi(document)
		num, err := strconv.Atoi(number)
		if err != nil {
			return nil, fmt.Errorf("error converting string to int: %v", err)
		}
		birthdate = strings.ReplaceAll(birthdate, "-", "")

		bet := Bet{
			AgencyId:  a.Id,
			FirstName: name,
			LastName:  lastName,
			Document:  doc,
			Birthdate: birthdate,
			Number:    num,
		}

		bets = append(bets, bet)
	}
	return bets, nil
}

func (a *AgencyImpl) StoreWinner(winningBets []Bet) error {
	file, err := os.OpenFile(a.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, bet := range winningBets {
		row := []string{
			strconv.Itoa(bet.AgencyId),
			bet.FirstName,
			bet.LastName,
			strconv.Itoa(bet.Document),
			bet.Birthdate,
			strconv.Itoa(bet.Number),
		}

		err := writer.Write(row)
		if err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
	}
	return nil
}
