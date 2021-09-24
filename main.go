package main

import "log"

func main() {
	g, err := NewGame(500, 500)
	if err != nil {
		log.Fatal(err)
	}

	err = g.Run()

	if err != nil {
		log.Fatal(err)
	}
}
