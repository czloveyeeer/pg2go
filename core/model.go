package core

type Table struct {
	TableName string `gorm:"column:table_name"` //table name
}

type Column struct {
	ColumnNumber int    `gorm:"column_number"`  // column index
	ColumnName   string `gorm:"column_name"`    // column_name
	ColumnType   string `gorm:"column_type"`    // column_type
	IsPrimaryKey string `gorm:"is_primary_key"` // is_primary_key
	Comment      string `gorm:"comment"`        // comment
}
