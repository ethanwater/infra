package app

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	BUCKET_LIMITER_SIZE      uint32        = 10
	BUCKET_LIMITER_LEAK_AMT  uint32        = 1
	BUCKET_LIMITER_LEAK_RATE time.Duration = 500 * time.Millisecond
)

var RequestChannel *chan uint32
var RequestChannelCounter *uint32
var KillRequestTickerChannel = make(chan uint16)

type Limiter struct {
	requestTicker         time.Ticker
	requestChannel        chan uint32
	requestChannelCounter uint32
	requestBlockerState   atomic.Uint32
}

func init() {
	limiter := Limiter{
		requestTicker:         *time.NewTicker(BUCKET_LIMITER_LEAK_RATE),
		requestChannel:        make(chan uint32, BUCKET_LIMITER_SIZE),
		requestChannelCounter: 0,
		requestBlockerState:   atomic.Uint32{},
	}
	RequestChannel = &limiter.requestChannel
	RequestChannelCounter = &limiter.requestChannelCounter
	limiter.RateLimiter()
}

func (l *Limiter) RateLimiter() {
	go func() {
		for {
			select {
			case <-l.requestTicker.C:
				if l.requestChannelCounter >= 10 {
					VivianServerLogger.LogWarning(fmt.Sprintf("Blocking channel {status code:%v}", http.StatusTooManyRequests))
				}
				//fmt.Println("Channel Len:", len(l.requestChannel), "Channel Cap:", cap(l.requestChannel), "Pool:", l.requestChannelCounter)
				if l.requestChannelCounter > 0 {
					l.requestChannelCounter -= BUCKET_LIMITER_LEAK_AMT
				} else {
					l.requestChannelCounter = 0
					l.requestChannel = make(chan uint32, BUCKET_LIMITER_SIZE)
				}
			case <-KillRequestTickerChannel:
				VivianServerLogger.LogDebug("killing request ticker channel")
				return
			}
		}
	}()
}
