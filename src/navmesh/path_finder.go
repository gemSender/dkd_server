package navmesh

import (
	"../utility/gen_heap"
	"../utility/math"
	"fmt"
	"container/heap"
	//"log"
	"time"
)

const (
	Unchecked = byte(0)
	InCloseList = byte(1)
	InOpenList = byte(2)
)

type PathNode struct{
	edge *NavMeshEdge
	degree int
	preNode *PathNode
	cost float32
	point math.Vec2
}

type PathFinder struct {
	Mesh     *NavMesh
	openList *gen_heap.Heap
	target   math.Vec2
	flag     []byte
	initFlag []byte
}

func CreatePathFinder(mesh *NavMesh) *PathFinder{
	ret := &PathFinder{Mesh:mesh}
	ret.initFlag = make([]byte, len(mesh.Triangles))
	ret.flag = make([]byte, len(mesh.Triangles))
	ret.openList = gen_heap.Create(func(a interface{}, b interface{}) bool{
		nodea := a.(*PathNode)
		nodeb := b.(*PathNode)
		return nodea.cost + math.VecDist(nodea.point, ret.target) < nodeb.cost + math.VecDist(nodeb.point, ret.target)
		//return nodea.cost < nodeb.cost
	})
	return  ret
}

func (this *PathFinder) FindPath(start math.Vec2, end math.Vec2) ([]math.Vec2, []math.GenSegment){
	this.target = end
	copy(this.flag, this.initFlag)
	this.openList.Clear()
	startTriangle := this.Mesh.GetTriangleByPoint(start)
	endTriangle := this.Mesh.GetTriangleByPoint(end)
	fmt.Printf("start triangle: %v, end triangle %v\n", startTriangle.ArrIndex, endTriangle.ArrIndex)
	if startTriangle == endTriangle{
		return []math.Vec2{start, end}, make([]math.GenSegment, 0)
	}
	heap.Push(this.openList, &PathNode{cost:0, point:start, preNode:nil, edge:nil, degree:0})
	this.flag[startTriangle.ArrIndex] = InOpenList
	for this.openList.Len() > 0{
		node := heap.Pop(this.openList).(*PathNode)
		var nodeTriangle *NavMeshTriangle
		if node.preNode == nil {
			nodeTriangle = startTriangle
		}else {
			if node.edge.Next == endTriangle {
				edgeList := make([]math.GenSegment, node.degree + 1)
				for tempNode := node; tempNode.preNode != nil; tempNode = tempNode.preNode{
					edgeList[tempNode.degree - 1] = tempNode.edge
				}
				edgeList[node.degree] = &math.Segment{Start:end, End:end}
				return this.GetVectices(start, end, edgeList), edgeList
			}
			nodeTriangle = node.edge.Next
		}
		this.flag[nodeTriangle.ArrIndex] = InCloseList
		for _, adjEdge := range nodeTriangle.Adjs{
			var fromPos math.Vec2
			if node.preNode == nil{
				fromPos = start
			}else {
				fromPos = node.edge.Center
			}
			adjTri := adjEdge.Next
			if adjTri.Area != Area_Walkable{
				continue
			}
			switch this.flag[adjTri.ArrIndex] {
			case Unchecked:
				this.flag[adjTri.ArrIndex] = InOpenList
				//addCost, nodePoint := this.mesh.DistPosToEdge(fromPos, adjEdge)
				newNode := &PathNode{cost: node.cost + math.VecDist(fromPos, adjEdge.Center), preNode:node, point:adjEdge.Center, edge:adjEdge ,degree:node.degree + 1}
				heap.Push(this.openList, newNode)
			case InOpenList:
				index, oldNodeI := this.openList.Find(func(item interface{}) bool{
					return item.(*PathNode).edge.Next == adjTri
				})
				oldNode := oldNodeI.(*PathNode)
				//addCost, nodePoint := this.mesh.DistPosToEdge(fromPos, adjEdge)
				compareCost := node.cost + math.VecDist(fromPos, adjEdge.Center)
				if compareCost < oldNode.cost {
					oldNode.preNode = node
					oldNode.cost = compareCost
					oldNode.edge = adjEdge
					oldNode.point = adjEdge.Center
					oldNode.degree = node.degree + 1
				}
				this.openList.SetByIndex(index, oldNode)
			default:
			}
		}
	}
	return nil ,nil
}

func (this *PathFinder) GetVectices(start math.Vec2, end math.Vec2, edges []math.GenSegment) []math.Vec2{
	time1 := time.Now().UnixNano()
	defer fmt.Println("path search time: ", time.Now().UnixNano() - time1)
	vertTable := this.Mesh.Vertices
	vertices := make([]math.Vec2, 0, 32)
	startPoint := start
	vertices = append(vertices, startPoint)
	leftLineEndPoint, rightLineEndPoint := startPoint, startPoint
	firstEdgeIndex := 0
	for {
		if firstEdgeIndex >= len(edges) {
			break
		}
		nextEdgeIndex := firstEdgeIndex
		firstEdge := edges[firstEdgeIndex]
		leftLineEndPointEdgeIndex, rightLinePointEdgeIndex := firstEdgeIndex, firstEdgeIndex
		leftLineEndPoint, rightLineEndPoint = firstEdge.GetStartPoint(vertTable), firstEdge.GetEndPoint(vertTable)
		if math.VecDist(startPoint, leftLineEndPoint) < 0.00001{
			//log.Println("skip edge ", leftLineEndPointEdgeIndex)
			firstEdgeIndex++
			continue
		} else if math.VecDist(startPoint, rightLineEndPoint) < 0.00001{
			//log.Println("skip edge ", rightLinePointEdgeIndex)
			firstEdgeIndex++
			continue
		}
		for {
			nextEdgeIndex++
			if nextEdgeIndex >= len(edges){
				firstEdgeIndex = nextEdgeIndex
				break
			}
			leftLine := math.Vec2Minus(leftLineEndPoint, startPoint)
			rightLine := math.Vec2Minus(rightLineEndPoint, startPoint)
			if math.SameDir(leftLine, rightLine){
				if (leftLine.Magnitude() < rightLine.Magnitude()) {
					startPoint = leftLineEndPoint
					vertices = append(vertices, startPoint)
					firstEdgeIndex = leftLineEndPointEdgeIndex + 1
					//log.Println("startPoine set on edge for same dir", leftLineEndPointEdgeIndex)
				}else {
					startPoint = rightLineEndPoint
					vertices = append(vertices, startPoint)
					firstEdgeIndex = rightLinePointEdgeIndex + 1
					//log.Println("startPoine set on edge same dir", rightLinePointEdgeIndex)
				}
				//log.Println("check same dir ", leftLineEndPointEdgeIndex, " and ", rightLinePointEdgeIndex, " failed")
				break
			}
			//log.Println("check same dir ", leftLineEndPointEdgeIndex, " and ", rightLinePointEdgeIndex, " pass(", leftLineEndPoint, ", ", rightLineEndPoint, ")")
			nextEdge := edges[nextEdgeIndex]
			pl := nextEdge.GetStartPoint(vertTable)
			pr := nextEdge.GetEndPoint(vertTable)
			v1 := math.Vec2Minus(pl, startPoint) // startPoint -> pl
			v2 := math.Vec2Minus(pr, startPoint) // startPoint -> pr
			v3 := math.Vec2Minus(leftLineEndPoint, pl) // pl -> leftLineEndPoint
			v4 := math.Vec2Minus(leftLineEndPoint, pr) // pr -> leftLineEndPoint
			v5 := math.Vec2Minus(rightLineEndPoint, pl) // pl -> rightLineEndPoint
			v6 := math.Vec2Minus(rightLineEndPoint, pr) // pr -> rightLineEndPoint
			if math.Vec2CrossZ(v2, v4) * math.Vec2CrossZ(leftLine, math.Vec2Minus(rightLineEndPoint, leftLineEndPoint)) > 0{
				startPoint = leftLineEndPoint
				vertices = append(vertices, startPoint)
				firstEdgeIndex = leftLineEndPointEdgeIndex + 1
				//log.Println("startPoine set on edge ", leftLineEndPointEdgeIndex)
				break
			}else if math.Vec2CrossZ(v1, v5)  * math.Vec2CrossZ(rightLine, math.Vec2Minus(leftLineEndPoint, rightLineEndPoint)) > 0{
				startPoint = rightLineEndPoint
				vertices = append(vertices, startPoint)
				firstEdgeIndex = rightLinePointEdgeIndex + 1
				//log.Println("startPoine set on edge ", rightLinePointEdgeIndex)
				break
			}else {
				templeftLineEndPoint := leftLineEndPoint
				if math.Vec2CrossZ(v1, v3) * math.Vec2CrossZ(rightLine, math.Vec2Minus(leftLineEndPoint, rightLineEndPoint)) > 0{
					//log.Println("new leftline end point set on Edge ", nextEdgeIndex, " point is ", pl, " when left endpoint is ", leftLineEndPoint)
					leftLineEndPoint = pl
					leftLineEndPointEdgeIndex = nextEdgeIndex
				}
				if math.Vec2CrossZ(v2, v6) * math.Vec2CrossZ(leftLine, math.Vec2Minus(rightLineEndPoint, templeftLineEndPoint)) > 0{
					//log.Println("new rightline end point set on Edge ", nextEdgeIndex, " point is ", pr, " when right endpoint is ", rightLineEndPoint)
					rightLineEndPoint = pr
					rightLinePointEdgeIndex = nextEdgeIndex
				}
			}
		}
	}
	/*
	v1 := math.Vec2Minus(end, startPoint)
	v2 := math.Vec2Minus(leftLineEndPoint, end)
	v3 := math.Vec2Minus(rightLineEndPoint, end)
	if math.Vec2CrossZ(v1, v2) < 0 {
		vertices = append(vertices, leftLineEndPoint)
	}else if math.Vec2CrossZ(v1, v3) > 0 {
		vertices = append(vertices, rightLineEndPoint)
	}
	*/
	vertices = append(vertices, end)
	return  vertices
}

func (this *PathFinder) GetPoint(vertIndex int) math.Vec2{
	return this.Mesh.Vertices[vertIndex]
}