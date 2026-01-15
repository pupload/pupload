package runtime_repo

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/pupload/pupload/internal/controller/flows/runtime"

	"github.com/redis/go-redis/v9"
)

type RedisRuntimeRepo struct {
	client *redis.Client
}

func CreateRedisRuntimeRepo(client *redis.Client) *RedisRuntimeRepo {
	return &RedisRuntimeRepo{
		client: client,
	}
}

func (r *RedisRuntimeRepo) SaveRuntime(rt runtime.RuntimeFlow) error {
	key := fmt.Sprintf("flowrun:%s", rt.FlowRun.ID)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(rt); err != nil {
		return err
	}

	if err := r.client.Set(context.TODO(), key, buf.Bytes(), 0).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisRuntimeRepo) LoadRuntime(runID string) (runtime.RuntimeFlow, error) {
	key := fmt.Sprintf("flowrun:%s", runID)
	raw, err := r.client.Get(context.TODO(), key).Bytes()
	if err != nil {
		return runtime.RuntimeFlow{}, err
	}

	var val runtime.RuntimeFlow
	dec := gob.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&val); err != nil {
		return val, err
	}

	return val, nil
}

func (r *RedisRuntimeRepo) DeleteRuntime(runID string) error {
	key := fmt.Sprintf("flowrun:%s", runID)
	_, err := r.client.Del(context.TODO(), key).Result()
	return err
}

func (r *RedisRuntimeRepo) Close(ctx context.Context) error {
	return r.client.Close()
}

func (r *RedisRuntimeRepo) ListRuntimeIDs() ([]string, error) {

	list := make([]string, 0)

	keys, _, err := r.client.Scan(context.TODO(), 0, "flowrun:*", 1000).Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		id := key[8:]
		list = append(list, id)
	}

	return list, nil
}
