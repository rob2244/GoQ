# <center>GoQ</center>

Go Q is a reliable messaging service for applications in kubernetes written entirely in go. GoQ strives to deliver at least once delivery of messages to applications running in the kubernetes cluster.

## Architecture

Conceptually the main abstraction in GoQ is that of the queue manager. Every pod in the cluster using GoQ runs an instance of the GoQ queue manager as a sidecar. An application communicates with it's queue manager using grpc bindings, or alternativley a web api. Each container in the cluster must specify a unique identifier that can be used register it with the queuing system.

When an application wants to send a message to another application running in the cluster, it sends it to the queue manager side car via the rest api or grpc bindings. The queue manager is then responsible for reliable delivery of the message. Internally the queue manager maintains a queue of incoming an outgoing messages for the application.

### GoQ Messages

GoQ messages can be any binary data, GoQ doesn't enforce any messaging standards or constraints. This does put more responsibility on the application developer to know what format messages being sent and recieved are in.

In GoQ there are two types of messages:

- persistent
- transient

Persistent messages are replicated to an external data store. Since persistent messages are saved to a data store they may be re-delivered in case of pod failure (may change this to replicated across pods instead of extenral store). Transient messages on the other hand, will be lost if the pod fails before delivery.

Message size is currently capped at 4,294,967,296 bytes as this is the maximum size of a protobuf bytes scalar type. However this may change in the future.

All messages will have an hmac sha generated when they are queued which will be verified when the message is recieved by the recieving queue manager.

### Service Discovery

The kubernetes dns service is used for service discovery. On initialization each queue manager registers itself with the kubernets service given the unique identifier provided to it by the application. When a queue manager wishes to send a message to another queue manager, the cluster ip of that queue manager is first looked up by the unique identifier.

### Delivery Garauntees

Queue managers strive to garauntee at least once delivery to applications and other queue managers. To this end queue managers utilize retries and exponential back off when trying to deliver messsages. A queue managers retry and exponential back off period are configurable through a Retry Policy.

If a queue manager exhausts it's alloted retries, it will send the message to a poison queue external to GoQ, the poison queue is configurable and extensions can be written to support any data store as a poison queue

### Application Failure

On pod failure any non persistent mesages that have not been sent to another queue manager or delivered to the application will be lost. When the pod comes back online, the queue manager will load any persistent messages from the external store and resume delivery attempts.

When a pod fails the unique cluster identifier must be the same, as any messages that are still outstanding for delivery won't be deliverable if the unique identifier changes.

### Queue Overflow

Each queue manager has a limit on the number of messages it can hold in it's send and recieve queues. If for soome reason the queue is filled due to message volume, any additionall messages submitted to the filled queue will be put int the configured poison queue.

### Metrics

Each Queue Manager maintains performance metrics using the go-metrics library, these metrics are periodically flushed to pluggable "metrics sinks". The metrics api can be extended to write metrics to a number of different data stores.
