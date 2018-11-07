package ccache

import (
	"fmt"
	"time"
)

type RecursionTimer struct {
	Title string
	Durations map[string]time.Duration
	Recursion []string
	StartTimes []time.Time
}

func NewRecursionTimer(title string) *RecursionTimer {
	return &RecursionTimer{
		title,
	make(map[string]time.Duration),
		make([]string, 0),
		make([]time.Time, 0),
		}
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

	cur += time.Now().Sub(t.StartTimes[len(t.StartTimes) - 1])

	t.Durations[f] = cur
	t.StartTimes = t.StartTimes[:len(t.StartTimes) - 1]
}

func (t *RecursionTimer) Report() {
	fmt.Println("Title: ", t.Title)
	for k, v := range t.Durations {
		fmt.Println("", k + ":", v)
	}
}