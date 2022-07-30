package main

import (
	"fmt"
	"os"

	"github.com/madkins23/go-mongo/mdb"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage: dbping <dbname>")
	} else if access, err := mdb.Connect(os.Args[1], nil); err != nil {
		fmt.Printf("Unable to connect to %s: %s\n", os.Args[1], err)
	} else if err := access.Disconnect(); err != nil {
		fmt.Printf("Unable to disconnect from %s: %s\n", os.Args[1], err)
	}
}
