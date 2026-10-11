package cos_locks

import (
	"context"

	cos_storage "github.com/storage-lock/go-cos-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// cosClientConnectionManager 包装已有的 cos.Client
type cosClientConnectionManager struct {
	client *cos.Client
}

func (m *cosClientConnectionManager) Name() string {
	return "cos-client-connection-manager"
}

func (m *cosClientConnectionManager) Take(ctx context.Context) (*cos.Client, error) {
	return m.client, nil
}

func (m *cosClientConnectionManager) Return(ctx context.Context, connection *cos.Client) error {
	return nil
}

func (m *cosClientConnectionManager) Shutdown(ctx context.Context) error {
	return nil
}

var cosClientBeanFactory = storage_lock_factory.NewStorageLockFactoryBeanFactory[*cos.Client, *cos.Client]()

// newCOSStorage 指向 cos_storage.NewCOSStorage，包级变量以便单测注入构造失败，
// 覆盖 GetOrInit initFunc 的错误分支。client/config 双入口共用。
var newCOSStorage = cos_storage.NewCOSStorage

// NewCOSLockByClient 从已有 COS Client 创建分布式锁
func NewCOSLockByClient(ctx context.Context, client *cos.Client, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetCOSLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

// GetCOSLockFactoryByClient 获取基于 COS Client 的锁工厂
func GetCOSLockFactoryByClient(ctx context.Context, client *cos.Client) (*storage_lock_factory.StorageLockFactory[*cos.Client], error) {
	return cosClientBeanFactory.GetOrInit(ctx, client, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*cos.Client], error) {
		cm := &cosClientConnectionManager{client: client}
		options := cos_storage.NewCOSStorageOptions().
			SetConnectionManager(cm).
			SetTimeProvider(&cos_storage.LocalTimeProvider{})
		storage, err := newCOSStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		return storage_lock_factory.NewStorageLockFactory[*cos.Client](storage, nil), nil
	})
}
