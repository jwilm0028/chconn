package column

type UUID struct {
	column
	val [16]byte
}

func NewUUID(nullable bool) *UUID {
	return &UUID{
		column: column{
			nullable:    nullable,
			colNullable: newNullable(),
			size:        UUIDSize,
		},
	}
}

func (c *UUID) Next() bool {
	if c.i >= c.totalByte {
		return false
	}
	c.setVal(c.b[c.i : c.i+c.size])
	c.i += c.size
	return true
}

func (c *UUID) Value() [16]byte {
	return c.val
}

func (c *UUID) ValueP() *[16]byte {
	if c.colNullable.b[(c.i-c.size)/(c.size)] == 1 {
		return nil
	}
	val := c.val
	return &val
}

func (c *UUID) ReadAll(value *[][16]byte) {
	for i := 0; i < c.totalByte; i += c.size {
		c.setVal(c.b[i : i+c.size])
		*value = append(*value,
			c.val)
	}
}

func (c *UUID) Fill(value [][16]byte) {
	for i := range value {
		c.setVal(c.b[c.i : c.i+c.size])
		value[i] = c.val
		c.i += c.size
	}
}

func (c *UUID) ReadAllP(value *[]*[16]byte) {
	for i := 0; i < c.totalByte; i += c.size {
		if c.colNullable.b[i/c.size] != 0 {
			*value = append(*value, nil)
			continue
		}
		c.setVal(c.b[i : i+c.size])
		val := c.val
		*value = append(*value, &val)
	}
}

func (c *UUID) FillP(value []*[16]byte) {
	for i := range value {
		if c.colNullable.b[c.i/c.size] == 1 {
			value[i] = nil
			c.i += c.size
			continue
		}
		c.setVal(c.b[c.i : c.i+c.size])
		val := c.val
		value[i] = &val
		c.i += c.size
	}
}

func (c *UUID) AppendEmpty() {
	c.numRow++
	c.writerData = append(c.writerData, emptyByte[:c.size]...)
}

func (c *UUID) Append(v [16]byte) {
	c.numRow++
	c.writerData = append(c.writerData,
		v[7], v[6], v[5], v[4],
		v[3], v[2], v[1], v[0],
		v[15], v[14], v[13], v[12],
		v[11], v[10], v[9], v[8],
	)
}

func (c *UUID) AppendP(v *[16]byte) {
	if v == nil {
		c.AppendEmpty()
		c.colNullable.Append(1)
		return
	}
	c.colNullable.Append(0)
	c.Append(*v)
}

func (c *UUID) setVal(b []byte) {
	c.val[0], c.val[7] = b[7], b[0]
	c.val[1], c.val[6] = b[6], b[1]
	c.val[2], c.val[5] = b[5], b[2]
	c.val[3], c.val[4] = b[4], b[3]
	c.val[8], c.val[15] = b[15], b[8]
	c.val[9], c.val[14] = b[14], b[9]
	c.val[10], c.val[13] = b[13], b[10]
	c.val[11], c.val[12] = b[12], b[11]
}
