package main

import (
	"context"
	"log"
	"simple-app-chat/proto"

	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grpc.Dial: %v", err)
	}
	defer cc.Close()
	c := proto.NewChatServiceClient(cc)
	_, _ = c.ListAll(context.Background(), &proto.ListAllRequest{})
}
