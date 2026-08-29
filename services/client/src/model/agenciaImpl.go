type AgencyImpl struct {
	Id         int
	OutputFile string
	InputFile  string
}

func createAgency(id int, outputFile string, inputFile string) *AgencyImpl {
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
	number := 0
	for row, err := reader.Read(); err == nil; row, err = reader.Read(), number++ {
		name, lastName, document, birthdate := row[1], row[2], row[3], row[4]

		num, err := strconv.Atoi(document)
		if err != nil {
			return nil, fmt.Errorf("error converting string to int: %v", err)
		}

		bet := Bet{
			AgencyId:  a.Id,
			FirstName: name,
			LastName:  lastName,
			Document:  document,
			Birthdate: birthdate,
			Number:    number,
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
			bet.Document,
			bet.Birthdate
		}

		err := writer.Write(row)
		if err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
	}
	return nil
}
