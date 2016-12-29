package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func SaveAsFile(id int64, b []byte, dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		} else {
			log.Println(err)
		}
	}

	path := dir + strconv.FormatInt(id, 10)
	os.Remove(path)

	ioutil.WriteFile(path, b, 0644)
}

func CheckMsgDisk(msgid int64, dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		return false
	}

	path := dir + strconv.FormatInt(msgid, 10)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func GetMsgFromDisk(id int64, dir string) ([]byte, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	path := dir + strconv.FormatInt(id, 10)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(path)
}
