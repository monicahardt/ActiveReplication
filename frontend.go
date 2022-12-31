package main

import (
	proto "Activereplication/grpc"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Frontend struct {
	proto.UnimplementedBankServer
	port            int
	servers         []proto.BankClient
	acks            []*proto.Ack

}

var port = flag.Int("port", 0, "frontend port number") // create the port that recieves the port that the client wants to access to

func main() {

	fmt.Println("WTF")
	//getting the portnumber
	flag.Parse()

	//setting the log file
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	//creating the frontend
	frontend := &Frontend{
		port:            *port,
		servers:         make([]proto.BankClient, 0),
		acks:            make([]*proto.Ack, 0),
		
	}

	go startFrontend(frontend)

	//Making the frontend connect to servers at ports 5001, 5002 and 5003
	for i := 0; i < 3; i++ {
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(5001+i), grpc.WithTransportCredentials(insecure.NewCredentials()))
		log.Printf("Frontend connected to server at port: %v\n", 5001+i)

		//appending the servers to the slice called servers
		frontend.servers = append(frontend.servers, proto.NewBankClient(conn))
		if err != nil {
			log.Printf("Could not connect: %s", err)
		}
		defer conn.Close()
	}

	for {

	}
}

//Start function just like for a normal server. Make it listen on it's own port
func startFrontend(frontend *Frontend) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(frontend.port))

	if err != nil {
		log.Fatalf("Could not create the frontend %v", err)
	}
	log.Printf("Started frontend at port: %d\n", frontend.port)

	// Register the grpc server and serve its listener
	proto.RegisterBankServer(grpcServer, frontend)

	serveError := grpcServer.Serve(listener)
	fmt.Printf("nedern")
	if serveError != nil {
		log.Fatalf("Could not serve listener frontend")

	}
}

//we could use mutual exclusion in the server class with a channel
//to make sure that two clients cannot change the balance at the same time

//So in the methods in the frontend we want to ask all the servers to deposit
//it is here we discover if a server has crashed
//since this is active replication, if a server crashes we just continue the program

func (f *Frontend) Deposit(ctx context.Context, amountToDeposit *proto.Amount) (*proto.Ack, error){
	log.Printf("Client %v deposited %v", amountToDeposit.Id, amountToDeposit.Amount)
	f.acks = make([]*proto.Ack, 0)
	
	// for each server in the slice, we call the server's deposit method
	for index, s := range f.servers {
		ack, err := s.Deposit(ctx, amountToDeposit)

		// if err != nil the server has crashed and we have to remove it from the slice
		// else the server is still running, add it to the slice
		if err != nil {
			log.Printf("Server crashed, removing it from slice")
			f.servers = append(f.servers[:index], f.servers[index+1:]...)
		} else {
			f.acks = append(f.acks, ack)
		}
	}

	err:= f.ValidateAcks()
	if (err == nil){
		log.Printf("Client %d successfully deposited %d to the account ", amountToDeposit.Id, amountToDeposit.Amount)
		return &proto.Ack{Ack: success}, nil
	} else {
		//if there was an error
		log.Printf("Something went wrong - couldn't deposit %d to the account ", amountToDeposit.Amount)
		return &proto.Ack{Ack: fail}, nil
	}
}

func (f *Frontend) ValidateAcks() (error){
	
	var sCount = 0
	var fCount = 0

	// counts how many succesful, failed and exception bids there were in the servers
	for i := 0; i < len(f.servers); i++ {
		if f.acks[i].Ack == success {
			sCount++
		}
		if f.acks[i].Ack == fail {
			fCount++
		}
	}

	// checks if more than half of the servers respond were successfull
	// removes the one that was NOT succesful, since it's deprecated
	if sCount > (len(f.servers)/2) && sCount != 0 {
		for i := 0; i < len(f.servers); i++ {
			if f.acks[i].Ack != success {
				// disconnect the server on f.servers[i]
				f.servers = append(f.servers[:i], f.servers[i+1:]...)
			}
		}
		return nil
	}

	// checks if more than half of the servers respond were fail
	// removes the one that was NOT fail, since it's deprecated
	if fCount > (len(f.servers)/2) && fCount != 0 {
		for i := 0; i < len(f.servers); i++ {
			if f.acks[i].Ack != fail {
				// disconnect the server on f.servers[i]
				f.servers = append(f.servers[:i], f.servers[i+1:]...)
			}
		}
		return errors.New("Most of the servers responded with fail")
	}

	// else everyone answered something different and therefore they're all faulty
	return errors.New("All the servers are faulty! Run!!")
}

//calls each of the servers getBalance method
//finds the highest balance and returns it. This is a bit bully where we just return the highest
//we could have an update method that was called if they did not say the same!!!********************
func (f *Frontend) GetBalance(ctx context.Context, in *proto.Empty) (*proto.Balance, error){
	log.Println("Client asked for the balance")
	balance := int32(0)

	for _, s := range f.servers {
		tmp, _ := s.GetBalance(ctx, in)

		if int32(tmp.Balance) > balance {
			balance = tmp.Balance
		}
	}
	log.Printf("The balance is %d \n", balance)
	return &proto.Balance{Balance: balance}, nil
}

// our enum types
type ack string

const (
	fail    string = "fail"
	success string = "success"
)
