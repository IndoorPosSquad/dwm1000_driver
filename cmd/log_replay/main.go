package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/skelterjohn/go.matrix"
)

type Data struct {
	X int64
	Y int64
}

type Sync struct {
	data  map[string]([]Data)
	delta map[string]int64
	tdoa  map[string]int64
	tmp   [256]int
}

func (s *Sync) Sync(target string, T1, delta int64) {
	if s.data[target] == nil {
		s.data[target] = make([]Data, 0)
	}
	s.data[target] = append(s.data[target], Data{X: T1, Y: delta})
	if len(s.data[target]) > 10 {
		s.data[target] = s.data[target][len(s.data[target])-10:]
	}
}

func (s *Sync) Convert(target string, clk int64) int64 {
	// no use
	var theta int64
	var prev int64
	n := len(s.data[target])
	x := matrix.Zeros(1, n)
	y := matrix.Zeros(1, n)
	for key, d := range s.data[target] {
		if d.X < prev {
			theta = 0xffffffffff
		}
		prev = d.X
		x.Set(0, key, float64(d.X+theta))
		y.Set(0, key, float64(d.Y))
	}
	r := polyfit(x, y, 2)
	// fmt.Println(r)
	a := r.Get(0, 0)
	b := r.Get(0, 1)
	c := r.Get(0, 2)
	fmt.Println("ets")
	xx := float64(int64(24379955293) + 0xffffffffff)
	yy := a*xx*xx + b*xx + c
	fmt.Println(int64(yy))
	result1 := (-(b) + math.Sqrt((b)*(b)-4*a*(c-float64(clk)))) / (2.0 * a)
	result2 := (-(b) - math.Sqrt((b)*(b)-4*a*(c-float64(clk)))) / (2.0 * a)
	// fmt.Println(int64(result1))
	// fmt.Println(int64(result2))
	result := int64(result1)
	if result > 0xffffffffff {
		result -= 0xffffffffff
	} else if result < 0 {
		result += 0xffffffffff
	}
	fmt.Println(result)

	result = int64(result2)
	if result > 0xffffffffff {
		result -= 0xffffffffff
	} else if result < 0 {
		result += 0xffffffffff
	}
	fmt.Println(result)
	return result
}

func (s *Sync) PredictOffset(target string, clk int64) int64 {
	var theta int64
	var prev int64
	n := len(s.data[target])
	x := matrix.Zeros(1, n)
	y := matrix.Zeros(1, n)
	for key, d := range s.data[target] {
		if d.X < prev {
			theta = 0xffffffffff
		}
		prev = d.X
		x.Set(0, key, float64(d.X+theta))
		y.Set(0, key, float64(d.Y))
	}
	r := polyfit(x, y, 2)
	// fmt.Println(r)
	a := r.Get(0, 0)
	b := r.Get(0, 1)
	c := r.Get(0, 2)
	xx := float64(int64(clk) + theta)
	yy := a*xx*xx + b*xx + c
	return int64(yy)
}

var prev = -1

func (s *Sync) Calculate(anchor string, ts int64, id int) {
	s.tmp[id]++
	if id > prev {
		if prev != -1 {
			s.tmp[prev] = 0
		}
	}
	prev = id
	if anchor == "decaf1ff" {
		s.delta["decaf1ff"] = 0
		s.delta["decaf2ff"] = s.PredictOffset("decaf2ff", ts)
		s.delta["decaf3ff"] = s.PredictOffset("decaf3ff", ts)
		s.delta["decaf4ff"] = s.PredictOffset("decaf4ff", ts)
	}
	s.tdoa[anchor] = ts
	if s.tmp[id] == 4 {
		s.tdoa["decaf2ff"] = s.tdoa["decaf2ff"] - s.delta["decaf2ff"]
		s.tdoa["decaf3ff"] = s.tdoa["decaf3ff"] - s.delta["decaf3ff"]
		s.tdoa["decaf4ff"] = s.tdoa["decaf4ff"] - s.delta["decaf4ff"] + 0xffffffffff
		if s.tdoa["decaf2ff"] > 0xffffffffff {
			s.tdoa["decaf2ff"] -= 0xffffffffff
		} else if s.tdoa["decaf2ff"] < 0 {
			s.tdoa["decaf2ff"] += 0xffffffffff
		}
		if s.tdoa["decaf3ff"] > 0xffffffffff {
			s.tdoa["decaf3ff"] -= 0xffffffffff
		} else if s.tdoa["decaf3ff"] < 0 {
			s.tdoa["decaf3ff"] += 0xffffffffff
		}
		if s.tdoa["decaf4ff"] > 0xffffffffff {
			s.tdoa["decaf4ff"] -= 0xffffffffff
		} else if s.tdoa["decaf4ff"] < 0 {
			s.tdoa["decaf4ff"] += 0xffffffffff
		}
		fmt.Println(s.tdoa)
		fmt.Printf("id: %d, d1: %d, d2: %d, d3: %d\n", id, s.tdoa["decaf2ff"]-s.tdoa["decaf1ff"],
			s.tdoa["decaf3ff"]-s.tdoa["decaf1ff"], s.tdoa["decaf4ff"]-s.tdoa["decaf1ff"])
	}
}

// s 1467629052039 0 decaf2ff 1047656524340 139673476849
func (s *Sync) ReadStr(str string) {
	if s.data == nil {
		s.data = map[string]([]Data){}
	}
	if s.delta == nil {
		s.delta = map[string]int64{}
	}
	if s.tdoa == nil {
		s.tdoa = map[string]int64{}
	}
	res := strings.Split(str, " ")
	if res[0] == "s" {
		T1, err := strconv.ParseInt(res[4], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		delta, err := strconv.ParseInt(res[5], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		s.Sync(res[3], T1, delta)
	} else if res[0] == "t" {
		TS, err := strconv.ParseInt(res[4], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		id, err := strconv.ParseInt(res[2], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		s.Calculate(res[3], TS, int(id))
	}
}

func (s *Sync) dump(target string) {
	for _, d := range s.data[target] {
		fmt.Println(d)
	}
}

func main() {
	s := new(Sync)
	file1, _ := os.Open("1.log")
	file2, _ := os.Open("2.log")
	buf1 := bufio.NewReader(file1)
	buf2 := bufio.NewReader(file2)
	chan1 := make(chan string)
	chan2 := make(chan string)
	go func() {
		for {
			line1, err := buf1.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			chan1 <- line1[:len(line1)-1]
		}
	}()
	go func() {
		for {
			line2, err := buf2.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			chan2 <- line2[:len(line2)-1]
		}
	}()
	str1 := <-chan1
	str2 := <-chan2
	var s1, s2 []string
	for {
		s1 = strings.Split(str1, " ")
		s2 = strings.Split(str2, " ")
		t1, err := strconv.ParseInt(s1[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		t2, err := strconv.ParseInt(s2[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		if t1 < t2 {
			s.ReadStr(str1)
			str1 = <-chan1
		} else {
			s.ReadStr(str2)
			str2 = <-chan2
		}
	}
	// s.ReadStr("s 1467629052039 0 decaf2ff 1047656524340 139673476849")
	// s.ReadStr("s 1467629052209 1 decaf2ff 1058551622196 139673468728")
	// s.ReadStr("s 1467629052379 2 decaf2ff 1069443044404 139673460587")
	// s.ReadStr("s 1467629052550 3 decaf2ff 1080331486772 139673452448")
	// s.ReadStr("s 1467629052720 4 decaf2ff 1091218266164 139673444291")
	// s.ReadStr("s 1467629052890 5 decaf2ff 2598408244 139673436139")
	// s.ReadStr("s 1467629053061 6 decaf2ff 13486643252 139673427972")
	// s.ReadStr("s 1467629053231 7 decaf2ff 24374496308 139673419783")
	// // 24379955293
	// s.ReadStr("s 1467629053402 8 decaf2ff 35260407348 139673411624")
	// s.ReadStr("s 1467629053572 9 decaf2ff 46143857204 139673403445")
	// fmt.Println(s.PredictOffset("decaf2ff", 46143857204))
}
