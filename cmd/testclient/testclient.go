package main

import (
	"context"
	"log"
	"time"

	"github.com/rob2244/GoQ/pkg/queue"
	"google.golang.org/grpc"
)

// var recieverID *string

func main() {
	// flag.String(*recieverID, "receiverID", "The ID of the reciever the message should be send to")
	// flag.Parse()

	recieverID := "172.17.0.10"

	tmt, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	conn, err := grpc.DialContext(tmt, "localhost:10000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to establish connection to queue manager: %v", err)
	}
	defer conn.Close()

	client := queue.NewQueueManagerClient(conn)

	for {
		msg := queue.Message{Data: []byte("Ping"), RecieverID: recieverID}
		tmt, cancel = context.WithTimeout(context.Background(), time.Second*100)
		_, err := client.QueueMessage(tmt, &msg)
		if err != nil {
			log.Printf("Messsage recieved while queueing message: %v", err)
		}

		log.Println("Successfully sent message")

		time.Sleep(time.Second * 5)
	}
}
