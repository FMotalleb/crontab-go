// Package utils provides utility functions.
package utils

import "sync"

func ZipChannels[T interface{}](channels ...<-chan T) <-chan T {
	output := make(chan T)
	wg := new(sync.WaitGroup)
	for _, ch := range channels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				output <- i
			}
		}()
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}
