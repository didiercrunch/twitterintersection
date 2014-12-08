package main

import (
	"fmt"
	"reflect"
	"testing"
)

type MockFollowerGetter struct {
	T *testing.T
}

func (m *MockFollowerGetter) GetFollowerByCursor(screenName, cursor string) <-chan *FollowerList {
	followerListC := make(chan *FollowerList)
	go func() {
		u1 := &User{"nat", 78789}
		u2 := &User{"jude", 78789}
		u3 := &User{"alice", 78789}
		u4 := &User{"bob", 78789}
		switch cursor {
		case "-1":
			followerListC <- &FollowerList{"1", []*User{u1, u2}}
		case "1":
			followerListC <- &FollowerList{"0", []*User{u3, u4}}
		case "0":
		default:
			m.T.Error("bad cursor", cursor)
		}
	}()
	return followerListC
}

func (m *MockFollowerGetter) GetFollowerIdsByCursor(screenName, cursor string) <-chan *FollowerIDList {
	followerListC := make(chan *FollowerIDList)
	go func() {
		switch cursor {
		case "-1":
			followerListC <- &FollowerIDList{"1", []uint64{1, 2}}
		case "1":
			followerListC <- &FollowerIDList{"0", []uint64{3, 4}}
		case "0":
		default:
			m.T.Error("bad cursor", cursor)
		}
	}()
	return followerListC
}

func (m *MockFollowerGetter) GetScreenNameOfUsersByIds(ids []uint64) <-chan string {
	ret := make(chan string)
	go func() {
		for _, id := range ids {
			ret <- fmt.Sprintf("%v", id)
		}
		close(ret)
	}()
	return ret
}

func readAllStringFromChannel(c <-chan string) []string {
	ret := make([]string, 0)
	for s := range c {
		ret = append(ret, s)
	}
	return ret
}

func readAllUInt64FromChannel(c <-chan uint64) []uint64 {
	ret := make([]uint64, 0)
	for s := range c {
		ret = append(ret, s)
	}
	return ret
}

func TestGetFollowerScreenNames(t *testing.T) {
	fg := &MockFollowerGetter{t}
	followerC := GetFollowerScreenNames(fg, "justinBieber")
	followers := readAllStringFromChannel(followerC)
	expected := []string{"nat", "jude", "alice", "bob"}
	if !reflect.DeepEqual(followers, expected) {
		t.Fail()
	}
}

func TestGetFollowerIds(t *testing.T) {
	fg := &MockFollowerGetter{t}
	ids := readAllUInt64FromChannel(GetFollowerIds(fg, "justinBieber"))
	expected := []uint64{1, 2, 3, 4}
	if !reflect.DeepEqual(ids, expected) {
		t.Error(ids)
	}
}

func uint64Range(n int) []uint64 {
	ret := make([]uint64, n)
	for i := 0; i < n; i++ {
		ret[i] = uint64(i)
	}
	return ret
}

func TestGetScreenNameByIds(t *testing.T) {
	fg := &MockFollowerGetter{t}
	ids := uint64Range(233)
	ret := make(map[string]bool)
	for name := range GetScreenNameByIds(fg, makeUInt64Channel(ids...)) {
		ret[name] = true
	}
	if len(ret) != 233 {
		t.Fail()
	}
}
