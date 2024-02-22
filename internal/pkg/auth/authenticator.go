package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
	"vivian.infra/utils"
)

const (
	CHARSET       string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	AUTH_KEY_SIZE int    = 5
	HASH_COST     int    = 12 
)

type Authenticator2FA interface {
	GenerateAuthKey2FA(context.Context, *utils.VivianLogger) (string, error)
	VerifyAuthKey2FA(context.Context, string, *utils.VivianLogger) (bool, error)
	ExpireAuthentication2FA(context.Context, *utils.VivianLogger) error
}

type HashManager struct {
	atomicValue atomic.Value
	flag        uint16
}

var hashManagerAtomic HashManager

func init() {
	hashManagerAtomic.flag = 1
}

func GenerateAuthKey2FA(ctx context.Context, s *utils.VivianLogger) (string, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if hashManagerAtomic.flag == 0 {
		return "", errors.New("2FA has already been generated")
	}

	hashManagerAtomic.flag = 0

	start := time.Now()

	authKeyGeneration := func() string {
		source := rand.New(rand.NewSource(time.Now().Unix()))
		var authKey strings.Builder
		for i := 0; i < AUTH_KEY_SIZE; i++ {
			sample := source.Intn(len(CHARSET))
			authKey.WriteString(string(CHARSET[sample]))
		}
		return authKey.String()
	}
	authKey := authKeyGeneration()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		authKeyHash, err := HashKeyphrase(ctx, authKey)
		if err != nil {
			s.LogError("failure during hashing process", err)
			return
		}
		hashManagerAtomic.atomicValue.Store([]byte(authKeyHash))
	}()
	wg.Wait()

	elapsed := time.Since(start)
	hash := hashManagerAtomic.atomicValue.Load().([]byte)

	if hash == nil {
		s.LogError("failure fetching the authentication key", errors.New("no hash available"))
		return "", nil
	}

	s.LogSuccess(fmt.Sprintf("authentication key generated: %v | %v", authKey, elapsed))
	return authKey, nil
}

func VerifyAuthKey2FA(ctx context.Context, key string, s *utils.VivianLogger) (bool, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if hashManagerAtomic.flag == 1 {
		s.LogWarning("2FA has not been initialized")
		return false, errors.New("2FA has not been initialized")
	}

	hash := hashManagerAtomic.atomicValue.Load()
	if hash == nil {
		s.LogWarning("failure fetching data from hashmanager")
		return false, errors.New("HashManager failure")
	}

	key = sanitize(key)
	if !ensure2FA(key) {
		s.LogWarning("invalid key")
		return false, errors.New("invalid key")
	}

	err := bcrypt.CompareHashAndPassword(hash.([]byte), []byte(key))
	if err != nil {
		return false, err
	} else {
		s.LogSuccess("verified key")
		Expire2FA(ctx, s)
		return true, nil
	}
}

func Expire2FA(ctx context.Context, s *utils.VivianLogger) error {
	if hashManagerAtomic.atomicValue.Load() == nil {
		err := errors.New("invalid hashmanager load")
		return err
	}
	hashManagerAtomic.atomicValue = atomic.Value{}
	hashManagerAtomic.flag = 1

	s.LogDebug(fmt.Sprintf("killed 2FA key at: %v}", time.Now().UTC()))
	return nil
}
