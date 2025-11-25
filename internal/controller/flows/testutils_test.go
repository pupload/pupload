package flows

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"pupload/internal/models"
	locals3 "pupload/internal/stores/local_s3"
)

// NewMiniredisClient starts an in-memory Redis server and returns
// the miniredis instance and a go-redis client connected to it.
func NewMiniredisClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	m, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: m.Addr()})
	return m, rdb
}

// NewTestFlowService creates a FlowService configured for tests.
// It uses an in-memory Redis instance and returns a cleanup function.
func NewTestFlowService(t *testing.T) (*FlowService, func()) {
	t.Helper()
	m, rdb := NewMiniredisClient(t)

	f := &FlowService{
		FlowPath:       t.TempDir(),
		FlowList:       make(map[string]models.Flow),
		NodeDefs:       make(map[string]models.NodeDef),
		RedisClient:    rdb,
		AsynqClient:    asynq.NewClientFromRedisClient(rdb),
		AsynqServer:    nil,
		LocalStoreMap:  make(map[LocalStoreKey]models.Store),
		GlobalStoreMap: make(map[string]models.Store),
	}

	cleanup := func() {
		f.AsynqClient.Close()
		f.RedisClient.Close()
		m.Close()
	}

	return f, cleanup
}

// AddLocalS3Store registers a LocalS3Store for the given flow and store name.
// Returns the created store (as models.Store) so tests can interact with it.
func AddLocalS3Store(t *testing.T, f *FlowService, flowName, storeName, bucket string) models.Store {
	t.Helper()
	s, err := locals3.NewLocalS3Store(locals3.LocalS3StoreInput{BucketName: bucket})
	if err != nil {
		t.Fatalf("NewLocalS3Store: %v", err)
	}

	f.LocalStoreMap[LocalStoreKey{flowName, storeName}] = s
	return s
}
