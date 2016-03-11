package len_fixed_que

type LenFixedQue struct{
	slice []interface{}
	heapPos int
	count int
}

func New(len int) *LenFixedQue {
	return &LenFixedQue{slice:make([]interface{}, len), count:0, heapPos:0}
}

func (this *LenFixedQue) Count() int{
	return this.count
}

func (this *LenFixedQue) Len() int{
	return len(this.slice)
}

func (this *LenFixedQue) Enqueue(elem interface{}){
	cap := len(this.slice)
	if this.count < cap{
		this.slice[(this.heapPos + this.count) % cap] = elem
		this.count ++
	}else{
		this.slice[this.heapPos] = elem
		this.heapPos = (this.heapPos + 1) % cap
	}
}

func (this *LenFixedQue) Dequeue() interface{} {
	cap := len(this.slice)
	if this.count > 0{
		ret := this.slice[this.heapPos];
		this.slice[this.heapPos] = nil
		this.heapPos = (this.heapPos + 1) % cap
		this.count --
		return ret
	}
	panic("queue is empty")
	return nil
}

func (this *LenFixedQue) GetHeadElem() interface{}{
	if this.count > 0{
		return this.slice[this.heapPos]
	}
	panic("queue is empty")
	return nil
}

func (this *LenFixedQue) GetTailElem() interface{}{
	if this.count > 0{
		return this.slice[(this.heapPos + this.count - 1) % len(this.slice)]
	}
	panic("queue is empty")
	return nil
}

func (this *LenFixedQue) Get(index int) interface{} {
	if this.count > index{
		return this.slice[(this.heapPos + index) % len(this.slice)]
	}
	panic("index out of range")
	return nil
}

func (this *LenFixedQue) Set(index int, value interface{}){
	if this.count > index{
		this.slice[(this.heapPos + index) % len(this.slice)] = value
	}else {
		panic("index out of range")
	}
}