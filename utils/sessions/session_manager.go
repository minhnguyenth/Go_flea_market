package sessions

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type ISessionManager interface {
	SessionExists(token string) (bool, error)
	RegisterSession(token string) error
	DeleteSession(token string) error
}

var once sync.Once
var singleInstance *SessionManager = nil

const (
	SessionHashKey    = "session"
	SessionExpireHash = "session_expire"
	SessionLimitKey   = "session_limit"
	SessionTTL        = 30 * 60 * time.Second // 30 minutes

)

type SessionManager struct {
	redis *redis.Client
}

func GetSessionManager() (*SessionManager, error) {
	once.Do(func() {
		if singleInstance == nil {
			singleInstance = &SessionManager{
				redis: redis.NewClient(&redis.Options{
					Addr: os.Getenv("REDIS_HOST"),
				}),
			}
		}

		// cleaning up hash in redis mannaging session keys by go routine
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				ctx := context.Background()
				keys, err := singleInstance.redis.HKeys(ctx, SessionHashKey).Result()
				if err != nil {
					log.Printf("Error getting session keys for cleanup: %v", err)
					continue
				}

				for _, key := range keys {
					// Check if corresponding TTL key exists
					if exists, err := singleInstance.redis.Exists(ctx, key).Result(); err != nil {
						log.Printf("Error checking TTL key existence for %s: %v", key, err)
						continue
					} else if exists == 0 {
						// Remove from session hash if TTL key doesn't exist (expired)
						err := singleInstance.redis.HDel(ctx, SessionHashKey, key).Err()
						if err != nil {
							log.Printf("Error deleting expired session key %s from hash: %v", key, err)
						} else {
							log.Printf("Cleaned up expired session key: %s", key)
						}
					}
				}
			}
		}()

	})
	return singleInstance, nil
}

func (s *SessionManager) SessionExists(token string) (bool, error) {
	exists, err := s.redis.HExists(context.Background(), SessionHashKey, token).Result()
	if err != nil {
		return false, err
	}
	if exists {
		// for the user already has a session
		stillAlive, err := s.redis.Get(context.Background(), token).Result()
		if err != nil {
			if err == redis.Nil {
				return false, nil
			}
			return false, err
		}
		if stillAlive != "" {
			return true, nil
		}
	}

	return false, nil
}

func (s *SessionManager) RegisterSession(token string) (bool, error) {

	// check there are still some space for new session (total number of session is less than session_limit set in redis)
	sessionLimit, err := s.redis.Get(context.Background(), SessionLimitKey).Result()
	if err != nil {
		return false, err
	}

	// get current number of sessions
	sessionCount, err := s.redis.HLen(context.Background(), SessionHashKey).Result()
	if err != nil {
		return false, err
	}

	limit, err := strconv.Atoi(sessionLimit)
	if err != nil {
		return false, err
	}

	if int64(limit) < sessionCount {
		log.Println("session limit is reached : ", sessionCount, " / ", limit)
		return false, errors.New("session limit is reached")
	}

	// register new session
	s.redis.HSet(context.Background(), SessionHashKey, token, "1")
	// try
	// > TTL token
	// in redis cli to see session TTL(= Time to Live)
	s.redis.Set(context.Background(), token, "", SessionTTL)
	return true, nil
}

func (s *SessionManager) DeleteSession(token string) error {
	s.redis.HDel(context.Background(), SessionHashKey, token)
	s.redis.Del(context.Background(), token)
	return nil
}
