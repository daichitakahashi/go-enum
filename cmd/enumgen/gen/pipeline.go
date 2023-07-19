package gen

func iterate[T any](s []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, v := range s {
			out <- v
		}
	}()
	return out
}

type pipelineStageFunc[In any, Out any] func(in <-chan In) <-chan Out

func pipelineStage[In any, Out any](fn func(in In, out chan Out)) pipelineStageFunc[In, Out] {
	return pipelineStageFunc[In, Out](func(in <-chan In) <-chan Out {
		out := make(chan Out)
		go func() {
			defer close(out)
			for i := range in {
				fn(i, out)
			}
		}()
		return out
	})
}
