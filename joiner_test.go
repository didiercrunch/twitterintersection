package main

import (
	"reflect"
	"testing"
)

func makeUInt64Channel(numbers ...uint64) <-chan uint64 {
	c := make(chan uint64)
	go func() {
		for _, n := range numbers {
			c <- n
		}
		close(c)
	}()
	return c
}

func readUInt64Channel(c <-chan uint64) []uint64 {
	ret := make([]uint64, 0)
	for n := range c {
		ret = append(ret, n)
	}
	return ret
}

func equalsAsMultiSet(a, b []uint64) bool {
	aSet, bSet := make(map[uint64]int), make(map[uint64]int)
	for _, e := range a {
		if _, ok := aSet[e]; ok {
			aSet[e]++
		} else {
			aSet[e] = 1
		}
	}

	for _, e := range b {
		if _, ok := aSet[e]; ok {
			bSet[e]++
		} else {
			bSet[e] = 1
		}
	}
	return reflect.DeepEqual(aSet, bSet)
}

func TestUniq(t *testing.T) {
	c := makeUInt64Channel(1, 3, 2, 1, 4, 3, 3, 2, 1)
	expt := []uint64{1, 3, 2, 4}
	if ret := readUInt64Channel(uniq(c)); !reflect.DeepEqual(expt, ret) {
		t.Fail()
	}
}

func TestIntersection(t *testing.T) {
	c := Intersection(makeUInt64Channel(1, 2, 3, 4), makeUInt64Channel(1, 2, 3, 4))
	if s := readUInt64Channel(c); !equalsAsMultiSet(s, []uint64{1, 2, 3, 4}) {
		t.Error("needed {1, 2, 3 , 4} received ", s)
	}

	c = Intersection(makeUInt64Channel(1, 2), makeUInt64Channel(1, 2, 3, 4))
	if s := readUInt64Channel(c); !equalsAsMultiSet(s, []uint64{1, 2}) {
		t.Error("needed {1, 2} received ", s)
	}

	c = Intersection(makeUInt64Channel(3, 4), makeUInt64Channel(1, 2, 3, 4))
	if s := readUInt64Channel(c); !equalsAsMultiSet(s, []uint64{3, 4}) {
		t.Error("needed {3 , 4} received ", s)
	}
}
