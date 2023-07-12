package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestScheduler_AddJob(t *testing.T) {
	t.Parallel()
	//unstable test due to timing issues
	var counter int
	scheduler := NewScheduler(time.Millisecond)
	scheduler.AddJob(&Job{
		D: time.Millisecond,
		Apply: func() {
			counter++
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*6)
	defer cancel()
	scheduler.Run(ctx)
	<-ctx.Done()
	if counter != 3 {
		t.Fail()
	}
}
