package orm

import (
	"fmt"
	"reflect"
)

func flushInCache(engine *Engine, entities ...interface{}) error {

	invalidEntities := make([]interface{}, 0)
	validEntities := make([]interface{}, 0)
	redisValues := make(map[string][]interface{})

	for _, entity := range entities {

		value := reflect.ValueOf(entity)
		elem := value.Elem()
		orm := engine.initIfNeeded(value)
		elem.Field(0).Interface().(*ORM).elem = elem
		t := elem.Type()

		id := elem.Field(1).Uint()
		entityName := t.String()
		schema := orm.tableSchema
		cache := schema.GetRedisCacheContainer(engine)
		if cache == nil || id == 0 {
			invalidEntities = append(invalidEntities, entity)
		} else {
			isDirty, bind, err := isDirty(elem)
			if err != nil {
				return err
			}
			if !isDirty {
				continue
			}
			old := make(map[string]interface{}, len(orm.dBData))
			for k, v := range orm.dBData {
				old[k] = v
			}
			injectBind(elem, bind)
			entityCacheKey := schema.getCacheKey(id)
			entityCacheValue := buildRedisValue(entity, schema)
			if redisValues[cache.code] == nil {
				redisValues[cache.code] = make([]interface{}, 0)
			}
			redisValues[cache.code] = append(redisValues[cache.code], entityCacheKey, entityCacheValue)

			validEntities = append(validEntities, createDirtyQueueMember(entityName, id))
		}
	}
	if len(invalidEntities) > 0 {
		err := engine.Flush(invalidEntities...)
		if err != nil {
			return err
		}
	}
	if len(validEntities) > 0 {
		_, err := engine.getRedisForQueue("default").SAdd("dirty_queue", validEntities...)
		if err != nil {
			return err
		}
		for cacheCode, keys := range redisValues {
			err = engine.GetRedis(cacheCode).MSet(keys...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createDirtyQueueMember(entityName string, id uint64) interface{} {
	return fmt.Sprintf("%s:%d", entityName, id)
}
