package main

import (
	"MyGo/dynamo_db/pkg/config"
	"MyGo/dynamo_db/pkg/table"

	"fmt"
	"os"
)

func main() {

	out, err := table.CreateTable()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Created the DynamoDB table", config.Config.Table)
	fmt.Println(out)
}
