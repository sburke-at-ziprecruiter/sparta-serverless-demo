package main

import (
	"github.com/seanpburke/sparta-serverless-demo/pkg/config"
	"github.com/seanpburke/sparta-serverless-demo/pkg/table"

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
