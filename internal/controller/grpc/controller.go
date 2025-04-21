package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/semaphore"
	"tages/pb"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	mu              sync.RWMutex
	uploadSemaphore *semaphore.Weighted
	listSemaphore   *semaphore.Weighted
}

func NewFileServer(limitUpload, limitListFile int64) *FileServer {
	return &FileServer{
		uploadSemaphore: semaphore.NewWeighted(limitUpload),
		listSemaphore:   semaphore.NewWeighted(limitListFile),
	}
}

func (s *FileServer) UploadFile(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	if err := s.uploadSemaphore.Acquire(ctx, 1); err != nil {
		return nil, err
	}

	var Err error
	res := make(chan struct{}, 1)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		defer close(res)
		filename := filepath.Base(req.Filename)
		log.Printf("Uploading file: %s", filename)

		err := os.WriteFile(filename, req.FileData, 0644)
		if err != nil {
			Err = fmt.Errorf("failed to write file %s: %w", filename, err)
			res <- struct{}{}
			return
		}
		res <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-res:
		s.uploadSemaphore.Release(1)
		if Err != nil {
			return &pb.UploadResponse{Message: "File uploaded successfully"}, nil
		}
		return &pb.UploadResponse{Message: "Failed to upload file"}, Err
	}
}

func (s *FileServer) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	if err := s.listSemaphore.Acquire(ctx, 1); err != nil {
		return nil, err
	}

	var Err error
	res := make(chan []*pb.FileInfo, 1)

	go func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		defer close(res)

		files, err := os.ReadDir(".")
		if err != nil {
			Err = fmt.Errorf("failed to read directory: %w", err)
			res <- nil
			return
		}

		var fileInfos []*pb.FileInfo
		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				Err = fmt.Errorf("failed to get file info: %w", err)
				res <- nil
				return
			}

			if !info.IsDir() {
				fileInfos = append(fileInfos, &pb.FileInfo{
					Name:      info.Name(),
					CreatedAt: info.ModTime().String(),
					UpdatedAt: info.ModTime().String(),
				})
			}
		}
		res <- fileInfos
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-res:
		s.listSemaphore.Release(1)
		if v == nil {
			return nil, Err
		}
		return &pb.ListFilesResponse{Files: v}, nil
	}
}

func (s *FileServer) DownloadFile(ctx context.Context, req *pb.DownloadRequest) (*pb.DownloadResponse, error) {
	if err := s.uploadSemaphore.Acquire(ctx, 1); err != nil {
		return nil, err
	}

	var Err error
	res := make(chan []byte, 1)

	go func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		defer close(res)

		filename := filepath.Base(req.Filename)
		log.Printf("Downloading file: %s", filename)

		data, err := os.ReadFile(filename)
		if err != nil {
			Err = fmt.Errorf("failed to read file %s: %w", filename, err)
			res <- nil
			return
		}
		res <- data
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-res:
		s.uploadSemaphore.Release(1)
		if v == nil {
			return nil, Err
		}
		return &pb.DownloadResponse{FileData: v}, nil
	}
}
