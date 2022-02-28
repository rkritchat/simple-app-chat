package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"simple-app-chat/proto"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("start chat server on :50051")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	proto.RegisterChatServiceServer(s, &server{
		mem: make(map[string]proto.ChatService_LoginServer),
	})
	_ = s.Serve(lis)
}

type server struct {
	//TODO store stream
	mem map[string]proto.ChatService_LoginServer
}

func (s *server) Login(stream proto.ChatService_LoginServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			delete(s.mem, req.GetName()) //remove chat id from memory
			return nil
		}
		if err != nil {
			log.Fatalf("err while recieve message2: %v", err)
			return err
		}
		fmt.Printf("Login by :%v\n", req.GetName())
		s.mem[req.GetName()] = stream
		err = stream.Send(&proto.LoginResponse{
			Message: &proto.Message{
				Id:           req.GetName(),
				IsFromServer: true,
				Ok:           true,
			},
		})
		if err != nil {
			log.Fatalf("err while send response: %v", err)
			return err
		}
	}
}

func (s server) Logout(context context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	//remove from memory by Id
	fmt.Println("come to logout")
	delete(s.mem, req.GetId())
	return &proto.LogoutResponse{
		Ok: true,
	}, nil
}

func (s *server) Chat(context context.Context, req *proto.ChatRequest) (*proto.ChatResponse, error) {
	to := req.GetMessage().GetTo()
	fmt.Printf("send message to id %v\n", to)
	if des, ok := s.mem[to]; !ok {
		fmt.Printf("%v is not found on mem\n", to)
		return nil, fmt.Errorf("id: %v is not available", to)
	} else {
		//send message to destination
		err := des.Send(&proto.LoginResponse{
			Message: req.GetMessage(),
		})
		if err != nil {
			log.Fatalf("err while forward message: %v", err)
			return nil, err
		}
		//if not err
		return &proto.ChatResponse{
			Message: &proto.Message{
				IsFromServer: true,
				Ok:           true,
			},
		}, nil
	}
}

func (s *server) ListAll(context context.Context, req *proto.ListAllRequest) (*proto.ListAllResponse, error) {
	fmt.Printf("total user is %v\n", len(s.mem))
	for k := range s.mem { //key is uuid
		fmt.Printf("%v is Active\n", k)
	}
	return &proto.ListAllResponse{
		AdminId: "oK",
	}, nil
}
