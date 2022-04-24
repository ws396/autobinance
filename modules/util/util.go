package util

import (
	"encoding/json"
	"fmt"
)

func ShowJSON(data interface{}) {
	j, err := json.MarshalIndent(data, "", "    🐱") // 🐱🐱🐱🐱🐱🐱🐱🐱!!
	if err != nil {
		fmt.Println(err)
		//return
	}

	fmt.Println(string(j))
}
