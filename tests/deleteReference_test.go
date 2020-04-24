package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntityDeleteReference struct {
	orm.ORM `orm:"localCache"`
	ID      uint
}

type TestEntityDeleteReferenceRefRestrict struct {
	orm.ORM      `orm:"localCache"`
	ID           uint
	ReferenceOne *TestEntityDeleteReference
}

type TestEntityDeleteReferenceRefCascade struct {
	orm.ORM           `orm:"localCache"`
	ID                uint
	ReferenceOne      *TestEntityDeleteReference `orm:"cascade"`
	IndexReferenceOne *orm.CachedQuery           `query:":ReferenceOne = ?"`
}

func TestDeleteReference(t *testing.T) {
	engine := PrepareTables(t, &orm.Registry{}, TestEntityDeleteReference{},
		TestEntityDeleteReferenceRefRestrict{}, TestEntityDeleteReferenceRefCascade{})
	entity1 := &TestEntityDeleteReference{}
	engine.RegisterEntity(entity1)
	err := entity1.Flush()
	assert.Nil(t, err)
	entity2 := &TestEntityDeleteReference{}
	engine.RegisterEntity(entity2)
	err = entity2.Flush()
	assert.Nil(t, err)

	entityRestrict := &TestEntityDeleteReferenceRefRestrict{}
	engine.RegisterEntity(entityRestrict)
	entityRestrict.ReferenceOne = &TestEntityDeleteReference{ID: 1}
	err = entityRestrict.Flush()
	assert.Nil(t, err)

	entity1.MarkToDelete()
	err = entity1.Flush()
	assert.NotNil(t, err)
	assert.IsType(t, &orm.ForeignKeyError{}, err)
	assert.Equal(t, "test:TestEntityDeleteReferenceRefRestrict:ReferenceOne", err.(*orm.ForeignKeyError).Constraint)

	entityCascade := &TestEntityDeleteReferenceRefCascade{}
	entityCascade2 := &TestEntityDeleteReferenceRefCascade{}
	engine.RegisterEntity(entityCascade)
	engine.RegisterEntity(entityCascade2)
	entityCascade.ReferenceOne = &TestEntityDeleteReference{ID: 2}
	entityCascade2.ReferenceOne = &TestEntityDeleteReference{ID: 2}
	err = entityCascade.Flush()
	assert.Nil(t, err)
	err = entityCascade2.Flush()
	assert.Nil(t, err)

	var rows []*TestEntityDeleteReferenceRefCascade
	total, err := engine.CachedSearch(&rows, "IndexReferenceOne", nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, total)

	entity2.MarkToDelete()
	err = entity2.Flush()
	assert.Nil(t, err)

	total, err = engine.CachedSearch(&rows, "IndexReferenceOne", nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 0, total)
}
