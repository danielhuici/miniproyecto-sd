package main

import (
	"distconssim"
	"fmt"
	"github.com/melbahja/goph"
	"log"
	"strconv"
)

func SSHExecute(node distconssim.Node, command string) {
	// Start new ssh connection with private key.
	auth, err := goph.Key("/home/a758635/.ssh/id_rsa", "")
	if err != nil {
		log.Fatal(err)
	}

	client, err := goph.New("a758635", node.Address, auth)
	if err != nil {
		log.Fatal(err)
	}

	// Defer closing the network connection.
	defer client.Close()

	client.Run(command)
	out, err := client.Run(command)

	fmt.Println("SSH"+strconv.Itoa(node.ID)+":", string(out))
}
