package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "chitchat/grpc"
)

type Server struct {
	pb.UnimplementedChitChatServiceServer
	mu      sync.Mutex
	clients map[string]chan *pb.Message
	clock   int64
}

func (s *Server) tick(receivedClock int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock = max(s.clock, receivedClock) + 1
	return s.clock
}

func (s *Server) Connect(req *pb.User, stream pb.ChitChatService_ConnectServer) error {
	username := req.Username

	ch := make(chan *pb.Message, 10)

	s.mu.Lock()
	s.clients[username] = ch
	s.mu.Unlock()

	ts := s.tick(req.Timestamp)
	log.Printf("event=CONNECT client=%s lamport=%d", username, ts)

	s.broadcast(&pb.Message{
		Sender:    "Server",
		Content:   fmt.Sprintf("Participant %s joined Chit Chat at logical time %d", username, ts),
		Timestamp: ts,
	})
	log.Printf("event=BROADCAST kind=JOIN client=%s lamport=%d", username, ts)

	// dissconect user once done
	defer s.Disconnect(context.Background(), &pb.User{Username: username, Timestamp: ts})

	// loops forever
	for {
		select {
		case msg, ok := <-ch: // if a message is on the user channale send it to the clients through the stream
			if !ok { // channel closed by removeClient
				return nil
			}
			if err := stream.Send(msg); err != nil { // return error if something went wrong (stream.Send can only return an error)
				log.Printf("event=SEND_FAILED to=%s from=%s err=%v", username, msg.Sender, err)
				return err
			}
			log.Printf("event=SEND to=%s from=%s lamport=%d msg=%q", username, msg.Sender, msg.Timestamp, msg.Content)
		case <-stream.Context().Done(): // return and disconnect user once the stream is closed or dissconected
			return nil
		}
	}
}

func (s *Server) PostMessage(ctx context.Context, msg *pb.Message) (*pb.Nothing, error) {
	ts := s.tick(msg.Timestamp)
	log.Printf("event=RECEIVE from=%s lamport=%d msg=%q", msg.Sender, ts, msg.Content)
	if !utf8.ValidString(msg.Content) || utf8.RuneCountInString(msg.Content) > 128 {
		log.Printf("event=REJECTED from=%s lamport=%d reason=invalid_message", msg.Sender, ts)
		return nil, status.Error(codes.InvalidArgument, "message must be valid UTF-8 and at most 128 characters")
	}
	s.broadcast(&pb.Message{
		Sender:    msg.Sender,
		Content:   msg.Content,
		Timestamp: ts,
	})
	log.Printf("event=BROADCAST kind=MESSAGE client=%s lamport=%d", msg.Sender, ts)

	return &pb.Nothing{}, nil
}

func (s *Server) Disconnect(ctx context.Context, user *pb.User) (*pb.Nothing, error) {
	s.removeClient(user.Username, user.Timestamp)
	return &pb.Nothing{}, nil
}

// only removes the client
func (s *Server) removeClient(username string, receivedClock int64) {
	s.mu.Lock()
	c, ok := s.clients[username]
	if ok {
		delete(s.clients, username) // remove from map
		close(c)                    // closes the users channale and stopes the loop in conect
	}
	s.mu.Unlock()

	if !ok {
		return // already removed, don't broadcast a second leave
	}

	// move clock, broadcast, log
	ts := s.tick(receivedClock)
	log.Printf("event=DISCONNECT client=%s lamport=%d", username, ts)
	s.broadcast(&pb.Message{
		Sender:    "Server",
		Content:   fmt.Sprintf("Participant %s left Chit Chat at logical time %d", username, ts),
		Timestamp: ts,
	})
	log.Printf("event=BROADCAST kind=LEAVE client=%s lamport=%d", username, ts)
}

// brodsastes a message to all conected clients stream
func (s *Server) broadcast(msg *pb.Message) {
	// lock and unlock once done
	s.mu.Lock()
	defer s.mu.Unlock()

	// Post in all channels
	for _, ch := range s.clients {
		select {
		case ch <- msg:
		default: // drops message in case channels buffer is full. makes sure the system doesn't deadlock
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

	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatalf("event=STARTUP_FAILED addr=localhost:5050 err=%v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, server)

	// on Ctrl+C / SIGTERM: log shutdown and stop the server so Serve returns.
	// Stop (not GracefulStop) because Connect streams never finish on their own.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Printf("event=SHUTDOWN addr=localhost:5050 lamport=%d", server.tick(0))
		grpcServer.Stop()
	}()

	log.Printf("event=STARTUP addr=localhost:5050 lamport=%d", server.clock)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("event=SERVE_FAILED err=%v", err)
	}
	log.Printf("event=STOPPED")
}
