package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "chitchat/grpc"
)

type Client struct {
    name    string
    clock   int64
    mu      sync.Mutex
    service pb.ChitChatServiceClient
}

// receives messages from the server until the stream closes
func (c *Client) receive(stream pb.ChitChatService_ConnectClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Println("disconnected:", err)
			return
		}

		received := msg.Timestamp
		c.mu.Lock()
		c.clock = max(c.clock, received) + 1
		ts := c.clock
		c.mu.Unlock()
		fmt.Printf("[%s @ %d] %s\n", msg.Sender, ts, msg.Content)
	}
}

func (c *Client) send(text string) {
    c.mu.Lock()
    c.clock++
    ts := c.clock
    c.mu.Unlock()

    _, err := c.service.PostMessage(context.Background(), &pb.Message{
        Sender:    c.name,
        Content:   text,
        Timestamp: ts,
    })
    if err != nil {
        log.Println("failed to send:", err)
    }
}

func main() {
	// pick the username when starting: ex. go run . alice
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <username>")
	}
	name := os.Args[1]

	conn, err := grpc.NewClient("localhost:5050",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChitChatServiceClient(conn)
	c := &Client{name: name, service: client}

	stream, err := client.Connect(context.Background(), &pb.User{Username: name})
	if err != nil {
		log.Fatalf("failed to join: %v", err)
	}

	go c.receive(stream) // once receive is a method too

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
    	c.send(scanner.Text())
	}
}