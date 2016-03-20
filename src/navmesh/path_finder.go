package navmesh

import (
	"../utility/gen_heap"
	"../utility/math"
	"fmt"
	"container/heap"
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
	mesh *NavMesh
	openList *gen_heap.Heap
	target math.Vec2
	flag []byte
	initFlag []byte
}

func CreatePathFinder(mesh *NavMesh) *PathFinder{
	ret := &PathFinder{mesh:mesh}
	ret.initFlag = make([]byte, len(mesh.Triangles))
	ret.flag = make([]byte, len(mesh.Triangles))
	ret.openList = gen_heap.Create(func(a interface{}, b interface{}) bool{
		nodea := a.(*PathNode)
		nodeb := b.(*PathNode)
		return nodea.cost + math.VecDist(nodea.point, ret.target) < nodeb.cost + math.VecDist(nodeb.point, ret.target)
	})
	return  ret
}

func (this *PathFinder) FindPath(start math.Vec2, end math.Vec2) []math.Vec2{
	this.target = end
	copy(this.flag, this.initFlag)
	this.openList.Clear()
	startTriangle := this.mesh.GetTriangleByPoint(start)
	endTriangle := this.mesh.GetTriangleByPoint(end)
	fmt.Printf("start triangle: %v, end triangle %v\n", startTriangle.ArrIndex, endTriangle.ArrIndex)
	if startTriangle == endTriangle{
		return []math.Vec2{start, end}
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
				edgeList := make([]*NavMeshEdge, node.degree)
				for tempNode := node; tempNode.preNode != nil; tempNode = tempNode.preNode{
					edgeList[tempNode.degree - 1] = tempNode.edge
				}
				return this.GetVectices(start, end, edgeList)
			}
			nodeTriangle = node.edge.Next
		}
		this.flag[nodeTriangle.ArrIndex] = InCloseList
		for _, adjEdge := range nodeTriangle.Adjs{
			adjTri := adjEdge.Next
			if adjTri.Area != Area_Walkable{
				continue
			}
			switch this.flag[adjTri.ArrIndex] {
			case Unchecked:
				this.flag[adjTri.ArrIndex] = InOpenList
				newNode := &PathNode{cost: node.cost + adjEdge.Cost, preNode:node, point:adjTri.Center, edge:adjEdge ,degree:node.degree + 1}
				heap.Push(this.openList, newNode)
			case InOpenList:
				index, oldNodeI := this.openList.Find(func(item interface{}) bool{
					return item.(*PathNode).edge.Next == adjTri
				})
				oldNode := oldNodeI.(*PathNode)
				compareCost := node.cost + adjEdge.Cost
				if compareCost < oldNode.cost {
					oldNode.preNode = node
					oldNode.cost = compareCost
					oldNode.edge = adjEdge
					oldNode.degree = node.degree + 1
				}
				this.openList.SetByIndex(index, oldNode)
			default:
			}
		}
	}
	return nil
}

func (this *PathFinder) GetVectices(start math.Vec2, end math.Vec2, edges []*NavMeshEdge) []math.Vec2{
	vertTable := this.mesh.Vertices
	vertices := make([]math.Vec2, 0, 32)
	startPoint := start
	vertices = append(vertices, startPoint)
	leftLineEndPoint, rightLineEndPoint := startPoint, startPoint
	nextEdgeIndex := 0
	for {
		if nextEdgeIndex >= len(edges) {
			break
		}
		firstEdge := edges[nextEdgeIndex].vertices
		leftLineEndPoint, rightLineEndPoint = vertTable[firstEdge[0]], vertTable[firstEdge[1]]
		if startPoint == leftLineEndPoint {
			nextEdgeIndex++
			continue
		} else if startPoint == rightLineEndPoint{
			nextEdgeIndex++
			continue
		}
		for {
			nextEdgeIndex++
			if nextEdgeIndex >= len(edges){
				break
			}
			nexeEdge := edges[nextEdgeIndex].vertices
			idx1 := nexeEdge[0]
			idx2 := nexeEdge[1]
			pl := vertTable[idx1]
			pr := vertTable[idx2]
			v1 := math.Vec2Minus(pl, startPoint) // startPoint -> pl
			v2 := math.Vec2Minus(pr, startPoint) // startPoint -> pr
			v3 := math.Vec2Minus(leftLineEndPoint, pl) // pl -> leftLineEndPoint
			v4 := math.Vec2Minus(leftLineEndPoint, pr) // pr -> leftLineEndPoint
			v5 := math.Vec2Minus(rightLineEndPoint, pl) // pl -> rightLineEndPoint
			v6 := math.Vec2Minus(rightLineEndPoint, pr) // pr -> rightLineEndPoint
			if math.Vec2CrossZ(v2, v4) < 0{
				startPoint = leftLineEndPoint
				vertices = append(vertices, startPoint)
				break
			}else if math.Vec2CrossZ(v1, v5) > 0{
				startPoint = rightLineEndPoint
				vertices = append(vertices, startPoint)
				break
			}else {
				if math.Vec2CrossZ(v1, v3) > 0{
					leftLineEndPoint = pl
				}
				if math.Vec2CrossZ(v2, v6) < 0{
					rightLineEndPoint = pr
				}
			}
		}
	}
	v1 := math.Vec2Minus(end, startPoint)
	v2 := math.Vec2Minus(leftLineEndPoint, end)
	v3 := math.Vec2Minus(rightLineEndPoint, end)
	if math.Vec2CrossZ(v1, v2) < 0 {
		vertices = append(vertices, leftLineEndPoint)
	}else if math.Vec2CrossZ(v1, v3) > 0 {
		vertices = append(vertices, rightLineEndPoint)
	}
	vertices = append(vertices, end)
	return  vertices
}
