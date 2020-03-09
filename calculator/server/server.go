package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/nestor94/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Calculate(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	fmt.Printf("Calculate Service Invoked %v", req)
	firstNumber := req.GetCalculator().GetFirstNumber()
	secondNumber := req.GetCalculator().GetSecondNumber()
	result := firstNumber + secondNumber
	res := &calculatorpb.CalculatorResponse{
		Result: result,
	}
	return res, nil
}

func (*server) PrimeDecomposition(req *calculatorpb.PrimeDecompositionRequest, stream calculatorpb.CalculatorService_PrimeDecompositionServer) error {
	fmt.Printf("Prime Decomposition Service Invoked %v\n", req)
	n := req.GetPrimeNumber().GetNumber()
	var k int32 = 2
	for n > 1 {
		if n%k == 0 {
			res := &calculatorpb.PrimeDecompositionResponse{
				Factor: k,
			}
			stream.Send(res)
			n = n / k
			time.Sleep(1 * time.Second)
		} else {
			k++
		}
	}
	return nil
}

func (*server) Average(stream calculatorpb.CalculatorService_AverageServer) error {
	fmt.Println("Average Service was Invoked!")
	n := 0
	var sum int32 = 0
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Average: float64(sum) / float64(n),
			})
		}
		if err != nil {
			log.Printf("Error on Streaming From Client %v\n", err)
		}
		sum += msg.GetNumber()
		n++
	}
}

func (*server) Maximum(stream calculatorpb.CalculatorService_MaximumServer) error {
	fmt.Println("Maximum Service was Invoked")
	max := int32(math.Inf(-1))
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error on Streaming %v\n", err)
		}
		if req.GetNumber() > max {
			max = req.GetNumber()
			sentErr := stream.Send(&calculatorpb.MaximumResponse{
				Max: max,
			})
			if sentErr != nil {
				log.Fatalf("Error on Streaming Data to Client%v\n", sentErr)
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquaredRoot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Negative Number Received: %v", number))
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {

	log.Println("Starting Server at 50051")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server Started Listening at 50051")
	log.Println("Register Calculator Service")
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to Serve %v", err)
	}

	log.Println("Server Ready")
}
