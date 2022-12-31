package main

import (
	proto "Activereplication/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id         int
	portNumber int
}

var (
	clientPort   = flag.Int("cPort", 0, "client port number")
	frontendPort = flag.Int("fPort", 0, "frontend port number")
)

func main() {

	flag.Parse()

	client := &Client{
		id:         *clientPort,
		portNumber: *clientPort,
	}

	go connectToFrontend(client)
	log.Printf("Client: %v connected to frontend", client.id)

	for {
	}
}


func connectToFrontend(client *Client) {
	FrontendClient := getFrontendConnection()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//this is the input of the client who want to make a bid or see the result of the auction
		input := scanner.Text()
		

		if input == "deposit" {
			scanner.Scan() //The next scan is the amount to deposit
			amountToDeposit, _ := strconv.ParseInt(scanner.Text(), 10, 0)
			_, err := FrontendClient.Deposit(context.Background(), &proto.Amount{Amount: int32(amountToDeposit), Id: int32(client.id)})

			if err != nil {
				fmt.Printf("Deposit failed")
			}

		} else if input == "balance" {
			balance, _ := FrontendClient.GetBalance(context.Background(), &proto.Empty{})
			fmt.Printf("The balance is %v\n", balance.Balance)
		} else {
			fmt.Println("Invalid")
		}
	}
}

func getFrontendConnection() proto.BankClient {
	//calling the frontend
	//when creating a client, you specify at what port the frontend it. This of course has to match the frontend you started
	connection, err := grpc.Dial(":"+strconv.Itoa(*frontendPort), grpc.WithTransportCredentials(insecure.NewCredentials())) // remember to put the last line in the dial function
	if err != nil {
		log.Fatalln("Client could not make connection to frontend")
	}
	return proto.NewBankClient(connection)
}
