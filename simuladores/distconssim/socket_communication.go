package distconssim

import (
	"bytes"
	"encoding/gob"
	"github.com/DistributedClocks/GoVector/govec"
	"log"
	"net"
)

var Logger *govec.GoLog

var REQUEST_TIME = "REQUEST_TIME"
var NEW_EVENTS = "NEW_EVENTS"
var RESPONSE_TIME = "RESPONSE_TIME"

type Message struct {
	ID        int
	Type      string
	Source    Node
	EventList EventList
}

type Node struct {
	ID      int
	Address string
	Port    string
}

var P0 = Node{
	ID:      0,
	Address: "155.210.154.194",
	Port:    ":5000",
}

var P1 = Node{
	ID:      1,
	Address: "155.210.154.195",
	Port:    ":5001",
}

var P2 = Node{
	ID:      2,
	Address: "155.210.154.198",
	Port:    ":5002",
}

var NODES = [3]Node{P0, P1, P2}

// SendBroadcast Envía a todos los nodos vecinos
func (n Node) SendBroadcast(m Message) error {
	var conn net.Conn
	var err error

	for _, node := range NODES {
		if n.ID != node.ID {
			//log.Println("Broadcast send to: " + fmt.Sprint(node.ID))
			conn, err = net.Dial("tcp", node.Address+node.Port)
			if err != nil {
				panic("Client connection error: ")
				log.Println("PANIC: ")
			}
			//		Logger.PrepareSend("Sending Message", nil, govec.GetDefaultLogOptions())
			if err != nil {
				panic(err)
			}

			binBuffer := new(bytes.Buffer)
			gobobj := gob.NewEncoder(binBuffer)
			gobobj.Encode(m)

			conn.Write(binBuffer.Bytes())
			conn.Close()
		}
	}

	return err
}

// Send Envía a un nodo en concreto
func (n Node) Send(m Message, destination Node) error {
	var conn net.Conn
	var err error

	//log.Println("Normal send to: " + fmt.Sprint(destination.ID))
	conn, err = net.Dial("tcp", destination.Address+destination.Port)
	if err != nil {
		panic("Client connection error: ")
		log.Println("PANIC: ")
	}
	//		Logger.PrepareSend("Sending Message", nil, govec.GetDefaultLogOptions())
	if err != nil {
		panic(err)
	}

	binBuffer := new(bytes.Buffer)
	gobobj := gob.NewEncoder(binBuffer)
	gobobj.Encode(m)

	conn.Write(binBuffer.Bytes())
	conn.Close()

	return err
}

func (n Node) Receive(listener net.Listener, se *SimulationEngine) error {
	for {
		var conn net.Conn
		var err error
		reply := make([]byte, 1024)

		if err != nil {
			panic("Server listen error")
		}
		conn, err = listener.Accept()
		if err != nil {
			log.Println("Connection receiver closed")
			return err
		}
		conn.Read(reply)

		buffer := bytes.NewBuffer(reply)
		decodedMessage := new(Message)
		gobobjdec := gob.NewDecoder(buffer)
		gobobjdec.Decode(decodedMessage)
		//log.Println("------------- LISTA DE EVENTOS RECIBIDOS --------------")
		//log.Println("------------- FROM: " + fmt.Sprint(decodedMessage.Source.ID) + " --------------")
		//decodedMessage.EventList.Imprime()
		//log.Println("----------- FIN LISTA DE EVENTOS RECIBIDOS ------------")

		n.decodeMessage(*decodedMessage, se)
	}
}

func (n Node) decodeMessage(decodedMessage Message, se *SimulationEngine) {
	switch decodedMessage.Type {
	// He recibido eventos de un vecino
	case NEW_EVENTS:
		for _, event := range decodedMessage.EventList {
			if event.IsNull {
				n.insertarEventoRemoto(decodedMessage.EventList[0], se, decodedMessage.Source.ID)
				se.CanalComunicacion <- true
			} else {
				if se.esEventoMio(event) { // Si no es un evento null (Look Ahead), solo me interesa si lo tengo que procesar en esta subred
					se.IlEventos.inserta(se.eventoDeRemotoAlocal(event))
				}
			}
		}

	// Alguien está solicitando un LOOK AHEAD
	case REQUEST_TIME:
		lookAhead := se.calcularLookAhead()
		e := Event{IiTiempo: lookAhead,
			IiTransicion: se.ilMislefs.IaRed[0].IiIndLocal, IsNull: true}
		n.Send(Message{ID: 0, Source: n, Type: NEW_EVENTS, EventList: EventList{e}}, decodedMessage.Source)
		//log.Println("Respondiendo ante solicitud de tiempos desde " + fmt.Sprint(n.ID) +
		//" hacia " + fmt.Sprint(decodedMessage.Source) + ": " + fmt.Sprint(e))
		// El receptor debe saber quién le está comunicando el tiempo (vecino 1 o vecino 2). Le indicamos una de las transiciones (ID)
		// que tiene para que pueda identificarlo.

	}
}

func (se *SimulationEngine) calcularLookAhead() TypeClock {
	return se.iiRelojlocal + TIEMPO_SEGURIDAD
}

func (n Node) insertarEventoRemoto(event Event, se *SimulationEngine, source int) {
	se.IlEventosRemotos[source].inserta(event)
}

func (n Node) LaunchReceiver(se *SimulationEngine) {
	listener, err := net.Listen("tcp", n.Port)
	if err != nil {
		panic(err)
	}
	log.Println("Listening at port", n.Port)
	n.Receive(listener, se)
}
