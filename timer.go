package ccache

import "time"

type RecursionTimer struct {
	Durations map[string]time.Duration
	Recursion []string
	StartTimes []time.Time
}

func (t *RecursionTimer) Enter(fun string) {
	t.Recursion = append(t.Recursion, fun)
	t.StartTimes = append(t.StartTimes, time.Now())
}

func (t *RecursionTimer) Leave() {
	f := t.Recursion[len(t.Recursion) - 1]
	t.Recursion = t.Recursion[:len(t.Recursion) - 1]

	cur, ok := t.Durations[f]
	if !ok {
		cur = 0
	}

	cur += time.Now().Sub(t.StartTimes[len(t.StartTimes)])

	t.Durations[f] = cur
	t.StartTimes = t.StartTimes[:len(t.StartTimes) - 1]
}

