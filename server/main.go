package main

import (
	"basics/pb"
	"bytes"
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

// unary rpc
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

// server streaming rpc
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

func (*server) Upload(stream pb.FileService_UploadServer) error {
	var buf bytes.Buffer
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := &pb.UploadResponse{Size: int32(buf.Len())}
			return stream.SendAndClose(res)
		}
		if err != nil {
			return err
		}

		data := req.GetData()
		log.Printf("%v", string(data))
		buf.Write(data)
	}
}

// bi streaming rpc
func (*server) UploadAndNotifyProgress(stream pb.FileService_UploadAndNotifyProgressServer) error {
	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
		data := req.GetData()
		size += len(data)

		res := &pb.UploadAndNotifyProgressResponse{
			Msg: fmt.Sprintf("%v", size),
		}
		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
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
