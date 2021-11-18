package main

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
)

func main() {
	token := os.Getenv("TOKEN")
	client := slack.New(token)
	info, err := client.GetUserInfo("XX")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", info)
}
