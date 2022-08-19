package core

import (
	"bufio"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"pg2go/db"
	"pg2go/util"
	"strings"
)

var findTablesSql = `
SELECT
c.relkind AS type,
c.relname AS table_name
FROM pg_class c
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
AND c.relkind = 'r'
ORDER BY c.relname
`

var FindColumnsSql = `
SELECT
    a.attnum AS column_number,
    a.attname AS column_name,
    --format_type(a.atttypid, a.atttypmod) AS column_type,
    a.attnotnull AS not_null,
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') AS default_value,
    COALESCE(ct.contype = 'p', false) AS  is_primary_key,
    CASE
        WHEN a.atttypid = ANY ('{int,int8,int2}'::regtype[])
          AND EXISTS (
             SELECT 1 FROM pg_attrdef ad
             WHERE  ad.adrelid = a.attrelid
             AND    ad.adnum   = a.attnum
             )
            THEN CASE a.atttypid
                    WHEN 'int'::regtype  THEN 'serial'
                    WHEN 'int8'::regtype THEN 'bigserial'
                    WHEN 'int2'::regtype THEN 'smallserial'
                 END
        WHEN a.atttypid = ANY ('{uuid}'::regtype[]) AND COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') != ''
            THEN 'autogenuuid'
        ELSE format_type(a.atttypid, a.atttypmod)
    END AS column_type,
    COALESCE(b.description,'') AS comment
FROM pg_attribute a
JOIN ONLY pg_class c ON c.oid = a.attrelid
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
AND a.attnum = ANY(ct.conkey) AND ct.contype = 'p'
LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
LEFT JOIN pg_description b ON a.attrelid=b.objoid AND a.attnum = b.objsubid
WHERE a.attisdropped = false
AND n.nspname = 'public'
AND c.relname = ?
AND a.attnum > 0
ORDER BY a.attnum
`

func FindTables() []Table {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(fmt.Sprintf("recover from a fatal error : %v", e))
		}
	}()

	var tables = make([]Table, 0, 10)
	db.DB.Raw(findTablesSql).Find(&tables)
	return tables
}

// FindColumns find columns' property by specific dataSource and table name
func FindColumns(tableName string) []Column {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(fmt.Sprintf("recover from a fatal error : %v", e))
		}
	}()

	var columns = make([]Column, 0, 10)
	db.DB.Raw(FindColumnsSql, tableName).Find(&columns)
	return columns
}

// TableToStruct generate go model y specific the dataSource and table name
func TableToStruct(tableName string) (string, string) {

	columnString := ""
	tmp := ""
	columns := FindColumns(tableName)
	pk := ""
	for _, column := range columns {
		if column.IsPrimaryKey == "true" {
			pk = column.ColumnName
		}
		tmp = fmt.Sprintf("    %s  %s //%s \n", util.UnderLineToHump(util.HumpToUnderLine(column.ColumnName)), util.TypeConvert(column.ColumnType), column.Comment)
		columnString = columnString + tmp
	}

	rs := fmt.Sprintf("type %s struct{\n%s}", util.UnderLineToHump(util.HumpToUnderLine(tableName)), columnString)

	return rs, pk
}

// AddJSONFormGormTag 添加json格式
func AddJSONFormGormTag(in, pk string) string {
	var result string
	scanner := bufio.NewScanner(strings.NewReader(in))
	var oldLineTmp = ""
	var lineTmp = ""
	var propertyTmp = ""
	var seperateArr []string
	for scanner.Scan() {
		oldLineTmp = scanner.Text()
		lineTmp = strings.Trim(scanner.Text(), " ")
		if strings.Contains(lineTmp, "{") || strings.Contains(lineTmp, "}") {
			result = result + oldLineTmp + "\n"
			continue
		}
		seperateArr = util.Split(lineTmp, " ")
		if len(seperateArr) < 3 {
			fmt.Println("============================", seperateArr, "=============================")
			continue
		}
		header := fmt.Sprintf("    %s  %s", seperateArr[0], seperateArr[1])
		footer := fmt.Sprintf("    %s", seperateArr[2])

		propertyTmp = util.HumpToUnderLine(seperateArr[0])

		if pk == propertyTmp {
			oldLineTmp = header + fmt.Sprintf("    `gorm:\"column:%s,%s\" json:\"%s,%s\" form:\"%s\"`", propertyTmp, "primary_key", propertyTmp, "omitempty", propertyTmp) + footer
		} else {
			oldLineTmp = header + fmt.Sprintf("    `gorm:\"column:%s\" json:\"%s,%s\" form:\"%s\"`", propertyTmp, propertyTmp, "omitempty", propertyTmp) + footer
		}
		result = result + oldLineTmp + "\n"
	}
	return result
}
