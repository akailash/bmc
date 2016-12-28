package main

import (
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}
func MsgServer() {
	log.Println("Starting MsgServer")
	//	http.HandleFunc("/", JsonResponse)
	http.Handle("/", http.FileServer(http.Dir(StoreDir)))
	log.Fatal(http.ListenAndServe(FetcherPort, nil))
}

func Fetcher(store *msgstore, wanted *wantedstore) {
	log.Println("Starting Fetcher")
	client := NewTimeoutClient(HttpConnectTO*time.Second, HttpReadWriteTO*time.Second)
	for {
		time.Sleep(FetchDuration * time.Millisecond)
		//TODO should use reader lock on wanted but since writer lock is acquired within the loop its not done
		for k, srcs := range wanted.cache {
			log.Println("Fetcher fetching MsgID:", k)
			for _, src := range srcs {
				log.Println("Fetcher fetching MsgID:", k, "from", src)

				resp, err := client.Get("http://" + src + FetcherPort + "/" + strconv.FormatInt(k, 10))
				if err != nil {
					log.Println("Error fetching ", k, "from", src, err)
					continue
				}
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					resp.Body.Close()
					continue
				}
				m := new(NewMsg)
				proto.Unmarshal(b, m)
				log.Println("Downloaded message: ", m)
				resp.Body.Close()
				wanted.Delete(k)
				store.Add(m)
				SaveAsFile(m, StoreDir)
				//Fetch next key
				break
			}

		}

	}

}
