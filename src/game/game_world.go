package game

import (
	"../messages/proto_files"
	"github.com/golang/protobuf/proto"
	"log"
	"math/rand"
	"time"
	"../data_access"
	"gopkg.in/mgo.v2/bson"
	"../scheduler"
	"fmt"
	"../navmesh"
	math1  "../utility/math"
)



type GameWorld struct{
	DBO             *data_access.DBOperatorObj
	rand            *rand.Rand
	idIndexMap	map[string]int32
	indexPlayerMap  map[int32]*PlayerState
	actionDict      map[string]func (int32, []byte)
	lastUpdateTime  int64
	nextPlayerIndex int32
	scheduler scheduler.Scheduler
	pathFinder 	*navmesh.PathFinder
}


func GetTimeStampMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func CreateWorld(dbCmdChan chan data_access.DBCommand) *GameWorld{
	world := &GameWorld{
		idIndexMap:make(map[string]int32),
		indexPlayerMap:make(map[int32]*PlayerState),
		actionDict:make(map[string]func(int32, []byte)),
	}
	nm, err := navmesh.GetNavMeshFromFile("navmesh/navmesh.bytes")
	if err != nil{
		log.Panic(err)
	}
	world.pathFinder = navmesh.CreatePathFinder(nm)
	world.RegisterCallback("MoveTo", world.OnPlayerMoveTo)
	world.RegisterCallback("StartPath", world.OnPlayerStartPath)
	world.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	world.lastUpdateTime = GetTimeStampMs()
	world.DBO = data_access.CreateOperatorObj(dbCmdChan)
	world.scheduler = NewHeapScheduler(128)
	return world
}

func (world *GameWorld) RemovePlayerByIndex(index int32) *PlayerState{
	var ret = world.indexPlayerMap[index]
	id := ret.Id
	delete(world.indexPlayerMap, index)
	delete(world.idIndexMap, id)
	return ret
}

func (world *GameWorld) AddPlayer(player *PlayerState)  *PlayerState{
	world.indexPlayerMap[player.Index] = player
	world.idIndexMap[player.Id] = player.Index
	return player
}

func (world *GameWorld) NeedUpdate(now int64, framePerSec int32)  bool {
	timeSinceLastUpdate := now - world.lastUpdateTime
	if int32(timeSinceLastUpdate) * framePerSec > 1000{
		return true
	}
	return  false
}

func (world *GameWorld)  OnPlayerStartPath(index int32, binData[] byte){
	msg := messages.StartPath{}
	err := proto.Unmarshal(binData, &msg)
	if err != nil{
		log.Panic(err)
	}
	player := world.indexPlayerMap[index]
	player.StartPath(*msg.Sx, *msg.Sy)
	pushMsg := messages.PlayerStartPath{Sx:msg.Sx, Sy:msg.Sy, Dx:msg.Dx, Dy:msg.Dy, Index:&index, Timestamp:msg.Timestamp}
	pushMsgBytes, err1 := proto.Marshal(&pushMsg)
	if err1 != nil{
		log.Panic(err1)
	}
	msgType := "PlayerStartPath"
	packedMsg := messages.GenReplyMsg{Type:&msgType, Data:pushMsgBytes}
	//{
		start := math1.Vec2{X:*msg.Sx, Y:*msg.Sy}
		end := math1.Vec2{X:*msg.Dx, Y:*msg.Dy}
		path := world.pathFinder.FindPath(start, end)
		if path != nil{
			fmt.Println("find path: ")
			for _, vert := range path{
				fmt.Printf("(%v, %v)", vert.X, vert.Y)
			}
		}else{
			fmt.Println("path not found")
		}
	//}
	for key, val := range world.indexPlayerMap {
		if key != index {
			val.MsgChan <- packedMsg
		}else {
			vertices := make([]*messages.Vec2, len(path))
			for i, v := range path {
				X := v.X
				Y := v.Y
				vertices[i] = &messages.Vec2{X:&X, Y:&Y}
			}
			replyMsg := messages.StartPathReply{Vertices:vertices}
			replyMsgBytes, err := proto.Marshal(&replyMsg)
			if err != nil{
				log.Panic(err)
			}
			repMsgType := "StartPathReply"
			isReply := true
			val.MsgChan <- messages.GenReplyMsg{Type:&repMsgType, IsReply:&isReply, Data:replyMsgBytes}
		}
	}
}

func (world *GameWorld) OnPlayerMoveTo(index int32, binData[] byte){
	msg := messages.MoveTo{}
	err := proto.Unmarshal(binData, &msg)
	if(err != nil){
		log.Println(err)
	}
	player := world.indexPlayerMap[index]
	player.MoveTo(*msg.X, *msg.Y)
	timeStamp := GetTimeStampMs()
	pushMsg := messages.PlayerMoveTo{Index:&index, X:msg.X, Y:msg.Y, DirX:msg.DirX, DirY:msg.DirY, Timestamp:&timeStamp}
	pushMsgBytes, err := proto.Marshal(&pushMsg)
	if err != nil{
		log.Panic(err)
	}
	isReply := false
	msgType := "PlayerMoveTo"
	packedMsg := messages.GenReplyMsg{Type:&msgType, Data:pushMsgBytes, IsReply:&isReply}
	for _, p := range world.indexPlayerMap {
		p.MsgChan <- packedMsg
	}
}

func (world *GameWorld) RegisterCallback(msgType string, callBack func (int32, []byte)){
	world.actionDict[msgType] = callBack
}

func (world *GameWorld) Update(now int64)  {
	world.scheduler.TrySchedule()
	world.lastUpdateTime = now;
}

func (world *GameWorld) OnPlayerQuit(index int32)  {
	if index != -1{
		player := world.RemovePlayerByIndex(index)
		pushMsg := &messages.PlayerQuit{Index:&index}
		pushMsgBytes, err := proto.Marshal(pushMsg)
		if err != nil{
			log.Panic(err)
		}
		repType := "PlayerQuit"
		msg := messages.GenReplyMsg{Type:&repType, Data:pushMsgBytes}
		for _,p  := range world.indexPlayerMap {
			p.MsgChan <- msg
		}
		log.Println("player ", index, " exit")
		world.DBO.Update("player", bson.M{"_id":player.Id}, bson.M{"$set" : bson.M{"x" : player.X, "y" : player.Y}})
	}
}

func (world *GameWorld) AllocIndex() int32{
	ret := world.nextPlayerIndex
	world.nextPlayerIndex++
	return ret
}

func (world *GameWorld) CreatePlayer(playerId string) bson.M{
	x := (world.rand.Float64() - float64(0.5)) * 2 * 100
	y := (world.rand.Float64() - float64(0.5)) * 2 * 100
	colorIndex := world.rand.Intn(8)
	ret := bson.M{"_id" : playerId, "x" : x, "y" : y, "colorIndex" : colorIndex, "level" : 1, "exp" : 0}
	world.DBO.Insert("player", ret).AddCallback(
		func(result interface{}, err error) {
			if err != nil{
				log.Panic(err)
			}else{
				log.Println("CreatePlayer: ", result)
			}
		})
	return ret
}

func (world *GameWorld) CreatePlayerByDoc(doc bson.M) *PlayerState{
	return &PlayerState{
		Level:int32(doc["level"].(int)),
		Exp:int32(doc["level"].(int)),
		Id:doc["_id"].(string),
		Index:world.AllocIndex(),
		X:float32(doc["x"].(float64)),
		Y:float32(doc["y"].(float64)),
		ColorIndex:int32(doc["colorIndex"].(int)),
		lastPosSyncTime:GetTimeStampMs(),
	}
}

func (world *GameWorld) OnLogin(binData []byte, msgChannel chan messages.GenReplyMsg, idChannel chan int32){
	loginMsg := messages.Login{}
	proto.Unmarshal(binData, &loginMsg)
	equipId := loginMsg.GetEquipId()
	timestamp := GetTimeStampMs()
	afterCreatePlayer := func(player *PlayerState){
		player.MsgChan = msgChannel
		idChannel <- player.Index
		errorCode := int32(0)
		replyState := &messages.PlayerState{
			Index:&player.Index,
			X: &player.X,
			Y:&player.Y,
			ColorIndex:&player.ColorIndex,
			Level:&player.Level,
			Exp:&player.Exp,
		}
		replyMsg := messages.LoginReply{
			ErrorCode:&errorCode,
			MyState:replyState,
			Timestamp:&timestamp,
			Players:make([]*messages.PlayerState, 0, len(world.indexPlayerMap))}
		for _, val := range world.indexPlayerMap {
			replyMsg.Players = append(replyMsg.Players, &messages.PlayerState{
				Index:&val.Index,
				X:&val.X,
				Y:&val.Y,
				ColorIndex:&val.ColorIndex,
				Level:&val.Level,
				Exp:&val.Exp,
			})
		}
		replyMsgBytes, err0 := proto.Marshal(&replyMsg)
		if err0 != nil{
			log.Panic(err0)
		}
		isReply := true
		replyType := "LoginReply"
		msgChannel <- messages.GenReplyMsg{Type:&replyType, Data:replyMsgBytes, IsReply:&isReply}
		if len(world.indexPlayerMap) > 0 {
			pushMsg := messages.PlayerLogin{ Timestamp:&timestamp, PlayerData:replyState}
			pushMsgBytes, err := proto.Marshal(&pushMsg)
			if err != nil {
				log.Panic(err)
			}
			pushType := "PlayerLogin"
			msg := messages.GenReplyMsg{Type:&pushType, Data:pushMsgBytes}
			for _, v := range world.indexPlayerMap {
				v.MsgChan <- msg
			}
		}
		world.AddPlayer(player)
	}
	world.DBO.FindOne("mac_user", bson.M{"_id" : equipId}, bson.M{"user_id" : 1}).AddCallback(
		func(r1 interface{}, e1 error){
			if e1 == nil{
				playerId := r1.(bson.M)["user_id"].(string)
				world.DBO.FindOne("player", bson.M{"_id" : playerId}).AddCallback(func(r2 interface{}, e2 error) {
					if e2 == nil {
						log.Println("find user: ", r2)
						playerDoc := r2.(bson.M)
						_, contain := world.idIndexMap[playerDoc["_id"].(string)]
						if !contain {
							player := world.CreatePlayerByDoc(playerDoc)
							afterCreatePlayer(player)
						}else{
							errorCode := int32(1)
							replyMsg := messages.LoginReply{ErrorCode:&errorCode}
							replyBytes, err := proto.Marshal(&replyMsg)
							if err != nil{
								log.Panic(err)
							}else {
								isReply := true
								repType := "LoginReply"
								msgChannel <- messages.GenReplyMsg{Type:&repType, Data:replyBytes, IsReply:&isReply}
							}
						}
					}else{
						playerDoc := world.CreatePlayer(playerId)
						log.Println("create user: ", playerDoc)
						afterCreatePlayer(world.CreatePlayerByDoc(playerDoc))
					}
				})
			}else{
				playerId := string(bson.NewObjectId())
				world.DBO.Insert("mac_user", bson.M{"_id" : equipId, "user_id" : playerId}).AddCallback(func(r2 interface{}, e3 error) {
					if e3 == nil{
						playerDoc := world.CreatePlayer(playerId)
						log.Println("create user: ", playerDoc)
						afterCreatePlayer(world.CreatePlayerByDoc(playerDoc))
					}else{
						log.Panic(e3)
					}
				})
			}
		})
}

func (world *GameWorld)  OnBinaryMessage(msgBody []byte, msgChannel chan messages.GenReplyMsg, idChannel chan int32, index int32){
	if  msgBody == nil {
		world.OnPlayerQuit(index)
		return;
	}
	genMsg := messages.GenMessage{}
	parseErr := proto.Unmarshal(msgBody, &genMsg)
	if parseErr != nil{
		log.Println(parseErr)
		return;
	}
	if index == -1{
		if *genMsg.Type == "Login"{
			world.OnLogin(genMsg.Data, msgChannel, idChannel)
		}
	}else{
		world.actionDict[*genMsg.Type](index, genMsg.Data)
	}
}
