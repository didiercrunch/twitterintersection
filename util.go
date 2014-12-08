package main

import "sync"

// returns all the element x that are in c1 and c2.  can retuns many time
// the same x.
func intersection(c1, c2 <-chan uint64) <-chan uint64 {
	var wg sync.WaitGroup
	out := make(chan uint64)

	mc1, mc2 := make(map[uint64]bool), make(map[uint64]bool)

	filter := func(c <-chan uint64, mapToAdd, mapToVerify map[uint64]bool) {
		for n := range c {
			mapToAdd[n] = true
			if _, ok := mapToVerify[n]; ok {
				out <- n
			}
		}
		wg.Done()
	}
	wg.Add(2)
	go filter(c1, mc1, mc2)
	go filter(c2, mc2, mc1)

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// remove all duplicates in inputC
func uniq(inputC <-chan uint64) <-chan uint64 {
	m := make(map[uint64]bool)
	c := make(chan uint64)
	go func() {
		for n := range inputC {
			if _, ok := m[n]; !ok {
				m[n] = true
				c <- n
			}
		}
		close(c)
	}()
	return c
}

// returns the intersection of a and b without repetitions
func Intersection(a, b <-chan uint64) <-chan uint64 {
	return uniq(intersection(a, b))

}
