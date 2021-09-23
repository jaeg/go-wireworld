package main

import "log"

func main() {
	g, err := NewGame(100, 100, 640, 480)
	if err != nil {
		log.Fatal(err)
	}

	err = g.Run()

	if err != nil {
		log.Fatal(err)
	}
}
