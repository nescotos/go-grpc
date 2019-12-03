package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

	//doUnary(c)
	//doStreaming(c)
	doClientStreaming(c)

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

func doStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting RPC Call to Prime Decomposition")
	req := &calculatorpb.PrimeDecompositionRequest{
		PrimeNumber: &calculatorpb.PrimeNumber{
			Number: 120,
		},
	}
	res, err := c.PrimeDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error Calling Stream Server %v", err)
	}

	for {
		msg, err := res.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error Reading Streaming %v", err)
		}

		log.Printf("Result from Streaming %v\n", msg.GetFactor())
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting RPC Call to Get Average")
	requests := []*calculatorpb.AverageRequest{
		&calculatorpb.AverageRequest{
			Number: 1,
		},
		&calculatorpb.AverageRequest{
			Number: 2,
		},
		&calculatorpb.AverageRequest{
			Number: 3,
		},
		&calculatorpb.AverageRequest{
			Number: 4,
		},
	}

	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Error on Client Streaming %v\n", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending Request to Server %v\n", req)
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	msg, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error on Client Streaming %v\n", err)
	}
	log.Printf("Average is %v\n", msg.GetAverage())
}
