package cache

import (
	"time"

	"github.com/seargo/seargo/pkg/models"
)

type Cache interface {
	Get(key string) (*models.Response, bool)
	Set(key string, value *models.Response, ttl time.Duration)
	Delete(key string)
}
