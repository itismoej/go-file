package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	ID        uint64 `gorm:"primarykey"`
	Name      string `gorm:"size:127"`
	Size      uint32
	Path      string `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (file *File) FromProtoBuf(pbFile *pb.File) {
	file.ID = pbFile.ID
	file.Name = pbFile.Name
	file.Size = pbFile.Size
}

func (file *File) ConvertToProtoBuf() *pb.File {
	return &pb.File{
		ID:   file.ID,
		Name: file.Name,
		Size: file.Size,
	}
}

func (file *File) Save(
	db *gorm.DB,
	data bytes.Buffer,
) (*pb.File, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&file)
		if errors.Is(result.Error, gorm.ErrInvalidData) {
			return status.Errorf(codes.InvalidArgument, "invalid data has been entered")
		}

		dirPath := filepath.Join(".", "media")
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("cannot create media directory in the path: %w", err)
		}

		filePath := filepath.Join(dirPath, fmt.Sprintf("%d-%s", file.ID, file.Name))
		createdFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("cannot create file in the path: %w", err)
		}

		_, err = data.WriteTo(createdFile)
		if err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error in uploading file: %s", err)
	}

	return file.ConvertToProtoBuf(), nil
}
