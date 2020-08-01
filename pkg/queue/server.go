package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server implements a queue manager grpc server
type Server struct {
	sendBuffer     chan *Message
	deliveryBuffer chan *Message
	uuid           string
	config         ServerConfig
	// TODO: not sure if clientCredential is thread safe
	clientCredential *credentials.TransportCredentials
}

// ServerConfig encapsulates the available server options
// for the queue manager grpc server
type ServerConfig struct {
	DeliveryBuffLen int
	SendBuffLen     int
	UUID            string
	MaximumBackoff  time.Duration
	DialTimeout     time.Duration
	TLS             bool
	KeyFilepath     string
}

// NewServer creates a new instance of a Queue Manager grpc server
func NewServer(cfg ServerConfig) (*Server, error) {
	rand.Seed(time.Now().UnixNano())

	if err := validateConfiguration(&cfg); err != nil {
		return nil, err
	}

	var creds credentials.TransportCredentials
	var err error
	if cfg.TLS {
		creds, err = credentials.NewClientTLSFromFile(cfg.KeyFilepath, "")

		if err != nil {
			return nil, fmt.Errorf("Could not process the credentials in .pem file: %v ", err)
		}
	}

	srv := &Server{
		sendBuffer:       make(chan *Message, cfg.DeliveryBuffLen),
		deliveryBuffer:   make(chan *Message, cfg.DeliveryBuffLen),
		uuid:             cfg.UUID,
		config:           cfg,
		clientCredential: &creds,
	}

	go srv.send(context.Background())
	return srv, nil
}

func validateConfiguration(cfg *ServerConfig) error {
	if cfg.TLS && cfg.KeyFilepath == "" {
		return errors.New("tls specified, but missing path to pem file")
	}

	if cfg.DeliveryBuffLen == 0 {
		return errors.New("Delivery buffer length must be greater than 0")
	}

	if cfg.SendBuffLen == 0 {
		return errors.New("Send buffer length must be greater than 0")
	}

	if cfg.UUID == "" {
		return errors.New("UUID for server must be specified")
	}

	if cfg.MaximumBackoff == 0 {
		return errors.New("Maximum backoff must be specified")
	}

	if cfg.DialTimeout == 0 {
		return errors.New("Dial timeout must be specified")
	}

	return nil
}

// QueueMessage tells the queue manager to put a message in either its send or
// its delivery buffer based on the unique id of the reciever specified in the
// message body
func (s *Server) QueueMessage(ctx context.Context, msg *Message) (*Empty, error) {
	log.Printf("Queueing message destined for: %s", msg.RecieverID)

	if msg.RecieverID == s.uuid {
		if qLen := len(s.deliveryBuffer); qLen >= s.config.DeliveryBuffLen {
			return nil, MaxCapacity(s.uuid, qLen, s.config.DeliveryBuffLen, Deliver)
		}

		s.deliveryBuffer <- msg
		return &Empty{}, nil
	}

	if qLen := len(s.sendBuffer); qLen >= s.config.SendBuffLen {
		return nil, MaxCapacity(s.uuid, qLen, s.config.SendBuffLen, Send)
	}
	s.sendBuffer <- msg
	return &Empty{}, nil
}

func (s *Server) send(ctx context.Context) {
	// Try to send message, if sending fails retry with exponential backoff
	// See https://cloud.google.com/iot/docs/how-tos/exponential-backoff
	// TODO: The backoff functionality will currently retry everything
	// even if only one part of the operation fails, fix this to seperate cnxn backoff
	// from message sending backoff

	for {
		msg := <-s.sendBuffer
		fmt.Println("Pulling message off of send buffer")
		if err := s.transfer(ctx, msg); err != nil {
			// TODO: write to poison queue
			fmt.Println("Error while sending message")
		}

	}
}

func (s *Server) transfer(ctx context.Context, msg *Message) error {
	start := time.Now()
	exponent := 1

	backoff := func() {
		dur, _ := time.ParseDuration(fmt.Sprintf("%d ms", exponent+rand.Intn(1000)))
		time.Sleep(dur)
		exponent = exponent + 1
	}

	for {
		remoteIP := msg.RecieverID
		remotePort := 10000

		tmt, cancel := context.WithTimeout(ctx, s.config.DialTimeout)
		defer cancel()

		opts := []grpc.DialOption{grpc.WithBlock()}
		if s.config.TLS {
			opts = append(opts, grpc.WithTransportCredentials(*s.clientCredential))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		// TODO: Use connection pool for more efficient transfer
		conn, err := grpc.DialContext(tmt, fmt.Sprintf("%s:%d", remoteIP, remotePort), opts...)
		if err != nil {
			log.Printf("Recieved error while connecting to remote queue manager: %v", err)

			if start.Sub(time.Now()) > s.config.MaximumBackoff {
				// TODO: return latest error
				return errors.New("Maximum backoff reached, writing message to poison queue")
			}

			backoff()
			continue
		}
		defer conn.Close()

		client := NewQueueManagerClient(conn)
		_, err = client.QueueMessage(ctx, msg)
		if err == nil {
			log.Printf("Successfully sent message")
			return nil
		}

		log.Printf("Recieved error while sending message: %v", err)
		if start.Sub(time.Now()) > s.config.MaximumBackoff {
			// TODO: return latest error
			return errors.New("Maximum backoff reached, writing message to poison queue")
		}

		backoff()
		continue
	}
}

func deliver(s *Server) {
	// Pull message off delivery queue

	// establish grpc client cnxn

	// Try to deliber message, if sending fails retry with exponential backoff
	// See https://cloud.google.com/iot/docs/how-tos/exponential-backoff

	// If all retries fail send message to poson queue
}
