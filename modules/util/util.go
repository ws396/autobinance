package util

import (
	"encoding/json"
	"fmt"
	"log"
)

func ShowJSON(data interface{}) {
	j, err := json.MarshalIndent(data, "", "    🐱") // 🐱🐱🐱🐱🐱🐱🐱🐱!!
	if err != nil {
		log.Panicln(err)
	}

	fmt.Println(string(j))
}
