package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"strconv"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	targetFileId, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		errorMessage := "the number should be given after running main.go file.\n" +
			"if you want to download the file with id 4, for example:\n" +
			"\n      go run main.go 4\n"
		log.Fatalf(errorMessage)
	}
	stream, err := filesClient.Download(ctx, &pb.File{ID: targetFileId})
	if err != nil {
		log.Fatalf("%s", err)
	}

	req, err := stream.Recv()
	if err != nil {
		logError(status.Errorf(codes.Unknown, "cannot receive file info: %s", err))
	} else {
		fileName := req.GetInfo().GetName()
		log.Printf("receive an upload-file request with name %s", fileName)

		fileData := bytes.Buffer{}

		for {
			err := contextError(stream.Context())
			if err != nil {
				logError(err)
			}

			log.Print("waiting to receive more data")

			req, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more data")
				break
			}
			if err != nil {
				logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
			}

			chunk := req.GetContent()

			_, err = fileData.Write(chunk)
			if err != nil {
				logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
			}
		}

		filePath := fmt.Sprintf("./%s", req.GetInfo().GetName())
		createdFile, err := os.Create(filePath)
		if err != nil {
			logError(status.Errorf(codes.Internal, "cannot create file: %w", err))
		}

		_, err = fileData.WriteTo(createdFile)
		if err != nil {
			logError(status.Errorf(codes.Internal, "cannot write to file: %w", err))
		}

		err = stream.CloseSend()
		if err != nil {
			logError(status.Errorf(codes.Unknown, "cannot close send: %v", err))
		}

		log.Printf("saved file with id: %d - name: %s", targetFileId, createdFile.Name())
	}
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}

func logError(err error) {
	if err != nil {
		log.Print(err)
	}
	os.Exit(1)
}
