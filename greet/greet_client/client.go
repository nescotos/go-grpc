package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/nestor94/grpc-go/greet/greetpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello from Client!")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Unable to connect %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	//doUnary(c)
	doServerStreaming(c)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting RPC Call")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nestor",
			LastName:  "Escoto",
		},
	}
	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("Error %v", err)
	}

	log.Printf("Response from Greet Server %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting Streaming RPC Call")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nestor",
			LastName:  "Escoto",
		},
	}
	res, err := c.GreetManyTimes(context.Background(), req)

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

		log.Printf("Result from Streaming %v", msg.GetResult())
	}

}
