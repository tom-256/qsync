package main

import (
	"os"
	"fmt"
)

func main(){
	token := os.Getenv("QIITA_ACCESS_TOKEN")
	fmt.Println(token)
}