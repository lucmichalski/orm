package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEntityDeleteReference struct {
	ORM `orm:"localCache"`
	ID  uint
}

type testEntityDeleteReferenceRefRestrict struct {
	ORM          `orm:"localCache"`
	ID           uint
	ReferenceOne *testEntityDeleteReference
}

type testEntityDeleteReferenceRefCascade struct {
	ORM               `orm:"localCache"`
	ID                uint
	ReferenceOne      *testEntityDeleteReference `orm:"cascade;index=TestIndex"`
	IndexReferenceOne *CachedQuery               `query:":ReferenceOne = ?"`
}

func TestDeleteReference(t *testing.T) {
	engine := PrepareTables(t, &Registry{}, testEntityDeleteReference{},
		testEntityDeleteReferenceRefRestrict{}, testEntityDeleteReferenceRefCascade{})
	entity1 := &testEntityDeleteReference{}
	engine.Track(entity1)
	engine.Flush()
	entity2 := &testEntityDeleteReference{}
	engine.Track(entity2)
	engine.Flush()

	entityRestrict := &testEntityDeleteReferenceRefRestrict{}
	engine.Track(entityRestrict)
	entityRestrict.ReferenceOne = &testEntityDeleteReference{ID: 1}
	engine.Flush()

	engine.MarkToDelete(entity1)
	err := engine.FlushWithCheck()
	assert.NotNil(t, err)
	assert.Equal(t, "test:testEntityDeleteReferenceRefRestrict:ReferenceOne", err.(*ForeignKeyError).Constraint)
	engine.ClearTrackedEntities()

	entityCascade := &testEntityDeleteReferenceRefCascade{}
	entityCascade2 := &testEntityDeleteReferenceRefCascade{}
	engine.Track(entityCascade)
	engine.Track(entityCascade2)
	entityCascade.ReferenceOne = &testEntityDeleteReference{ID: 2}
	entityCascade2.ReferenceOne = &testEntityDeleteReference{ID: 2}
	engine.Flush()
	var rows []*testEntityDeleteReferenceRefCascade
	total := engine.CachedSearch(&rows, "IndexReferenceOne", nil, 2)
	assert.Equal(t, 2, total)

	engine.MarkToDelete(entity2)
	engine.Flush()

	total = engine.CachedSearch(&rows, "IndexReferenceOne", nil, 2)
	assert.Equal(t, 0, total)
}
