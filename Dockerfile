# create image from the official Go image
FROM golang:1.10.3-alpine3.8

RUN apk add --update tzdata \
    bash curl git;

# Create binary directory, install glide
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# define work directory
ADD . /go/src/github.com/go-squads/floodgate-worker
WORKDIR /go/src/github.com/go-squads/floodgate-worker
RUN apk add make
# serve the app
CMD make run 
