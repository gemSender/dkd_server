package main

import (
	"net"
	"io"
	"log"
	"os"
	"./game"
)

type ClientMsg struct {
	body []byte
	channel chan []byte
	id string
}

func GameMainLoop(msgRecvChan chan ClientMsg){
	world := game.CreateWorld()
	for {
		select {
		case msg := <- msgRecvChan:
			world.OnBinaryMessage(msg.body, msg.channel, msg.id)
			log.Println("recv msg, length is ", len(msg.body))
		default :
			world.Update()
		}
	}
}

func handleConnection(conn net.Conn, gameChan chan ClientMsg){
	defer conn.Close()
	channel := make(chan []byte)
	countBuf := make([]byte, 4)
	id := ""
	onExit := func() {gameChan <- ClientMsg{body:nil, id:id, channel:channel}}
	defer onExit()
	doRecvMsg := func() {
		n := int(countBuf[0]) << 24 + int(countBuf[1]) << 16 + int(countBuf[2]) << 8 + int(countBuf[3])
		data := make([]byte, n)
		sum := 0
		for sum < n{
			addLen, err2 := conn.Read(data[sum:])
			switch  {
			case err2 == io.EOF:
				log.Println(conn.RemoteAddr(), " disconnectd")
				return
			case err2 != nil:
				log.Println(err2)
				return
			default:
				sum += addLen
			}
		}
		gameChan <- ClientMsg{body:data, channel:channel, id:id}
	}
	doRecvMsg()
	id = string(<-channel)
	for {
		select {
		case sendMsg := <-channel:
			byteLen := int32(len(sendMsg))
			lenBytes := [4]byte{byte(byteLen >> 24), byte(byteLen >> 16), byte(byteLen >> 8), byte(byteLen)}
			conn.Write(lenBytes[0:4])
			conn.Write(sendMsg)
		default:
			_, err1 := conn.Read(countBuf)
			switch {
			case err1 == io.EOF:
				log.Println(conn.RemoteAddr(), " disconnectd")
				return
			case err1 != nil:
				log.Println(err1)
				return
			default:
				doRecvMsg()
			}
		}
	}
}

func main(){
	port := "1234"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	gameChan := make(chan ClientMsg)
	go GameMainLoop(gameChan)
	listenSock, listenErr := net.Listen("tcp", ":" + port)
	if(listenErr != nil){
		log.Panic(listenErr)
	}
	for {
		conn, accErr := listenSock.Accept()
		if(accErr != nil){
			log.Panic(accErr)
		}
		log.Println(conn.RemoteAddr(), " connected")
		go handleConnection(conn, gameChan)
	}
}