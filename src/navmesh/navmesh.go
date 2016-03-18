package navmesh

import (
	math_utility "../utility/math"
	"os"
	"bufio"
	"encoding/binary"
	"math"
	"fmt"
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
	Center math_utility.Vec2
}

func (this *NavMeshTriangle) GetCenter(vertices []math_utility.Vec2) math_utility.Vec2{
	return math_utility.VecDivide(math_utility.VecAdd(math_utility.VecAdd(vertices[this.Indices[0]], vertices[this.Indices[1]]), vertices[this.Indices[2]]), 3)
}

func dist(t1 *NavMeshTriangle, t2 *NavMeshTriangle, vertices []math_utility.Vec2) float32{
	p1, p2, p3 := vertices[t1.Indices[0]], vertices[t1.Indices[1]], vertices[t1.Indices[2]]
	q1, q2, q3 := vertices[t2.Indices[0]], vertices[t2.Indices[1]], vertices[t2.Indices[2]]
	c1 := math_utility.VecDivide(math_utility.VecAdd(math_utility.VecAdd(p1, p2), p3), 3)
	c2 := math_utility.VecDivide(math_utility.VecAdd(math_utility.VecAdd(q1, q2), q3), 3)
	return math_utility.VecDist(c1, c2)
}

type NavMeshEdge struct{
	Cost float32
	Next *NavMeshTriangle
	vertices [2]int
}


type NavMesh struct {
	Triangles []*NavMeshTriangle
	Vertices []math_utility.Vec2
}

func get_hash_key(a int, b int) int64 {
	if a > b{
		return int64(b) << 32 | int64(a)
	}
	return int64(a) << 32 | int64(b)
}

func CreateNavMesh(vertices []math_utility.Vec2, indices []int, areas []int)  (*NavMesh, error){
	lenIndices := len(indices)
	if lenIndices % 3 != 0{
		return nil, NavMeshError("indeces count error")
	}
	triCount := lenIndices / 3
	triangles := make([]*NavMeshTriangle, triCount)
	tempMap := make(map[int64][]int)
	for i:=0; i < len(vertices); i++{
		for j:= i+1; j < len(vertices); j++{
			if(vertices[i] == vertices[j]){
				fmt.Printf("same vertex, %v: %v, %v: %v\n", i, vertices[i], j, vertices[j])
			}
		}
	}
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
		triIndex := i/3
		triItem := NavMeshTriangle{}
		i1, i2, i3 := indices[i], indices[i + 1], indices[i + 2]
		triItem.Indices[0], triItem.Indices[1], triItem.Indices[2] = i1, i2, i3
		triItem.Area = areas[triIndex]
		triItem.ArrIndex = triIndex
		addIndex(i1, i2, triIndex)
		addIndex(i1, i3, triIndex)
		addIndex(i2, i3, triIndex)
		triangles[triIndex] = &triItem
	}
	for idx, tri := range triangles{
		edges := make([]*NavMeshEdge, 0, 3)
		i1, i2, i3 := tri.Indices[0], tri.Indices[1], tri.Indices[2]
		fmt.Printf("%v ->", idx)
		addEdge := func(idx1 int, idx2 int) {
			key := get_hash_key(idx1, idx2)
			for _, otherIdx := range tempMap[key]{
				if otherIdx != idx{
					adjTri := triangles[otherIdx]
					cost := dist(tri, adjTri, vertices)
					edge := &NavMeshEdge{Cost:cost, Next:adjTri}
					edge.vertices[0], edge.vertices[1] = idx1, idx2
					edges = append(edges, edge)
					fmt.Printf(" %v", otherIdx)
				}
			}
		}
		addEdge(i1, i2)
		addEdge(i2, i3)
		addEdge(i3, i1)
		fmt.Println()
		tri.Adjs = edges
	}
	return &NavMesh{Triangles:triangles, Vertices:vertices}, nil
}

func (this *NavMesh) GetTriangleByPoint(point math_utility.Vec2) *NavMeshTriangle {
	for _, t := range this.Triangles{
		A, B, C := this.Vertices[t.Indices[0]], this.Vertices[t.Indices[1]], this.Vertices[t.Indices[2]]
		if math_utility.PointInTriangle(A, B, C, point){
			return t
		}
	}
	return nil
}

func GetNavMeshFromFile(path string) (*NavMesh, error){
	fp, err := os.OpenFile(path, os.O_RDONLY, 0660)
	if err != nil{
		return nil, err
	}
	defer fp.Close()
	lenBuf := make([]byte, 4)
	reader := bufio.NewReader(fp)
	readNbytes := func(n int, buf []byte) error{
		sum := 0
		for sum < n{
			addLen, readErr:= reader.Read(buf[sum:n])
			if readErr != nil{
				return readErr
			}
			sum += addLen
		}
		return nil
	}
	err1 := readNbytes(4, lenBuf)
	if err1 != nil{
		return nil, err1
	}
	intBuf := make([]byte, 4)
	floatBuf := make([]byte, 8)
	vertLen := GetIntFromBytes(lenBuf)
	vertices := make([]math_utility.Vec2, vertLen)
	for i := 0; i < vertLen; i++{
		if err2 := readNbytes(8, floatBuf); err2 != nil{
			return nil, err2
		}
		point := math_utility.Vec2{
			X:GetFloat32FromBytes(floatBuf[:4]),
			Y:GetFloat32FromBytes(floatBuf[4:]),
		}
		vertices[i] = point
	}
	err3 := readNbytes(4, lenBuf)
	if err3 != nil{
		return nil, err3
	}
	triLen := GetIntFromBytes(lenBuf)
	triIndices := make([]int, triLen)
	for i := 0; i < triLen; i ++{
		err4 := readNbytes(4, intBuf)
		if err4 != nil{
			return nil, err4
		}
		triIndices[i] = GetIntFromBytes(intBuf)
	}
	err5 := readNbytes(4, lenBuf)
	if err5 != nil{
		return nil, err5
	}
	areaLen := GetIntFromBytes(lenBuf)
	areas := make([]int, triLen)
	for i := 0; i < areaLen; i ++{
		err6 := readNbytes(4, intBuf)
		if err6 != nil{
			return nil, err6
		}
		areas[i] = GetIntFromBytes(intBuf)
	}
	return CreateNavMesh(vertices, triIndices, areas)
}

func GetFloat32FromBytes(buf []byte)  float32{
	uints := binary.LittleEndian.Uint32(buf)
	return math.Float32frombits(uints)
}

func GetIntFromBytes(bytes []byte)  int{
	return (int(bytes[3]) << 24) | (int(bytes[2]) << 16) | (int(bytes[1]) << 8) | int(bytes[0])
}