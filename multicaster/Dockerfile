FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go get "github.com/golang/protobuf/proto"
RUN go build -o main . 
CMD ["/app/main"]
