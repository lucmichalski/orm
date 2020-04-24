package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bsm/redislock"
)

type EntityNotRegisteredError struct {
	Name string
}

func (e EntityNotRegisteredError) Error() string {
	return fmt.Sprintf("entity '%s' is not registered", strings.Trim(e.Name, "*[]"))
}

type ValidatedRegistry interface {
	CreateEngine() *Engine
	GetTableSchema(entityName string) TableSchema
	GetTableSchemaForEntity(entity Entity) TableSchema
	GetDirtyQueueCodes() []string
	GetLogQueueCodes() []string
	GetLazyQueueCodes() []string
}

type validatedRegistry struct {
	tableSchemas         map[reflect.Type]*tableSchema
	entities             map[string]reflect.Type
	sqlClients           map[string]*DBConfig
	dirtyQueues          map[string]DirtyQueueSender
	logQueues            map[string]QueueSenderReceiver
	lazyQueues           map[string]QueueSenderReceiver
	localCacheContainers map[string]*LocalCacheConfig
	redisServers         map[string]*RedisCacheConfig
	lockServers          map[string]string
	enums                map[string]reflect.Value
}

func (r *validatedRegistry) CreateEngine() *Engine {
	e := &Engine{registry: r}
	e.dbs = make(map[string]*DB)
	e.trackedEntities = make([]reflect.Value, 0)
	if e.registry.sqlClients != nil {
		for key, val := range e.registry.sqlClients {
			e.dbs[key] = &DB{engine: e, code: val.code, databaseName: val.databaseName, db: &sqlDBStandard{db: val.db}}
		}
	}
	e.localCache = make(map[string]*LocalCache)
	if e.registry.localCacheContainers != nil {
		for key, val := range e.registry.localCacheContainers {
			e.localCache[key] = &LocalCache{engine: e, code: val.code, lru: val.lru, ttl: val.ttl}
		}
	}
	e.redis = make(map[string]*RedisCache)
	if e.registry.redisServers != nil {
		for key, val := range e.registry.redisServers {
			e.redis[key] = &RedisCache{engine: e, code: val.code, client: val.client}
		}
	}
	e.locks = make(map[string]*Locker)
	if e.registry.lockServers != nil {
		for key, val := range e.registry.lockServers {
			locker := redislock.New(e.registry.redisServers[val].client)
			e.locks[key] = &Locker{locker: locker}
		}
	}
	return e
}

func (r *validatedRegistry) GetTableSchema(entityName string) TableSchema {
	t, has := r.entities[entityName]
	if !has {
		return nil
	}
	return getTableSchema(r, t)
}

func (r *validatedRegistry) GetTableSchemaForEntity(entity Entity) TableSchema {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableSchema := getTableSchema(r, t)
	if tableSchema == nil {
		panic(EntityNotRegisteredError{Name: t.String()})
	}
	return tableSchema
}

func (r *validatedRegistry) GetDirtyQueueCodes() []string {
	codes := make([]string, len(r.dirtyQueues))
	i := 0
	for code := range r.dirtyQueues {
		codes[i] = code
		i++
	}
	return codes
}

func (r *validatedRegistry) GetLogQueueCodes() []string {
	codes := make([]string, len(r.logQueues))
	i := 0
	for code := range r.logQueues {
		codes[i] = code
		i++
	}
	return codes
}

func (r *validatedRegistry) GetLazyQueueCodes() []string {
	codes := make([]string, len(r.lazyQueues))
	i := 0
	for code := range r.lazyQueues {
		codes[i] = code
		i++
	}
	return codes
}
