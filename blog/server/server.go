package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/nestor94/grpc-go/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

var collection *mongo.Collection

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func dataToBlog(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	result, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error: %v", err))
	}
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"))
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
			Id:       oid.Hex(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	//Empty Struct

	data := &blogItem{}
	filter := bson.M{
		"_id": oid,
	}
	result := collection.FindOne(context.Background(), filter)
	if err := result.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find blog with that id %v", err))
	}
	return &blogpb.ReadBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to parse ID"))
	}
	data := &blogItem{}
	filter := bson.M{
		"_id": oid,
	}
	result := collection.FindOne(context.Background(), filter)
	if err := result.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find blog with that id %v", err))
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updatedError := collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unable to update on database %v\n", updatedError))
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlog(data),
	}, nil

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[*] Blog Service Starting")
	log.Println("[*] Starting Database")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to the client: %v", err)
	}

	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("[*] Server Listening at 50051")
	log.Println("[-] Register Blog Service")

	s := grpc.NewServer()

	blogpb.RegisterBlogServiceServer(s, &server{})

	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to Serve %v", err)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	<-ch
	log.Println("[*] Stopping Server")
	s.Stop()
	log.Println("[*] Closing Listener")
	lis.Close()
	log.Println("[*] Closing Database")
	client.Disconnect(context.TODO())
	log.Println("[*] Server is Closing Down...")

}
