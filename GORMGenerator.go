package GORMGenerator

import (
    "bytes"
    "github.com/jinzhu/gorm"
    "log"
    "strings"
    "text/template"
    _ "github.com/lib/pq"
)

type DBInfo struct {
    DBName string
    DBUser string
    DBType string
    DBLogMode bool
}

type Column struct {
    Name string
    Type string
}

var (
    columnMap map[string]map[string]string
)

func init() {
    columnMap = make(map[string]map[string]string)
    columnMap["postgres"]["integer"] = "int"
    columnMap["postgres"]["character varying"] = "string"
    columnMap["postgres"]["timestamp without time zone"] = "time.Time"
    columnMap["postgres"]["boolean"] = "bool"
    columnMap["postgres"]["double precision"] = "float64"
    columnMap["postgres"]["text"] = "string"
}

func openDBHandle(dbInfo DBInfo) gorm.DB {
    if dbInfo.DBType != "postgres" {
        log.Println("[WARNING]: GORMGenerator may not work with databases other than postgres. We'll attempt to generate the structs for you, but just be warned.")
    }
    db, err := gorm.Open(dbInfo.DBType, "user=" + dbInfo.DBUser +
    " dbname=" + dbInfo.DBName + " sslmode=disable")

    if err != nil {
        log.Println("[ERROR]: ", err)
        return db
    }

    if dbInfo.DBLogMode {
        db.LogMode(dbInfo.DBLogMode)
    }
    db.DB()

    return db
}

func GenerateStructsForTables(dbInfo DBInfo, tableNames []string) string {
    db := openDBHandle(dbInfo)
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
        convertedColumns = append(convertedColumns, []string{element[0], columnMap["postgres"][columnType]})
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

