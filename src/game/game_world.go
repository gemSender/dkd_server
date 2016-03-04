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
	MsgChan chan messages.GenReplyMsg
}

type GameWorld struct{
	playerDict map[string]*PlayerState
	actionDict map[string]func (string, []byte)
}

func CreateWorld() *GameWorld{
	world := &GameWorld{playerDict:make(map[string]*PlayerState), actionDict:make(map[string]func(string, []byte))}
	world.RegisterCallback("MoveTo", world.OnPlayerMoveTo)
	return world
}

func (world *GameWorld) OnPlayerMoveTo(id string, binData[] byte){
	msg := messages.MoveTo{}
	err := proto.Unmarshal(binData, &msg)
	if(err != nil){
		log.Println(err)
	}
	player := world.playerDict[id]
	player.X, player.Y = *msg.X, *msg.Y
	pushMsg := messages.PlayerMoveTo{Id:&id, X:msg.X, Y:msg.Y}
	pushMsgBytes, err := proto.Marshal(&pushMsg)
	if err != nil{
		log.Panic(err)
	}
	isReply := false
	msgType := "PlayerMoveTo"
	for _, p := range world.playerDict{
		p.MsgChan <- messages.GenReplyMsg{Type:&msgType, Data:pushMsgBytes, IsReply:&isReply}
	}
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

func (world *GameWorld) OnLogin(binData []byte, msgChannel chan messages.GenReplyMsg, idChannel chan string){
	loginMsg := messages.Login{}
	proto.Unmarshal(binData, &loginMsg)
	log.Println("player ", loginMsg.GetId(), " login");
	idChannel <- loginMsg.GetId()
	newPlayer := new(PlayerState)
	*newPlayer = PlayerState{Id: *loginMsg.Id, X:0, Y:0, MsgChan:msgChannel}
	world.playerDict[*loginMsg.Id] = newPlayer
	x := float32(0)
	y := float32(0)
	pushMsg := messages.PlayerLogin{Id:loginMsg.Id, X:&x, Y:&y}
	pushMsgBytes, err := proto.Marshal(&pushMsg)
	if err != nil{
		log.Panic(err)
	}
	isReply := false
	replyType := "PlayerLogin"
	for _, v := range world.playerDict{
		v.MsgChan <- messages.GenReplyMsg{Type:&replyType, Data:pushMsgBytes, IsReply:&isReply}
	}
}

func (world *GameWorld)  OnBinaryMessage(msgBody []byte, msgChannel chan messages.GenReplyMsg, idChannel chan string, id string){
	if  msgBody == nil {
		world.OnPlayerExit(id)
		return;
	}
	genMsg := messages.GenMessage{}
	parseErr := proto.Unmarshal(msgBody, &genMsg)
	if parseErr != nil{
		log.Println(parseErr)
		return;
	}
	if id == ""{
		if *genMsg.Type == "Login"{
			world.OnLogin(genMsg.Data, msgChannel, idChannel)
		}
	}else{
		world.actionDict[*genMsg.Type](id, genMsg.Data)
	}
}