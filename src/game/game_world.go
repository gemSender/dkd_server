package game

import (
	"../messages/proto_files"
	"github.com/golang/protobuf/proto"
	"log"
	"math/rand"
	"time"
)



type GameWorld struct{
	rand *rand.Rand
	playerDict map[string]*PlayerState
	actionDict map[string]func (string, []byte)
	lastUpdateTime int64
}

func GetTimeStampMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func (world *GameWorld) NeedUpdate(now int64, framePerSec int32)  bool {
	timeSinceLastUpdate := now - world.lastUpdateTime
	if int32(timeSinceLastUpdate) * framePerSec > 1000{
		return true
	}
	return  false
}

func CreateWorld() *GameWorld{
	world := &GameWorld{playerDict:make(map[string]*PlayerState), actionDict:make(map[string]func(string, []byte))}
	world.RegisterCallback("MoveTo", world.OnPlayerMoveTo)
	world.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	world.lastUpdateTime = GetTimeStampMs()
	return world
}

func (world *GameWorld) OnPlayerMoveTo(id string, binData[] byte){
	msg := messages.MoveTo{}
	err := proto.Unmarshal(binData, &msg)
	if(err != nil){
		log.Println(err)
	}
	player := world.playerDict[id]
	player.MoveTo(*msg.X, *msg.Y)
	timeStamp := GetTimeStampMs()
	pushMsg := messages.PlayerMoveTo{Id:&id, X:msg.X, Y:msg.Y, DirX:msg.DirX, DirY:msg.DirY, Timestamp:&timeStamp}
	pushMsgBytes, err := proto.Marshal(&pushMsg)
	if err != nil{
		log.Panic(err)
	}
	isReply := false
	msgType := "PlayerMoveTo"
	packedMsg := messages.GenReplyMsg{Type:&msgType, Data:pushMsgBytes, IsReply:&isReply}
	for _, p := range world.playerDict{
		p.MsgChan <- packedMsg
	}
}

func (world *GameWorld) RegisterCallback(msgType string, callBack func (string, []byte)){
	world.actionDict[msgType] = callBack
}

func (world *GameWorld) Update(now int64)  {
	world.lastUpdateTime = now;
}

func (world *GameWorld) OnPlayerQuit(id string)  {
	if id != ""{
		delete(world.playerDict, id)
		pushMsg := &messages.PlayerQuit{Id:&id}
		pushMsgBytes, err := proto.Marshal(pushMsg)
		if err != nil{
			log.Panic(err)
		}
		repType := "PlayerQuit"
		msg := messages.GenReplyMsg{Type:&repType, Data:pushMsgBytes}
		for _,p  := range world.playerDict{
			p.MsgChan <- msg
		}
		log.Println("player ", id, " exit")
	}
}

func (world *GameWorld) OnLogin(binData []byte, msgChannel chan messages.GenReplyMsg, idChannel chan string){
	loginMsg := messages.Login{}
	proto.Unmarshal(binData, &loginMsg)
	log.Println("player ", loginMsg.GetId(), " login");
	idChannel <- loginMsg.GetId()
	x := float32(0)
	y := float32(0)
	colorIndex := world.rand.Int31n(8)
	timestamp := time.Now().UnixNano() / 1e6
	replyMsg := messages.LoginReply{X:&x, Y:&y, ColorIndex:&colorIndex, Timestamp:&timestamp, Players:make([]*messages.PlayerState, 0, len(world.playerDict))}
	for _, val := range world.playerDict {
		replyMsg.Players = append(replyMsg.Players, &messages.PlayerState{Id:&val.Id, X:&val.X, Y:&val.Y, ColorIndex:&val.ColorIndex})
	}
	replyMsgBytes, err0 := proto.Marshal(&replyMsg)
	if err0 != nil{
		log.Panic(err0)
	}
	isReply := true
	replyType := "LoginReply"
	msgChannel <- messages.GenReplyMsg{Type:&replyType, Data:replyMsgBytes, IsReply:&isReply}
	if len(world.playerDict) > 0 {
		pushMsg := messages.PlayerLogin{Id:loginMsg.Id, X:&x, Y:&y, ColorIndex:&colorIndex, Timestamp:&timestamp}
		pushMsgBytes, err := proto.Marshal(&pushMsg)
		if err != nil {
			log.Panic(err)
		}
		pushType := "PlayerLogin"
		msg := messages.GenReplyMsg{Type:&pushType, Data:pushMsgBytes}
		for _, v := range world.playerDict {
			v.MsgChan <- msg
		}
	}
	newPlayer := new(PlayerState)
	*newPlayer = PlayerState{Id: *loginMsg.Id, X:x, Y:y, MsgChan:msgChannel, lastPosSyncTime:timestamp, ColorIndex:colorIndex}
	world.playerDict[*loginMsg.Id] = newPlayer
}

func (world *GameWorld)  OnBinaryMessage(msgBody []byte, msgChannel chan messages.GenReplyMsg, idChannel chan string, id string){
	if  msgBody == nil {
		world.OnPlayerQuit(id)
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