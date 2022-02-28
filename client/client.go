package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"simple-app-chat/proto"
	"time"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Welcome to my chat")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grp.Dial: %v", err)
		return
	}
	defer cc.Close()
	c := proto.NewChatServiceClient(cc)

	//init client
	client := clientChat{
		c: c,
	}

	//login
	client.login()
	defer client.logout()

	// start recieve message
	go client.recieveMsg()

	time.Sleep(time.Second)
	client.start()

}

type clientChat struct {
	c        proto.ChatServiceClient
	stream   proto.ChatService_LoginClient
	id       string
	username string
	to       string
}

func (client *clientChat) start() {
	for {
		bio := bufio.NewReader(os.Stdin)
		msg, _ := bio.ReadString('\n')
		_, err := client.c.Chat(context.Background(), &proto.ChatRequest{
			Message: &proto.Message{
				From:    client.username,
				To:      client.to,
				Message: msg,
			},
		})
		if err != nil {
			log.Fatalf("c.Chat: %v", err)
		}
	}
}

func (client *clientChat) login() {
	var username string
	fmt.Print("Enter your username:")
	fmt.Scan(&username)
	client.username = username

	var to string
	fmt.Print("Enter target name:")
	fmt.Scan(&to)
	fmt.Printf("connect to %v room\n", to)
	client.to = to

	stream, err := client.c.Login(context.Background())
	if err != nil {
		log.Fatalf("c.Login: %v", err)
		return
	}
	err = stream.Send(&proto.LoginRequest{
		Name: username,
	})
	if err != nil {
		log.Fatalf("login: %v", err)
		panic(err)
	}
	client.stream = stream
}

func (client *clientChat) recieveMsg() { //message from otherside will come here
	for {
		resp, err := client.stream.Recv()
		if err != nil {
			log.Fatalf("stream.Recv :%v", err)
			return
		}
		message := resp.GetMessage()
		if message.GetIsFromServer() {
			if len(message.GetId()) > 0 {
				fmt.Printf("id is %v\n", message.GetId())
				fmt.Println("================================================================================")
				client.id = message.GetId()
			}
			//will not print message from server
			continue
		}
		fmt.Printf("%v: %v", message.GetFrom(), message.GetMessage())
	}

}

func (client *clientChat) logout() {
	fmt.Println("logout...")
	resp, err := client.c.Logout(context.Background(), &proto.LogoutRequest{
		Id: client.id,
	})
	if err != nil {
		log.Fatalf("err while c.Logout%v", err)
	}
	if resp.GetOk() {
		fmt.Println("logout sucessfully.")
	}
	_ = client.stream.CloseSend()
}
