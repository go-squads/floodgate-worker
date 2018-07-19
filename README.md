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

## Installation Instructions


## Version History  
