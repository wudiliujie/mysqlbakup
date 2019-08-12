package backup

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/wudiliujie/common/convert"
	"github.com/wudiliujie/common/log"
	"github.com/wudiliujie/common/progressbar"
	"github.com/wudiliujie/mysqlbakup/infoObjects"
	"github.com/wudiliujie/mysqlbakup/mysqlObjects"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

type ProcessEndType int32

const Version = "1.0"
const (
	_ ProcessEndType = iota
	ProcessEndType_UnknownStatus
	ProcessEndType_Complete
	ProcessEndType_Cancelled
	ProcessEndType_Error
)

type ProcessType int32

const (
	_ ProcessType = iota
	ProcessType_Export
	ProcessType_Import
)

type NextImportAction int32

const (
	_ NextImportAction = iota
	NextImportAction_Ignore
	NextImportAction_SetNames
	NextImportAction_CreateNewDatabase
	NextImportAction_AppendLine
	NextImportAction_ChangeDelimiter
	NextImportAction_AppendLineAndExecute
)

type KV struct {
	K string
	V string
}
type MySqlBackup struct {
	dataBase                      *mysqlObjects.MySqlDatabase
	server                        *mysqlObjects.MySqlServer
	db                            *sql.DB
	exportInfo                    *infoObjects.ExportInformations
	timeStart                     time.Time
	stopProcess                   bool
	processCompletionType         ProcessEndType
	currentProcess                ProcessType
	lastError                     error
	currentTableName              string
	totalRowsInCurrentTable       int64
	totalRowsInAllTables          int64
	currentRowIndexInCurrentTable int64
	currentRowIndexInAllTable     int64
	totalTables                   int64
	currentTableIndex             int64
	writeObj                      *bufio.Writer
	readerObj                     *bufio.Reader
	totalBytes                    int64
	currentBytes                  int64
	buffImport                    *bytes.Buffer
	delimiter                     string
	lastErrorSql                  string
	OnExportProcessEvent          func(per int32, msg string)

	pgb *progressbar.ProgressBar
}

func NewMySqlBackup(_db *sql.DB) *MySqlBackup {
	backup := new(MySqlBackup)
	backup.Init(_db)
	return backup
}
func (m *MySqlBackup) Init(_db *sql.DB) {
	m.db = _db
	m.dataBase = mysqlObjects.NewMySqlDatabase()
	m.exportInfo = infoObjects.NewExportInformations()
	m.server = mysqlObjects.NewMysqlServer()
}

//导出初始化变量
func (m *MySqlBackup) Export_InitializeVariables() error {
	if m.db == nil {
		return errors.New("sql.DB is not initialized. Object not set to an instance of an object.")
	}
	if m.exportInfo.BlobExportMode == infoObjects.BlobDataExportMode_BinaryChar &&
		!m.exportInfo.BlobExportModeForBinaryStringAllow {
		return errors.New("[ExportInfo.BlobExportMode = BlobDataExportMode.BinaryString] is still under development.")
	}

	m.timeStart = time.Now()
	m.stopProcess = false
	m.processCompletionType = ProcessEndType_UnknownStatus
	m.currentProcess = ProcessType_Export
	m.lastError = nil
	//timerReport.Interval = ExportInfo.IntervalForProgressReport;
	m.dataBase.GetDatabaseInfo(m.db, m.exportInfo.GetTotalRowsMode)
	m.server.GetServerInfo(m.db)
	m.currentTableName = ""
	m.totalRowsInCurrentTable = 0
	m.totalRowsInAllTables = 0

	dicTables := m.Export_GetTablesToBeExported()
	for _, v := range dicTables {
		m.totalRowsInAllTables += m.dataBase.GetTable(v.K).GetTotalRows()
	}
	m.currentRowIndexInCurrentTable = 0
	m.currentRowIndexInAllTable = 0
	m.totalTables = 0
	m.currentTableIndex = 0
	return nil
}

func (m *MySqlBackup) Export_GetTablesToBeExported() []*KV {
	dic := make(map[string]string)

	if len(m.exportInfo.TablesToBeExportedDic) == 0 {
		for _, v := range m.dataBase.GetTables() {
			dic[v.GetName()] = fmt.Sprintf("SELECT * FROM `%v`;", v.GetName())
		}
	} else {
		for k, v := range m.exportInfo.TablesToBeExportedDic {
			for _, table := range m.dataBase.GetTables() {
				if k == table.GetName() {
					dic[k] = v
					continue
				}
			}
		}
	}
	//处理外键
	dic2 := make(map[string]string)
	for k := range dic {
		dic2[k] = m.dataBase.GetTable(k).GetCreateTableSql()
	}
	lst := m.Export_ReArrangeDependencies(dic2, "foreign key", "`")
	ret := make([]*KV, 0)
	for _, v := range lst {
		ret = append(ret, &KV{K: v, V: dic[v]})
	}
	return ret
}

func (m *MySqlBackup) Export_ReArrangeDependencies(dic map[string]string, splitKeyword string, keyNameWrapper string) []string {
	splitKeyword = fmt.Sprintf(" %v ", splitKeyword)
	ret := make([]string, 0)
	requireLoop := true
	for requireLoop {
		requireLoop = false
		for k, v := range dic {
			if Contains(ret, k) {
				continue
			}

			allReferencedAdded := true
			createSql := strings.ToLower(v)
			referenceInfo := ""
			referenceTaken := false
			if splitKeyword != "" {
				if strings.Contains(createSql, splitKeyword) {
					sa := strings.Split(createSql, splitKeyword)
					referenceInfo = sa[len(sa)-1]
					referenceTaken = true
				}
			}
			if !referenceTaken {
				referenceInfo = createSql
			}
			for k1 := range dic {
				if k == k1 {
					continue
				}
				if Contains(ret, k1) {
					continue
				}
				_thisTBname := keyNameWrapper + strings.ToLower(k1) + keyNameWrapper
				if strings.Contains(referenceInfo, _thisTBname) {
					allReferencedAdded = false
					break
				}
			}

			if allReferencedAdded {
				if !Contains(ret, k) {
					ret = append(ret, k)
					requireLoop = true
					break
				}
			}

		}

	}

	for k := range dic {
		if !Contains(ret, k) {
			ret = append(ret, k)
		}
	}
	return ret
}

func (m *MySqlBackup) ExportStart() {
	err := m.Export_InitializeVariables()
	if err != nil {
		log.Error("Export_InitializeVariables  %v", err)
		return
	}
	stage := 1
	for stage < 11 {
		if m.stopProcess {
			break
		}
		switch stage {
		case 1:

			m.Export_BasicInfo()
			break
		case 2:

			m.Export_CreateDatabase()
			break
		case 3:
			m.Export_DocumentHeader()
			break
		case 4:
			m.Export_TableRows()
			break
		case 5:
			m.Export_Functions()
			break
		case 10:
			m.Export_DocumentFooter()
		default:
			//todo Export_Procedures
			//todo Export_Events
			//todo Export_Views
			//todo Export_Triggers
			//todo Export_Procedures

			break
		}

		stage = stage + 1
	}

}
func (m *MySqlBackup) OnExportProcess() {
	per := m.currentRowIndexInAllTable * 100 / m.totalRowsInAllTables
	if m.OnExportProcessEvent != nil {
		m.OnExportProcessEvent(int32(per), m.currentTableName)
	}
}

func (m *MySqlBackup) ExportToFile(fileName string) {
	fileObj, err := os.Create(fileName)
	if err != nil {
		log.Error("ExportToFile:%v", err)
		return
	}
	m.writeObj = bufio.NewWriterSize(fileObj, 4096)
	m.ExportStart()
	err = m.writeObj.Flush()
	if err != nil {
		log.Error("ExportToFile flush %v", err)
	}
	err = fileObj.Close()
	if err != nil {

		log.Error("fileObj close  %v", err)
	}
}
func (m *MySqlBackup) Export_BasicInfo() {
	log.Release("导出基础信息")
	m.WriteComment(fmt.Sprintf("MySqlBackup %v", Version))
	if m.exportInfo.RecordDumpTime {
		m.WriteComment(fmt.Sprintf("Dump Time:%v", m.timeStart.Format("2006-01-02 15:04:05")))
	} else {
		m.WriteComment("")
	}
	m.WriteComment("--------------------------------------")
	m.WriteComment(fmt.Sprintf("Server Version %v", m.server.GetVersion()))
}
func (m *MySqlBackup) Export_CreateDatabase() {
	if !m.exportInfo.AddCreateDatabase && !m.exportInfo.AddDropDatabase {
		return
	}
	log.Release("导出创建数据库信息")
	m.WriteLine("")
	m.WriteLine("")
	if m.exportInfo.AddDropDatabase {
		m.WriteLine(fmt.Sprintf("DROP DATABASE `%v`;", m.dataBase.GetName()))
	}
	if m.exportInfo.AddCreateDatabase {
		m.WriteLine(m.dataBase.GetCreateDatabaseSQL())
		m.WriteLine(fmt.Sprintf("Use `%v`;", m.dataBase.GetName()))
	}
	m.WriteLine("")
	m.WriteLine("")
}
func (m *MySqlBackup) Export_DocumentHeader() {
	log.Release("导出文件头信息")
	m.WriteLine("")
	for _, v := range m.exportInfo.GetDocumentHeaders(m.db) {
		m.WriteLine(v)
	}
	m.WriteLine("")
	m.WriteLine("")
}
func (m *MySqlBackup) Export_DocumentFooter() {
	m.WriteLine("")
	for _, v := range m.exportInfo.GetDocumentFooters() {
		m.WriteLine(v)
	}
	m.WriteLine("")
	m.WriteLine("")
}

func (m *MySqlBackup) Export_TableRows() {
	tableList := m.Export_GetTablesToBeExported()
	m.totalTables = int64(len(tableList))
	log.Release("导出表和数据")
	m.pgb = progressbar.NewOptions64(m.totalRowsInAllTables, progressbar.OptionShowIts(), progressbar.OptionShowCount())
	if m.exportInfo.ExportTableStructure || m.exportInfo.ExportRows {

		for _, v := range tableList {

			tableName, selectSQL := v.K, v.V
			exclude := m.Export_ThisTableIsExcluded(tableName)
			if exclude {
				continue
			}
			m.pgb.Describe(tableName)
			m.currentTableName = tableName
			m.currentTableIndex++
			m.totalRowsInCurrentTable = m.dataBase.GetTable(tableName).GetTotalRows()
			if m.exportInfo.ExportTableStructure {
				m.Export_TableStructure(tableName)
			}
			if m.exportInfo.ExportRows {
				m.Export_Rows(tableName, selectSQL)
			}

		}
	}
}
func (m *MySqlBackup) Export_TableStructure(tableName string) {
	if m.stopProcess {
		return
	}
	m.WriteComment("")
	m.WriteComment(fmt.Sprintf("Definition of %v", tableName))
	m.WriteComment("")
	m.WriteLine("")
	if m.exportInfo.AddDropTable {
		m.WriteLine(fmt.Sprintf("DROP TABLE IF EXISTS `%v`;", tableName))
	}
	if m.exportInfo.ResetAutoIncrement {
		m.WriteLine(m.dataBase.GetTable(tableName).GetCreateTableSql())
	} else {
		m.WriteLine(m.dataBase.GetTable(tableName).GetCreateTableSql())
	}
	m.WriteLine("")
	m.writeObj.Flush()

}
func (m *MySqlBackup) Export_Rows(tableName string, selectSQL string) {
	m.WriteComment("")
	m.WriteComment(fmt.Sprintf("Dumping data for table %v", tableName))
	m.WriteComment("")
	m.WriteLine("")
	m.WriteLine(fmt.Sprintf("/*!40000 ALTER TABLE `%v` DISABLE KEYS */;", tableName))
	if m.exportInfo.WrapWithinTransaction {
		m.WriteLine("START TRANSACTION;")
	}

	m.Export_RowsData(tableName, selectSQL)

	if m.exportInfo.WrapWithinTransaction {
		m.WriteLine("COMMIT;")
	}
	m.WriteLine(fmt.Sprintf("/*!40000 ALTER TABLE `%v` ENABLE KEYS */;", tableName))
	m.WriteLine("")
	m.writeObj.Flush()

}

func (m *MySqlBackup) Export_RowsData(tableName string, selectSQL string) {
	m.currentRowIndexInCurrentTable = 0
	if m.exportInfo.RowsExportMode == infoObjects.RowsDataExportMode_Insert ||
		m.exportInfo.RowsExportMode == infoObjects.RowsDataExportMode_InsertIgnore ||
		m.exportInfo.RowsExportMode == infoObjects.RowsDataExportMode_Replace {
		m.Export_RowsData_Insert_Ignore_Replace(tableName, selectSQL)
	} else if m.exportInfo.RowsExportMode == infoObjects.RowsDataExportMode_OnDuplicateKeyUpdate {

	}
}
func (m *MySqlBackup) Export_RowsData_Insert_Ignore_Replace(tableName string, selectSQL string) {
	table := m.dataBase.GetTable(tableName)
	stmt, err := m.db.Prepare(selectSQL)
	if err != nil {
		log.Error("mysql query:%v>>%v", selectSQL, err)
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Error("mysql query:%v>>%v", selectSQL, err)
		return
	}
	defer func() {
		_ = stmt.Close()
		_ = rows.Close()
	}()
	insertStatementHeader := ""
	buffer := bytes.Buffer{}
	for rows.Next() {
		if m.stopProcess {
			return
		}
		m.currentRowIndexInAllTable++
		m.pgb.Add64(1)
		m.currentRowIndexInCurrentTable++
		m.OnExportProcess()
		if insertStatementHeader == "" {
			insertStatementHeader = m.Export_GetInsertStatementHeader(m.exportInfo.RowsExportMode, tableName, rows)
		}
		sqlDataRow := m.Export_GetValueString(rows, table)
		if buffer.Len() == 0 {
			buffer.WriteString(insertStatementHeader)
			buffer.WriteString(sqlDataRow)
		} else if buffer.Len()+len(sqlDataRow) < m.exportInfo.MaxSqlLength {
			buffer.WriteString(",")
			buffer.WriteString(sqlDataRow)
		} else {
			buffer.WriteString(";")
			m.WriteLine(buffer.String())
			_ = m.writeObj.Flush()
			buffer.Reset()
			buffer.WriteString(insertStatementHeader)
			buffer.WriteString(sqlDataRow)
		}
	}
	if buffer.Len() > 0 {
		buffer.WriteString(";")
	}
	m.WriteLine(buffer.String())
	_ = m.writeObj.Flush()
}

func (m *MySqlBackup) Export_GetInsertStatementHeader(rowsExportMode infoObjects.RowsDataExportMode, tableName string, rows *sql.Rows) string {
	buffer := bytes.Buffer{}
	if rowsExportMode == infoObjects.RowsDataExportMode_Insert {
		buffer.WriteString("INSERT INTO `")
	} else if rowsExportMode == infoObjects.RowsDataExportMode_InsertIgnore {
		buffer.WriteString("INSERT IGNORE INTO `")
	} else if rowsExportMode == infoObjects.RowsDataExportMode_Replace {
		buffer.WriteString("REPLACE INTO `")
	}
	buffer.WriteString(tableName)
	buffer.WriteString("`(")
	columns, err := rows.Columns()
	if err != nil {
		log.Error("%v", err)
		return ""
	}
	for i := 0; i < len(columns); i++ {
		colName := columns[i]
		if m.dataBase.GetTable(tableName).GetColumns(colName).GetIsGeneratedColumn() {
			continue
		}
		if i > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString("`")
		buffer.WriteString(colName)
		buffer.WriteString("`")
	}
	buffer.WriteString(") VALUES")
	return buffer.String()

	return ""
}

func (m *MySqlBackup) Export_GetValueString(rows *sql.Rows, table *mysqlObjects.MySqlTable) string {
	buffer := bytes.Buffer{}
	columns, err := rows.Columns()
	if err != nil {
		log.Error("Export_GetValueString %v", err)
		return ""
	}
	size := len(columns)
	pts := make([]interface{}, size)
	container := make([]interface{}, size)

	for i := range pts {
		pts[i] = &container[i]
	}
	err = rows.Scan(pts...)
	if err != nil {
		log.Error("rows scan :%v", err)
		return ""
	}

	for i := 0; i < len(columns); i++ {
		columnName := columns[i]
		if table.GetColumns(columnName).GetIsGeneratedColumn() {
			continue
		}
		if i == 0 {
			buffer.WriteString("(")
		} else {
			buffer.WriteString(",")
		}
		buffer.WriteString(ConvertToSqlFormat(container[i], true, true, table.GetColumns(columnName), m.exportInfo.BlobExportMode))

	}
	buffer.WriteString(")")
	return buffer.String()
}

func (m *MySqlBackup) Export_ThisTableIsExcluded(tableName string) bool {
	tableNameLower := strings.ToLower(tableName)
	for _, v := range m.exportInfo.ExcludeTables {
		if strings.ToLower(v) == tableNameLower {
			return true
		}
	}
	return false
}
func (m *MySqlBackup) Export_Functions() {
	if !m.exportInfo.ExportFunctions || len(m.dataBase.GetFunctions()) == 0 {
		return
	}
	m.WriteComment("")
	m.WriteComment("Dumping functions")
	m.WriteComment("")
	m.WriteLine("")
	for _, function := range m.dataBase.GetFunctions() {
		if m.stopProcess {
			return
		}
		if strings.Trim(function.GetCreateFunctionSQL(), " ") == "" ||
			strings.Trim(function.GetCreateFunctionSqlWithoutDefiner(), " ") == "" {
			continue
		}
		m.WriteLine(fmt.Sprintf("DROP FUNCTION IF EXISTS `%v`;", function.GetName()))
		m.WriteLine("DELIMITER " + m.exportInfo.ScriptsDelimiter)
		if m.exportInfo.ExportRoutinesWithoutDefiner {
			m.WriteLine(function.GetCreateFunctionSqlWithoutDefiner() + " " + m.exportInfo.ScriptsDelimiter)
		} else {
			m.WriteLine(function.GetCreateFunctionSQL() + " " + m.exportInfo.ScriptsDelimiter)
		}
		m.WriteLine("DELIMITER ;")
	}
	m.writeObj.Flush()
}

func (m *MySqlBackup) WriteLine(text string) {
	_, _ = m.writeObj.WriteString(text)
	_, _ = m.writeObj.WriteString("\r\n")
}
func (m *MySqlBackup) WriteComment(text string) {
	m.WriteLine("-- " + text)
}

//通过文件导入
func (m *MySqlBackup) ImportFromFile(fileName string) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Error("ImportFromFile:%v", err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Error("ImportFromFile  %v", err)
		return
	}
	m.totalBytes = fileInfo.Size()
	m.readerObj = bufio.NewReader(file)
	//m.readerObj.ReadString('\n')
	m.Import_Start()
}
func (m *MySqlBackup) Import_Start() {
	err := m.Import_InitializeVariables()
	if err != nil {
		log.Error("Import_Start:%v", err)
		return
	}
	m.pgb = progressbar.NewOptions64(m.totalBytes, progressbar.OptionSetBytes64(m.totalBytes))
	line := ""
	for true {
		if m.stopProcess {
			m.processCompletionType = ProcessEndType_Cancelled
			break
		}

		line = m.Import_GetLine()
		if line == "end" {
			break
		}
		if len(line) == 0 {
			continue
		}
		m.Import_ProcessLine(line)
	}

}
func (m *MySqlBackup) Import_InitializeVariables() error {
	if m.db == nil {
		return errors.New("sql.DB is not initialized. Object not set to an instance of an object.")
	}
	m.stopProcess = false
	m.lastError = nil
	m.timeStart = time.Now()
	m.currentBytes = 0
	m.buffImport = &bytes.Buffer{}
	m.currentProcess = ProcessType_Import
	m.processCompletionType = ProcessEndType_Complete
	m.delimiter = ";"
	m.lastErrorSql = ""
	return nil
}
func (m *MySqlBackup) Import_GetLine() string {
	line, err := m.readerObj.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "end"
		}
		log.Error("Import_GetLine:%v", err)
		return ""
	}
	m.currentBytes += int64(len(line))
	m.pgb.Add64(int64(len(line)))
	line = strings.TrimSpace(line)
	return line
}
func (m *MySqlBackup) Import_ProcessLine(line string) {
	nextAction := m.Import_AnalyseNextAction(line)
	switch nextAction {
	case NextImportAction_Ignore:
		break
	case NextImportAction_AppendLine:
		m.buffImport.WriteString(line)
		m.buffImport.WriteString("\r\n")
		break
	case NextImportAction_ChangeDelimiter:
		m.delimiter = "DELIMITER"
		break
	case NextImportAction_AppendLineAndExecute:
		m.Import_AppendLineAndExecute(line)
		break
	}
}
func (m *MySqlBackup) Import_AnalyseNextAction(line string) NextImportAction {
	len := len(line)
	if len == 0 {
		return NextImportAction_Ignore
	}
	if strings.HasSuffix(line, m.delimiter) {
		return NextImportAction_AppendLineAndExecute
	}
	if strings.HasPrefix(line, "DELIMITER ") {
		return NextImportAction_ChangeDelimiter
	}
	return NextImportAction_AppendLine
}
func (m *MySqlBackup) Import_AppendLineAndExecute(line string) {
	m.buffImport.WriteString(line)
	m.buffImport.WriteString("\r\n")
	importQuery := m.buffImport.String()
	_, err := m.db.Exec(importQuery)
	if err != nil {
		log.Error("Import_AppendLineAndExecute:%v", err)
	}
	m.buffImport.Reset()
}

func ConvertToSqlFormat(ob interface{}, wrapStringWithSingleQuote bool, escapeStringSequence bool, col *mysqlObjects.MySqlColumn, blobExportMode infoObjects.BlobDataExportMode) string {
	buffer := &bytes.Buffer{}
	if ob == nil {
		buffer.WriteString("NULL")
	} else if str, ok := ob.(string); ok {
		if escapeStringSequence {
			str = EscapeStringSequence(str)
		}
		if wrapStringWithSingleQuote {
			buffer.WriteString("'")
		}
		buffer.WriteString(str)
		if wrapStringWithSingleQuote {
			buffer.WriteString("'")
		}
	} else if v, ok := ob.(bool); ok {
		if v {
			buffer.WriteString("1")
		} else {
			buffer.WriteString("0")
		}
	} else if v, ok := ob.([]uint8); ok {
		if col.MySqlDataType == "varchar" || col.MySqlDataType == "text" {
			str := string(v)
			if escapeStringSequence {
				str = EscapeStringSequence(str)
			}
			if wrapStringWithSingleQuote {
				buffer.WriteString("'")
			}
			buffer.WriteString(str)
			if wrapStringWithSingleQuote {
				buffer.WriteString("'")
			}
		} else {
			if col.MySqlDataType != "blob" {
				log.Debug(col.MySqlDataType)
			}
			if len(v) == 0 {
				if wrapStringWithSingleQuote {
					return "''"
				} else {
					return ""
				}
			} else {
				if blobExportMode == infoObjects.BlobDataExportMode_HexString {
					buffer.WriteString(ConvertByteArrayToHexString(v))
				} else if blobExportMode == infoObjects.BlobDataExportMode_BinaryChar {
					if wrapStringWithSingleQuote {
						buffer.WriteString("'")
					}
					for _, v := range v {
						escape_string(buffer, string(v))
					}
					if wrapStringWithSingleQuote {
						buffer.WriteString("'")
					}
				}
			}
		}

	} else if v, ok := ob.(time.Time); ok {
		if wrapStringWithSingleQuote {
			buffer.WriteString("'")
		}
		buffer.WriteString(v.Format("2006-01-02 15:04:05"))
		if wrapStringWithSingleQuote {
			buffer.WriteString("'")
		}
	} else if v, ok := ob.(int); ok {
		buffer.WriteString(convert.ToString(v))
	} else if v, ok := ob.(int64); ok {
		buffer.WriteString(convert.ToString(v))
	} else {
		log.Fatal("类型不存在：%v", reflect.TypeOf(ob))
	}

	return buffer.String()
}
func EscapeStringSequence(data string) string {
	buff := &bytes.Buffer{}
	for _, v := range data {
		escape_string(buff, string(v))
	}
	return buff.String()
}

func escape_string(buff *bytes.Buffer, c string) {
	switch c {
	case "\\":
		buff.WriteString("\\\\")
		break
	case "\r":
		buff.WriteString("\\r")
		break
	case "\n":
		buff.WriteString("\\n")
		break
	case "\b":
		buff.WriteString("\\b")
		break
	case "\"":
		buff.WriteString("\\\"")
		break
	case "'":
		buff.WriteString("''")
		break
	default:
		buff.WriteString(c)
	}
}

func ConvertByteArrayToHexString(ba []byte) string {
	if len(ba) == 0 {
		return ""
	}
	buffer := &bytes.Buffer{}
	buffer.WriteString("0x")
	for _, v := range ba {
		b := v >> 4
		if b > 9 {
			buffer.WriteByte(b + 0x37)
		} else {
			buffer.WriteByte(b + 0x30)
		}
		b = v & 0xf
		if b > 9 {
			buffer.WriteByte(b + 0x37)
		} else {
			buffer.WriteByte(b + 0x30)
		}
	}
	return buffer.String()
}
func Contains(data []string, value string) bool {
	for _, v := range data {
		if v == value {
			return true
		}
	}
	return false
}
