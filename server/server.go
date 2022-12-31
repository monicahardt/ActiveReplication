package main

import (
	proto "Activereplication/grpc"
	"context"
	"flag"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedBankServer
	name          string
	port          int
	balance			int32
}

var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func main() {
	flag.Parse()

	server := &Server{
		name: "serverName",
		port: *port,
	}

	go startServer(server)

	for {

	}
}

func startServer(server *Server) {
	grpcServer := grpc.NewServer()                                           // create a new grpc server
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.port)) // creates the listener

	if err != nil {
		log.Fatalln("Could not start listener")
	}

	log.Printf("Server started at port %v", server.port)

	proto.RegisterBankServer(grpcServer, server)
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register server")
	}

}

//simple depositi method. You just insert a positive amount
func (s *Server) Deposit(ctx context.Context, in *proto.Amount) (*proto.Ack, error){
	if(in.Amount < 0){
		log.Println("Fail, you must deposit positive amount")
		return &proto.Ack{Ack: fail}, nil
	} else {
		s.balance += in.Amount
		log.Printf("Server %d: Money successfully added to the account", s.port)
		return &proto.Ack{Ack: success}, nil
	}
}

//returning the balance on the server
func (s *Server) GetBalance(ctx context.Context, in *proto.Empty) (*proto.Balance, error){
	return &proto.Balance{Balance: s.balance}, nil
}



// our enum types
type ack string

const (
	fail    string = "fail"
	success string = "success"
)
