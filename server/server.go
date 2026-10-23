package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"

	pb "chitchat/grpc"
)

// TODO: Check and see if everything that needs to be logged is logged correcrly

type Server struct {
	pb.UnimplementedChitChatServiceServer
	mu      sync.Mutex
	clients map[string]chan *pb.Message
	clock   int64
}

func (s *Server) Connect(req *pb.User, stream pb.ChitChatService_ConnectServer) error {
	username := req.Username

	ch := make(chan *pb.Message, 10)

	s.mu.Lock()
	s.clients[username] = ch
	s.mu.Unlock()

	s.broadcast(&pb.Message{
		Sender:    "Server",
		Content:   username + " joined the chat",
		Timestamp: s.clock,
	})

	// TODO: make server clock update acording with reviced clock

	s.clock++

	// dissconect user once done
	defer s.Disconnect(context.Background(), req)

	// loops forever
	for {
		select {
		case msg := <-ch: // if a message is on the user channale send it to the clients through the stream
			if err := stream.Send(msg); err != nil { // return error if something went wrong (stream.Send can only return an error)
				log.Printf("event=SEND_FAILED to=%s from=%s err=%v", username, msg.Sender, err)
				return err
			}
			log.Printf("event=SEND to=%s from=%s lamport=%d msg=%q", username, msg.Sender, s.clock, msg.Content)
		case <-stream.Context().Done(): // return and disconnect user once the stream is closed or dissconected
			return nil
		}
	}

}

func (s *Server) PostMessage(ctx context.Context, msg *pb.Message) (*pb.Nothing, error) {
	log.Printf("event=RECEIVE from=%s lamport=%d msg=%q", msg.Sender, s.clock, msg.Content)
	s.broadcast(msg)
	log.Printf("event=BROADCAST from=%s lamport=%d", msg.Sender, s.clock)

	return &pb.Nothing{}, nil
}

func (s *Server) Disconnect(ctx context.Context, user *pb.User) (*pb.Nothing, error) {
	s.mu.Lock()
	delete(s.clients, user.Username)
	s.mu.Unlock()

	s.broadcast(&pb.Message{
		Sender:    "Server",
		Content:   user.Username + " left the chat",
		Timestamp: s.clock,
	})

	s.clock++

	return &pb.Nothing{}, nil
}

// brodsastes a message to all conected clients stream
func (s *Server) broadcast(msg *pb.Message) {
	// lock and unlock once done
	s.mu.Lock()
	defer s.mu.Unlock()

	// go through the channales and skip the senders name
	for name, ch := range s.clients {
		if name == msg.Sender {
			continue
		}
		select {
		case ch <- msg:
		default: // drops message incase channale buffer is full. makes sure the system dosnt lock up
		}
	}
}

func main() {
	// opens the log file and creates the template
	f, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close() // close it once done

	log.SetOutput(f)
	log.SetPrefix("[Server] ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	server := &Server{
		clients: make(map[string]chan *pb.Message),
	}

	fmt.Printf("@ %d is live", server.clock)
	log.Println()

	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, server)

	log.Printf("event=STARTUP addr=localhost:5050 lamport=%d", server.clock)
	grpcServer.Serve(listener)

	fmt.Printf("Server is live")
	// start countdown

	// listen in for users and usermessages
	// if new user send message
}
