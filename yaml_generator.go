package main

import (
	"fmt"
	"os"
	"strconv"
)

const AGENCY_QUORUM_MIN = 5
const baseCompose = `services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - AGENCY_QUORUM_MIN=%d
    ports:
      - "5678:5678"
%s`

// @ TODO: Hacer variable input y output
const baseClient = `
  client_%d:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_%d
    environment:
      - AGENCY_ID=%d
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-%d.csv 
      - OUTPUT_FILE=/output/output-%d.csv
      - BATCH_SIZE=4
    depends_on:
      - server
    volumes:
      - ./input:/input
      - ./output:/output
`

func buildClients(clientAmount int) string {
	clients := ""
	for i := 1; i <= clientAmount; i++ {
		clients += fmt.Sprintf(baseClient, i, i, i, i, i)
	}
	return clients
}

func buildYaml(clientAmount int) error {
	content := fmt.Sprintf(baseCompose, AGENCY_QUORUM_MIN, buildClients(clientAmount))
	const filename = "docker-compose.yaml"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(content)
	return nil
}

func main() {
	const argsLen = 2

	if len(os.Args) != argsLen {
		fmt.Fprintf(os.Stderr, "Not enough args, received %d, expected %d\n", len(os.Args), argsLen)
		return
	}

	clientAmount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("You must include a valid amount of clients")
		return
	}
	err = buildYaml(clientAmount)
	if err != nil {
		fmt.Println(err)
	}

}
