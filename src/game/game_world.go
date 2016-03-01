package game

import (
	"../messages/proto_files"
	"github.com/golang/protobuf/proto"
	"log"
)

type PlayerState struct {
	Id string
	X float32
	Y float32
	MsgChan chan []byte
}

type GameWorld struct{
	playerDict map[string]*PlayerState
	actionDict map[string]func (string, []byte)
}

func CreateWorld() *GameWorld{
	world := &GameWorld{playerDict:make(map[string]*PlayerState), actionDict:make(map[string]func(string, []byte))}
	world.RegisterCallback("MoveTo",
		func(id string, binData []byte) {
			msg := messages.MoveTo{}
			err := proto.Unmarshal(binData, &msg)
			if(err != nil){
				log.Println(err)
			}
			player := world.playerDict[id]
			player.X, player.Y = *msg.X, *msg.Y
		})
	return world
}

func (world *GameWorld) RegisterCallback(msgType string, callBack func (string, []byte)){
	world.actionDict[msgType] = callBack
}

func (world *GameWorld) Update()  {

}

func (world *GameWorld) OnPlayerExit(id string)  {
	if id != ""{
		delete(world.playerDict, id)
		log.Println("player ", id, " exit")
	}
}

func (world *GameWorld) OnLogin(binData []byte, msgChannel chan []byte){
	loginMsg := messages.Login{}
	proto.Unmarshal(binData, &loginMsg)
	msgChannel <- []byte(*loginMsg.Id)
	newPlayer := new(PlayerState)
	*newPlayer = PlayerState{Id: *loginMsg.Id, X:0, Y:0, MsgChan:msgChannel}
	world.playerDict[*loginMsg.Id] = newPlayer
}

func (world *GameWorld)  OnBinaryMessage(msgBody []byte, msgChannel chan []byte, id string){
	genMsg := messages.GemMessage{}
	parseErr := proto.Unmarshal(msgBody, &genMsg)
	if parseErr != nil{
		log.Println(parseErr)
	}
	if  msgBody == nil{
		world.OnPlayerExit(id)
	}else if id == ""{
		if *genMsg.Type == "Login"{
			world.OnLogin(msgBody, msgChannel)
		}
	}else{
		world.actionDict[*genMsg.Type](id, genMsg.Data)
	}
}