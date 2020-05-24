package main

import (
	"fmt"
	"os"

	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/customer"
	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/movie"
	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/store"
)

func main() {

	// Get my customer
	customerPhone := "828-234-1717"
	cus, err := customer.Get(customerPhone)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Customer:", cus)

	sto, err := store.Get(cus.StorePhone)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Store:", sto)

	rental, err := cus.PutRental()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Rental:", rental)

	mov, err := movie.Get(2013, "Rush")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Movie:", mov)

	movren, err := mov.PutRental(rental.Phone, rental.Date)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("Successfully added rental %s (%d) to %s on %s\n", movren.Title, movren.Year, movren.Phone, movren.Date)
}
