package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
	"testing"
)

type TestEntityFakeDelete struct {
	Orm        *orm.ORM `orm:"localCache"`
	Id         uint16
	Name       string
	FakeDelete bool
	IndexAll   *orm.CachedQuery `query:""`
	IndexName  *orm.CachedQuery `query:":Name = ?"`
}

func TestFakeDelete(t *testing.T) {
	registry := &orm.Registry{}
	engine := PrepareTables(t, registry, TestEntityFakeDelete{})

	entity := &TestEntityFakeDelete{}
	entity.Name = "one"
	err := engine.Flush(entity)
	assert.Nil(t, err)
	entity2 := &TestEntityFakeDelete{}
	entity2.Name = "two"
	err = engine.Flush(entity2)
	assert.Nil(t, err)

	var rows []*TestEntityFakeDelete
	total, err := engine.CachedSearch(&rows, "IndexAll", nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, total)
	total, err = engine.CachedSearch(&rows, "IndexName", nil, "two")
	assert.Nil(t, err)
	assert.Equal(t, 1, total)

	entity2.Orm.MarkToDelete()
	assert.True(t, entity2.FakeDelete)
	assert.True(t, engine.IsDirty(entity2))
	err = engine.Flush(entity2)
	assert.Nil(t, err)
	assert.False(t, engine.IsDirty(entity2))

	total, err = engine.SearchWithCount(orm.NewWhere("1"), nil, &rows)
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "one", rows[0].Name)

	err = engine.GetById(1, entity)
	assert.Nil(t, err)
	assert.False(t, entity.FakeDelete)

	err = engine.GetById(2, entity2)
	assert.Nil(t, err)
	assert.True(t, entity2.FakeDelete)

	total, err = engine.CachedSearch(&rows, "IndexAll", nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "one", rows[0].Name)
	total, err = engine.CachedSearch(&rows, "IndexName", nil, "two")
	assert.Nil(t, err)
	assert.Equal(t, 0, total)

	entity2.Orm.ForceMarkToDelete()
	err = engine.Flush(entity2)
	assert.Nil(t, err)
	has, err := engine.TryById(2, entity2)
	assert.Nil(t, err)
	assert.False(t, has)

}