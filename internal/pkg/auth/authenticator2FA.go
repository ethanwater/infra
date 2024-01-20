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
	charset     string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	authKeySize int    = 5
)

type T interface {
	GenerateAuthKey2FA(context.Context, *utils.VivianLogger) (string, error)
	VerifyAuthKey2FA(context.Context, string, *utils.VivianLogger) (bool, error)
	ExpireAuthentication2FA(context.Context, *utils.VivianLogger) error
}

type HashManager struct {
	atomicValue atomic.Value
	flag        uint16
}

var HashManagerAtomic HashManager

func GenerateAuthKey2FA(ctx context.Context, s *utils.VivianLogger) (string, error) {
	source := rand.New(rand.NewSource(time.Now().Unix()))
	var authKey strings.Builder

	for i := 0; i < authKeySize; i++ {
		sample := source.Intn(len(charset))
		authKey.WriteString(string(charset[sample]))
	}

	HashManagerAtomic.flag = 0

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		authKeyHash, err := HashKeyphrase(ctx, authKey.String())
		if err != nil {
			s.LogError("failure hashing the authentication key", err)
			return
		}
		HashManagerAtomic.atomicValue.Store([]byte(authKeyHash))
	}()
	wg.Wait()

	hash := HashManagerAtomic.atomicValue.Load().([]byte)

	if hash == nil {
		s.LogError("failure hashing the authentication key", errors.New("empty hash"))
		return "", nil
	}

	s.LogSuccess(fmt.Sprintf("authentication key generated: %v", authKey.String()))
	//t.sender.Get().SendVerificationCodeEmail(ctx, email, authKey.String())
	return authKey.String(), nil
}

func VerifyAuthKey2FA(ctx context.Context, key string, s *utils.VivianLogger) (bool, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if HashManagerAtomic.flag == 1 {
		return false, nil
	}

	hash := HashManagerAtomic.atomicValue.Load()
	if hash == nil {
		// Handle the case where the value is nil
		s.LogWarning("2FA hash is not initialized")
		return false, nil
	}

	if SanitizeCheck(key) {
		status := bcrypt.CompareHashAndPassword(hash.([]byte), []byte(key))
		if status != nil {
			s.LogWarning("invalid key")
			return false, status
		} else {
			s.LogSuccess("verified key")
			return true, status
		}
	}

	return false, nil
}

func Expire2FA(ctx context.Context, s *utils.VivianLogger) error {
	if HashManagerAtomic.atomicValue.Load() == nil {
		err := errors.New("HashManagerAtomic is already nil")
		return err
	}
	HashManagerAtomic.atomicValue = atomic.Value{}
	HashManagerAtomic.flag = 1

	s.LogDebug(fmt.Sprintf("killing 2FA -> expired at: %v", time.Now().UTC()))
	return nil
}
