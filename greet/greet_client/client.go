package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/nestor94/grpc-go/greet/greetpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBidirectionalStreaming(c)
	doUnaryDeadline(c, 5)
	doUnaryDeadline(c, 2)

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

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting Client Streaming RPC")
	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nestor",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Ramon",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Escoto",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Padilla",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Otro",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error calling Client Streaming %v\n", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending Request to Server %v\n", req)
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	msg, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error on Streaming %v\n", err)
	}
	log.Printf("Response from Server: %v", msg.GetResult())
}

func doBidirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting BiDirectional Streaming")
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nestor",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Ramon",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Escoto",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Padilla",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Otro",
			},
		},
	}
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error on Streaming %v\n", err)
	}

	waitChannel := make(chan struct{})

	go func() {
		for _, req := range requests {
			fmt.Printf("Sending Request to the Server %v\n", req)
			stream.Send(req)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error on Bidirectional Streaming %v\n", err)
				break
			}
			log.Printf("Message Received from Server %v\n", res.GetResult())
		}
		close(waitChannel)
	}()

	<-waitChannel
}

func doUnaryDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting RPC with Deadline Call")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nestor",
			LastName:  "Escoto",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	res, err := c.GreetWithDeadline(ctx, req)

	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout!")
			} else {
				fmt.Printf("Unexpected Error %v\n", err)
			}
		} else {
			log.Fatalf("Error %v", err)
		}

	} else {
		log.Printf("Response from Greet Server %v", res.Result)
	}

}
