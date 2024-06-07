package goutils

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
