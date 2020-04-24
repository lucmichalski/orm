package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntityDirtyQueueAll struct {
	orm.ORM `orm:"dirty=test"`
	ID      uint
	Name    string `orm:"length=100"`
}

type TestEntityDirtyQueueAge struct {
	orm.ORM
	ID   uint
	Name string `orm:"dirty=test"`
	Age  uint16 `orm:"dirty=test"`
}

func TestDirtyQueue(t *testing.T) {
	entityAll := &TestEntityDirtyQueueAll{Name: "Name"}
	entityAge := &TestEntityDirtyQueueAge{Name: "Name", Age: 18}
	registry := &orm.Registry{}
	registry.RegisterDirtyQueue("test", &orm.RedisDirtyQueueSender{PoolName: "default_queue"})
	engine := PrepareTables(t, registry, entityAll, entityAge)

	LoggerRedisQueue := &TestCacheLogger{}
	cache := engine.GetRedis("default_queue")
	cache.RegisterLogger(LoggerRedisQueue)

	engine.RegisterEntity(entityAll)
	err := entityAll.Flush()
	assert.Nil(t, err)
	engine.RegisterEntity(entityAge)
	err = entityAge.Flush()
	assert.Nil(t, err)

	assert.Len(t, LoggerRedisQueue.Requests, 2)
	assert.Equal(t, "SADD 1 values test", LoggerRedisQueue.Requests[0])

	receiver := orm.NewDirtyReceiver(engine, "test")

	entities := receiver.GetEntities()
	assert.Len(t, entities, 2)

	size, err := receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), size)
	has, err := receiver.Digest(2, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		assert.Len(t, data, 2)
		assert.Equal(t, uint64(1), data[0].ID)
		assert.Equal(t, uint64(1), data[1].ID)
		assert.True(t, data[0].Inserted)
		assert.True(t, data[1].Inserted)
		assert.False(t, data[0].Updated)
		assert.False(t, data[1].Updated)
		assert.False(t, data[0].Deleted)
		assert.False(t, data[1].Deleted)
		return nil, nil
	})
	assert.Nil(t, err)
	assert.True(t, has)
	has, err = receiver.Digest(2, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		return nil, nil
	})
	assert.Nil(t, err)
	assert.False(t, has)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)

	entityAll.Name = "Name 2"
	err = entityAll.Flush()
	assert.Nil(t, err)
	assert.Len(t, LoggerRedisQueue.Requests, 7)
	assert.Equal(t, "SADD 1 values test", LoggerRedisQueue.Requests[6])

	entityAge.Name = "Name 2"
	err = entityAll.Flush()
	assert.Nil(t, err)
	assert.Len(t, LoggerRedisQueue.Requests, 7)

	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), size)
	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		assert.Len(t, data, 1)
		assert.Equal(t, "TestEntityDirtyQueueAll", data[0].TableSchema.GetTableName())
		assert.Equal(t, uint64(1), data[0].ID)
		assert.False(t, data[0].Inserted)
		assert.True(t, data[0].Updated)
		assert.False(t, data[0].Deleted)
		return nil, nil
	})
	assert.Nil(t, err)
	assert.True(t, has)
	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		return nil, nil
	})
	assert.Nil(t, err)
	assert.False(t, has)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)

	entityAge.Age = 10
	err = entityAge.Flush()
	assert.Nil(t, err)
	assert.Len(t, LoggerRedisQueue.Requests, 12)
	assert.Equal(t, "SADD 1 values test", LoggerRedisQueue.Requests[11])

	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), size)
	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		assert.Len(t, data, 1)
		assert.Equal(t, "TestEntityDirtyQueueAge", data[0].TableSchema.GetTableName())
		assert.Equal(t, uint64(1), data[0].ID)
		assert.False(t, data[0].Inserted)
		assert.True(t, data[0].Updated)
		assert.False(t, data[0].Deleted)
		return nil, nil
	})
	assert.Nil(t, err)
	assert.True(t, has)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)

	entityAge.MarkToDelete()
	err = entityAge.Flush()
	assert.Nil(t, err)

	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), size)
	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		assert.Len(t, data, 1)
		assert.Equal(t, "TestEntityDirtyQueueAge", data[0].TableSchema.GetTableName())
		assert.Equal(t, uint64(1), data[0].ID)
		assert.False(t, data[0].Inserted)
		assert.False(t, data[0].Updated)
		assert.True(t, data[0].Deleted)
		return nil, nil
	})
	assert.Nil(t, err)
	assert.True(t, has)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)

	err = receiver.MarkDirty("tests.TestEntityDirtyQueueAge", 1, 2)
	assert.Nil(t, err)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), size)

	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		assert.Len(t, data, 2)
		assert.True(t, data[0].Updated)
		assert.True(t, data[1].Updated)
		assert.False(t, data[0].Inserted)
		assert.False(t, data[1].Inserted)
		assert.False(t, data[0].Deleted)
		assert.False(t, data[1].Deleted)
		return []interface{}{"a", "tests.TestEntityDirtyQueueAge:u:f", "c:d:f"}, fmt.Errorf("has invalid")
	})
	assert.True(t, has)
	assert.NotNil(t, err)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(3), size)

	has, err = receiver.Digest(100, func(data []orm.DirtyData) (invalid []interface{}, err error) {
		return nil, nil
	})
	assert.Nil(t, err)
	assert.True(t, has)
	size, err = receiver.Size()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), size)
}
