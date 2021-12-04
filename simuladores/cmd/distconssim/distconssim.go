// Este programa requiere 2 parámetros de entrada :
//      - Nombre fichero json de Lefs
//        - Número de ciclo final
//
// Ejemplo : distconssim  testdata/PrimerEjemplo.rdp.subred0.json  5

// RUN: go run *.go -d 4

package main

import (
	"distconssim"
	"fmt"
	"log"
	"os"
	"strconv"
)

var DISTRIBUTED_MODE = "-d"

func main() {
	if os.Args[1] == DISTRIBUTED_MODE {
		fmt.Println("Ejecutando subred 0")
		go SSHExecute(distconssim.P0, "source $HOME/.profile && cd $HOME/MASTER/miniproyecto/cmd/distconssim && go run *.go 0 4") //+ os.Args[2])
		fmt.Println("Ejecutando subred 1")
		go SSHExecute(distconssim.P1, "source $HOME/.profile && cd $HOME/MASTER/miniproyecto/cmd/distconssim && go run *.go 1 4") // + os.Args[2])
		fmt.Println("Ejecutando subred 2")
		SSHExecute(distconssim.P2, "source $HOME/.profile && cd $HOME/MASTER/miniproyecto/cmd/distconssim && go run *.go 2 4") // + os.Args[2])
	} else {
		nodeId, _ := strconv.Atoi(os.Args[1])
		f, err := os.OpenFile("logfile"+os.Args[1]+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)

		// cargamos un fichero de estructura Lef en formato json para centralizado
		// os.Args[0] es el nombre del programa que no nos interesa

		lefs, err := distconssim.Load("testdata/3subredes.subred" + os.Args[1] + ".json")

		if err != nil {
			println("Couln't load the Petri Net file !")
		}

		ms := distconssim.MakeMotorSimulation(lefs)
		// ciclo 0 hasta ciclo os.args[2]
		cicloFinal, _ := strconv.Atoi(os.Args[2])

		ms.SimularPeriodo(0, distconssim.TypeClock(cicloFinal), nodeId)
	}

}
