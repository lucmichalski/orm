package orm

import (
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/golang/groupcache/lru"
	"math"
	"reflect"
	"time"
)

var mySqlClients = make(map[string]*DB)
var localCacheContainers = make(map[string]*LocalCache)
var redisServers = make(map[string]*RedisCache)
var entities = make(map[string]reflect.Type)
var dirtyQueuesCodes = make(map[string]string)
var queueRedisName = "default"

func RegisterEntity(entity ...interface{}) {
	for _, e := range entity {
		t := reflect.TypeOf(e)
		entities[t.String()] = t
	}
}

func Init(entity ...interface{}) {
	for _, e := range entity {
		value := reflect.Indirect(reflect.ValueOf(e))
		_, err := initIfNeeded(value, e)
		if err != nil {
			panic(err.Error())
		}
	}
}

func initIfNeeded(value reflect.Value, entity interface{}) (*ORM, error) {
	orm := value.Field(0).Interface().(*ORM)
	if orm == nil {
		orm = &ORM{dBData: make(map[string]interface{}), e: entity}
		value.Field(0).Set(reflect.ValueOf(orm))
		tableSchema := getTableSchema(value.Type())
		for i := 2; i < value.NumField(); i++ {
			field := value.Field(i)
			isOne := field.Type().String() == "*orm.ReferenceOne"
			isTwo := !isOne && field.Type().String() == "*orm.ReferenceMany"
			if isOne || isTwo {
				f := value.Type().Field(i)
				reference, has := tableSchema.tags[f.Name]["ref"]
				if !has {
					return nil, fmt.Errorf("missing ref tag")
				}
				if isOne {
					def := ReferenceOne{t: getEntityType(reference)}
					value.FieldByName(f.Name).Set(reflect.ValueOf(&def))
				} else {
					def := ReferenceMany{t: getEntityType(reference)}
					value.FieldByName(f.Name).Set(reflect.ValueOf(&def))
				}
			}
		}
	}
	orm.e = entity
	return orm, nil
}

func RegisterMySqlPool(dataSourceName string, code ...string) error {
	sqlDB, _ := sql.Open("mysql", dataSourceName)
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	db := &DB{code: dbCode, db: sqlDB}
	mySqlClients[dbCode] = db

	var variable string
	var maxConnections float64
	var maxTime float64
	err := db.QueryRow("SHOW VARIABLES LIKE 'max_connections'").Scan(&variable, &maxConnections)
	if err != nil {
		return nil
	}
	err = db.QueryRow("SHOW VARIABLES LIKE 'interactive_timeout'").Scan(&variable, &maxTime)
	if err != nil {
		return nil
	}
	maxConnectionsOrm := math.Ceil(maxConnections * 0.9)
	maxIdleConnections := math.Ceil(maxConnections * 0.05)
	maxConnectionsTime := math.Ceil(maxTime * 0.7)
	if maxIdleConnections < 10 {
		maxIdleConnections = maxConnectionsOrm
	}

	db.db.SetMaxOpenConns(int(maxConnectionsOrm))
	db.db.SetMaxIdleConns(int(maxIdleConnections))
	db.db.SetConnMaxLifetime(time.Duration(int(maxConnectionsTime)) * time.Second)

	return nil
}

func UnregisterMySqlPools() {
	mySqlClients = make(map[string]*DB)
}

func RegisterLocalCache(size int, code ...string) {
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	localCacheContainers[dbCode] = &LocalCache{code: dbCode, lru: lru.New(size)}
}

func EnableContextCache(size int, ttl int64) {
	localCacheContainers["_context_cache"] = &LocalCache{code: "_context_cache", lru: lru.New(size), ttl: ttl, created: time.Now().Unix()}
}

func DisableContextCache() {
	delete(localCacheContainers, "_context_cache")
}

func RegisterRedis(address string, db int, code ...string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   db,
	})
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	redisCache := &RedisCache{code: dbCode, client: client}
	redisServers[dbCode] = redisCache
	return redisCache
}

func RegisterDirtyQueue(code string, redisCode string) {
	dirtyQueuesCodes[code] = redisCode
}

func SetRedisForQueue(redisCode string) {
	queueRedisName = redisCode
}

func GetMysql(code ...string) *DB {
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	db, has := mySqlClients[dbCode]
	if !has {
		panic(fmt.Errorf("unregistered database code: %s", dbCode))
	}
	return db
}

func GetLocalCache(code ...string) *LocalCache {
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	cache, has := localCacheContainers[dbCode]
	if has == true {
		return cache
	}
	panic(fmt.Errorf("unregistered local cache: %s", dbCode))
}

func GetRedis(code ...string) *RedisCache {
	dbCode := "default"
	if len(code) > 0 {
		dbCode = code[0]
	}
	client, has := redisServers[dbCode]
	if !has {
		panic(fmt.Errorf("unregistered redis code: %s", dbCode))
	}
	return client
}

func getEntityType(name string) reflect.Type {
	t, has := entities[name]
	if !has {
		panic(fmt.Errorf("unregistered entity %s", name))
	}
	return t
}

func GetContextCache() *LocalCache {
	contextCache, has := localCacheContainers["_context_cache"]
	if !has {
		return nil
	}
	if time.Now().Unix()-contextCache.created+contextCache.ttl <= 0 {
		contextCache.lru.Clear()
		contextCache.created = time.Now().Unix()
	}
	return contextCache
}

func (db *DB) AddDatabaseLogger(logger DatabaseLogger) {
	for _, db := range mySqlClients {
		db.AddLogger(logger)
	}
}

func (db *DB) AddRedisLogger(logger CacheLogger) {
	for _, red := range redisServers {
		red.AddLogger(logger)
	}
}
