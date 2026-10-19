package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"uuid"

	"google.golang.org/grpc"

	pb "chitchat/grpc"
)

type Server struct {
	pb.UnimplementedChitChatServiceServer
	mu      sync.Mutex
	clients map[string]chan *pb.Message
}

func (s *Server) Connect(ctx context.Context, req *pb.ConnectRequest) (*pb.User, error) {
	username := *pb.ConnectRequest.username
	user := pb.User{
		username: username,
		userId:   uuid.NewV4(),
	}
	return &user, nil
}

func (s *Server) Disconnect(ctx context.Context, req *pb.User) (*pb.Nothing, error) {
	fmt.Println(req.username + " disconnected.")
	s.users = remove
	return pb.Nothing{}, nil
}

func (s *Server) GetMessage(ctx context.Context, req *pb.Message) (*pb.Message, error) {

}

func (s *Server) PostMessage(ctx context.Context, req *pb.Message) (*pb.Nothing, error) {

}

func doStuff() {
	// wating for something on the line
}

func main() {
	server := Server{
		message: *pb.Message{&pb.Message{
			content:   "Server is live",
			timestamp: time.Now().String(),
		}},
		users: []*pb.User{&pb.User{}},
	}

	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	pb.RegisterChitChatServiceServer(grpcServer, &server)

	grpcServer.Serve(listener)

	// start countdown

	// listen in for users and usermessages
	// if new user send message
}
