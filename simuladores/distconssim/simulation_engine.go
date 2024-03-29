/** Package centralsim with several files to offer a centralized simulation
PROPOSITO: Tipo abstracto para realizar la simulacion de una (sub)RdP.
COMENTARIOS:
	- El resultado de una simulacion local sera un slice dinamico de
	componentes, de forma que cada una de ella sera una structura estatica de
	dos enteros, el primero de ellos sera el codigo de la transicion disparada y
	el segundo sera el valor del reloj local para el que se disparo.
*/
package distconssim

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

var TIEMPO_SEGURIDAD TypeClock = 1
var TOTAL_REDES = 3

// TypeClock defines integer size for holding time.
type TypeClock int64

// ResultadoTransition holds fired transition id and time of firing
type ResultadoTransition struct {
	CodTransition     IndLocalTrans
	ValorRelojDisparo TypeClock
}

// SimulationEngine is the basic data type for simulation execution
type SimulationEngine struct {
	iiRelojlocal      TypeClock // Valor de mi reloj local
	ilMislefs         Lefs      // Estructura de datos del simulador
	IlEventos         EventList //Lista de eventos a procesar
	IlEventosRemotos  []EventList
	ivTransResults    []ResultadoTransition // slice dinamico con los resultados
	EventNumber       float64               // cantidad de eventos ejecutados
	CanalComunicacion chan bool
	NodeId            int
}

// MakeMotorSimulation : inicializar SimulationEngine struct
func MakeMotorSimulation(alLaLef Lefs) SimulationEngine {
	m := SimulationEngine{}

	m.iiRelojlocal = 0
	m.ilMislefs = alLaLef
	m.IlEventos = MakeEventList(100) //aun siendo dinamicos...
	m.IlEventosRemotos = make([]EventList, TOTAL_REDES, TOTAL_REDES)
	for i := 0; i < TOTAL_REDES; i++ {
		m.IlEventosRemotos[i] = MakeEventList(100)
	}

	m.ivTransResults = make([]ResultadoTransition, 0, 100)
	m.EventNumber = 0

	return m
}

// disparar una transicion. Esto es, generar todos los eventos
//	   ocurridos por el disparo de una transicion
//   RECIBE: Indice en el vector de la transicion a disparar
func (se *SimulationEngine) dispararTransicion(ilTr IndLocalTrans, nodeId int) {
	var ListaEventosEnviar EventList
	log.Println("Se va a disparar la transición: " + fmt.Sprint(ilTr))
	// Prepare 5 local variables
	trList := se.ilMislefs.IaRed              // transition list
	timeTrans := trList[ilTr].IiTiempo        // time to spread to new events
	timeDur := trList[ilTr].IiDuracionDisparo // firing time length
	listIul := trList[ilTr].TransConstIul     // Iul list of pairs Trans, Ctes
	listPul := trList[ilTr].TransConstPul     // Pul list of pairs Trans, Ctes

	// First apply Iul propagations (Inmediate : 0 propagation time)
	for _, trCo := range listIul {
		/*
			log.Println("A ver trList: " + fmt.Sprint(trList))
			log.Println("A ver trList[0]: " + fmt.Sprint(trList[0]))
			log.Println("A ver listIul: " + fmt.Sprint(listIul))
			log.Println("A ver trCo: " + fmt.Sprint(trCo))
			log.Println("A ver trCo[0]: " + fmt.Sprint(trCo[0]))
			log.Println("A ver IndLocalTrans 0: " + fmt.Sprint(trList[IndLocalTrans(0)]))

		*/
		trCo[0] = se.iulDeRemotoAlocal(trCo[0])
		trList[IndLocalTrans(trCo[0])].updateFuncValue(TypeConst(trCo[1]))
	}

	// Generamos eventos ocurridos por disparo de transicion ilTr
	for _, trCo := range listPul {
		// tiempo = tiempo de la transicion + coste disparo
		// Si es una transición negativa, preparala para enviar
		eventoGenerado := Event{timeTrans + timeDur,
			IndLocalTrans(trCo[0]),
			TypeConst(trCo[1]), false}

		if se.esEventoMio(eventoGenerado) {
			log.Println("El evento generado nos pertenece.")
			se.IlEventos.inserta(eventoGenerado)

		} else {
			log.Println("El evento pertenece a otra subred: " + fmt.Sprintf("%v", eventoGenerado.IiTransicion))
			ListaEventosEnviar.inserta(eventoGenerado)
		}
	}

	// Send ListaEventosEnviar broadcast (neighbours)
	if len(ListaEventosEnviar) > 0 {
		log.Println("Broadcasting eventos generados en tiempo " + fmt.Sprintf("%v", timeDur))
		NODES[nodeId].SendBroadcast(Message{ID: 0, Type: "NEW_EVENTS", Source: NODES[nodeId], EventList: ListaEventosEnviar})
	}
}

/* fireEnabledTransitions dispara todas las transiciones sensibilizadas
   		PROPOSITO: Accede a lista de transiciones sensibilizadas y procede con
	   	su disparo, lo que generara nuevos eventos y modificara el marcado de
		transicion disparada. Igualmente anotara en resultados el disparo de
		cada transicion para el reloj actual dado
*/
func (se *SimulationEngine) fireEnabledTransitions(aiLocalClock TypeClock, nodeId int) {
	for se.ilMislefs.haySensibilizadas() { //while
		liCodTrans := se.ilMislefs.getSensibilizada()
		se.dispararTransicion(liCodTrans, nodeId)

		// Anotar el Resultado que disparo la liCodTrans en tiempoaiLocalClock
		se.ivTransResults = append(se.ivTransResults,
			ResultadoTransition{liCodTrans, aiLocalClock})
	}
}

// tratarEventos : Accede a lista eventos y trata todos con tiempo aiTiempo
func (se *SimulationEngine) tratarEventos() {
	var leEvento Event
	aiTiempo := se.iiRelojlocal

	for se.IlEventos.hayEventos(aiTiempo) {
		leEvento = se.IlEventos.popPrimerEvento() // extraer evento más reciente
		idTr := leEvento.IiTransicion             // obtener transición del evento
		trList := se.ilMislefs.IaRed              // obtener lista de transiciones de Lefs

		// Establecer nuevo valor de la funcion
		trList[idTr].updateFuncValue(leEvento.IiCte)
		// Establecer nuevo valor del tiempo
		trList[idTr].actualizaTiempo(leEvento.IiTiempo)

		se.EventNumber++
	}

}

// avanzarTiempo : Modifica reloj local con minimo tiempo de entre
//	   recibidos del exterior o del primer evento en lista de eventos
func (se *SimulationEngine) avanzarTiempo() TypeClock {
	time.Sleep(50 * time.Millisecond)
	//	tiempoLocal := se.iiRelojlocal
	tiempoLocalLista := se.IlEventos.tiempoPrimerEvento()
	tiempoRemotoMenor := se.obtenerMenorTiempoEntreListas()

	nextTime := TypeClock(-1)
	// Si tengo eventos en mi lista local, los tendré en cuenta.
	if tiempoLocalLista != -1 {
		//log.Println("Debo tener en cuenta la lista local")
		nextTime = Mins([]TypeClock{tiempoRemotoMenor, tiempoLocalLista})
		// Si no -> Directamente implica que tengo Look Ahead.
	} else {
		//log.Println("NO Debo tener en cuenta la lista local")
		nextTime = tiempoRemotoMenor
	}

	//log.Println("Tiempo remoto 0 más bajo:" + fmt.Sprint(tiempoRemoto0))
	//log.Println("Tiempo remoto 1 más bajo:" + fmt.Sprint(tiempoRemoto1))
	log.Println("Tiempo en LISTA LOCAL:" + fmt.Sprint(tiempoLocalLista))
	//log.Println("Tiempo local:" + fmt.Sprint(tiempoLocal))

	// Limpiar eventos...
	for i := 0; i < TOTAL_REDES; i++ {
		if i != se.NodeId {
			se.IlEventosRemotos[i].eliminaPrimerEvento()
		}
	}

	log.Println("NEXT CLOCK...... : ", nextTime)

	return nextTime
}

func (se *SimulationEngine) obtenerMenorTiempoEntreListas() TypeClock {
	var tiemposRemotos string
	var relojesRemotos []TypeClock
	for i := 0; i < TOTAL_REDES; i++ {
		if i != se.NodeId {
			relojesRemotos = append(relojesRemotos, se.IlEventosRemotos[i].tiempoPrimerEvento())
			tiemposRemotos += "|" + fmt.Sprint(se.IlEventosRemotos[i].tiempoPrimerEvento())
		}
	}
	log.Println("Tiempos en LISTAS REMOTAS: " + tiemposRemotos)
	return Mins(relojesRemotos)
}

func Mins(values []TypeClock) TypeClock {
	minValue := values[0]
	for _, value := range values {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

// devolverResultados : Mostrar los resultados de la simulacion
func (se SimulationEngine) devolverResultados() {
	resultados := "----------------------------------------\n"
	resultados += "Resultados del simulador local\n"
	resultados += "----------------------------------------\n"
	if len(se.ivTransResults) == 0 {
		resultados += "No esperes ningun resultado...\n"
	}

	for _, liResult := range se.ivTransResults {
		resultados +=
			"TIEMPO: " + fmt.Sprintf("%v", liResult.ValorRelojDisparo) +
				" -> TRANSICION: " + fmt.Sprintf("%v", liResult.CodTransition) + "\n"
	}

	resultados += "\n ========== TOTAL DE TRANSICIONES DISPARADAS = " +
		fmt.Sprintf("%d", len(se.ivTransResults)) + "\n"

	log.Println(resultados)
}

// SimularUnpaso de una RdP con duración disparo >= 1
func (se *SimulationEngine) simularUnpaso(CicloFinal TypeClock) {
	se.ilMislefs.actualizaSensibilizadas(se.iiRelojlocal)

	log.Println("-----------Stack de transiciones sensibilizadas---------")
	se.ilMislefs.IsTransSensib.ImprimeTransStack()
	log.Println("-----------Final Stack de transiciones---------")

	// Fire enabled transitions and produce events
	se.fireEnabledTransitions(se.iiRelojlocal, se.NodeId)

	log.Println("-----------Lista eventos después de disparos---------")
	se.IlEventos.Imprime()
	log.Println("-----------Final lista eventos---------")

	for i := 0; i < TOTAL_REDES; i++ {
		if i != se.NodeId && len(se.IlEventosRemotos[i]) == 0 {
			NODES[se.NodeId].Send(Message{ID: 0, Type: REQUEST_TIME, Source: NODES[se.NodeId], EventList: EventList{}}, NODES[i])
			<-se.CanalComunicacion
		}
	}

	log.Println("Todas mis listas REMOTAS están ready. Avanzamos reloj a tiempo menor...\n")
	//fmt.Sprint(se.IlEventosRemoto0) + " | " + fmt.Sprint(se.IlEventosRemoto1))
	se.iiRelojlocal = se.avanzarTiempo()
	if se.iiRelojlocal == -1 {
		se.iiRelojlocal = CicloFinal
	}
	log.Println("Se ha adelantado el reloj: ", fmt.Sprint(se.iiRelojlocal))

	se.tratarEventos()

	// enviar mensajes null a todos los PLs vecinos con un
	//estampilla de tiempo que indique el límite mínimo de tiempo
	//en futuros mensajes enviados a ese PL ( tiempo en curso +
	//previsión_tiempo_mínimo_futuro)
	log.Println("--------------------------------------")
	log.Println("--------------------------------------")
	log.Println("--------------------------------------")
}

func (se *SimulationEngine) eventoDeRemotoAlocal(evento Event) Event {
	for i, trans := range se.ilMislefs.IaRed {
		if evento.IiTransicion == trans.IiIndLocal {
			evento.IiTransicion = IndLocalTrans(i)
		}
	}
	return evento
}

func (se *SimulationEngine) iulDeRemotoAlocal(id int) int {
	for _, trans := range se.ilMislefs.IaRed {
		for i, iul := range trans.TransConstIul {
			if iul[0] == id {
				return i
			}
		}
	}

	return id
}

// Verifica si un evento concreto ha de ser procesado en
// la red local
func (se *SimulationEngine) esEventoMio(evento Event) bool {
	for _, trans := range se.ilMislefs.IaRed {
		if trans.IiIndLocal == evento.IiTransicion {
			return true
		}
	}

	return false
}

// SimularPeriodo de una RdP
// RECIBE: - Ciclo inicial (por si marcado recibido no se corresponde al
//				inicial sino a uno obtenido tras simular ai_cicloinicial ciclos)
//		   - Ciclo con el que terminamos
func (se *SimulationEngine) SimularPeriodo(CicloInicial, CicloFinal TypeClock, NodeId int) {
	log.Println("--------------- DISTRIBUTED SIMULATION STARTS! ------------------")

	se.NodeId = NodeId
	se.CanalComunicacion = make(chan bool)
	go NODES[NodeId].LaunchReceiver(se)

	time.Sleep(10 * time.Second)
	log.Println("Subnet " + strconv.Itoa(NodeId) + " started")
	ldIni := time.Now()

	// Inicializamos el reloj local
	// ------------------------------------------------------------------
	se.iiRelojlocal = CicloInicial

	for se.iiRelojlocal < CicloFinal {
		log.Println("RELOJ LOCAL !!!  = ", se.iiRelojlocal)
		//se.ilMislefs.ImprimeLefs()

		se.simularUnpaso(CicloFinal)
	}

	elapsedTime := time.Since(ldIni)

	fmt.Printf("Eventos por segundo = %f",
		se.EventNumber/elapsedTime.Seconds())

	/*	// Devolver los resultados de la simulacion
		se.devolverResultados()
		result := "\n---------------------\n"
		result += "TIEMPO SIMULADO en ciclos: " +
			fmt.Sprintf("%d", Nciclos-CicloInicial) + "\n"
		result += "TIEMPO ejecución REAL simulación: " +
			fmt.Sprintf("%v", elapsedTime.String()) + "\n"
		log.Println(result)
	*/
}
