package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/summer-solutions/orm"
)

type AddressByIDLocal struct {
	Street   string
	Building uint16
}

type TestEntityByIDLocal struct {
	orm.ORM              `orm:"localCache"`
	ID                   uint
	Name                 string `orm:"length=100;index=FirstIndex"`
	BigName              string `orm:"length=max"`
	Uint8                uint8  `orm:"unique=SecondIndex:2,ThirdIndex"`
	Uint16               uint16
	Uint24               uint32 `orm:"mediumint=true"`
	Uint32               uint32
	Uint64               uint64 `orm:"unique=SecondIndex"`
	Int8                 int8
	Int16                int16
	Int32                int32
	Int64                int64
	Rune                 rune
	Int                  int
	Bool                 bool
	Float32              float32
	Float32Precision     float32 `orm:"precision=10"`
	Float64              float64
	Float32Decimal       float32 `orm:"decimal=8,2"`
	Float64DecimalSigned float64 `orm:"decimal=8,4;unsigned=false"`
	Year                 uint16  `orm:"year=true"`
	Date                 time.Time
	DateTime             time.Time `orm:"time=true"`
	Address              AddressByIDLocal
	JSON                 interface{}
	Uint8Slice           []uint8
	Ignored              []time.Time `orm:"ignore"`
	ReferenceOne         *TestEntityByIDLocal
}

func TestGetByIDLocal(t *testing.T) {
	var entity TestEntityByIDLocal
	engine := PrepareTables(t, &orm.Registry{}, entity)

	found, err := engine.LoadByID(100, &entity)
	assert.Nil(t, err)
	assert.False(t, found)
	found, err = engine.LoadByID(100, &entity)
	assert.Nil(t, err)
	assert.False(t, found)

	entity = TestEntityByIDLocal{}
	engine.RegisterEntity(&entity)
	err = entity.Flush()
	assert.Nil(t, err)

	assert.False(t, entity.IsDirty())

	has, err := engine.LoadByID(1, &entity)
	assert.True(t, has)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), entity.ID)

	entity.ReferenceOne = &TestEntityByIDLocal{ID: 1}
	err = entity.Flush()
	assert.Nil(t, err)
	assert.False(t, entity.IsDirty())

	DBLogger := &TestDatabaseLogger{}
	pool := engine.GetMysql()
	pool.RegisterLogger(DBLogger)

	found, err = engine.LoadByID(1, &entity, "ReferenceOne")
	assert.Nil(t, err)
	assert.True(t, found)
	assert.NotNil(t, entity)
	assert.Equal(t, uint(1), entity.ID)
	assert.Equal(t, "", entity.Name)
	assert.Equal(t, "", entity.BigName)
	assert.Equal(t, uint8(0), entity.Uint8)
	assert.Equal(t, uint16(0), entity.Uint16)
	assert.Equal(t, uint32(0), entity.Uint24)
	assert.Equal(t, uint32(0), entity.Uint32)
	assert.Equal(t, int8(0), entity.Int8)
	assert.Equal(t, int16(0), entity.Int16)
	assert.Equal(t, int32(0), entity.Int32)
	assert.Equal(t, int64(0), entity.Int64)
	assert.Equal(t, rune(0), entity.Rune)
	assert.Equal(t, 0, entity.Int)
	assert.Equal(t, false, entity.Bool)
	assert.Equal(t, float32(0), entity.Float32)
	assert.Equal(t, float64(0), entity.Float64)
	assert.Equal(t, float32(0), entity.Float32Decimal)
	assert.Equal(t, float64(0), entity.Float64DecimalSigned)
	assert.Equal(t, uint16(0), entity.Year)
	assert.IsType(t, time.Time{}, entity.Date)
	assert.NotNil(t, entity.ReferenceOne)
	assert.True(t, entity.Date.Equal(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.IsType(t, time.Time{}, entity.DateTime)
	assert.True(t, entity.Date.Equal(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, AddressByIDLocal{Street: "", Building: uint16(0)}, entity.Address)
	assert.Nil(t, entity.JSON)
	assert.Equal(t, "", string(entity.Uint8Slice))
	assert.Nil(t, err)

	assert.False(t, entity.IsDirty())
	assert.Len(t, DBLogger.Queries, 0)

	entity.Name = "Test name"
	entity.BigName = "Test big name"
	entity.Uint8 = 2
	entity.Int = 14
	entity.Bool = true
	entity.Float32 = 1.11
	entity.Float32Precision = 1.13
	entity.Float64 = 7.002
	entity.Float32Decimal = 123.13
	entity.Float64DecimalSigned = -12.01
	entity.Year = 1982
	entity.Date = time.Date(1982, 4, 6, 0, 0, 0, 0, time.UTC)
	entity.DateTime = time.Date(2019, 2, 11, 12, 34, 11, 0, time.UTC)
	entity.Address.Street = "wall street"
	entity.Address.Building = 12
	entity.JSON = map[string]string{"name": "John"}
	entity.Uint8Slice = []byte("test me")

	assert.True(t, entity.IsDirty())

	err = entity.Flush()
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.False(t, entity.IsDirty())
	assert.Len(t, DBLogger.Queries, 1)

	var entity2 TestEntityByIDLocal
	found, err = engine.LoadByID(1, &entity2)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.NotNil(t, entity2)
	assert.Equal(t, "Test name", entity2.Name)
	assert.Equal(t, "Test big name", entity2.BigName)
	assert.Equal(t, uint8(2), entity2.Uint8)
	assert.Equal(t, 14, entity2.Int)
	assert.Equal(t, true, entity2.Bool)
	assert.Equal(t, float32(1.11), entity2.Float32)
	assert.Equal(t, float32(1.13), entity2.Float32Precision)
	assert.Equal(t, 7.002, entity2.Float64)
	assert.Equal(t, float32(123.13), entity2.Float32Decimal)
	assert.Equal(t, -12.01, entity2.Float64DecimalSigned)
	assert.Equal(t, uint16(1982), entity2.Year)
	assert.Equal(t, time.Date(1982, 4, 6, 0, 0, 0, 0, time.UTC), entity2.Date)
	assert.Equal(t, time.Date(2019, 2, 11, 12, 34, 11, 0, time.UTC), entity2.DateTime)
	assert.Equal(t, "wall street", entity2.Address.Street)
	assert.Equal(t, uint16(12), entity2.Address.Building)
	assert.Equal(t, map[string]interface{}{"name": "John"}, entity2.JSON)
	assert.Len(t, DBLogger.Queries, 1)
}

func BenchmarkGetByIDLocal(b *testing.B) {
	var entity TestEntityByIDLocal
	engine := PrepareTables(&testing.T{}, &orm.Registry{}, entity)

	entity = TestEntityByIDLocal{}
	engine.RegisterEntity(&entity)
	err := entity.Flush()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		_, _ = engine.LoadByID(1, &entity)
	}
}
