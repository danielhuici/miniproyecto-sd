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
var GENERATED_EVENTS = "GENERATED_EVENTS"
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
	Address: "127.0.0.1",
	Port:    ":5000",
}

var P1 = Node{
	ID:      1,
	Address: "127.0.0.1",
	Port:    ":5001",
}

var P2 = Node{
	ID:      2,
	Address: "127.0.0.1",
	Port:    ":5002",
}

var NODES = [3]Node{P0, P1, P2}

func (n Node) SendBroadcast(m Message) error {
	var conn net.Conn
	var err error

	for _, node := range NODES {
		if n.ID != node.ID {
			//fmt.Println("Broadcast send to: " + fmt.Sprint(node.ID))
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

func (n Node) Send(m Message, destination Node) error {
	var conn net.Conn
	var err error

	//fmt.Println("Normal send to: " + fmt.Sprint(destination.ID))
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
		//fmt.Println("Listening...!")
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
		//fmt.Println("------------- LISTA DE EVENTOS RECIBIDOS --------------")
		//fmt.Println("------------- FROM: " + fmt.Sprint(decodedMessage.Source.ID) + " --------------")
		//decodedMessage.EventList.Imprime()
		//fmt.Println("----------- FIN LISTA DE EVENTOS RECIBIDOS ------------")

		n.decodeMessage(*decodedMessage, se)
	}
}

func (n Node) decodeMessage(decodedMessage Message, se *SimulationEngine) {
	switch decodedMessage.Type {
	case GENERATED_EVENTS:
		//fmt.Println("------------- LISTA DE EVENTOS TRAS RECIBIR --------------")
		//se.IlEventos.Imprime()
		//fmt.Println("------------- FIN LISTA DE EVENTOS TRAS RECIBIR --------------")
		for _, event := range decodedMessage.EventList {
			if event.IsNull {
				n.insertarEventoRemoto(decodedMessage.EventList[0], se, decodedMessage.Source.ID)
			} else {
				if se.esEventoMio(event) {
					se.IlEventos.inserta(se.eventoDeRemotoAlocal(event))
					//fmt.Println("El ID transicion del evento ahora es: " + fmt.Sprint(event.IiTransicion))
				}
			}
		}

	case REQUEST_TIME:
		e := Event{IiTiempo: se.iiRelojlocal + SAFE_TIME,
			IiTransicion: se.ilMislefs.IaRed[0].IiIndLocal, IsNull: true}
		n.Send(Message{ID: 0, Source: n, Type: GENERATED_EVENTS, EventList: EventList{e}}, decodedMessage.Source)
		//fmt.Println("Respondiendo ante solicitud de tiempos desde " + fmt.Sprint(n.ID) +
		//" hacia " + fmt.Sprint(decodedMessage.Source) + ": " + fmt.Sprint(e))
		// El receptor debe saber quién le está comunicando el tiempo (vecino 1 o vecino 2). Le indicamos una de las transiciones (ID)
		// que tiene para que pueda identificarlo.

	}
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
