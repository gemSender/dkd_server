package navmesh

import (
	"../utility/math"
)

const (
	Area_Walkable = 0
	Area_Unwalkable = 1
)
type NavMeshError string

func (err NavMeshError) Error() string{
	return  string(err)
}

type NavMeshTriangle struct {
	Adjs []*NavMeshEdge
	Indices [3]int
	Area int
	ArrIndex int
	Center math.Vec2
}

func (this *NavMeshTriangle) GetCenter(vertices []math.Vec2) math.Vec2{
	return math.VecDivide(math.VecAdd(math.VecAdd(vertices[this.Indices[0]], vertices[this.Indices[1]]), vertices[this.Indices[2]]), 3)
}

func dist(t1 *NavMeshTriangle, t2 *NavMeshTriangle, vertices []math.Vec2) float32{
	p1, p2, p3 := vertices[t1.Indices[0]], vertices[t1.Indices[1]], vertices[t1.Indices[2]]
	q1, q2, q3 := vertices[t2.Indices[0]], vertices[t2.Indices[1]], vertices[t2.Indices[2]]
	c1 := math.VecDivide(math.VecAdd(math.VecAdd(p1, p2), p3), 3)
	c2 := math.VecDivide(math.VecAdd(math.VecAdd(q1, q2), q3), 3)
	return math.VecDist(c1, c2)
}

type NavMeshEdge struct{
	Cost float32
	Next *NavMeshTriangle
}


type NavMesh struct {
	Triangles []*NavMeshTriangle
	Vertices []math.Vec2
}

func get_hash_key(a int, b int) int64 {
	if a > b{
		return int64(b) << 32 | int64(a)
	}
	return int64(a) << 32 | int64(b)
}

func CreateNavMesh(vertices []math.Vec2, indices []int, areas []int)  (*NavMesh, error){
	lenIndices := len(indices)
	if lenIndices % 3 != 0{
		return nil, NavMeshError("indeces count error")
	}
	triCount := lenIndices / 3
	triangles := make([]*NavMeshTriangle, triCount)
	tempMap := make(map[int64][]int)
	addIndex := func(idx1 int, idx2 int, tIndex int) {
		key := get_hash_key(idx1, idx2)
		switch v := tempMap[key] ; v{
		case nil:
			v = make([]int, 0, 2)
			v = append(v, tIndex)
			tempMap[key] = v
		default:
			v = append(v, tIndex)
			tempMap[key] = v
		}
	}
	for i := 0; i < lenIndices; i += 3 {
		triItem := NavMeshTriangle{}
		i1, i2, i3 := indices[i], indices[i + 1], indices[i + 2]
		triItem.Indices[0], triItem.Indices[1], triItem.Indices[2] = i1, i2, i3
		triItem.Area = areas[i / 3]
		triItem.ArrIndex = i
		addIndex(i1, i2, i)
		addIndex(i1, i3, i)
		addIndex(i2, i3, i)
		triangles[i] = &triItem
	}
	for idx, tri := range triangles{
		edge := make([]*NavMeshEdge, 0, 3)
		i1, i2, i3 := tri.Indices[0], tri.Indices[1], tri.Indices[2]
		addEdge := func(idx1 int, idx2 int) {
			key := get_hash_key(idx1, idx2)
			for _, otherIdx := range tempMap[key]{
				if otherIdx != idx{
					adjTri := triangles[otherIdx]
					cost := dist(tri, adjTri, vertices)
					edge = append(edge, &NavMeshEdge{Cost:cost, Next:adjTri})
				}
			}
		}
		addEdge(i1, i2)
		addEdge(i1, i3)
		addEdge(i2, i3)
		tri.Adjs = edge
	}
	return &NavMesh{Triangles:triangles, Vertices:vertices}, nil
}

func (this *NavMesh) GetTriangleByPoint(point math.Vec2) *NavMeshTriangle {
	return nil
}