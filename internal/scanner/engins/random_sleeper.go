package engines

import (
	"math/rand"
	"time"
)

type RandomSleeperEngine interface {
	RandomSleep() time.Duration
	GetRandomDuration() time.Duration
}

type randomSleeperEngine struct {
	minDuration, maxDuration time.Duration
	rng                      *rand.Rand
}

func NewRandomSleeperEngine(minDuration, maxDuration time.Duration) RandomSleeperEngine {
	return &randomSleeperEngine{
		minDuration: minDuration,
		maxDuration: maxDuration,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *randomSleeperEngine) RandomSleep() time.Duration {
	duration := r.GetRandomDuration()
	time.Sleep(duration)
	return duration
}

func (r *randomSleeperEngine) GetRandomDuration() time.Duration {
	if r.minDuration >= r.maxDuration {
		return r.minDuration
	}

	diff := r.maxDuration - r.minDuration
	randomNano := r.rng.Int63n(int64(diff))
	return r.minDuration + time.Duration(randomNano)
}
