// Este programa requiere 2 parámetros de entrada :
//      - Nombre fichero json de Lefs
//        - Número de ciclo final
//
// Ejemplo : censim  testdata/PrimerEjemplo.rdp.subred0.json  5
package main

import (
	"distconssim"
	"os"
	"strconv"
)

func main() {
	// cargamos un fichero de estructura Lef en formato json para centralizado
	// os.Args[0] es el nombre del programa que no nos interesa
	node, _ := strconv.Atoi(os.Args[1])
	lefs, err := distconssim.Load("testdata/3subredes.subred" + os.Args[1] + ".json")

	if err != nil {
		println("Couln't load the Petri Net file !")
	}

	ms := distconssim.MakeMotorSimulation(lefs)
	// ciclo 0 hasta ciclo os.args[2]
	cicloFinal, _ := strconv.Atoi(os.Args[2])

	ms.SimularPeriodo(0, distconssim.TypeClock(cicloFinal), node)
}
