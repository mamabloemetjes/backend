package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"mamabloemetjes_server/config"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
	redisCtx    = context.Background()
)

// CacheService provides Redis caching functionality with connection pooling and retry logic
type CacheService struct {
	logger *gecho.Logger
	config *structs.Config
	client *redis.Client
}

func NewCacheService(logger *gecho.Logger, cfg *structs.Config) *CacheService {
	return &CacheService{
		logger: logger,
		config: cfg,
		client: getRedisClient(),
	}
}

// GetRedisClient returns a singleton Redis client with proper connection pooling
func getRedisClient() *redis.Client {
	redisOnce.Do(func() {
		cfg := config.GetConfig()
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Cache.Address,
			Username: cfg.Cache.Username,
			Password: cfg.Cache.Password,
			DB:       cfg.Cache.DB,

			// Connection pool settings
			PoolSize:        cfg.Cache.PoolSize,
			MinIdleConns:    cfg.Cache.MinIdleConns,
			MaxIdleConns:    cfg.Cache.MaxIdleConns,
			PoolTimeout:     cfg.Cache.PoolTimeout,
			ConnMaxIdleTime: cfg.Cache.IdleTimeout,

			// Timeouts
			DialTimeout:  cfg.Cache.DialTimeout,
			ReadTimeout:  cfg.Cache.ReadTimeout,
			WriteTimeout: cfg.Cache.WriteTimeout,

			// Retry settings
			MaxRetries:      cfg.Cache.MaxRetries,
			MinRetryBackoff: cfg.Cache.MinRetryBackoff,
			MaxRetryBackoff: cfg.Cache.MaxRetryBackoff,
		})
	})
	return redisClient
}

// CloseRedisConnection closes the Redis connection pool
func (cs *CacheService) Close() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// withRetry executes a Redis operation with exponential backoff retry logic
func (cs *CacheService) withRetry(operation func() error, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Only retry on network/connection errors, not on logical errors like key not found
		if !isRetryableError(err) {
			return err
		}

		maxBackoff := 2000 // max 2000ms = 2s
		base := 100        // 100ms base

		backoff := base * (1 << attempt) // exponential
		backoff = min(backoff, maxBackoff)

		// add jitter ±50%
		jitterBytes := make([]byte, 4)
		_, err = rand.Read(jitterBytes)
		if err != nil {
			// fallback to no jitter if random fails
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			continue
		}
		jitter := int(uint32(jitterBytes[0])<<24 | uint32(jitterBytes[1])<<16 | uint32(jitterBytes[2])<<8 | uint32(jitterBytes[3]))
		// No need to handle negative values; uint32 avoids sign extension
		// jitter is always non-negative

		jitter = jitter % (backoff/2 + 1)
		backoffWithJitter := backoff/2 + jitter

		time.Sleep(time.Duration(backoffWithJitter) * time.Millisecond)
	}

	return fmt.Errorf("redis operation failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry on nil results (key not found)
	if err == redis.Nil {
		return false
	}

	// Retry on network/connection errors
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"broken pipe",
		"no such host",
		"network is unreachable",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// Set sets a key with TTL and automatic retry logic
func (cs *CacheService) Set(key string, value any, ttl time.Duration) error {
	return cs.withRetry(func() error {
		return cs.client.Set(redisCtx, key, value, ttl).Err()
	}, 3)
}

// Get retrieves a key with automatic retry logic
func (cs *CacheService) Get(key string) (string, error) {
	var result string
	var resultErr error

	err := cs.withRetry(func() error {
		val, err := cs.client.Get(redisCtx, key).Result()
		if err == redis.Nil {
			result = ""
			resultErr = nil
			return nil // Don't retry on key not found
		}
		if err != nil {
			return err
		}
		result = val
		resultErr = nil
		return nil
	}, 3)

	if err != nil {
		return "", err
	}

	return result, resultErr
}

// Delete removes a key with automatic retry logic
func (cs *CacheService) Delete(key string) error {
	return cs.withRetry(func() error {
		return cs.client.Del(redisCtx, key).Err()
	}, 3)
}

// Exists checks if a key exists with automatic retry logic
func (cs *CacheService) Exists(key string) (bool, error) {
	var result bool

	err := cs.withRetry(func() error {
		count, err := cs.client.Exists(redisCtx, key).Result()
		if err != nil {
			return err
		}
		result = count > 0
		return nil
	}, 3)

	return result, err
}

// BlacklistToken adds a token's jti to the blacklist with expiration and retry logic
func (cs *CacheService) BlacklistToken(jti uuid.UUID, exp time.Time) error {
	ttl := cs.config.Auth.BlacklistCacheTTL
	if exp.After(time.Now()) {
		ttl = time.Until(exp)
	}

	key := fmt.Sprintf("blacklist:%s", jti)
	return cs.Set(key, "true", ttl)
}

// IsTokenBlacklisted checks if a JTI exists in Redis with retry logic
func (cs *CacheService) IsTokenBlacklisted(jti uuid.UUID) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", jti.String())
	val, err := cs.Get(key)
	if err != nil {
		return false, err
	}

	return val == "true", nil
}

// Get UserFromCache retrieves a user object from cache using userID
func (cs *CacheService) GetUserFromCache(userID uuid.UUID) (*tables.User, error) {
	key := fmt.Sprintf("user:%s", userID.String())
	val, err := cs.Get(key)
	if err != nil {
		return nil, err
	}

	if val == "" {
		return nil, nil // not found in cache
	}

	user := &tables.User{}
	err = json.Unmarshal([]byte(val), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// SetUserInCache stores a user object in cache with TTL
func (cs *CacheService) SetUserInCache(user *tables.User) error {
	if user == nil {
		// Nothing to cache
		return nil
	}
	key := fmt.Sprintf("user:%s", user.Id.String())
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return cs.Set(key, data, cs.config.Auth.CacheUserTTL)
}

// DeleteUserFromCache removes a user object from cache
func (cs *CacheService) DeleteUserFromCache(userID uuid.UUID) error {
	key := fmt.Sprintf("user:%s", userID.String())
	return cs.Delete(key)
}

// SetRateLimit sets a rate limit counter for an IP/endpoint combination
func (cs *CacheService) SetRateLimit(ip, endpoint string, count int, ttl time.Duration) error {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)
	return cs.Set(key, count, ttl)
}

// GetRateLimit retrieves the current rate limit count for an IP/endpoint
func (cs *CacheService) GetRateLimit(ip, endpoint string) (int, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)
	val, err := cs.Get(key)
	if err != nil {
		return 0, err
	}

	if val == "" {
		return 0, nil
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit value: %w", err)
	}

	return count, nil
}

// IncrementRateLimit atomically increments a rate limit counter
func (cs *CacheService) IncrementRateLimit(ip, endpoint string, ttl time.Duration) (int, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)

	var result int64
	err := cs.withRetry(func() error {
		val, err := cs.client.Incr(redisCtx, key).Result()
		if err != nil {
			return err
		}
		result = val

		// Set expiration only on first increment
		if val == 1 {
			return cs.client.Expire(redisCtx, key, ttl).Err()
		}

		return nil
	}, 3)

	return int(result), err
}

// Ping tests the Redis connection
func (cs *CacheService) Ping() error {
	return cs.withRetry(func() error {
		return cs.client.Ping(redisCtx).Err()
	}, 3)
}

// GetConnectionStats returns Redis connection pool statistics
func (cs *CacheService) GetConnectionStats() map[string]any {
	stats := cs.client.PoolStats()

	return map[string]any{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}

// GetRateLimitStatus returns current rate limit information for debugging
func (cs *CacheService) GetRateLimitStatus(ip, endpoint string) (map[string]any, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)

	var result map[string]any

	err := cs.withRetry(func() error {
		// Get current count
		val, err := cs.client.Get(redisCtx, key).Result()
		if err == redis.Nil {
			result = map[string]any{
				"count": 0,
				"ttl":   0,
			}
			return nil
		}
		if err != nil {
			return err
		}

		// Get TTL
		ttl, err := cs.client.TTL(redisCtx, key).Result()
		if err != nil {
			return err
		}

		// Parse count
		count, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid rate limit value: %w", err)
		}

		result = map[string]any{
			"count": count,
			"ttl":   int(ttl.Seconds()),
		}
		return nil
	}, 3)

	return result, err
}

// ============================================================================
// Product Caching Methods
// ============================================================================

// GetActiveProductsList retrieves cached active products list
func (cs *CacheService) GetActiveProductsList(page, pageSize int, includeImages bool) ([]tables.Product, error) {
	key := fmt.Sprintf("products:active:page:%d:size:%d:images:%v", page, pageSize, includeImages)

	products, err := getJSON[[]tables.Product](cs, key)
	if err != nil {
		cs.logger.Warn("Failed to get active products from cache", "error", err, "key", key)
		return nil, err
	}

	if products == nil {
		return nil, nil
	}

	return *products, nil
}

// SetActiveProductsList caches active products list
func (cs *CacheService) SetActiveProductsList(page, pageSize int, includeImages bool, products []tables.Product) error {
	key := fmt.Sprintf("products:active:page:%d:size:%d:images:%v", page, pageSize, includeImages)
	ttl := cs.getProductListTTL()

	return setJSON(cs, key, products, ttl)
}

// GetProductBySKU retrieves a cached product by SKU
func (cs *CacheService) GetProductBySKU(sku string) (*tables.Product, error) {
	key := fmt.Sprintf("product:sku:%s", sku)

	product, err := getJSON[tables.Product](cs, key)
	if err != nil {
		cs.logger.Warn("Failed to get product from cache", "error", err, "sku", sku)
		return nil, err
	}

	if product == nil {
		return nil, nil
	}

	return product, nil
}

// SetProductBySKU caches a product by SKU
func (cs *CacheService) SetProductBySKU(product *tables.Product) error {
	key := fmt.Sprintf("product:sku:%s", product.SKU)
	ttl := cs.getProductListTTL()

	cs.logger.Debug("Caching product by SKU", "sku", product.SKU, "ttl", ttl)

	return setJSON(cs, key, product, ttl)
}

// GetProductByID retrieves a cached product by ID
func (cs *CacheService) GetProductByID(id string, includeImages bool) (*tables.Product, error) {
	key := fmt.Sprintf("product:id:%s:images:%v", id, includeImages)

	product, err := getJSON[tables.Product](cs, key)
	if err != nil {
		cs.logger.Warn("Failed to get product from cache", "error", err, "id", id)
		return nil, err
	}

	if product == nil {
		return nil, nil
	}

	return product, nil
}

// SetProductByID caches a product by ID
func (cs *CacheService) SetProductByID(product *tables.Product, includeImages bool) error {
	key := fmt.Sprintf("product:id:%s:images:%v", product.ID.String(), includeImages)
	ttl := cs.getProductListTTL()

	return setJSON(cs, key, product, ttl)
}

// GetProductCount retrieves cached product count
func (cs *CacheService) GetProductCount(filterKey string) (*int, error) {
	key := fmt.Sprintf("products:count:%s", filterKey)

	count, err := getJSON[int](cs, key)
	if err != nil {
		cs.logger.Warn("Failed to get product count from cache", "error", err, "key", key)
		return nil, err
	}

	if count == nil {
		return nil, nil
	}

	return count, nil
}

// SetProductCount caches product count
func (cs *CacheService) SetProductCount(filterKey string, count int) error {
	key := fmt.Sprintf("products:count:%s", filterKey)
	ttl := cs.getProductCountTTL()

	return setJSON(cs, key, count, ttl)
}

// ============================================================================
// Cache Invalidation Methods
// ============================================================================

// InvalidateUserCache removes a user from cache
func (cs *CacheService) InvalidateUserCache(userID uuid.UUID) error {
	key := fmt.Sprintf("user:%s", userID.String())
	return cs.Delete(key)
}

// InvalidateProductCaches removes all product-related caches
// This should be called when any product is created, updated, or deleted
func (cs *CacheService) InvalidateProductCaches(productID uuid.UUID) error {
	cs.logger.Info("Invalidating product caches", "product_id", productID)

	// First, get the product to find its SKU (if it exists in cache)
	// This is best-effort - if it fails, we still delete pattern-based caches
	productKey := fmt.Sprintf("product:id:%s:*", productID.String())
	if err := cs.DeletePattern(productKey); err != nil {
		cs.logger.Warn("Failed to delete product ID cache", "product_id", productID, "error", err)
	}

	// Delete all active product lists (they may contain this product)
	if err := cs.DeletePattern("products:active:*"); err != nil {
		cs.logger.Warn("Failed to delete active products cache", "error", err)
		return err
	}

	// Delete all product counts
	if err := cs.DeletePattern("products:count:*"); err != nil {
		cs.logger.Warn("Failed to delete product counts cache", "error", err)
		return err
	}

	cs.logger.Info("Product caches invalidated successfully", "product_id", productID)
	return nil
}

// InvalidateProductCacheBySKU removes a specific product cache by SKU
func (cs *CacheService) InvalidateProductCacheBySKU(sku string) error {
	key := fmt.Sprintf("product:sku:%s", sku)
	return cs.Delete(key)
}

// InvalidateAllProductCaches removes ALL product-related caches
// Use with caution - this is a heavy operation
func (cs *CacheService) InvalidateAllProductCaches() error {
	cs.logger.Warn("Invalidating ALL product caches")

	patterns := []string{
		"product:*",
		"products:*",
	}

	for _, pattern := range patterns {
		if err := cs.DeletePattern(pattern); err != nil {
			cs.logger.Error("Failed to delete cache pattern", "pattern", pattern, "error", err)
			return err
		}
	}

	cs.logger.Info("All product caches invalidated successfully")
	return nil
}

// DeletePattern removes all keys matching a pattern using SCAN
func (cs *CacheService) DeletePattern(pattern string) error {
	return cs.withRetry(func() error {
		var cursor uint64
		deletedCount := 0

		for {
			keys, nextCursor, err := cs.client.Scan(redisCtx, cursor, pattern, 100).Result()
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if len(keys) > 0 {
				if err := cs.client.Del(redisCtx, keys...).Err(); err != nil {
					return fmt.Errorf("delete failed: %w", err)
				}
				deletedCount += len(keys)
			}

			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}

		return nil
	}, 3)
}

func (cs *CacheService) ClearAll() error {
	return cs.withRetry(func() error {
		return cs.client.FlushDB(redisCtx).Err()
	}, 3)
}

// ============================================================================
// Helper Methods
// ============================================================================

// getProductListTTL returns the TTL for product lists from config
func (cs *CacheService) getProductListTTL() time.Duration {
	if cs.config.Cache.ProductListTTL > 0 {
		return cs.config.Cache.ProductListTTL
	}
	return 5 * time.Minute // fallback default
}

// getProductCountTTL returns the TTL for product counts from config
func (cs *CacheService) getProductCountTTL() time.Duration {
	if cs.config.Cache.ProductCountTTL > 0 {
		return cs.config.Cache.ProductCountTTL
	}
	return 10 * time.Minute // fallback default
}

func setJSON[T any](cs *CacheService, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cs.Set(key, data, ttl)
}

func getJSON[T any](cs *CacheService, key string) (*T, error) {
	val, err := cs.Get(key)
	if err != nil {
		return nil, err
	}

	if val == "" {
		return nil, nil // not found in cache
	}

	var result T
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
