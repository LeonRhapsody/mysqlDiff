package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	"log"
	"os"
	"strings"
)

// TableColumn 表结构
type TableColumn struct {
	ID      string
	Table   string
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

type DBInfo struct {
	DBID   string
	Tables []string
	//map[field]struct
	TableInfo map[string][]TableColumn
}

// 收集库的所有表结构，return DBInfo
func initDbInfo(connStr string) DBInfo {

	//root/Alipms@123/10.21.47.70:3307/DCP_ISMS
	str := strings.Split(connStr, "/")
	id := str[2] + "/" + str[3]
	user := str[0]
	password := str[1]
	ipPort := str[2]
	dbname := str[3]

	dbStr := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, ipPort, dbname)
	dbc, err := sql.Open("mysql", dbStr)
	if err != nil {
		panic(err.Error())
	}
	defer dbc.Close()

	TableColumnsMap := make(map[string][]TableColumn)

	rows, err := dbc.Query("show Tables")
	if err != nil {
		panic(err.Error())
	}

	defer rows.Close()

	//get Tables
	var tableLists []string

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			panic(err.Error())
		}
		tableLists = append(tableLists, tableName)
	}

	//get columns
	for _, t := range tableLists {
		var columns []TableColumn

		columnRows, err := dbc.Query("DESCRIBE " + t)
		if err != nil {
			panic(err.Error())
		}

		for columnRows.Next() {
			var column TableColumn
			column.Table = t
			column.ID = id

			err = columnRows.Scan(&column.Field, &column.Type, &column.Null, &column.Key, &column.Default, &column.Extra)
			if err != nil {
				panic(err.Error())
			}
			columns = append(columns, column)
		}
		TableColumnsMap[t] = columns
	}

	return DBInfo{
		DBID:      id,
		Tables:    tableLists,
		TableInfo: TableColumnsMap,
	}
}

func sliceToMap0(s []string) map[string]int {
	map1 := make(map[string]int)

	for _, v := range s {

		map1[v] = 0
	}

	return map1

}

//输出两个库的相同表名和不同表名
func diffTableName(db1 DBInfo, db2 DBInfo) (map[string]string, map[string]string) {

	sameTables := make(map[string]string)
	missTables := make(map[string]string)

	map1 := sliceToMap0(db1.Tables)
	map2 := sliceToMap0(db2.Tables)

	for k, _ := range map1 {
		if _, ok := map2[k]; !ok {
			missTables[k] = db1.DBID
		} else {
			if _, ok := sameTables[k]; !ok {
				sameTables[k] = db1.DBID
			}

		}
	}
	for k, _ := range map2 {
		if _, ok := map1[k]; !ok {
			missTables[k] = db2.DBID
		} else {
			if _, ok := sameTables[k]; !ok {
				sameTables[k] = db2.DBID
			}

		}
	}

	return missTables, sameTables

}

// ColumnsToMap map[table]info --> map[field]info
func ColumnsToMap(tableStructs []TableColumn) map[string]TableColumn {
	map1 := make(map[string]TableColumn)

	for _, tableStruct := range tableStructs {
		map1[tableStruct.Field] = tableStruct
	}
	return map1
}

func main() {
	// 定义命令行参数
	var (
		op           string
		templateName string
		connStr      string
		help         string
	)

	// 定义命令行参数的具体含义
	flag.StringVar(&help, "help", "./mysqlDiff -op diff -templateName DCP_ISMS -connStr root/Alipms@123/10.21.47.70:3307/DCP_ISMS_1", "usage description")
	flag.StringVar(&op, "op", "diff", "load/diff")
	flag.StringVar(&templateName, "templateName", "", "templateName")
	flag.StringVar(&connStr, "connStr", "root/Alipms@123@10.21.47.70:3307/DCP_ISMS", "connStr")

	// 解析命令行参数
	flag.Parse()

	// 保存模板库数据到文件
	if op == "save" {
		templateDB := initDbInfo(connStr)
		dbName := strings.Split(connStr, "/")[3]
		save(dbName+".db", templateDB)
		log.Printf("load %s successful ", dbName)
		os.Exit(0)
	}

	templateDB := load(templateName + ".db")
	targetDB := initDbInfo(connStr)

	//log.Println(templateDB, targetDB)
	missTables, sameTables := diffTableName(templateDB, targetDB)

	var missTableRows []table.Row
	var diffFields []table.Row
	var missFieldsRow []table.Row

	for t, _ := range missTables {
		missTableRows = append(missTableRows, table.Row{"+", missTables[t], t})
	}

	for t, _ := range sameTables {

		missFields := make(map[string]TableColumn)
		sameFields := make(map[string]TableColumn)

		tableStruct1 := templateDB.TableInfo[t]
		tableStruct2 := targetDB.TableInfo[t]

		ColumnMap1 := ColumnsToMap(tableStruct1)
		ColumnMap2 := ColumnsToMap(tableStruct2)

		for k, _ := range ColumnMap1 {
			if _, ok := ColumnMap2[k]; !ok {
				missFields[k] = ColumnMap1[k]
			} else {
				if _, ok = sameFields[k]; !ok {
					sameFields[k] = ColumnMap1[k]
				}
			}
		}
		for k, _ := range ColumnMap2 {
			if _, ok := ColumnMap1[k]; !ok {
				missFields[k] = ColumnMap2[k]
			} else {
				if _, ok = sameFields[k]; !ok {
					sameFields[k] = ColumnMap2[k]
				}

			}
		}

		for k, _ := range missFields {
			missFieldsRow = append(missFieldsRow, table.Row{"+", missFields[k].ID, missFields[k].Table, k})

		}

		for k, _ := range sameFields {
			if ColumnMap1[k].Type != ColumnMap2[k].Type {
				diffFields = append(diffFields, table.Row{"?", ColumnMap1[k].ID, ColumnMap1[k].Table, k, ColumnMap1[k].Type})
				diffFields = append(diffFields, table.Row{"?", ColumnMap2[k].ID, ColumnMap2[k].Table, k, ColumnMap2[k].Type})
			}
		}

	}

	demo := NewDemo()
	demo.MakeHeader()
	demo.AppendRows(missTableRows)
	demo.AppendRows(missFieldsRow)
	demo.AppendRows(diffFields)
	demo.ColumnMerge([]string{"Table", "Field"})
	demo.Print()
}
