package column_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn"
	"github.com/vahid-sohrabloo/chconn/column"
)

func TestUint8(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := chconn.Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_uint8`)
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = conn.Exec(context.Background(), `CREATE TABLE test_uint8 (
				uint8 UInt8,
				uint8_nullable Nullable(UInt8),
				uint8_array Array(UInt8),
				uint8_array_nullable Array(Nullable(UInt8))
			) Engine=Memory`)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewUint8(false)

	colArrayValues := column.NewUint8(false)
	colArray := column.NewArray(colArrayValues)

	colArrayValuesNil := column.NewUint8(true)
	colArrayNil := column.NewArray(colArrayValuesNil)

	colNil := column.NewUint8(true)

	var colInsert []uint8
	var colInsertArray [][]uint8
	var colInsertArrayNil [][]*uint8
	var colNilInsert []*uint8

	rows := 10
	for i := 0; i < rows; i++ {
		val := uint8(i)
		valArray := []uint8{val, uint8(i) + 1}
		valArrayNil := []*uint8{&val, nil}

		col.Append(val)
		colInsert = append(colInsert, val)

		// example insert array
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
		test_uint8 (uint8,uint8_nullable,uint8_array,uint8_array_nullable)
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
		uint8,uint8_nullable,uint8_array,uint8_array_nullable
	FROM test_uint8`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewUint8(false)
	colNilRead := column.NewUint8(true)
	colArrayReadData := column.NewUint8(false)
	colArrayRead := column.NewArray(colArrayReadData)
	colArrayReadDataNil := column.NewUint8(true)
	colArrayReadNil := column.NewArray(colArrayReadDataNil)
	var colData []uint8
	var colNilData []*uint8
	var colArrayData [][]uint8
	var colArrayDataNil [][]*uint8

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
			arr := make([]uint8, l)
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)

		for _, l := range colArrayLens {
			arr := make([]*uint8, l)
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
		uint8,uint8_nullable,uint8_array,uint8_array_nullable
	FROM test_uint8`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead = column.NewUint8(false)
	colNilRead = column.NewUint8(true)
	colArrayReadData = column.NewUint8(false)
	colArrayRead = column.NewArray(colArrayReadData)
	colArrayReadDataNil = column.NewUint8(true)
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
			arr := make([]uint8, colArrayRead.Value())
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		for colArrayReadNil.Next() {
			arr := make([]*uint8, colArrayReadNil.Value())
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

	conn.Close(context.Background())
}
