package main

import (
	"context"
	"log"
	"net"

	accountspb "chariottakehome/api/services/accounts"
	userspb "chariottakehome/api/services/users"
	"chariottakehome/internal/accounts"
	"chariottakehome/internal/database"
	"chariottakehome/internal/users"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	userspb.RegisterUserServiceServer(s, &userspb.UserService{Repo: users.NewRepo(database.ConnPool())})
	accountspb.RegisterAccountServiceServer(s, &accountspb.AccountService{Repo: accounts.NewRepo(database.ConnPool())})

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("Received request: %s", info.FullMethod)

	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("Error handling request: %s", err)
	}
	return resp, err
}
