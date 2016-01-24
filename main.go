package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/lib/pq"
    "log"
    "text/template"
    "bytes"
    "strings"
    "os"
)

var (
    pgColumnMap map[string]string
)

func main() {
    args := os.Args[1:]
    dbUser := args[0]
    dbName := args[1]
    tableNames := args[2:]

    db, err := gorm.Open("postgres", "user=" + dbUser + " dbname=" + dbName + " sslmode=disable")
    if err != nil {
      panic(err)
    }

    db.LogMode(true)
    db.DB()

    db.DB().Ping()
    db.DB().SetMaxIdleConns(10)
    db.DB().SetMaxOpenConns(100)

    result := generateStructsForTables(db, tableNames)
    log.Println("Result: ", result)
}

func generateStructsForTables(db gorm.DB, tableNames []string) string {
    var structs string = `
        import(
            'time'
        )

    `
    for _,tableName := range tableNames {
        columnNames := getColumnNames(db, tableName)
        log.Println("Column names: ", columnNames)
        columnNames = convertDataTypes(columnNames)
        columnNames = convertColumnNames(columnNames)
        log.Println("Column names: ", columnNames)

        compiledTemplate := generate(columnNames, strings.Title(tableName))
        log.Println("Template: ", compiledTemplate)
        structs += compiledTemplate
    }

    return structs
}


func init() {
    pgColumnMap = make(map[string]string)
    pgColumnMap["integer"] = "int"
    pgColumnMap["character varying"] = "string"
    pgColumnMap["timestamp without time zone"] = "time.Time"
    pgColumnMap["boolean"] = "bool"
    pgColumnMap["double precision"] = "float64"
    pgColumnMap["text"] = "string"
}

func convertColumnNames(columns [][]string) [][]string {
    var convertedColumns [][]string
    for _,element := range columns {
        columnName := element[0]
        // Convert snake case
        columnName = convertSnakeCaseToCamelCase(columnName)
        // ID is a special column
        if columnName == "Id" {
            columnName = "ID"
        }
        convertedColumns = append(convertedColumns, []string{columnName, element[1]})
    }
    return convertedColumns
}

func convertDataTypes(columns [][]string) [][]string {
    var convertedColumns [][]string

    // First entry in each array is the column name, second is the data type
    for _,element := range columns {
        columnType := element[1]
        convertedColumns = append(convertedColumns, []string{element[0], pgColumnMap[columnType]})
    }

    return convertedColumns
}

func convertSnakeCaseToCamelCase(str string) string {
    replaced := strings.Replace(str, "_", " ", -1)
    cased := strings.Title(replaced)
    return strings.Replace(cased, " ", "", -1)
}

func generate(columns [][]string, structName string) string {
    tmpl := `
        type {{.StructName}} struct {
            {{range $i, $r := .Columns}}{{range .}}{{.}} {{end}} {{if ne $i $.ColumnLength}}
            {{end}}{{end}}
        }
    `

    data := map[string]interface{}{
        "StructName": structName,
        "Columns": columns,
        "ColumnLength": len(columns) - 1,
    }

    t := template.Must(template.New("model").Parse(tmpl))
    buf := &bytes.Buffer{}
    if err := t.Execute(buf, data); err != nil {
        panic(err)
    }
    s := buf.String()
    return s
}

func getColumnNames(db gorm.DB, table string) ([][]string) {
    query := `select column_name, data_type from information_schema.columns 
              where table_name = '` + table + `'`
    rows, err := db.Raw(query).Rows()
    if err != nil {
        panic(err)
    }
    defer rows.Close()

    var rowData [][]string
    for rows.Next() {
        var columnName string
        var dataType string

        rows.Scan(&columnName, &dataType)
        rowData = append(rowData, []string{columnName, dataType})
    }
    return rowData
}

