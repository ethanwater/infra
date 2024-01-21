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
	"vivian.infra/internal/utils"
)

const (
	CHARSET       string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	AUTH_KEY_SIZE int    = 5
)

type Authenticator interface {
	//Login
	LoginUser(context.Context, string, string, *utils.VivianLogger) (bool, error)

	//2FA
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
	if hashManagerAtomic.flag == 0 {
		return "", errors.New("2FA has already been generated")
	}

	hashManagerAtomic.flag = 0

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
			s.LogError("Failure during hashing process", err)
			return
		}
		hashManagerAtomic.atomicValue.Store([]byte(authKeyHash))
	}()
	wg.Wait()

	hash := hashManagerAtomic.atomicValue.Load().([]byte)

	if hash == nil {
		s.LogError("Failure fetching the authentication key", errors.New("no hash available"))
		return "", nil
	}

	s.LogSuccess(fmt.Sprintf("Authentication key generated: %v", authKey))
	return authKey, nil
}

func VerifyAuthKey2FA(ctx context.Context, key string, s *utils.VivianLogger) (bool, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if hashManagerAtomic.flag == 1 {
		s.LogWarning("2FA has not been initialized")
		return false, nil
	}

	hash := hashManagerAtomic.atomicValue.Load()
	if hash == nil {
		s.LogWarning("HashManager failure")
		return false, nil
	}

	key = sanitize(key)
	if sanitizeCheck(key) {
		err := bcrypt.CompareHashAndPassword(hash.([]byte), []byte(key))
		if err != nil {
			s.LogWarning("Invalid key")
			return false, err
		} else {
			s.LogSuccess("Verified key")
			Expire2FA(ctx, s)
			return true, nil
		}
	}

	return false, errors.New("unable to Sanitize") //how the fuck would you get this anyways
}

func Expire2FA(ctx context.Context, s *utils.VivianLogger) error {
	if hashManagerAtomic.atomicValue.Load() == nil {
		err := errors.New("HashManager is already nil")
		return err
	}
	hashManagerAtomic.atomicValue = atomic.Value{}
	hashManagerAtomic.flag = 1

	s.LogDebug(fmt.Sprintf("Killing 2FA Key {expired at: %v}", time.Now().UTC()))
	return nil
}
