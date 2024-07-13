// Package utils provides utility functions.
package utils

func ZipChannels[T interface{}](channels ...<-chan T) <-chan T {
	output := make(chan T)
	for _, ch := range channels {
		go func() {
			for i := range ch {
				output <- i
			}
		}()
	}
	return output
}
