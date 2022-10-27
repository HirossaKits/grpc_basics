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
)

func main() {
	conn, err := grpc.Dial("localhost:3111", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)
	// callListFiles(client)
	// callDownload(client)
	// CallUpload(client)
	CallUploadAndNotifyProgress(client)
}

// unary rpc
func callListFiles(client pb.FileServiceClient) {
	res, err := client.ListFiles(context.Background(), &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(res.GetFilenames())
}

// server streaming rpc
func callDownload(client pb.FileServiceClient) {

	req := &pb.DownloadRequest{Filename: "user.txt"}
	stream, err := client.Download(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(1 * time.Second)
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
func CallUploadAndNotifyProgress(client pb.FileServiceClient) {
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
}
