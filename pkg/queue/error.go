package queue

import "fmt"

// Type represents the two available queue manager queue types
type Type string

const (
	// Send is the queue for outgoing messages
	Send Type = "Send"
	// Deliver is the queue for incoming messages
	Deliver Type = "Deliver"
)

// MaxCapacityError is returned when either the send or deliver
// queue manager queue is filed to capacity
type MaxCapacityError struct {
	ManagerUUID string
	QueueLength int
	QueueCap    int
	QueueType   Type
}

func (err *MaxCapacityError) Error() string {
	return fmt.Sprintf("Queue manager: %s - %s queue full, length: %d capacity: %d",
		err.ManagerUUID, err.QueueType, err.QueueLength, err.QueueCap)
}

// MaxCapacity returns a new instance of the MaxCapacity error type
func MaxCapacity(managerUUID string, qLen, qCap int, qType Type) *MaxCapacityError {
	return &MaxCapacityError{
		managerUUID,
		qLen,
		qCap,
		qType,
	}
}
