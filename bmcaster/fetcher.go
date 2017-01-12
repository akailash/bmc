package main

import (
	"github.com/golang/protobuf/proto"
	"log"
	"net"
)

func sendMsgToClient(conn net.Conn) {
	log.Println("Connection established")
	//Close the connection when the function exits
	defer conn.Close()
	data := make([]byte, MaxDatagramSize)
	var repeat int32 = 1
	for repeat != 0 {
		//Read the data waiting on the connection and put it in the data buffer
		n, err := conn.Read(data)
		if err != nil {
			log.Println(err)
			return
		}
		protodata := new(Header)
		err = proto.Unmarshal(data[0:n], protodata)
		if err != nil {
			log.Println(err)
			return
		}
		b, err := GetMsgFromDisk(protodata.GetMsgId(), StoreDir)
		if err != nil {
			log.Println(err)
			return
		}
		n, err = conn.Write(b)
		log.Println("Sent", n, "bytes")
		repeat = protodata.GetMsgLength()
		//log.Println("Repeat = ", repeat)
	}
}

func TcpMsgServer() {
	log.Println("Starting TcpMsgServer")
	// listen on all interfaces
	ln, err := net.Listen("tcp", FetcherPort)
	if err != nil {
		// handle error
		log.Println(err)
		return
	}
	// accept connection on port
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go sendMsgToClient(conn)
	}

}

func TcpFetcher(store *msgstore, keys []int64, src string) {
	log.Println("Starting TcpFetcher")
	if len(keys) != 0 {
		conn, err := net.Dial("tcp", src+FetcherPort)
		if err != nil {
			log.Println("Error fetching from", src, err)
			return
		}
		defer conn.Close()
		buf := make([]byte, MaxDatagramSize)

		for i, k := range keys {
			log.Println("Fetcher fetching MsgID:", k)
			if CheckMsgDisk(k, StoreDir) {
				log.Println(k, "exists on disk now")
				continue
			}
			head := new(Header)
			head.MsgId = proto.Int64(k)
			head.MsgLength = proto.Int32(int32(len(keys) - i - 1))
			head.MsgType = Header_CONFIG.Enum()
			b, err := proto.Marshal(head)
			if err != nil {
				log.Println(err)
				break
			}
			log.Println("Fetcher fetching MsgID:", k, "from", src)

			n, err := conn.Write(b)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Wrote header ", n)
			n, err = conn.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Read message ", n)
			m := new(NewMsg)
			err = proto.Unmarshal(buf[:n], m)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("CONFIG-UPDATE-RECEIVED { \"update_id\" =", k, "}")
			store.Add(k)
			SaveAsFile(k, b, StoreDir)
		}

	}

}
