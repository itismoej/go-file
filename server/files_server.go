package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/mjafari98/go-file/models"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
)

const maxFileSize = 1 << 20

type FilesServer struct {
	pb.UnimplementedFilesServer
}

func (server *FilesServer) Upload(stream pb.Files_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive file info"))
	}

	fileName := req.GetInfo().GetName()
	log.Printf("receive an upload-file request with name %s", fileName)

	newFile := &models.File{
		Name: filepath.Base(fileName),
	}

	fileData := bytes.Buffer{}
	fileSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetContent()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)

		fileSize += size
		if fileSize > maxFileSize {
			return logError(status.Errorf(
				codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxFileSize,
			))
		}

		_, err = fileData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	newFile.Size = uint32(fileSize)
	DB.Create(&newFile)
	newFile.Path = fmt.Sprintf("%d-%s", newFile.ID, newFile.Name)
	DB.Save(&newFile)

	fileProtoBuf, err := newFile.Save(DB, fileData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save file to the store: %v", err))
	}

	err = stream.SendAndClose(fileProtoBuf)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved file with id: %d, size: %d", newFile.ID, newFile.Size)
	return nil
}

func (server *FilesServer) Download(fileData *pb.File, stream pb.Files_DownloadServer) error {
	var file models.File
	DB.Take(&file, "id = ?", fileData.ID)
	openedFile, err := os.Open("media/" + file.Path)
	if err != nil {
		log.Fatalf("cannot open file %s", err)
	}
	defer openedFile.Close()

	fileInfo := &pb.FileChunk{
		Data: &pb.FileChunk_Info{
			Info: &pb.FileInfo{
				Name: file.Name,
			},
		},
	}
	err = stream.Send(fileInfo)
	if err != nil {
		log.Fatalf("%s", err)
	}

	reader := bufio.NewReader(openedFile)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		err = stream.Send(&pb.FileChunk{Data: &pb.FileChunk_Content{Content: buffer[:n]}})
		if err != nil {
			log.Fatalf("cannot send chunk to server: %s", err)
			return err
		}
	}

	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
