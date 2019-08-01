package mysqlObjects

import (
	"common/convert"
	"common/db"
	"database/sql"
	"fmt"
	"strings"
)

type MySqlServer struct {
	_versionNumber          string
	_edition                string
	_majorVersionNumber     float64
	_characterSetServer     string
	_characterSetSystem     string
	_characterSetConnection string
	_characterSetDatabase   string
	_currentUser            string
	_currentUserClientHost  string
	_currentClientHost      string
}

func NewMysqlServer() *MySqlServer {
	return new(MySqlServer)
}
func (s *MySqlServer) GetServerInfo(_db *sql.DB) {
	s._edition = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'version_comment';", 1)
	s._versionNumber = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'version';", 1)
	s._characterSetServer = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'character_set_server';", 1)
	s._characterSetSystem = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'character_set_system';", 1)
	s._characterSetConnection = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'character_set_connection';", 1)
	s._characterSetDatabase = db.ExecuteScalarStr(_db, "SHOW variables LIKE 'character_set_database';", 1)
	s._currentUserClientHost = db.ExecuteScalarStr(_db, "SELECT current_user;", 0)
	ca := strings.Split(s._currentUserClientHost, "@")
	s._currentUser = ca[0]
	s._currentClientHost = ca[1]

	vsa := strings.Split(s._versionNumber, ".")
	v := ""
	if len(vsa) > 1 {
		v = vsa[0] + "." + vsa[1]
	} else {
		v = vsa[0]
	}
	s._majorVersionNumber = convert.ToFloat64(v)
}
func (s *MySqlServer) GetVersion() string {
	return fmt.Sprintf("%v %v", s._versionNumber, s._edition)
}
