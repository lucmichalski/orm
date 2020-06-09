package orm

import (
	"encoding/json"
	"fmt"
)

func loadByID(engine *Engine, id uint64, entity Entity, useCache bool, references ...string) (found bool) {
	orm := initIfNeeded(engine, entity)
	schema := orm.tableSchema
	var cacheKey string
	localCache, hasLocalCache := schema.GetLocalCache(engine)

	if hasLocalCache && useCache {
		cacheKey = schema.getCacheKey(id)
		e, has := localCache.Get(cacheKey)
		if has {
			if e == "nil" {
				return false
			}
			fillFromDBRow(id, engine, e.([]string), entity)
			if len(references) > 0 {
				warmUpReferences(engine, schema, orm.attributes.elem, references, false)
			}
			return true
		}
	}
	redisCache, hasRedis := schema.GetRedisCache(engine)
	if hasRedis && useCache {
		cacheKey = schema.getCacheKey(id)
		row, has := redisCache.Get(cacheKey)
		if has {
			if row == "nil" {
				return false
			}
			var decoded []string
			_ = json.Unmarshal([]byte(row), &decoded)
			fillFromDBRow(id, engine, decoded, entity)
			if len(references) > 0 {
				warmUpReferences(engine, schema, orm.attributes.elem, references, false)
			}
			return true
		}
	}
	found = searchRow(false, engine, NewWhere("`ID` = ?", id), entity, nil)
	if !found {
		if localCache != nil {
			localCache.Set(cacheKey, "nil")
		}
		if redisCache != nil {
			redisCache.Set(cacheKey, "nil", 60)
		}
		return false
	}
	if localCache != nil && useCache {
		localCache.Set(cacheKey, buildLocalCacheValue(entity))
	}
	if redisCache != nil && useCache {
		redisCache.Set(cacheKey, buildRedisValue(entity), 0)
	}
	if len(references) > 0 {
		warmUpReferences(engine, schema, orm.attributes.elem, references, false)
	}
	return true
}

func buildRedisValue(entity Entity) string {
	encoded, _ := json.Marshal(buildLocalCacheValue(entity))
	return string(encoded)
}

func buildLocalCacheValue(entity Entity) []string {
	bind := entity.getORM().dBData
	columns := entity.getORM().tableSchema.columnNames
	length := len(columns)
	value := make([]string, length-1)
	j := 0
	for i := 1; i < length; i++ { //skip id
		v := bind[columns[i]]
		if v == nil {
			v = ""
		}
		value[j] = fmt.Sprintf("%s", v)
		j++
	}
	return value
}
