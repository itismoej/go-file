package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

const GRPCPort = "50061"

func main() {
	serverAddress := fmt.Sprintf("0.0.0.0:%s", GRPCPort)
	cc, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}
	filesClient := pb.NewFilesClient(cc)

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("file does not exist in this path: %s", filePath)
	}
	defer file.Close()
	filename := file.Name()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := filesClient.Upload(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

	req := &pb.FileChunk{
		Data: &pb.FileChunk_Info{
			Info: &pb.FileInfo{
				Name: filename,
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
	log.Printf("%s - id: %d - size: %d bytes", res.GetName(), res.GetID(), res.GetSize())
}
