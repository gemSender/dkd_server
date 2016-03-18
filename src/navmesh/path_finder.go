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

func (this *PathFinder) GetVectices(start math.Vec2, end math.Vec2, edges []*NavMeshEdge) []math.Vec2{
	vertTable := this.mesh.Vertices
	vertices := make([]math.Vec2, 0, 16)
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
