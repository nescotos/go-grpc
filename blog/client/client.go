package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nestor94/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	log.Println("[*] Starting Blog Client")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("[!] Unable to connect %v", err)
	}

	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	blog := &blogpb.Blog{
		AuthorId: "Nestor",
		Title:    "My very first post",
		Content:  "This is the content of my blog",
	}

	response, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})

	blogID := response.GetBlog().GetId()

	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	log.Printf("Blog has been created: %v\n", response)

	_, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: "akasdlkasldk",
	})

	if err2 != nil {
		fmt.Printf("Error while reading: %v\n", err2)
	}

	result, err3 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: blogID,
	})

	if err3 != nil {
		fmt.Printf("Error while reading: %v", err3)
	}

	log.Printf("There is a Blog: %v\n", result.GetBlog())

}
