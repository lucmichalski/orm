package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntityFakeDelete struct {
	orm.ORM    `orm:"localCache"`
	ID         uint16
	Name       string
	FakeDelete bool
	Uint       uint
	IndexAll   *orm.CachedQuery `query:""`
	IndexName  *orm.CachedQuery `query:":Name = ?"`
}

func TestFakeDelete(t *testing.T) {
	registry := &orm.Registry{}
	engine := PrepareTables(t, registry, TestEntityFakeDelete{})

	entity := &TestEntityFakeDelete{}
	engine.RegisterEntity(entity)
	entity.Name = "one"
	err := entity.Flush()
	assert.Nil(t, err)
	entity2 := &TestEntityFakeDelete{}
	engine.RegisterEntity(entity2)
	entity2.Name = "two"
	err = entity2.Flush()
	assert.Nil(t, err)

	var rows []*TestEntityFakeDelete
	total, err := engine.CachedSearch(&rows, "IndexAll", nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, total)
	total, err = engine.CachedSearch(&rows, "IndexName", nil, "two")
	assert.Nil(t, err)
	assert.Equal(t, 1, total)

	entity2.MarkToDelete()
	assert.True(t, entity2.FakeDelete)
	assert.True(t, entity2.IsDirty())
	err = entity2.Flush()
	assert.Nil(t, err)
	assert.False(t, entity2.IsDirty())

	total, err = engine.SearchWithCount(orm.NewWhere("1"), nil, &rows)
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "one", rows[0].Name)

	has, err := engine.LoadByID(1, entity)
	assert.True(t, has)
	assert.Nil(t, err)
	assert.False(t, entity.FakeDelete)

	has, err = engine.LoadByID(2, entity2)
	assert.True(t, has)
	assert.Nil(t, err)
	assert.True(t, entity2.FakeDelete)

	total, err = engine.CachedSearch(&rows, "IndexAll", nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "one", rows[0].Name)
	total, err = engine.CachedSearch(&rows, "IndexName", nil, "two")
	assert.Nil(t, err)
	assert.Equal(t, 0, total)

	entity2.ForceMarkToDelete()
	err = entity2.Flush()
	assert.Nil(t, err)
	has, err = engine.LoadByID(2, entity2)
	assert.Nil(t, err)
	assert.False(t, has)

	entity.MarkToDelete()
	err = entity.Flush()
	assert.Nil(t, err)

	has, err = engine.SearchOne(orm.NewWhere("1"), entity)
	assert.False(t, has)
	assert.Nil(t, err)
}
