package main

import (
	"basics/pb"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

func (*server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {

	cd, _ := os.Getwd()
	d := filepath.Join(cd, "public")

	ps, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, p := range ps {
		if !p.IsDir() {
			fileNames = append(fileNames, p.Name())
		}
	}

	res := &pb.ListFilesResponse{
		Filenames: fileNames,
	}

	return res, nil
}

func (*server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {

	n := req.GetFilename()
	cd, _ := os.Getwd()
	p := filepath.Join(cd, "public", n)

	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 8)
	for {
		n, err := f.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		res := &pb.DownloadReesponse{Data: buf[:n]}
		sendErr := stream.Send(res)
		if sendErr != nil {
			return sendErr
		}
	}

	return nil
}

func main() {
	l, err := net.Listen("tcp", "localhost:3111")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &server{})

	fmt.Println("server is running...")
	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
