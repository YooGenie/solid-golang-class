package ratelimit

import (
	"time"

	"golang.org/x/time/rate"
)

type RateLimit struct {
	Limit float64 `json:"limit,omitempty"`
	Burst int     `json:"burst,omitempty"`
}

func NewRateLimiter(config RateLimit) *rate.Limiter { // *rate.Limiter 라이브러리 사용한다. rateLimiter 인스턴스 반환을 한다.
	// 초기화를 해서 파일시스템 rateLimiter에 넣어 주는 것이다.

	limit := config.Limit // 초당 몇개의 리퀘스트를 조절할 것인가
	burst := config.Burst // 버퍼와 같은 것이다. 초반에 리퀘스트가 밀려 들어올 수 있어서 그럴 때 몇 개까지 허용할 것인가.

	// calculate interval in ms to achieve the target limit per second
	interval := 1000.0 / float64(limit)
	rateLimiter := rate.NewLimiter(rate.Every(time.Duration(interval)*time.Millisecond), burst)

	return rateLimiter
}
