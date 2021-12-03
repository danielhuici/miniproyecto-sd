//Package centralsim with several files to offer a centralized simulation
// This file deals with the low level lefs encoding of a petri net
package centralsim

import (
	"encoding/json"
	"fmt"
	"os"
)

//type TypeIndexSubnet int32

//----------------------------------------------------------------------------

// Lefs es el tipo de datos principal que gestiona el disparo de transiciones.
type Lefs struct {
	// Slice de transiciones de esta subred
	IaRed TransitionList `json:"ia_red"`
	//ii_indice int32	// Contador de transiciones agnadidas, Necesario ???
	// Identificadores de las transiciones sensibilizadas para
	// T = Reloj local actual. Slice que funciona como Stack
	IsTransSensib TransitionStack
}

// Load obtains Lefs from a json file
func Load(filename string) (Lefs, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open json lefs file: %v\n", err)
		return Lefs{}, err
	}
	defer file.Close()

	result := Lefs{}
	if err := json.NewDecoder(file).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "Decode json 		file: %v\n", err)
		return Lefs{}, err
	}

	result.IsTransSensib = MakeTransitionStack(100) //aun siendo dinamicos...

	return result, nil
}

/*
-----------------------------------------------------------------
   METODO: agnade_sensibilizada
   RECIBE: Transicion sensibilizada a a�adir
   DEVUELVE: OK si todo va bien o ERROR en caso contrario
   PROPOSITO: A�ade a la lista de transiciones sensibilizadas
   HISTORIA DE CAMBIOS:
COMENTARIOS:
-----------------------------------------------------------------
*/
func (l *Lefs) agnadeSensibilizada(aiTransicion IndLocalTrans) bool {
	l.IsTransSensib.push(aiTransicion)
	return true // OK
}

// haySensibilizadas permite saber si tenemos transiciones sensibilizadas;
// se supone que previamente se ha llamado a actualizaSensibilizadas(relojLocal)
func (l Lefs) haySensibilizadas() bool {
	return !l.IsTransSensib.isEmpty()
}

// getSensibilizada coge el primer identificador de la lista de transiciones
//	 		sensibilizadas
func (l *Lefs) getSensibilizada() IndLocalTrans {
	if (*l).IsTransSensib.isEmpty() {
		return -1
	}

	return (*l).IsTransSensib.pop()
}

// actualizaSensibilizadas recorre toda la lista de transiciones
//	   e inserta trans sensibilizadas, con el mismo tiempo que el reloj local,
//  en la pila de transiciones sensibilizadas
func (l *Lefs) actualizaSensibilizadas(aiRelojLocal TypeClock) bool {
	for IndT, t := range (*l).IaRed {
		if t.IiValorLef <= 0 && t.IiTiempo == aiRelojLocal {
			(*l).IsTransSensib.push(IndLocalTrans(IndT))
		}
	}
	return true
}

// ImprimeTransiciones para depurar errores
func (l Lefs) ImprimeTransiciones() {
	fmt.Println(" ")
	fmt.Println("------IMPRIMIMOS LA LISTA DE TRANSICIONES---------")
	for _, tr := range l.IaRed {
		tr.ImprimeValores()
	}
	fmt.Println("------FINAL DE LA LISTA DE TRANSICIONES---------")
	fmt.Println(" ")
}

// ImprimeLefs : Imprimir los atributos de la clase para depurar errores
func (l Lefs) ImprimeLefs() {

	fmt.Println("STRUCT LEFS")
	//fmt.Println ("\tNº transiciones: ", self.ii_indice)
	fmt.Println("\tNº transiciones: ", l.IaRed.length())

	fmt.Println("------Lista transiciones---------")
	for _, tr := range l.IaRed {
		tr.Imprime()
	}
	fmt.Println("------Final lista transiciones---------")

	fmt.Println("FINAL ESTRUCTURA LEFS")
}
