package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

func save(fileName string, dbInfo DBInfo) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
	}

	enc := gob.NewEncoder(file)
	err = enc.Encode(dbInfo)
	if err != nil {
		log.Println(err)
	}

}
func load(fileName string) DBInfo {

	var dbInfo DBInfo
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dec := gob.NewDecoder(file)
	err = dec.Decode(&dbInfo)
	if err != nil {
		fmt.Println(err)
	}

	return dbInfo

}
