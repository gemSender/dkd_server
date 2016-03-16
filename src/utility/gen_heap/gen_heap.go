package gen_heap

import (
	"container/heap"
)

type Heap struct {
	slice []interface{}
	LessFunc func (interface{}, interface{}) bool
}

func (this *Heap) Len() int {
	return len(this.slice)
}

func (this *Heap) Less(i, j int) bool{
	return this.LessFunc(this.slice[i], this.slice[j])
}

func (this *Heap) Swap(i, j int){
	this.slice[i], this.slice[j] = this.slice[j], this.slice[i]
}

func (this *Heap) Push(elem interface{}){
	this.slice = append(this.slice, elem)
}

func (this *Heap) Pop() interface{}{
	len := len(this.slice)
	ret := this.slice[len - 1]
	this.slice = this.slice[:len - 1]
	return ret
}

func (this *Heap) Clear() {
	this.slice = this.slice[:0]
}

func (this *Heap) Find(match func (interface{}) bool) (int, interface{}){
	for idx, item := range this.slice{
		if(match(item)){
			return idx, item
		}
	}
	return -1, nil
}

func (this *Heap) SetByIndex(idx int, value interface{})  {
	this.slice[idx] = value;
	heap.Fix(this, idx)
}

func Create(lessFunc func (interface{}, interface{}) bool) *Heap{
	ret := &Heap{slice:make([]interface{}, 0, 32)}
	ret.LessFunc = lessFunc
	return  ret
}

