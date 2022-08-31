package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"log/data"
	"log/logs"
	"net"
	"time"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	//	write a log
	logEntry := data.LogEntry{
		Name:      input.Name,
		Data:      input.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "Failed"}
		return res, err
	}
	return &logs.LogResponse{Result: "Logged via gRPC"}, nil
}

func (app *Config) gRPCListen() {
	ls, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listan for gRPC: %v", err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Printf("gRPC Server Started on port %s", gRpcPort)

	if err = s.Serve(ls); err != nil {
		log.Fatalf("Failed to listan for gRPC: %v", err)
	}
}
