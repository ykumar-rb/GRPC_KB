/*
 *
 * Copyright 2019 HPE Inc.
 */

// Package main implements a client for Deployer service.
package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "github.com/grpc_poc/grpc_examples/grpc_TLS_withClient_Auth/common/proto"
)

const (
	address     = "localhost:50051"
//	address     = "hpe.com:50051"
)

// Authentication holds the login/password
type Authentication struct {
  Login    string
  Password string
}

// GetRequestMetadata gets the current request metadata
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
  return map[string]string{
    "login":    a.Login,
    "password": a.Password,
  }, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (a *Authentication) RequireTransportSecurity() bool {
  return true
}

func main() {
	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile("common/cert/server.crt", "")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}

	// Setup the login/pass
	auth := Authentication{
		Login:    "grpc_user",
		Password: "grpc123#",
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDeployerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()

	deploymentReq := &pb.DeployRequest{DeployType: "vm", Flavor: "medium", Target: "compute-1"}
	log.Printf("Sending Deployment Request: %v", deploymentReq)
	r1, err := c.Deploy(ctx, deploymentReq)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("DeployResponse: %s", r1.Status)
}
