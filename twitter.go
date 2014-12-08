package main

import (
	"fmt"
	"sync"
)

type FollowerGetter interface {
	GetFollowerByCursor(screenName, cursor string) <-chan *FollowerList
	GetGetFollowerIdsByCursor(screenName, cursor string) <-chan *FollowerIDList
	GetScreenNameOfUsersByIds(ids []uint64) <-chan string
}

func GetFollowerScreenNames(followerGetter FollowerGetter, screenName string) <-chan string {
	followerC := make(chan string)
	go func() {
		nextCursor := "-1"
		for nextCursor != "0" && nextCursor != "" {
			followers := <-followerGetter.GetFollowerByCursor(screenName, nextCursor)
			nextCursor = followers.NextCursor
			for _, follower := range followers.GetFollowerScreenNames() {
				followerC <- follower
			}
		}
		close(followerC)
	}()
	return followerC
}

func GetFollowerIds(followerGetter FollowerGetter, screenName string) <-chan uint64 {
	followerC := make(chan uint64)
	go func() {
		nextCursor := "-1"
		for nextCursor != "0" && nextCursor != "" {
			followers := <-followerGetter.GetGetFollowerIdsByCursor(screenName, nextCursor)
			nextCursor = followers.NextCursor
			for _, follower := range followers.Followers {
				followerC <- follower
			}
		}
		close(followerC)
	}()
	return followerC
}

func GetTwitterScreenNameByIds(followerGetter FollowerGetter, idsC <-chan uint64) <-chan string {
	screenNameC := make(chan string)
	var wg sync.WaitGroup

	produceScreenName := func(buffer []uint64) {
		for screenName := range followerGetter.GetScreenNameOfUsersByIds(buffer) {
			screenNameC <- screenName
		}
		wg.Done()
	}

	go func() {
		buffer := make([]uint64, 0, 100)
		for id := range idsC {
			buffer = append(buffer, id)
			if len(buffer) >= 95 {
				wg.Add(1)
				go produceScreenName(buffer)
				buffer = make([]uint64, 0, 100)
			}
		}
		wg.Add(1)
		go produceScreenName(buffer)
		go func() {
			wg.Wait()
			close(screenNameC)
		}()
	}()
	return screenNameC
}

func GetFollowerIdsOfBothAccounts(followerGetter FollowerGetter, screenName1, screenName2 string) <-chan uint64 {
	return Intersection(GetFollowerIds(followerGetter, screenName1), GetFollowerIds(followerGetter, screenName1))
}

func main() {
	token := "AAAAAAAAAAAAAAAAAAAAAPwfcQAAAAAAzkou%2FHjJNJmwdepeRq0c%2Bi3Nx6o%3DXofLt7SVvc99ulETLRA3yS2lYo8smfc6tACxEYsLUmGsrNbc9J"
	t := NewTwitterApi(TWITTER_API_URL, token)
	ff := Intersection(GetFollowerIds(t, "Shopify"), GetFollowerIds(t, "ShopifyPicks"))
	for screenName := range GetTwitterScreenNameByIds(t, ff) {
		fmt.Println(screenName)
	}
}
