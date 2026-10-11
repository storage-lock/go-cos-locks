package cos_locks

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	cos_storage "github.com/storage-lock/go-cos-storage"
	"github.com/stretchr/testify/assert"
	"github.com/tencentyun/cos-go-sdk-v5"
)

func testCosClient(t *testing.T) *cos.Client {
	t.Helper()
	u, err := url.Parse("https://test-bucket.cos.ap-beijing.myqcloud.com")
	assert.Nil(t, err)
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, http.DefaultClient)
}

func TestCosClientConnectionManager(t *testing.T) {
	client := testCosClient(t)
	m := &cosClientConnectionManager{client: client}
	assert.Equal(t, "cos-client-connection-manager", m.Name())

	ctx := context.Background()
	got, err := m.Take(ctx)
	assert.Nil(t, err)
	assert.Equal(t, client, got)
	assert.Nil(t, m.Return(ctx, got))
	assert.Nil(t, m.Shutdown(ctx))
}

func TestGetCOSLockFactoryByClient(t *testing.T) {
	client := testCosClient(t)
	ctx := context.Background()

	// COS 存储构造不探活，本地即可成功（但 COS 未声明 CAS，建锁报能力缺失）
	factory, err := GetCOSLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetCOSLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	_, err = NewCOSLockByClient(ctx, client, "test-cos-lock")
	assert.NotNil(t, err)
}

func TestGetCOSLockFactoryByConfig(t *testing.T) {
	ctx := context.Background()
	config := &COSConfig{
		BucketURL: "https://test-bucket.cos.ap-beijing.myqcloud.com",
		SecretID:  "test-id",
		SecretKey: "test-key",
	}

	factory, err := GetCOSLockFactoryByConfig(ctx, config)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetCOSLockFactoryByConfig(ctx, config)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	_, err = NewCOSLockByConfig(ctx, config, "test-cos-lock")
	assert.NotNil(t, err)
}

func TestCOSLock_StorageError(t *testing.T) {
	old := newCOSStorage
	newCOSStorage = func(ctx context.Context, options *cos_storage.COSStorageOptions) (*cos_storage.COSStorage, error) {
		return nil, errors.New("boom")
	}
	defer func() { newCOSStorage = old }()

	ctx := context.Background()

	// 独立 client+独立 BucketURL，避免命中成功用例的缓存
	errClient := testCosClient(t)
	errClient.BaseURL.BucketURL.Host = "boom-bucket.cos.ap-beijing.myqcloud.com"
	_, err := GetCOSLockFactoryByClient(ctx, errClient)
	assert.NotNil(t, err)
	_, err = NewCOSLockByClient(ctx, errClient, "x")
	assert.NotNil(t, err)

	errConfig := &COSConfig{
		BucketURL: "https://boom-config.cos.ap-beijing.myqcloud.com",
		SecretID:  "test-id",
		SecretKey: "test-key",
	}
	_, err = GetCOSLockFactoryByConfig(ctx, errConfig)
	assert.NotNil(t, err)
	_, err = NewCOSLockByConfig(ctx, errConfig, "x")
	assert.NotNil(t, err)
}
