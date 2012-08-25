package main

func bisect(a []int, x int) int {
 	lo, hi := 0, len(a)
  	
  	for lo < hi {
    		mid := lo + hi >> 1
    		if a[mid] < x {
    			lo = mid + 1
    		} else {
    			hi = mid
    		}
  	}
  	
  	return lo
}