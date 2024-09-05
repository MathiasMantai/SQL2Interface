package convert

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	f "github.com/MathiasMantai/sql2interface/file"
	"github.com/MathiasMantai/sql2interface/ignore"
	"github.com/MathiasMantai/sql2interface/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SQL2Interface struct {
	SourceDirectory string
	TargetDirectory string
	Sql             SQL
	SqlIgnore       *ignore.S2Ignore
	Combiner []Combiner
}

type SQL struct {
	FileName string `json:"file_name"`
	TableName string   `json:"table_name"`
	Columns   []Column `json:"columns"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}


func NewSQL2Interface(confDir string, source string, target string) *SQL2Interface {
	s2i := &SQL2Interface{
		SourceDirectory: source,
		TargetDirectory: target,
		SqlIgnore:       ignore.NewS2Ignore(confDir),
	}

	s2i.LoadCombiner()

	return s2i
}

func (s2i *SQL2Interface) Convert(fileName string) {
	if s2i.SqlIgnore.IsFileIgnored(fileName) {
		return
	}

	fileContent, getContentErr := f.GetFileContent(s2i.SourceDirectory, fileName)

	if getContentErr != nil {
		fmt.Println(getContentErr)
		return
	}

	parsedData, err := s2i.ParseSQL(fileName, fileContent)

	if err != nil {
		fmt.Println(err)
		return
	}

	addedToCombiner, index := s2i.AddToCombiner(parsedData)

	if addedToCombiner && index != -1 {
		convertSingleTable := s2i.ConvertSingleTable(index)

		if !convertSingleTable {
			fmt.Println("  => conversion will be skipped since convert_single_tables is set to false for this file...")
			return
		}
	}

	targetFileName := parsedData.TableName + ".ts"
	parsedInterface := CreateInterface(parsedData)
	writeFileError := f.SaveFile(s2i.TargetDirectory, targetFileName, parsedInterface)

	if writeFileError != nil {
		fmt.Println("x> error detected: " + writeFileError.Error())
		return
	}

	fmt.Printf("  => creating interface %v and saving to %v\n", parsedData.TableName, filepath.Join(s2i.TargetDirectory, fileName))
}

func (s2i *SQL2Interface) ParseSQL(fileName string, rawSQL string) (SQL, error) {
	var sql SQL

	rawSQL = strings.ToUpper(rawSQL)

	chunks := strings.Split(rawSQL, "(")

	rawTableName := strings.TrimSpace(chunks[0])
	rawColumns := (strings.Split(chunks[1], ")"))[0]
	columns, parseColumnsError := s2i.ParseRowColumnDefinitions(fileName, rawColumns)

	if parseColumnsError != nil {
		return sql, parseColumnsError
	}
	sql.FileName = fileName
	sql.TableName = s2i.ParseRawTableName(rawTableName)
	sql.Columns = columns

	return sql, nil
}

func (s2i *SQL2Interface) ParseRawTableName(rawTableDefinition string) string {
	replacer := strings.NewReplacer("CREATE", "", "TABLE", "", "IF", "", "NOT", "", "EXISTS", "")
	caser := cases.Title(language.Und, cases.NoLower)
	return caser.String(strings.ToLower(strings.TrimSpace(replacer.Replace(rawTableDefinition))))
}

func (s2i *SQL2Interface) ParseRowColumnDefinitions(fileName string, rawColumnDefinitions string) ([]Column, error) {
	var columns []Column

	columnDefinitions := strings.Split(rawColumnDefinitions, ",")

	for _, columnDefinition := range columnDefinitions {

		if strings.TrimSpace(columnDefinition) == "" {
			continue
		}

		columnDefinition = strings.TrimSpace(columnDefinition)
		chunks := strings.Split(columnDefinition, " ")

		caser := cases.Title(language.Und, cases.NoLower)
		columnName := caser.String(strings.ToLower(strings.TrimSpace(chunks[0])))
		columnType := caser.String(strings.ToLower(strings.TrimSpace(TypeMapper(chunks[1]))))

		if s2i.SqlIgnore.IsColumnIgnored(fileName, columnName) {
			continue
		}

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

// main method to run the program
// checks whether the source directory is a file and will only convert the file if true.
func (s2i *SQL2Interface) Run() {

	sourceIsDir, checkSourceDirError := f.IsDir(s2i.SourceDirectory)

	if checkSourceDirError != nil {
		panic(errors.New("=> source directory or file could not be found"))
	}

	if !sourceIsDir {
		s2i.Convert("")
		return
	}

	files, err := f.GetFiles(s2i.SourceDirectory)

	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		fileName := file.Name()
		fmt.Printf("=> attempting to convert file %v to interface...\n", fileName)
		s2i.Convert(fileName)
	}

	combinerToInterfaceError := s2i.CombinerToInterface()
	if combinerToInterfaceError!= nil {
        fmt.Println("x> error detected: " + combinerToInterfaceError.Error())
    }
}


/* COMBINER */

type Combiner struct {
	Tables []string `json:"tables"`
	Amount int `json:"amount"`
	InterfaceName string `json:"interface_name"`
	TableDefinitions []SQL `json:"table_definitions"`
	ConvertSingleTables bool `json:"convert_single_tables"`
}

func (s2i *SQL2Interface) LoadCombiner() {
	combinerConf := s2i.SqlIgnore.Config.CombineTables

	//handling of errors and edge cases
	if combinerConf == nil {
        fmt.Println("=> no combine_tables configuration found. skipping...")
        return 
    }

	if len(combinerConf) == 0 {
		fmt.Println("=> no combine_tables configuration found. skipping...")
		return 
	}

	for _, singleCombinerConf := range combinerConf {
		s2i.Combiner = append(s2i.Combiner, Combiner{
			Tables: singleCombinerConf.Tables,
            Amount: len(singleCombinerConf.Tables),
            InterfaceName: singleCombinerConf.Name,
		})
	}
}

func (s2i *SQL2Interface) AddToCombiner(definition SQL) (bool, int) {
	for i, combiner := range s2i.Combiner {
		if util.ValueInSlice(interface{}(definition.FileName), util.StringToInterfaceSlice(combiner.Tables)) {
			s2i.Combiner[i].TableDefinitions = append(s2i.Combiner[i].TableDefinitions, definition)
			return true, i
		}
	}

	return false, -1
}

func (s2i *SQL2Interface) ConvertSingleTable(index int) bool {
	return s2i.Combiner[index].ConvertSingleTables
}

func CombineTables(interfaceName string, interfaceDefinitions ...SQL) []Column {
	var newSql []Column

	for _, definition := range interfaceDefinitions {
		newSql = append(newSql, definition.Columns...)
	}

	return newSql
}

func (s2i *SQL2Interface) CombinerToInterface() error {
	fmt.Println("=> attempting to convert combined tables to interfaces...")

	for _, combiner := range s2i.Combiner {
        interfaceName := combiner.InterfaceName
		fmt.Printf("  => attempting to create combined interface %v\n", interfaceName)
        tableDefinitions := combiner.TableDefinitions

        combinedColumns := CombineTables(interfaceName, tableDefinitions...)

        targetFileName := interfaceName

		if !strings.Contains(targetFileName, ".ts") {
			targetFileName += ".ts"
        }
		
        parsedInterface := CreateInterface(SQL{TableName: interfaceName, Columns: combinedColumns})
        writeFileError := f.SaveFile(s2i.TargetDirectory, targetFileName, parsedInterface)

        if writeFileError != nil {
            return writeFileError
        }

        fmt.Printf("  => creating combined interface %v and saving to %v\n", interfaceName, filepath.Join(s2i.TargetDirectory, targetFileName))
    }

    return nil
}