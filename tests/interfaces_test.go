package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type TestEntityInterfaces struct {
	orm.ORM
	ID           uint
	Uint         uint
	Name         string
	ReferenceOne *TestEntityInterfacesRef
	Calculated   int `orm:"ignore"`
}

type TestEntityInterfacesRef struct {
	orm.ORM
	ID uint
}

func (e *TestEntityInterfaces) SetDefaults() {
	e.Uint = 3
	e.Name = "hello"
	e.ReferenceOne.ID = 1
}

func (e *TestEntityInterfaces) Validate() error {
	if e.Uint < 5 {
		return fmt.Errorf("uint too low")
	}
	return nil
}

func (e *TestEntityInterfaces) AfterSaved(engine *orm.Engine) error {
	_ = engine
	e.Calculated = int(e.Uint) + int(e.ReferenceOne.ID)
	return nil
}

func TestInterfaces(t *testing.T) {
	engine := PrepareTables(t, &orm.Registry{}, TestEntityInterfaces{}, TestEntityInterfacesRef{})

	err := engine.Flush(&TestEntityInterfacesRef{})
	assert.Nil(t, err)

	entity := &TestEntityInterfaces{}
	engine.Init(entity)
	assert.Equal(t, uint(3), entity.Uint)
	assert.Equal(t, "hello", entity.Name)
	assert.Equal(t, uint(1), entity.ReferenceOne.ID)

	err = engine.Flush(entity)
	assert.NotNil(t, err)
	assert.Error(t, err, "uint too low")
	entity.Uint = 5
	err = engine.Flush(entity)
	assert.Nil(t, err)
	assert.Equal(t, 6, entity.Calculated)

	entity.Uint = 10
	err = engine.Flush(entity)
	assert.Nil(t, err)
	assert.Equal(t, 11, entity.Calculated)
}
