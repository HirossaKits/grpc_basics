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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return status.Error(codes.NotFound, "file was not found")
	}

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
		time.Sleep(1 * time.Second)
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
		log.Printf("%v", string(data))
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

func logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log.Printf("request: %+v", req)

		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}

		log.Printf("response: %+v", resp)

		return resp, nil
	}
}

func authorize(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}
	if token != "test" {
		return nil, status.Error(codes.Unauthenticated, "token is invalid")
	}
	return ctx, nil
}

func main() {
	l, err := net.Listen("tcp", "localhost:3111")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile("ssl/localhost.pem", "ssl/localhost-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				logging(),
				grpc_auth.UnaryServerInterceptor(authorize))),
	)

	pb.RegisterFileServiceServer(s, &server{})

	fmt.Println("server is running...")
	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
