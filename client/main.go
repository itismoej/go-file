package main

import (
	"bufio"
	"context"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	serverAddress := "0.0.0.0:50052"
	cc, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}
	filesClient := pb.NewFilesClient(cc)

	file, err := os.Open("some_file.md")
	if err != nil {
		log.Fatalf("cannot open file %s", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := filesClient.Upload(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

	req := &pb.FileChunk{
		Data: &pb.FileChunk_Info{
			Info: &pb.FileInfo{
				Name: "useless_file.md",
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatalf("%s", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		err = stream.Send(&pb.FileChunk{Data: &pb.FileChunk_Content{Content: buffer[:n]}})
		if err != nil {
			log.Fatalf("cannot send chunk to server: %s", err)
		}
	}
	res, _ := stream.CloseAndRecv()
	log.Printf("%s %d %d", res.GetName(), res.GetID(), res.GetSize())
}
