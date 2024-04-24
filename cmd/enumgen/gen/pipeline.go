package gen

import "sync"

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

func pipe[A any, B any, C any](first func(A) B, second func(B) C) func(A) C {
	return func(in A) C {
		return second(first(in))
	}
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

func merge[In any, Out1 any, Out2 any, Out3 any](
	fn1 pipelineStageFunc[In, Out1],
	fn2 pipelineStageFunc[In, Out2],
	merge func([]Out1, []Out2) Out3,
) func(in <-chan In) Out3 {
	return (func(in <-chan In) Out3 {
		var (
			in1  = make(chan In)
			in2  = make(chan In)
			out1 = fn1(in1)
			out2 = fn2(in2)
		)
		go func() {
			defer close(in1)
			defer close(in2)
			for i := range in {
				in1 <- i
				in2 <- i
			}
		}()

		var (
			wg   sync.WaitGroup
			acc1 = make([]Out1, 0, 10)
			acc2 = make([]Out2, 0, 10)
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			for o := range out1 {
				o := o
				acc1 = append(acc1, o)
			}
		}()
		go func() {
			defer wg.Done()
			for o := range out2 {
				o := o
				acc2 = append(acc2, o)
			}
		}()
		wg.Wait()

		return merge(acc1, acc2)
	})
}

type pair[A any, B any] struct {
	left  A
	right B
}
