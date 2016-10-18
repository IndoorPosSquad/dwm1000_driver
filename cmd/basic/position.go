package main

import (
	"fmt"
	"math"
)

var (
	ErrNoSolution = fmt.Errorf("no solution found")
)

var (
	A      = &Node{0.0, 0.0, 0.0}
	B      = &Node{1.2, 0.0, 0.0}
	C      = &Node{0.0, 1.8, 0.0}
	D      = &Node{0.0, 1.8, 0.0}
	OffetA = 0.0
	OffetB = 0.0
	OffetC = 0.0
	OffetD = 0.0
)

type Node struct {
	X float64
	Y float64
	Z float64
}

func square(v float64) float64 {
	return v * v
}

func normalize(p *Node) float64 {
	return math.Sqrt(square(p.X) + square(p.Y) + square(p.Z))
}

func dot(p1, p2 *Node) float64 {
	return p1.X*p2.X + p1.Y*p2.Y + p1.Z*p2.Z
}

func subtract(p1, p2 *Node) *Node {
	return &Node{
		X: p1.X - p2.X,
		Y: p1.Y - p2.Y,
		Z: p1.Z - p2.Z,
	}
}

func add(p1, p2 *Node) *Node {
	return &Node{
		X: p1.X + p2.X,
		Y: p1.Y + p2.Y,
		Z: p1.Z + p2.Z,
	}
}

func divide(p *Node, v float64) *Node {
	return &Node{
		X: p.X / v,
		Y: p.Y / v,
		Z: p.Z / v,
	}
}

func multiply(p *Node, v float64) *Node {
	return &Node{
		X: p.X * v,
		Y: p.Y * v,
		Z: p.Z * v,
	}
}

func cross(p1, p2 *Node) *Node {
	return &Node{
		X: p1.Y*p2.Z - p1.Z*p2.Y,
		Y: p1.Z*p2.X - p1.X*p2.Z,
		Z: p1.X*p2.Y - p1.Y*p2.X,
	}
}

func Solve(p1, p2, p3 float64) *Node {
	n, err := solve(A, B, C, p1+OffetA, p2+OffetB, p3+OffetC)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return n
}

func solve(p1, p2, p3 *Node, d1, d2, d3 float64) (*Node, error) {
	ex := divide(subtract(p2, p1), normalize(subtract(p2, p1)))

	i := dot(ex, subtract(p3, p1))
	a := subtract(subtract(p3, p1), multiply(ex, i))
	ey := divide(a, normalize(a))
	ez := cross(ex, ey)
	d := normalize(subtract(p2, p1))
	j := dot(ey, subtract(p3, p1))

	x := (square(d1) - square(d2) + square(d)) / (2 * d)
	y := (square(d1)-square(d3)+square(i)+square(j))/(2*j) - (i/j)*x

	b := square(d1) - square(x) - square(y)
	if math.Abs(b) < 0.0000000001 {
		b = 0
	}
	z := math.Sqrt(b)

	if math.IsNaN(z) {
		return nil, ErrNoSolution
	}

	a = add(p1, add(multiply(ex, x), multiply(ey, y)))
	p4a := add(a, multiply(ez, z))
	// p4b := subtract(a, multiply(ez, z))
	if z == 0 {
		return a, nil
	}
	return p4a, nil
}

func distance(p1, p2 *Node) float64 {
	return normalize(subtract(p1, p2))
}

func test() {

	p1 := &Node{X: 0.0, Y: 0.0, Z: -10.0}
	p2 := &Node{X: 40.0, Y: 0.0, Z: 10.0}
	p3 := &Node{X: 0.0, Y: 30.0, Z: 0.0}
	p4 := &Node{X: 10.0, Y: 10.0, Z: 100.0}
	d1 := distance(p1, p4)
	d2 := distance(p2, p4)
	d3 := distance(p3, p4)
	fmt.Println(p1)
	fmt.Println(p2)
	fmt.Println(p3)
	solution, err := solve(p1, p2, p3, d1, d2, d3)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(solution)
}
