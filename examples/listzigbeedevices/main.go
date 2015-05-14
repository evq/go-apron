package main

import (
	"fmt"
	"github.com/evq/go-apron"
)


func main() {
	db, _ := apron.Open("./apron.db")
	fmt.Println(db.GetZigbeeDevices())
}
