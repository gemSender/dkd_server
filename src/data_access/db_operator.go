package data_access

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	FindOne = 1
	Insert = 2
	Update = 3
	Upsert = 4
)

type DBCommand struct{
	CmdType int
	CmdId int
	Collection string
	Arguments []interface{}
}

type DBOperationReply  struct{
	CmdId int
	Result interface{}
	err error
}



func StartService(info *mgo.DialInfo, cmdChan chan DBCommand, resultChan chan DBOperationReply)  {
	session, dialErr:= mgo.DialWithInfo(info)
	if dialErr != nil{
		log.Panic(dialErr)
	}
	defer session.Close()
	opMap := map[int]func (*mgo.Collection, DBCommand) (interface{}, error){
		FindOne : func(collection *mgo.Collection, cmd DBCommand) (interface{}, error){
			result := bson.M{}
			queryErr := collection.Find(cmd.Arguments[0]).One(result)
			return result, queryErr
		},
		Insert : func(collection *mgo.Collection, cmd DBCommand) (interface{}, error){
			insertError := collection.Insert(cmd.Arguments...)
			return (insertError == nil), insertError
		},
		Update : func(collection *mgo.Collection, cmd DBCommand) (interface{}, error){
			updateError := collection.Update(cmd.Arguments[0], cmd.Arguments[1])
			return (updateError == nil), updateError
		},
		Upsert : func(collection *mgo.Collection, cmd DBCommand) (interface{}, error){
			changeInfo, upsertError := collection.Upsert(cmd.Arguments[0], cmd.Arguments[1])
			return changeInfo, upsertError
		},
	}
	db := session.DB(info.Database)
	for {
		command := <- cmdChan
		collection := db.C(command.Collection)
		result, err := opMap[command.CmdType](collection, command)
		resultChan <- DBOperationReply{CmdId:command.CmdId, Result:result, err:err}
	}
}
