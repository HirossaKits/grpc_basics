package main

import (
	"basics/pb"
	"testing"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_callListFiles(t *testing.T) {
	type args struct {
		client pb.FileServiceClient
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callListFiles(tt.args.client)
		})
	}
}

func Test_callDownload(t *testing.T) {
	type args struct {
		client pb.FileServiceClient
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callDownload(tt.args.client)
		})
	}
}

func TestCallUpload(t *testing.T) {
	type args struct {
		client pb.FileServiceClient
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CallUpload(tt.args.client)
		})
	}
}
