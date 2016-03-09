package data_access

import (
	"container/list"
	"log"
)
type DBWaitMsg struct{
	CmdId    int
	callback func(interface{}, error)
}

func (this *DBWaitMsg) AddCallback(callback func(interface{}, error)){
	this.callback = callback
}

type DBOperatorObj struct {
	CmdChan   chan DBCommand
	waitList  *list.List
	nextCmdId int
}

func CreateOperatorObj(cmdChan chan DBCommand) *DBOperatorObj {
	return &DBOperatorObj{CmdChan:cmdChan, waitList:list.New(), nextCmdId:1}
}

func (this *DBOperatorObj) DBOp(command DBCommand)  *DBWaitMsg{
	command.CmdId = this.nextCmdId
	ret := &DBWaitMsg{CmdId:command.CmdId}
	this.waitList.PushBack(ret)
	this.CmdChan <- command
	this.nextCmdId ++
	return ret
}

func (this *DBOperatorObj) DealReply(reply DBOperationReply){
	headElem := this.waitList.Front()
	waitElem := headElem.Value.(*DBWaitMsg)
	if waitElem.CmdId == reply.CmdId {
		if waitElem.callback != nil{
			waitElem.callback(reply.Result, reply.err)
		}
		this.waitList.Remove(headElem)
	}else {
		log.Panic("db operation msgid unordered")
	}
}

func (this *DBOperatorObj) FindOne(collection string, query_selector ...interface{}) *DBWaitMsg{
	return this.DBOp(DBCommand{CmdType:FindOne, Collection:collection, Arguments:query_selector})
}

func (this *DBOperatorObj) Insert(collection string, docs ...interface{}) *DBWaitMsg{
	return this.DBOp(DBCommand{CmdType:Insert, Collection:collection, Arguments:docs})
}

func (this *DBOperatorObj) Update(collection string, query_modify ...interface{}) *DBWaitMsg{
	return this.DBOp(DBCommand{CmdType:Update, Collection:collection, Arguments:query_modify})
}