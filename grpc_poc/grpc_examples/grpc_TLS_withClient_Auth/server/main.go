/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	pb "github.com/grpc_poc/grpc_examples/grpc_TLS_withClient_Auth/common/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// private type for Context keys
type contextKey int

const (
	clientIDKey contextKey = iota
	port = ":50051"

//	port = "10.10.45.51:50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// Deploy implements helloworld.GreeterServer
func (s *server) Deploy(ctx context.Context, in *pb.DeployRequest) (*pb.DeployResponse, error) {
	log.Printf("Received Deployment Request: %v", in)
	return &pb.DeployResponse{Status: "success"}, nil
}

// authenticateAgent check the client credentials
func authenticateClient(ctx context.Context, s *server) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientLogin != "grpc_user" {
			return "", fmt.Errorf("unknown user %s", clientLogin)
		}
		if clientPassword != "grpc123#" {
			return "", fmt.Errorf("bad password %s", clientPassword)
		}
		log.Printf("authenticated client: %s", clientLogin)
		return "42", nil
	}
	return "", fmt.Errorf("missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*server)
	if !ok {
		return nil, fmt.Errorf("unable to cast server")
	}
	clientID, err := authenticateClient(ctx, s)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile("common/cert/server.crt", "common/cert/server.key")
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor)}

	log.Print("Listening for request....")
	s := grpc.NewServer(opts...)

	pb.RegisterDeployerServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
