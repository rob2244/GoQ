package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/rob2244/GoQ/pkg/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type config struct {
	queueManagerPort int
	tls              bool
	certFilepath     string
	keyFilepath      string
}

func main() {
	cfg := parseCommandLineArgs()
	startGRPCServer(cfg)
}

func startGRPCServer(cfg *config) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.queueManagerPort))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", cfg.queueManagerPort, err)
	}

	var opts []grpc.ServerOption
	if cfg.tls {
		creds, err := credentials.NewServerTLSFromFile(cfg.certFilepath, cfg.keyFilepath)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}

		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	defer grpcServer.GracefulStop()

	srv, err := queue.NewServer(queue.ServerConfig{
		DeliveryBuffLen: 100,
		SendBuffLen:     100,
		UUID:            "12345",
		MaximumBackoff:  time.Millisecond * 10000,
		DialTimeout:     time.Millisecond * 10000,
		TLS:             false,
	})
	if err != nil {
		log.Fatalf("Error while creating grpc server: %v", err)
	}

	queue.RegisterQueueManagerServer(grpcServer, srv)

	log.Printf("Queue Manager listening on port %d", cfg.queueManagerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error staring control server: %v", err)
	}
}

func parseCommandLineArgs() *config {
	queueManagerPort := flag.Int("queueManagerPort", 10000, "Port the control server should listen on")
	tls := flag.Bool("tls", true, "Whether tls is enabled or not, defaults to true")
	certFilepath := flag.String("certFilepath", "", "Filepath to tls certificate to use")
	keyFilepath := flag.String("keyFilepath", "", "Filepath to pem file for tls certificate")

	flag.Parse()

	if *tls && (*certFilepath == "" || *keyFilepath == "") {
		log.Fatal("tls specified, but missing path to pem file or cert file")
	}

	return &config{*queueManagerPort, *tls, *certFilepath, *keyFilepath}
}
