package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	currentIn := in
	for _, stage := range stages {
		currentIn = stage(currentIn)
	}
	finalOut := make(chan interface{})
	go func() {
		defer close(finalOut)
		for {
			select {
			case <-done:
				return
			case val, ok := <-currentIn:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case finalOut <- val:
				}
			}
		}
	}()
	return finalOut
}
