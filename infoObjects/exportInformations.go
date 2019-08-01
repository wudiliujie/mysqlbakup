package infoObjects

import (
	"common/db"
	"database/sql"
	"fmt"
)

type BlobDataExportMode int32

const (
	BlobDataExportMode_HexString  BlobDataExportMode = 1
	BlobDataExportMode_BinaryChar BlobDataExportMode = 2
)

type GetTotalRowsMethod int32

const (
	GetTotalRowsMethod_Skip              GetTotalRowsMethod = 1
	GetTotalRowsMethod_InformationSchema GetTotalRowsMethod = 2
	GetTotalRowsMethod_SelectCount       GetTotalRowsMethod = 3
)

type RowsDataExportMode int32

const (
	RowsDataExportMode_Insert               RowsDataExportMode = 1
	RowsDataExportMode_InsertIgnore         RowsDataExportMode = 2
	RowsDataExportMode_Replace              RowsDataExportMode = 3
	RowsDataExportMode_OnDuplicateKeyUpdate RowsDataExportMode = 4
	RowsDataExportMode_Update               RowsDataExportMode = 5
)

type ExportInformations struct {
	//BlobExportMode = BlobDataExportMode.BinaryChar is disabled by default as this feature is under development. Set this value to true if you wish continue to export BLOB into binary string/char format. This is temporary available for debugging and development purposes.
	BlobExportModeForBinaryStringAllow bool
	//Gets or Sets a enum value indicates how the BLOB should be exported. HexString = Hexa Decimal String (default); BinaryChar = char format.
	BlobExportMode            BlobDataExportMode
	IntervalForProgressReport int
	GetTotalRowsMode          GetTotalRowsMethod
	// Gets or Sets the tables that will be exported with custom SELECT defined. If none or empty, all tables and rows will be exported. Key = Table's _name. Value = Custom SELECT Statement. Example 1: SELECT * FROM `product` WHERE `category` = 1; Example 2: SELECT `name`,`description` FROM `product`;
	TablesToBeExportedDic map[string]string
	RecordDumpTime        bool
	AddCreateDatabase     bool
	AddDropDatabase       bool
	documentHeaders       []string
	documentFooters       []string
	ExportTableStructure  bool
	ExportRows            bool
	ExcludeTables         []string
	AddDropTable          bool
	ResetAutoIncrement    bool
	WrapWithinTransaction bool

	RowsExportMode               RowsDataExportMode
	MaxSqlLength                 int
	ExportFunctions              bool
	ScriptsDelimiter             string
	ExportRoutinesWithoutDefiner bool
}

func (e *ExportInformations) Init() {
	e.BlobExportMode = BlobDataExportMode_HexString
	e.BlobExportModeForBinaryStringAllow = false
	e.IntervalForProgressReport = 100
	e.GetTotalRowsMode = GetTotalRowsMethod_InformationSchema
	e.TablesToBeExportedDic = make(map[string]string)
	e.RecordDumpTime = true
	e.AddCreateDatabase = false
	e.AddDropDatabase = false
	e.ExportRows = true
	e.ExportTableStructure = true
	e.ExcludeTables = make([]string, 0)
	e.AddDropTable = true
	e.ResetAutoIncrement = false
	e.WrapWithinTransaction = false
	e.RowsExportMode = RowsDataExportMode_Insert
	e.MaxSqlLength = 1024 * 1024
	e.ExportFunctions = true
	e.ScriptsDelimiter = "|"
	e.ExportRoutinesWithoutDefiner = true
}

func (e *ExportInformations) GetDocumentHeaders(_db *sql.DB) []string {
	if e.documentHeaders == nil {
		databaseCharSet := db.ExecuteScalarStr(_db, "SHOW variables LIKE 'character_set_database';", 1)
		e.documentHeaders = append(e.documentHeaders, "/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;")
		e.documentHeaders = append(e.documentHeaders, "/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;")
		e.documentHeaders = append(e.documentHeaders, "/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;")
		e.documentHeaders = append(e.documentHeaders, fmt.Sprintf("/*!40101 SET NAMES %v */;", databaseCharSet))
		e.documentHeaders = append(e.documentHeaders, "/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;")
		e.documentHeaders = append(e.documentHeaders, "/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;")
		e.documentHeaders = append(e.documentHeaders, "/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;")
		e.documentHeaders = append(e.documentHeaders, "/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;")
	}
	return e.documentHeaders
}
func (e *ExportInformations) GetDocumentFooters() []string {
	if e.documentFooters == nil {
		e.documentFooters = append(e.documentFooters, "/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;")
		e.documentFooters = append(e.documentFooters, "/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;")
		e.documentFooters = append(e.documentFooters, "/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;")
		e.documentFooters = append(e.documentFooters, "/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;")
		e.documentFooters = append(e.documentFooters, "/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;")
		e.documentFooters = append(e.documentFooters, "/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;")
		e.documentFooters = append(e.documentFooters, "/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;")
	}
	return e.documentFooters
}

func NewExportInformations() *ExportInformations {
	v := new(ExportInformations)
	v.Init()

	return v
}
