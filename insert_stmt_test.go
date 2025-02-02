package chconn

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn/column"
	"github.com/vahid-sohrabloo/chconn/setting"
)

func TestInsertError(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	config, err := ParseConfig(connString)
	require.NoError(t, err)

	// test lock error
	c, err := ConnectConfig(context.Background(), config)
	require.NoError(t, err)

	c.(*conn).status = connStatusUninitialized
	res, err := c.Insert(context.Background(), "insert into system.numbers VALUES")
	require.Nil(t, res)
	require.EqualError(t, err, "conn uninitialized")
	require.EqualError(t, c.(*conn).lock(), "conn uninitialized")
	c.Close(context.Background())

	// test write block info error
	config.WriterFunc = func(w io.Writer) io.Writer {
		return &writerErrorHelper{
			err:         errors.New("timeout"),
			w:           w,
			numberValid: 1,
		}
	}
	c, err = ConnectConfig(context.Background(), config)
	require.NoError(t, err)

	res, err = c.Insert(context.Background(), "insert into system.numbers VALUES")
	require.EqualError(t, err, "block: write block info (timeout)")
	require.Nil(t, res)

	// test insert server error
	config.WriterFunc = nil
	c, err = ConnectConfig(context.Background(), config)
	require.NoError(t, err)

	res, err = c.Insert(context.Background(), "insert into system.numbers VALUES")
	require.EqualError(t, err, " DB::Exception (48): Method write is not supported by storage SystemNumbers")
	require.Nil(t, res)

	// test not block data error
	c, err = ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	res, err = c.Insert(context.Background(), "SET enable_http_compression=1")
	require.EqualError(t, err, "Unexpected packet from server (expected serverData got <nil>)")
	require.Nil(t, res)

	// test read column error
	c, err = ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	_, err = c.Exec(context.Background(), `DROP TABLE IF EXISTS clickhouse_test_insert_error`)
	require.NoError(t, err)
	settings := setting.NewSettings()
	settings.JoinUseNulls(false)
	_, err = c.ExecWithSetting(context.Background(), `CREATE TABLE clickhouse_test_insert_error (
				int8  Int8
			) Engine=Memory`, settings)

	require.NoError(t, err)

	config.ReaderFunc = func(r io.Reader) io.Reader {
		return &readErrorHelper{
			err:         errors.New("timeout"),
			r:           r,
			numberValid: 27,
		}
	}
	c, err = ConnectConfig(context.Background(), config)
	require.NoError(t, err)
	res, err = c.Insert(context.Background(), `INSERT INTO clickhouse_test_insert_error (
				int8
			) VALUES`)
	require.EqualError(t, err, "block: read column name (timeout)")
	require.Nil(t, res)
}

func TestInsert(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_insert`)
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = conn.Exec(context.Background(), `CREATE TABLE test_insert (
				int8 Int8
			) Engine=Memory`)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewInt8(false)

	var colInsert []int8

	rows := 10
	for i := 0; i < rows; i++ {
		val := int8(i)

		col.Append(val)
		colInsert = append(colInsert, val)
	}

	insertstmt, err := conn.Insert(context.Background(), `INSERT INTO test_insert (int8) VALUES`)

	require.NoError(t, err)
	require.Nil(t, res)

	err = insertstmt.Commit(context.Background(),
		col,
	)
	require.NoError(t, err)

	// example read all
	selectStmt, err := conn.Select(context.Background(), `SELECT int8 FROM test_insert`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewInt8(false)

	var colData []int8

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colRead)
		require.NoError(t, err)
		colRead.ReadAll(&colData)
	}

	assert.Equal(t, colInsert, colData)
	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	conn.RawConn().Close()
}
