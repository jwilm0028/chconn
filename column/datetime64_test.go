package column_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn"
	"github.com/vahid-sohrabloo/chconn/column"
)

func TestDateTime64(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := chconn.Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_datetime64`)
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = conn.Exec(context.Background(), `CREATE TABLE test_datetime64 (
				datetime64 DateTime64,
				datetime64_nullable Nullable(DateTime64),
				datetime64_array Array(DateTime64),
				datetime64_array_nullable Array(Nullable(DateTime64))
			) Engine=Memory`)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewDateTime64(3, false)

	colArrayValues := column.NewDateTime64(3, false)
	colArray := column.NewArray(colArrayValues)

	colArrayValuesNil := column.NewDateTime64(3, true)
	colArrayNil := column.NewArray(colArrayValuesNil)

	colNil := column.NewDateTime64(3, true)

	var colInsert []time.Time
	var colInsertArray [][]time.Time
	var colInsertArrayNil [][]*time.Time
	var colNilInsert []*time.Time

	rows := 10
	for i := 1; i <= rows; i++ {
		val := zeroTime.AddDate(i, 0, 0).Truncate(time.Millisecond)
		valArray := []time.Time{val, zeroTime.AddDate(i+1, 0, 0).Truncate(time.Millisecond)}
		valArrayNil := []*time.Time{&val, nil}

		col.Append(val)
		colInsert = append(colInsert, val)

		// example insert time.Time array
		colInsertArray = append(colInsertArray, valArray)
		colArray.AppendLen(len(valArray))
		for _, v := range valArray {
			colArrayValues.Append(v)
		}

		// example insert nullable array
		colInsertArrayNil = append(colInsertArrayNil, valArrayNil)
		colArrayNil.AppendLen(len(valArrayNil))
		for _, v := range valArrayNil {
			colArrayValuesNil.AppendP(v)
		}

		// example add nullable
		if i%2 == 0 {
			colNilInsert = append(colNilInsert, &val)
			if i <= rows/2 {
				// example to add by poiner
				colNil.AppendP(&val)
			} else {
				// example to without poiner
				colNil.Append(val)
				colNil.AppendIsNil(false)
			}
		} else {
			colNilInsert = append(colNilInsert, nil)
			if i <= rows/2 {
				// example to add by poiner
				colNil.AppendP(nil)
			} else {
				// example to add without poiner
				colNil.AppendEmpty()
				colNil.AppendIsNil(true)
			}
		}
	}

	insertstmt, err := conn.Insert(context.Background(), `INSERT INTO
		test_datetime64 (datetime64,datetime64_nullable,datetime64_array,datetime64_array_nullable)
	VALUES`)

	require.NoError(t, err)
	require.Nil(t, res)

	err = insertstmt.Commit(context.Background(),
		col,
		colNil,
		colArray,
		colArrayNil,
	)
	require.NoError(t, err)

	// example read all
	selectStmt, err := conn.Select(context.Background(), `SELECT
		datetime64,datetime64_nullable,datetime64_array,datetime64_array_nullable
	FROM test_datetime64`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewDateTime64(3, false)
	colNilRead := column.NewDateTime64(3, true)
	colArrayReadData := column.NewDateTime64(3, false)
	colArrayRead := column.NewArray(colArrayReadData)
	colArrayReadDataNil := column.NewDateTime64(3, true)
	colArrayReadNil := column.NewArray(colArrayReadDataNil)
	var colData []time.Time
	var colNilData []*time.Time
	var colArrayData [][]time.Time
	var colArrayDataNil [][]*time.Time

	var colArrayLens []int

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colRead)
		require.NoError(t, err)
		colRead.ReadAll(&colData)

		err = selectStmt.NextColumn(colNilRead)
		require.NoError(t, err)
		colNilRead.ReadAllP(&colNilData)

		// read array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayRead)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)

		for _, l := range colArrayLens {
			arr := make([]time.Time, l)
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)

		for _, l := range colArrayLens {
			arr := make([]*time.Time, l)
			colArrayReadDataNil.FillP(arr)
			colArrayDataNil = append(colArrayDataNil, arr)
		}
	}

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)

	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	// example one by one
	selectStmt, err = conn.Select(context.Background(), `SELECT
		datetime64,datetime64_nullable,datetime64_array,datetime64_array_nullable
	FROM test_datetime64`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead = column.NewDateTime64(3, false)
	colNilRead = column.NewDateTime64(3, true)
	colArrayReadData = column.NewDateTime64(3, false)
	colArrayRead = column.NewArray(colArrayReadData)
	colArrayReadDataNil = column.NewDateTime64(3, true)
	colArrayReadNil = column.NewArray(colArrayReadDataNil)
	colData = colData[:0]
	colNilData = colNilData[:0]
	colArrayData = colArrayData[:0]
	colArrayDataNil = colArrayDataNil[:0]

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colRead)
		require.NoError(t, err)
		for colRead.Next() {
			colData = append(colData, colRead.Value())
		}

		// read nullable
		err = selectStmt.NextColumn(colNilRead)
		require.NoError(t, err)
		for colNilRead.Next() {
			colNilData = append(colNilData, colNilRead.ValueP())
		}

		// read array
		err = selectStmt.NextColumn(colArrayRead)
		require.NoError(t, err)
		for colArrayRead.Next() {
			arr := make([]time.Time, colArrayRead.Value())
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		for colArrayReadNil.Next() {
			arr := make([]*time.Time, colArrayReadNil.Value())
			colArrayReadDataNil.FillP(arr)
			colArrayDataNil = append(colArrayDataNil, arr)
		}
	}

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)

	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	conn.RawConn().Close()
}
