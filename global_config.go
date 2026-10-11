package cos_locks

import (
	"context"

	cos_storage "github.com/storage-lock/go-cos-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSConfig COS 端点配置
type COSConfig struct {
	BucketURL string
	SecretID  string
	SecretKey string
}

var cosEndpointBeanFactory = storage_lock_factory.NewStorageLockFactoryBeanFactory[string, *cos.Client]()

// NewCOSLockByConfig 从配置创建分布式锁
func NewCOSLockByConfig(ctx context.Context, config *COSConfig, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetCOSLockFactoryByConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

// GetCOSLockFactoryByConfig 获取基于配置的锁工厂
func GetCOSLockFactoryByConfig(ctx context.Context, config *COSConfig) (*storage_lock_factory.StorageLockFactory[*cos.Client], error) {
	cacheKey := config.BucketURL

	return cosEndpointBeanFactory.GetOrInit(ctx, cacheKey, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*cos.Client], error) {
		cm := cos_storage.NewCOSConnectionManagerFromConfig(config.BucketURL, config.SecretID, config.SecretKey)
		options := cos_storage.NewCOSStorageOptions().
			SetConnectionManager(cm).
			SetTimeProvider(&cos_storage.LocalTimeProvider{})
		// newCOSStorage 定义在 global_client.go，双入口共用以便单测统一注入
		storage, err := newCOSStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		return storage_lock_factory.NewStorageLockFactory[*cos.Client](storage, nil), nil
	})
}
