FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go get "github.com/hashicorp/memberlist"
RUN go get "github.com/golang/protobuf/proto"
RUN go build -o main .
EXPOSE 7373 7946
CMD ["/app/main"]
