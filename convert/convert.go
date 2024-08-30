package convert

import (
	"path/filepath"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"github.com/MathiasMantai/sql2interface/ignore"
	f	"github.com/MathiasMantai/sql2interface/file"
	"strings"
)

type SQL2Interface struct {
	SourceDirectory string
	TargetDirectory string
	Sql SQL 
	SqlIgnore *ignore.S2Ignore
}

type SQL struct {
	TableName string   `json:"table_name"`
	Columns   []Column `json:"columns"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func NewSQL2Interface(source string, target string) *SQL2Interface {
    return &SQL2Interface{
		SourceDirectory: source,
        TargetDirectory: target,
		SqlIgnore: ignore.NewS2Ignore(source),
	}
}

func (s2i *SQL2Interface) Run() {
	files, err := f.GetFiles(s2i.SourceDirectory)

	if err != nil {
		fmt.Println(err)
	}
	
	for _, file := range files {
		fileName := file.Name()

		if s2i.SqlIgnore.IsIgnored(fileName) || fileName == s2i.SqlIgnore.FileName {
            continue
        }

		fileContent, getContentErr := f.GetFileContent(s2i.SourceDirectory, fileName)
	
		if getContentErr != nil {
			fmt.Println(getContentErr)
			continue
		}	
	
		parsedData, err := ParseSQL(fileContent)
	
		if err != nil {
			fmt.Println(err)
			continue
		}

		targetFileName := parsedData.TableName + ".ts"
		parsedInterface := CreateInterface(parsedData)
		writeFileError := f.SaveFile(s2i.TargetDirectory, targetFileName, parsedInterface)
	
		if writeFileError != nil {
			fmt.Println("=> error detected: " + writeFileError.Error())
			continue
		}
	
		fmt.Printf("=> creating interface %v and saving to %v\n", parsedData.TableName, filepath.Join(s2i.TargetDirectory, fileName))
	}
}

func ParseSQL(rawSQL string) (SQL, error) {
	var sql SQL

	rawSQL = strings.ToUpper(rawSQL)

	chunks := strings.Split(rawSQL, "(")

	rawTableName := strings.TrimSpace(chunks[0])
	rawColumns := (strings.Split(chunks[1], ")"))[0]
	columns, parseColumnsError := ParseRowColumnDefinitions(rawColumns)

	if parseColumnsError != nil {
		return sql, parseColumnsError
	}

	sql.TableName = ParseRawTableName(rawTableName)
	sql.Columns = columns

	return sql, nil
}

func ParseRawTableName(rawTableDefinition string) string {
	replacer := strings.NewReplacer("CREATE", "", "TABLE", "", "IF", "", "NOT", "", "EXISTS", "")
	caser := cases.Title(language.Und, cases.NoLower)
	return caser.String(strings.ToLower(strings.TrimSpace(replacer.Replace(rawTableDefinition))))
}

func ParseRowColumnDefinitions(rawColumnDefinitions string) ([]Column, error) {
	var columns []Column

	columnDefinitions := strings.Split(rawColumnDefinitions, ",")

	for _, columnDefinition := range columnDefinitions {
		columnDefinition = strings.TrimSpace(columnDefinition)
		chunks := strings.Split(columnDefinition, " ")
		caser := cases.Title(language.Und, cases.NoLower)
		columnName := caser.String(strings.ToLower(strings.TrimSpace(chunks[0])))
		columnType := caser.String(strings.ToLower(strings.TrimSpace(TypeMapper(chunks[1]))))
		columns = append(columns, Column{
			Name: columnName,
			Type: columnType,
		})
	}

	return columns, nil
}

func TypeMapper(colType string) string {

	if strings.Contains(strings.ToUpper(colType), "VARCHAR") {
		colType = "VARCHAR"
	}

	switch strings.ToUpper(colType) {
	case "VARCHAR", "TEXT", "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR", "ENUM":
		return "String"
	case "INT", "INTEGER", "SMALLINT", "TINYINT", "MEDIUMINT", "BIGINT", "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "SERIAL":
		return "Number"
	case "BOOLEAN":
		return "Boolean"
	default:
		return "String"
	}
}

func CreateInterface(sql SQL) string {
	interfaceFields := ""
	length := len(sql.Columns)
	for i, column := range sql.Columns {
		interfaceFields += fmt.Sprintf("\t%v: %v", column.Name, column.Type)

		if i < length-1 {
			interfaceFields += ", "
		}

		interfaceFields += "\r\n"
	}

	return fmt.Sprintf("interface %v {\n%v}", sql.TableName, interfaceFields)
}
