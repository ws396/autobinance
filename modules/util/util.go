package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func ShowJSON(data interface{}) {
	j, err := json.MarshalIndent(data, "", "    🐱") // 🐱🐱🐱🐱🐱🐱🐱🐱!!
	if err != nil {
		fmt.Println(err)
		//return
	}

	fmt.Println(string(j))
}

func WriteToLog(message string) {
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create("./log.txt")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()

	//writer := bufio.NewWriter(file)
	_, err = file.WriteString(message)
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}

	//writer.Flush()
}