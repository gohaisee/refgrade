package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_, _ = grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
}
