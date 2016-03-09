package main

import (
	"net"
	"io"
	"log"
	"os"
	"./game"
	"bufio"
	"./messages/proto_files"
	"github.com/golang/protobuf/proto"
	"gopkg.in/mgo.v2"
	"./data_access"
)

type ClientMsg struct {
	body []byte
	channel chan messages.GenReplyMsg
	id string
	idChannel chan string
}


func GameMainLoop(msgRecvChan chan ClientMsg, dbCmdChan chan data_access.DBCommand, dbReplyChan chan data_access.DBOperationReply){
	world := game.CreateWorld(dbCmdChan)
	for {
		select {
		case msg := <- msgRecvChan:
			world.OnBinaryMessage(msg.body, msg.channel, msg.idChannel, msg.id)
			//log.Println("recv msg, length is ", len(msg.body))
		case dbReply := <- dbReplyChan:
			world.DBO.DealReply(dbReply)
		default :
		}
	}
	now := game.GetTimeStampMs()
	if world.NeedUpdate(now, 30) {
		world.Update(now)
	}
}

func sender(conn net.Conn, channel chan messages.GenReplyMsg, quitChan chan int){
	writer := bufio.NewWriter(conn)
	for{
		select {
		case sendMsg := <-channel:
			body, encodeErr := proto.Marshal(&sendMsg)
			if encodeErr != nil{
				log.Panic(encodeErr)
			}
			byteLen := int32(len(body))
			lenBytes := [4]byte{byte(byteLen & 0xff), byte(byteLen >> 8 & 0xff), byte(byteLen >> 16 & 0xff), byte(byteLen >> 24 & 0xff)}
			writer.Write(lenBytes[0:4])
			writer.Write(body)
			writer.Flush()
		case <-quitChan:
			//log.Println("writer process of ", conn.RemoteAddr(), " quit")
			return
		}
	}
}

func receiver(conn net.Conn, sendChannel chan messages.GenReplyMsg, gameChan chan ClientMsg, quitChannel chan int){
	defer conn.Close()
	idChannel := make(chan string)
	id := ""
	onQuit := func(){
		quitChannel <- 1
		gameChan <- ClientMsg{id:id, body:nil}
	}
	defer onQuit()
	receiveNbytes := func(n int, buf []byte) error{
		sum := 0
		for sum < n {
			readLen, err := conn.Read(buf[sum:n])
			if err != nil{
				return err
			}
			sum += readLen
		}
		return nil
	}
	doRecvMsg := func() error{
		countArr := [4]byte{}
		countBuf := countArr[:4]
		err1 := receiveNbytes(4, countBuf)
		switch err1 {
		case nil:
			n := int(countBuf[0]) + int(countBuf[1]) << 8 + int(countBuf[2]) << 16 + int(countBuf[3]) << 24
			//log.Println("receive ", n, " bytes from client ", conn.RemoteAddr())
			dataBuf := make([]byte, n)
			err2 := receiveNbytes(n, dataBuf)
			switch  err2{
			case nil:
				gameChan <- ClientMsg{body:dataBuf, channel:sendChannel, id:id, idChannel:idChannel}
			default:
				return err2
			}
		default:
			return err1
		}
		return nil
	}
	switch err1 := doRecvMsg(); err1 {
	case nil:
		id = <- idChannel
		close(idChannel)
		for{
			switch err2:= doRecvMsg(); err2{
			case nil:
			case io.EOF:
				log.Println(conn.RemoteAddr(), " disconnectd")
				return
			default:
				log.Println(err2)
				return
			}
		}
	case io.EOF:
		log.Println(conn.RemoteAddr(), " disconnectd")
		return
	default:
		log.Println(err1)
		return
	}
}

func start_database(cmdChan chan data_access.DBCommand, replyChan chan data_access.DBOperationReply)  {
	dialInfo := mgo.DialInfo{Database:"test", Addrs:[]string{"192.168.0.245"}, Username:"test", Password:"test"}
	go data_access.StartService(&dialInfo, cmdChan, replyChan)
}

func start_listener(gameChan chan ClientMsg)  {
	port := "1234"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	listenSock, listenErr := net.Listen("tcp", ":" + port)
	if(listenErr != nil){
		log.Panic(listenErr)
	}
	for {
		conn, accErr := listenSock.Accept()
		if(accErr != nil){
			log.Panic(accErr)
		}
		sendChannel := make(chan messages.GenReplyMsg)
		quitChannel := make(chan int)
		log.Println(conn.RemoteAddr(), " connected")
		go receiver(conn, sendChannel, gameChan, quitChannel)
		go sender(conn, sendChannel, quitChannel)
	}
}

func start_game_looop(gameChan chan ClientMsg, dbCmdChan chan data_access.DBCommand, dbReplyChan chan data_access.DBOperationReply){
	go GameMainLoop(gameChan, dbCmdChan, dbReplyChan)
}

func main(){
	gameChan := make(chan ClientMsg)
	dbCmdChan := make(chan data_access.DBCommand)
	dbReplyChan := make(chan data_access.DBOperationReply)
	start_database(dbCmdChan, dbReplyChan)
	start_game_looop(gameChan, dbCmdChan, dbReplyChan)
	start_listener(gameChan)
}