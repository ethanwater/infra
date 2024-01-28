package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateDeploymentID() string {
	randomUUID := uuid.New()
	shortUUID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		randomUUID[:4], randomUUID[4:6], randomUUID[6:8],
		randomUUID[8:10], randomUUID[10:])

	return shortUUID
}
