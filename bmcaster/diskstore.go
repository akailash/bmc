package main

import (
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func SaveAsFile(m *NewMsg, dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := dir + strconv.FormatInt(m.GetHead().GetMsgId(), 10)
	os.Remove(path)

	b, err := proto.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	ioutil.WriteFile(path, b, 0644)
}

func CheckJsonDisk(msgid int64, dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := dir + strconv.FormatInt(msgid, 10)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
