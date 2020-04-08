package db

import (
    "strconv"
    "database/sql"
    "fmt"
    "log"
    "time"
    // _ "github.com/go-sql-driver/mysql"
)


var mysqlDb *sql.DB
var mysqlErr error

const (
    USER_NAME = "root"
    PASS_WORD = "123456"
    HOST      = "localhost"
    PORT      = "3306"
    DATABASE  = "minigame"
    CHARSET   = "utf8"
)

func query(sql string, callback func(err string, rows *sql.Rows)) {
    if callback != nil{
		callback("error", nil)
    }
    
    rows, err := mysqlDb.Query(sql)
    log.Println(err)
    if callback != nil{
        if err != nil{
            callback("error", nil)
        }else{
            callback("", rows)
        }
    }
}

func createDB() {
    query(`create table IF NOT EXISTS user (
        id INTEGER PRIMARY  KEY     AUTOINCREMENT,
        username            TEXT    NOT NULL,
        nickname            TEXT,
        avatar              CHAR(128),
        gender              INT
        )`, nil)

    query(`create table IF NOT EXISTS userinfo (
        userid              INT     NOT NULL,
        gameid              INT     NOT NULL,
        score               INT     NOT NULL,
        exp                 INT     NOT NULL
        )`,nil)
}

func InitDB () {
    dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", USER_NAME, PASS_WORD, HOST, PORT, DATABASE, CHARSET)

    mysqlDb, mysqlErr = sql.Open("mysql", dbDSN)
    //defer mysqlDb.Close();
    if mysqlErr != nil {
        log.Println("dbDSN: " + dbDSN)
        panic("db config invalid: " + mysqlErr.Error())
    }

    mysqlDb.SetMaxOpenConns(100)
    mysqlDb.SetMaxIdleConns(20)
    mysqlDb.SetConnMaxLifetime(100*time.Second)

    if mysqlErr = mysqlDb.Ping(); nil != mysqlErr {
        panic("connect db failed: " + mysqlErr.Error())
    }

    createDB()
}

func CreateUser (username string, nickname string, gender int, avatar string, callback func(id int)) {
	query("INSERT INTO user (username,nickname,avatar,gender) VALUES (\""+ username + "\", \"" + nickname + "\", " + strconv.Itoa(gender) + ", \"" + avatar + "\" )",
		func (err string, rows *sql.Rows) {
        if err != ""{
			callback(-1)
		}else{
            query("SELECT * FROM user", func (err string, rows *sql.Rows) {
                if err != ""{
					callback(-1)
				}else{
					// callback(len(rows))
				}
			})
		}
    })
}

func GetUserInfo  (userid int, gameid int, callback func(data map[string]interface{})) {
	query("SELECT * FROM userinfo WHERE userid = " + strconv.Itoa(userid) + " AND gameid = " + strconv.Itoa(gameid) + "",
	 	func (err string, rows *sql.Rows) {
        if err != ""{
			callback(nil)
		}else{
			// callback(rows[0])
		}
    })
}

func GetScore (userid int, gameid int, callback func(score int)) {
	query("SELECT score FROM userinfo WHERE userid = " + strconv.Itoa(userid) + " AND gameid = " + strconv.Itoa(gameid) + "", 
		func (err string, rows *sql.Rows) {
        if err != ""{
			callback(-1)
		}else{
			// callback(rows[0])
		}
    })
}

func GetExp (userid int, gameid int, callback func(exp int)) {
	query("SELECT exp FROM userinfo WHERE userid = " + strconv.Itoa(userid) + " AND gameid = " + strconv.Itoa(gameid) + "", 
		func (err string, rows *sql.Rows) {
        if err != ""{
			callback(-1)
		}else{
			// callback(rows[0])
		}
    })
}

func UpdateScore (userid int, gameid int, score int) {
    GetScore(userid, gameid, func (ret int) {
        if ret >= 0 && ret < score {
            GetUserInfo(userid, gameid, func (ret map[string]interface{}) {
                if ret != nil {
                    query("INSERT INTO userinfo (userid,gameid,score,exp) VALUES (" + strconv.Itoa(userid) + ", " + strconv.Itoa(gameid) + ", " + strconv.Itoa(score) + ", 0)", nil)
                }else {
                    query("UPDATE userinfo SET score = " + strconv.Itoa(score) + " WHERE userid = " + strconv.Itoa(userid) + " AND gameid = " + strconv.Itoa(gameid) + "", nil)
                }
            })
        }
    })
}

func AddExp (userid int, gameid int, add int) {
    if add < 0{
        return
	}

    GetUserInfo(userid, gameid, func (ret map[string]interface{}) {
        if ret != nil {
			query("INSERT INTO userinfo (userid,gameid,score,exp) VALUES (" + strconv.Itoa(userid) + ", " + strconv.Itoa(gameid) + ", 0, " + strconv.Itoa(add) + ")", nil)
        }else {
            GetExp(userid, gameid, func (ret int) {
                if ret >= 0 {
					query("UPDATE userinfo SET exp = " + strconv.Itoa(ret) + strconv.Itoa(add) + " WHERE userid = " + strconv.Itoa(userid) + " AND gameid = " + strconv.Itoa(gameid) + "", nil)
                }
            })
        }
    })
}
