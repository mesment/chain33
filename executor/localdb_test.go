package executor_test

import (
	"testing"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/util/testnode"
	"github.com/stretchr/testify/assert"
)

func TestLocalDBGet(t *testing.T) {
	mock33 := testnode.New("", nil)
	defer mock33.Close()
	db := executor.NewLocalDB(mock33.GetClient())
	testDBGet(t, db)
}

func BenchmarkLocalDBGet(b *testing.B) {
	mock33 := testnode.New("", nil)
	defer mock33.Close()
	db := executor.NewLocalDB(mock33.GetClient())
	err := db.Set([]byte("k1"), []byte("v1"))
	assert.Nil(b, err)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v, err := db.Get([]byte("k1"))
		assert.Nil(b, err)
		assert.Equal(b, v, []byte("v1"))
	}
}

func TestLocalDBTxGet(t *testing.T) {
	mock33 := testnode.New("", nil)
	defer mock33.Close()
	db := executor.NewLocalDB(mock33.GetClient())
	testTxGet(t, db)
}

func testDBGet(t *testing.T, db dbm.KV) {
	err := db.Set([]byte("k1"), []byte("v1"))
	assert.Nil(t, err)
	v, err := db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v1"))

	err = db.Set([]byte("k1"), []byte("v11"))
	assert.Nil(t, err)
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v11"))
}

func testTxGet(t *testing.T, db dbm.KV) {
	//新版本
	db.Begin()
	err := db.Set([]byte("k1"), []byte("v1"))
	assert.Nil(t, err)
	v, err := db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v1"))

	db.Commit()
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v1"))

	//在非transaction中set，直接set成功，不能rollback
	err = db.Set([]byte("k1"), []byte("v11"))
	assert.Nil(t, err)

	db.Begin()
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v11"))

	err = db.Set([]byte("k1"), []byte("v12"))
	assert.Nil(t, err)
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v12"))

	db.Rollback()
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v11"))
}

func TestLocalDB(t *testing.T) {
	mock33 := testnode.New("", nil)
	defer mock33.Close()
	db := executor.NewLocalDB(mock33.GetClient())
	err := db.Set([]byte("k1"), []byte("v1"))
	assert.Nil(t, err)
	v, err := db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v1"))

	err = db.Set([]byte("k1"), []byte("v11"))
	assert.Nil(t, err)
	v, err = db.Get([]byte("k1"))
	assert.Nil(t, err)
	assert.Equal(t, v, []byte("v11"))

	//beigin and rollback not imp
	db.Begin()
	db.Rollback()
	db.Commit()
	db.List([]byte("a"), []byte("b"), 1, 1)
}