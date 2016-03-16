package navmesh

import (
	"../utility/gen_heap"
	"../utility/math"
)

const (
	Unchecked = byte(0)
	InCloseList = byte(1)
	InOpenList = byte(2)
)
type PathNode struct{
	triangle *NavMeshTriangle
	preTriangle *NavMeshTriangle
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
		nodea := a.(PathNode)
		nodeb := b.(PathNode)
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
	this.openList.Push(PathNode{cost:0, point:start, preTriangle:nil, triangle:startTriangle})
	this.flag[startTriangle.ArrIndex] = InOpenList
	for this.openList.Len() > 0{
		node := this.openList.Pop().(PathNode)
		this.flag[node.triangle.ArrIndex] = InCloseList
		if node.triangle == endTriangle{
			break
		}
		for _, adjEdges := range node.triangle.Adjs{
			adjTri := adjEdges.Next
			if adjTri.Area != Area_Walkable{
				continue
			}
			switch this.flag[adjTri.ArrIndex] {
			case Unchecked:
				this.flag[adjTri.ArrIndex] = InOpenList
				newNode := PathNode{triangle:adjTri, cost: node.cost + adjEdges.Cost, preTriangle:node.triangle, point:adjTri.Center}
				this.openList.Push(newNode)
			case InOpenList:
				index, oldNodeI := this.openList.Find(func(item interface{}) bool{
					return item.(PathNode).triangle == adjTri
				})
				oldNode := oldNodeI.(PathNode)
				compareCost := node.cost + adjEdges.Cost
				if compareCost < oldNode.cost {
					oldNode.preTriangle = node.triangle
					oldNode.cost = compareCost
				}
				this.openList.SetByIndex(index, oldNode)
			default:
			}
		}
	}
	return nil
}