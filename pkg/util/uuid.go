package util

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GenUUID xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func GenUUID() string {
	return uuid.New().String()
}

// GenApiUUID ww-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func GenApiUUID() string {
	return fmt.Sprintf("ww-%s", strings.Replace(uuid.New().String(), "-", "", -1))
}
