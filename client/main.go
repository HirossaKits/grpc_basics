package main

import (
	"basics/pb"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	h, _ := os.UserHomeDir()
	cert := filepath.Join(h, "Library/Application Support/mkcert/rootCA.pem")
	fmt.Println(cert)
	creds, err := credentials.NewClientTLSFromFile(cert, "")
	if err != nil {
		log.Fatalf("failed to new creds")
	}

	conn, err := grpc.Dial("localhost:3111", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)
	// callListFiles(client)
	callDownload(client)
	// CallUpload(client)
	// CallUploadAndNotifyProgress(client)
}

// unary rpc
func callListFiles(client pb.FileServiceClient) {
	md := metadata.New(map[string]string{"authorization": "Bearer tests"})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(res.GetFilenames())
}

// server streaming rpc
func callDownload(client pb.FileServiceClient) {
	// time out
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.DownloadRequest{Filename: "user.txt"}
	stream, err := client.Download(ctx, req)
	if err != nil {

		resErr, ok := status.FromError(err)

		if ok {
			log.Fatalf("code = %v desc = %v", resErr.Code(), resErr.Message())
		} else {
			log.Fatal(err)
		}
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}
		fmt.Print(string(res.GetData()))
	}
}

// client streaming rpc
func CallUpload(client pb.FileServiceClient) {
	cd, _ := os.Getwd()
	p := filepath.Join(cd, "public", "date.txt")

	f, err := os.Open(p)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	buf := make([]byte, 8)
	for {
		n, err := f.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		req := &pb.UploadRequest{Data: buf[:n]}
		sendErr := stream.Send(req)
		if sendErr != nil {
			log.Fatal(sendErr)
		}
		time.Sleep(1 * time.Second)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v", res.GetSize())
}

// bi streaming rpc
func CallUploadAndNotifyProgress(client pb.FileServiceClient) ([]byte, error) {
	cd, _ := os.Getwd()
	p := path.Join(cd, "public", "date.txt")

	f, err := os.Open(p)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stream, err := client.UploadAndNotifyProgress(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	buf := make([]byte, 8)
	go func() {
		for {
			n, err := f.Read(buf)
			if n == 0 || err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			req := &pb.UploadAndNotifyProgressRequest{Data: buf[:n]}
			sendErr := stream.Send(req)
			if sendErr != nil {
				log.Fatal(sendErr)
			}
			time.Sleep(1 * time.Second)
		}
		err := stream.CloseSend()
		if err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan struct{})
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("%v", res.GetMsg())
		}
		close(ch)
	}()
	<-ch

	return buf, err
}
