package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func SaveAsJson(m Msg, dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := dir + strconv.Itoa(m.MsgID) + ".json"
	os.Remove(path)

	b, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	ioutil.WriteFile(path, b, 0644)
}

func CheckJsonDisk(msgid int, dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := dir + strconv.Itoa(msgid) + ".json"
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
