package sqlutils

import (
	"strconv"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/outbrain/log"	
)

type RowMap map[string]string

func (this *RowMap) GetString(key string) string {
	return (*this)[key]
}

func (this *RowMap) GetInt64(key string) int64 {
	res, _ := strconv.ParseInt((*this)[key], 10, 0)
	return res
}

func (this *RowMap) GetInt(key string) int {
	res, _ := strconv.Atoi((*this)[key])
	return res
}

func (this *RowMap) GetIntD(key string, def int) int {
	res, err := strconv.Atoi((*this)[key])
	if err != nil {return def}
	return res
}

func (this *RowMap) GetUint(key string) uint {
	res, _ := strconv.Atoi((*this)[key])
	return uint(res)
}

func (this *RowMap) GetBool(key string) bool {
	return (*this)[key] == "1"
}


var knownDBs map[string]*sql.DB = make(map[string]*sql.DB)

func GetDB(mysql_uri string) (*sql.DB, error) {
	
	if _, exists := knownDBs[mysql_uri]; !exists {
	    if db, err := sql.Open("mysql", mysql_uri); err == nil {
	    	knownDBs[mysql_uri] = db
	    } else {
	    	return db, err
	    }	    	    
	}
	return knownDBs[mysql_uri], nil
}


func RowToArray(rows *sql.Rows, columns []string) []string {
    buff := make([]interface{}, len(columns))
    data := make([]string, len(columns))
    for i, _ := range buff {
        buff[i] = &data[i]
    }
	rows.Scan(buff...)
	return data
}

func ScanRowsToArrays(rows *sql.Rows, on_row func([]string) error) error {
    columns, _ := rows.Columns()
    for rows.Next() {
    	arr := RowToArray(rows, columns)
	    
	    err := on_row(arr)
	    if err != nil {
	    	return err
	    }
    }
    return nil
}

func ScanRowsToMaps(rows *sql.Rows, on_row func(RowMap) error) error {
	columns, _ := rows.Columns()
	err := ScanRowsToArrays(rows, func(arr []string) error {
    	m := make(map[string]string)	 
	    for k, data_col := range arr {
	        m[columns[k]] = data_col
	    }
	    err := on_row(m)
	    if err != nil {
	    	return err
	    }
	    return nil
	})
	return err
}


func QueryRowsMap(db *sql.DB, query string, on_row func(RowMap) error) error {
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil && err != sql.ErrNoRows {
		return log.Errore(err)
	} 	
	err = ScanRowsToMaps(rows, on_row)
	return err
}



func Exec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
    stmt, err := db.Prepare(query);
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var res sql.Result
	res, err = stmt.Exec(args...)
	if err != nil {
		log.Errore(err)
	}
	return res, err
}
