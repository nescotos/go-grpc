package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nestor94/grpc-go/calculator/calculatorpb"

	"google.golang.org/grpc"
)

func main() {
	log.Println("Connection at localhost:50051")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Unable to connect %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	doUnary(c)

}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting RPC Call")
	req := &calculatorpb.CalculatorRequest{
		Calculator: &calculatorpb.Calculator{
			FirstNumber:  10,
			SecondNumber: 25,
		},
	}
	res, err := c.Calculate(context.Background(), req)

	if err != nil {
		log.Fatalf("Error %v", err)
	}

	log.Printf("Response from Greet Server %v", res.Result)
}
