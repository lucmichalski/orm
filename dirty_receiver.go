package orm

import (
	"fmt"
	"strconv"
	"strings"
)

type DirtyReceiver struct {
	engine    *Engine
	queueCode string
}

type DirtyData struct {
	TableSchema *tableSchema
	ID          uint64
	Inserted    bool
	Updated     bool
	Deleted     bool
}

func NewDirtyReceiver(engine *Engine, queueCode string) *DirtyReceiver {
	return &DirtyReceiver{engine: engine, queueCode: queueCode}
}

type DirtyHandler func([]DirtyData) (invalid []interface{}, err error)

func (r *DirtyReceiver) Size() (int64, error) {
	red := r.getRedis()
	return red.SCard(r.queueCode)
}

func (r *DirtyReceiver) GetEntities() []string {
	results := make([]string, 0)
	if r.engine.registry.entities != nil {
	Exit:
		for name, t := range r.engine.registry.entities {
			schema := getTableSchema(r.engine.registry, t)
			for _, tags := range schema.tags {
				queues, has := tags["dirty"]
				if has {
					queueNames := strings.Split(queues, ",")
					for _, queueName := range queueNames {
						if r.queueCode == queueName {
							results = append(results, name)
							continue Exit
						}
					}
				}
			}
		}
	}
	return results
}

func (r *DirtyReceiver) Digest(max int, handler DirtyHandler) (has bool, err error) {
	cache := r.getRedis()
	values, err := cache.SPopN(r.queueCode, int64(max))
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return false, nil
	}
	results := make([]DirtyData, 0, len(values))
	for _, v := range values {
		val := strings.Split(v, ":")
		if len(val) != 3 {
			continue
		}
		t, has := r.engine.registry.entities[val[0]]
		if !has {
			continue
		}
		tableSchema := getTableSchema(r.engine.registry, t)
		id, err := strconv.ParseUint(val[2], 10, 64)
		if err != nil {
			continue
		}
		data := DirtyData{
			TableSchema: tableSchema,
			ID:          id,
			Inserted:    val[1] == "i",
			Updated:     val[1] == "u",
			Deleted:     val[1] == "d",
		}
		results = append(results, data)
	}
	invalid, err := handler(results)
	if err != nil {
		if invalid != nil {
			_, _ = cache.SAdd(r.queueCode, invalid...)
		}
		return true, err
	}
	return true, nil
}

func (r *DirtyReceiver) MarkDirty(entityName string, ids ...uint64) error {
	cache := r.getRedis()
	data := make([]interface{}, len(ids))
	for index, id := range ids {
		data[index] = fmt.Sprintf("%s:u:%d", entityName, id)
	}
	_, err := cache.SAdd(r.queueCode, data...)
	return err
}

func (r *DirtyReceiver) getRedis() *RedisCache {
	queue, _ := r.engine.registry.dirtyQueues[r.queueCode].(*RedisDirtyQueueSender)
	redis := r.engine.GetRedis(queue.PoolName)
	return redis
}
