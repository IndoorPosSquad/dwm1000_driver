package main

import (
	"fmt"
	"math"

	"github.com/skelterjohn/go.matrix"
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
	X, Y, Z float64
}

func distance(A, B *Node) float64 {
	return math.Sqrt((B.X-A.X)*(B.X-A.X) + (B.Y-A.Y)*(B.Y-A.Y) + (B.Z-A.Z)*(B.Z-A.Z))
}

func projdistance(A, B *Node) float64 {
	return math.Sqrt((B.X-A.X)*(B.X-A.X) + (B.Y-A.Y)*(B.Y-A.Y))
}

func test(A, B, C, D *Node) {
	result := solve(A, B, C, distance(A, D), distance(B, D), distance(C, D))
	fmt.Printf("正确结果应该是(%3.2f, %3.2f, %3.2f), 结算结果是(%3.2f, %3.2f, %3.2f)\n", D.X, D.Y, D.Z, result.X, result.Y, result.Z)
}

func Solve(p1, p2, p3 float64) *Node {
	return solve(A, B, C, p1, p2, p3)
}

// func main() {
// 	A := &Node{0.0, 0.0, 0.0}
// 	B := &Node{1.2, 0.0, 0.0}
// 	C := &Node{0.0, 1.8, 0.0}
// 	fmt.Println(solve(A, B, C, 2.0, 1.0, 1.0))
// }

func solve(AnchorA, AnchorB, AnchorC *Node, p1, p2, p3 float64) (result *Node) {
	// 先计算坐标变换矩阵
	p1 += OffetA
	p2 += OffetB
	p3 += OffetC
	CordTrans := matrix.Zeros(3, 3)
	_AB := projdistance(AnchorA, AnchorB)
	_AC := projdistance(AnchorA, AnchorC)
	_BC := projdistance(AnchorB, AnchorC)
	HB := AnchorB.Z
	HC := AnchorC.Z
	AB := distance(AnchorA, AnchorB)
	AC := distance(AnchorA, AnchorC)
	BC := distance(AnchorB, AnchorC)
	_CosA := (_AC*_AC + _AB*_AB - _BC*_BC) / (2 * _AC * _AB)
	CosA := (AC*AC + AB*AB - BC*BC) / (2 * AC * AB)
	_dAB := _AC * _CosA
	dAB := AC * CosA
	_H := math.Sqrt(_AC*_AC - _dAB*_dAB)
	H := math.Sqrt(AC*AC - dAB*dAB)
	// X 轴单位向量
	CordTrans.Set(0, 0, _AB/AB)
	CordTrans.Set(0, 2, HB/AB)
	// Y 轴单位向量
	CordTrans.Set(1, 0, (_dAB-(_AB*dAB/AB))/H)
	CordTrans.Set(1, 1, _H/H)
	CordTrans.Set(1, 2, (HC-(HB*dAB/AB))/H)
	// Z 轴单位向量
	CordTrans.Set(2, 0, CordTrans.Get(0, 1)*CordTrans.Get(1, 2)-CordTrans.Get(1, 1)*CordTrans.Get(0, 2))
	CordTrans.Set(2, 1, CordTrans.Get(1, 0)*CordTrans.Get(0, 2)-CordTrans.Get(0, 0)*CordTrans.Get(1, 2))
	CordTrans.Set(2, 2, CordTrans.Get(0, 0)*CordTrans.Get(1, 1)-CordTrans.Get(1, 0)*CordTrans.Get(0, 1))
	temp := math.Sqrt(math.Pow(CordTrans.Get(2, 0), 2) + math.Pow(CordTrans.Get(2, 1), 2) + math.Pow(CordTrans.Get(2, 2), 2))
	CordTrans.Set(2, 0, CordTrans.Get(2, 0)/temp)
	CordTrans.Set(2, 1, CordTrans.Get(2, 1)/temp)
	CordTrans.Set(2, 2, CordTrans.Get(2, 2)/temp)

	// 开始计算坐标
	p := (AB + AC + BC) / 2
	S := math.Sqrt(p * (p - AB) * (p - AC) * (p - BC))
	alpha := p1
	bravo := p2
	charlie := p3
	delta := BC
	echo := AC
	fox := AB
	Delta := bravo*bravo + charlie*charlie - delta*delta
	Echo := alpha*alpha + charlie*charlie - echo*echo
	Fox := alpha*alpha + bravo*bravo - fox*fox
	V := math.Sqrt(4*alpha*alpha*bravo*bravo*charlie*charlie-alpha*alpha*Delta*Delta-bravo*bravo*Echo*Echo-charlie*charlie*Fox*Fox+Delta*Echo*Fox) / 12
	Height := V * 3 / S
	d1 := math.Sqrt(p1*p1 - Height*Height)
	d2 := math.Sqrt(p2*p2 - Height*Height)
	d3 := math.Sqrt(p3*p3 - Height*Height)
	CosAlpha := (d1*d1 + AB*AB - d2*d2) / (2 * d1 * AB)
	Y := 0.0
	X := d1 * CosAlpha
	if (1 - CosAlpha) < 0.000001 {
		Y = 0.0
	} else {
		Y = math.Sqrt(d1*d1 - X*X)
	}

	d4 := math.Pow((dAB-X), 2) + math.Pow((H-Y), 2)
	d5 := math.Pow((dAB-X), 2) + math.Pow((H+Y), 2)
	e1 := math.Abs(math.Sqrt(d4) - d3)
	e2 := math.Abs(math.Sqrt(d5) - d3)
	if e2 < e1 {
		Y = -Y
	}

	XYZ := matrix.Zeros(1, 3)
	XYZ.Set(0, 0, X)
	XYZ.Set(0, 1, Y)
	XYZ.Set(0, 2, Height)
	POS := matrix.Product(XYZ, CordTrans)
	result = &Node{X: AnchorA.X + POS.Get(0, 0), Y: AnchorA.Y + POS.Get(0, 1), Z: AnchorA.Z + POS.Get(0, 2)}
	return
}
