package main

import "log"

func main() {
	g, err := NewGame(40, 30, 640, 480)
	if err != nil {
		log.Fatal(err)
	}

	err = g.Run()

	if err != nil {
		log.Fatal(err)
	}
}
