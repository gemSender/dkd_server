package data_access

import (
	"container/list"
	"log"
)
type DBWaitMsg struct{
	CmdId int
	Callback func(interface{}, error)
}

type DBOperatorObj struct {
	CmdChan   chan DBCommand
	waitList  *list.List
	nextCmdId int
}

func CreateOperatorObj(cmdChan chan DBCommand) *DBOperatorObj {
	return &DBOperatorObj{CmdChan:cmdChan, waitList:list.New(), nextCmdId:1}
}

func (this *DBOperatorObj) DBOp(command DBCommand, callback func(interface{}, error))  *DBOperatorObj{
	command.CmdId = this.nextCmdId
	this.waitList.PushBack(DBWaitMsg{CmdId:command.CmdId, Callback:callback})
	this.CmdChan <- command
	this.nextCmdId ++
	return this
}

func (this *DBOperatorObj) DealReply(reply DBOperationReply){
	headElem := this.waitList.Front()
	waitElem := headElem.Value.(DBWaitMsg)
	if waitElem.CmdId == reply.CmdId {
		if waitElem.Callback != nil{
			waitElem.Callback(reply.Result, reply.err)
		}
		this.waitList.Remove(headElem)
	}else {
		log.Panic("db operation msgid unordered")
	}
}

func (this *DBOperatorObj) FindOne(collection string, query interface{}, callback func(interface{}, error)){
	this.DBOp(DBCommand{CmdType:FindOne, Collection:collection, Arguments:[]interface{}{query}}, callback)
}

func (this *DBOperatorObj) Insert(collection string, doc interface{}, callback func(interface{}, error)){
	this.DBOp(DBCommand{CmdType:Insert, Collection:collection, Arguments:[]interface{}{doc}}, callback)
}

func (this *DBOperatorObj) Update(collection string, query interface{}, modify interface{}, callback func(interface{}, error)){
	this.DBOp(DBCommand{CmdType:Update, Collection:collection, Arguments:[]interface{}{query, modify}}, callback)
}