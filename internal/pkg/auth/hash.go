package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

const cost int = 12

func HashKeyphrase(_ context.Context, password string) (string, error) {
	hashChannel := make(chan struct {
		hash string
		err  error
	})
	defer close(hashChannel)

	go func() {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		hashChannel <- struct {
			hash string
			err  error
		}{string(hash), err}
	}()

	result := <-hashChannel
	return result.hash, result.err
}

func VerfiyHashKeyphrase(hash, password string) bool {
	verificationChannel := make(chan bool)
	defer close(verificationChannel)

	go func() {
		status := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		verificationChannel <- status == nil
	}()

	return <-verificationChannel
}
