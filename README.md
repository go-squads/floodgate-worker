## Project Ciliwung FloodGate - worker  
This repo contains the code that implements the analytic worker. This worker 
will simulate a stream processor by becoming a consumer that will subscribe to
all topics. The worker is planned to have the following features:  
* The worker will be able to aggregate the log count based on the context &
context of the logs  
* The worker will be able to send alerts if it detects that there is an anomaly
in the data or the number of error logs breached the error threshold limit  
* The worker will be able to write the aggregated log count into InfluxDB  

## Tech Stack  
This program will be written in golang  

## Program Workflow  
The worker will request the messages from Kafka. The worker will sort the
messages based on the message's topics, it will create another goroutine
if it detects that the message contains a new topic  

cluster_analyser is an interface for *sarama.consumer to allow us to mock it for testing.

AnalyticServices creates a Service which will spawn a Consumer group for every topic. When starting the AnalyticService, it will spawn two types of worker. Firstly, it will spawn a worker that will watch for incoming new topic events that is sent by the barito-flow producer. If this worker(we named it topic refresher) detects a new topic, it will spawn a new analytic worker for that particular topic. 

## Installation/Running Instructions
* Clone this repository    
* cd into the project directory: cd floodgate-worker  
* Download kafka from the link provided [here](https://kafka.apache.org/quickstart) 
* Unzip it and cd into the kafka directory  
* Run the zookeeper with: ```bin/zookeeper-server-start.sh config/zookeeper.properties ```
* Run the server with:```bin/kafka-server-start.sh config/server.properties ```(in another terminal window)   
* Run with: "go main.go aw" in the root of the project directory to run the worker (do this in another terminal window) 
* Run the barito-flow producer  
* Send messages to the producer by posting to it  

## Testing Instructions 
go test -coverprofile cover.out  
go tool cover -html=cover.out -o cover.html  
Commands above to see test coverage  

## Version History  
