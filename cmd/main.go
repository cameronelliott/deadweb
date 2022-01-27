package main

import "github.com/cameronelliott/deadweb"

func main() {

	go deadweb.Server()

	select {}

}
