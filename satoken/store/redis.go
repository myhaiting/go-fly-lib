package store

import (
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/myhaiting/go-fly-lib/satoken"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	_    satoken.TokenStore = &TokenStore{}
	json                    = sonic.ConfigStd
	// Marshal is exported by gin/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by gin/json package.
	Unmarshal = json.Unmarshal
)

// NewRedisStore create an instance of a redis store
func NewRedisStore(opts *redis.Options) *TokenStore {
	if opts == nil {
		panic("options cannot be nil")
	}
	return NewRedisStoreWithCli(redis.NewClient(opts))
}

// NewRedisStoreWithCli create an instance of a redis store
func NewRedisStoreWithCli(cli *redis.Client) *TokenStore {
	return NewRedisStoreWithInterface(cli)
}

// NewRedisClusterStore create an instance of a redis cluster store
func NewRedisClusterStore(opts *redis.ClusterOptions) *TokenStore {
	if opts == nil {
		panic("options cannot be nil")
	}
	return NewRedisClusterStoreWithCli(redis.NewClusterClient(opts))
}

// NewRedisClusterStoreWithCli create an instance of a redis cluster store
func NewRedisClusterStoreWithCli(cli *redis.ClusterClient) *TokenStore {
	return NewRedisStoreWithInterface(cli)
}

// NewRedisStoreWithInterface create an instance of a redis store
func NewRedisStoreWithInterface(cli clienter) *TokenStore {
	store := &TokenStore{
		cli: cli,
	}
	return store
}

type clienter interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Exists(ctx context.Context, key ...string) *redis.IntCmd
	TxPipeline() redis.Pipeliner
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

// TokenStore redis token store
type TokenStore struct {
	cli clienter
}

func (s *TokenStore) checkError(result redis.Cmder) (bool, error) {
	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (s *TokenStore) getValue(result *redis.StringCmd) (string, error) {
	if ok, err := s.checkError(result); err != nil {
		return "", err
	} else if ok {
		return "", nil
	}
	return result.Val(), nil
}

func (s *TokenStore) Get(ctx context.Context, key string) (string, error) {
	return s.getValue(s.cli.Get(ctx, key))
}

func (s *TokenStore) Set(ctx context.Context, key string, value string, exp time.Duration) error {
	return s.cli.Set(ctx, key, value, exp).Err()
}

func (s *TokenStore) Update(ctx context.Context, key string, val string) error {
	ttl, err := s.cli.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	return s.cli.Set(ctx, key, val, ttl).Err()
}

func (s *TokenStore) Delete(ctx context.Context, key string) error {
	return s.cli.Del(ctx, key).Err()
}

func (s *TokenStore) GetTimeout(ctx context.Context, key string) (time.Duration, error) {
	return s.cli.TTL(ctx, key).Result()
}

func (s *TokenStore) UpdateTimeout(ctx context.Context, key string, exp time.Duration) error {
	return s.cli.Expire(ctx, key, exp).Err()
}

func (s *TokenStore) GetObj(ctx context.Context, key string, ret any) error {
	val, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return satoken.ErrObjectNotExist
	}
	return Unmarshal([]byte(val), ret)
}

func (s *TokenStore) SetObj(ctx context.Context, key string, val any, exp time.Duration) error {
	data, err := Marshal(val)
	if err != nil {
		return err
	}
	return s.Set(ctx, key, string(data), exp)
}

func (s *TokenStore) UpdateObj(ctx context.Context, key string, val any) error {
	data, err := Marshal(val)
	if err != nil {
		return err
	}
	return s.Update(ctx, key, string(data))
}

func (s *TokenStore) DeleteObj(ctx context.Context, key string) error {
	return s.cli.Del(ctx, key).Err()
}

func (s *TokenStore) GetObjTimeout(ctx context.Context, key string) (time.Duration, error) {
	return s.cli.TTL(ctx, key).Result()
}

func (s *TokenStore) UpdateObjTimeout(ctx context.Context, key string, exp time.Duration) error {
	return s.cli.Expire(ctx, key, exp).Err()
}
