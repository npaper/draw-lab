package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Wallpaper ...
type Wallpaper struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	R          string `json:"r"`
	G          string `json:"g"`
	B          string `json:"b"`
	Tags       string `json:"tags"`
	Ctime      int64  `json:"ctime"`
	Zang       int    `json:"zang"`
	DefaultSet string `json:"default_set"`
	Extra      string `json:"extra"`
	Publish    bool   `json:"publish"`
}

// NewWallpaper ...
func NewWallpaper() *Wallpaper {
	return &Wallpaper{
		Ctime:   time.Now().Unix(),
		Publish: false,
	}
}

// CheckWallpaperTable ...
func CheckWallpaperTable(conn *sqlx.DB) {
	result, err := conn.Exec("create table if not exists `wallpaper` (  `id` int(11)  primary key  auto_increment, `name` varchar(20), `r` varchar(255), `g` varchar(255), `b` varchar(255), `tags` varchar(128), `ctime` bigint, `zang` int(11), `default_set` varchar(128), `extra` varchar(128), `publish` tinyint ) ENGINE=InnoDB DEFAULT CHARSET=utf8;")

	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("影响行数为", num)
}

// Insert ...
func (w *Wallpaper) Insert(conn *sqlx.DB) (int64, error) {
	tx, err := conn.Begin()
	if err != nil {
		return 0, errors.New("开启事物失败")
	}
	sql := "insert into `wallpaper`(`name`,`r`,`g`,`b`,`tags`,`ctime`,`zang`,`default_set`,`extra`,`publish`)values(?,?,?,?,?,?,?,?,?,?)"
	result, err := conn.Exec(sql, w.Name, w.R, w.G, w.B, w.Tags, w.Ctime, w.Zang, w.DefaultSet, w.Extra, w.Publish)
	if err != nil {
		fmt.Println(err)

		//回滚
		err = tx.Rollback()
		if err != nil {
			fmt.Println(err)
		}
		return 0, err
	}
	tx.Commit()
	return result.LastInsertId()
}

// ZangWallpaper ...
func ZangWallpaper(conn *sqlx.DB, id int64) (int64, error) {
	tx, err := conn.Begin()
	if err != nil {
		return 0, errors.New("开启事物失败")
	}
	sql := "update `wallpaper` set `zang` = `zang` + 1 where `id` = ? "
	result, err := conn.Exec(sql, id)
	if err != nil {
		fmt.Println(err)

		//回滚
		err = tx.Rollback()
		if err != nil {
			fmt.Println(err)
		}
		return 0, err
	}
	tx.Commit()
	return result.RowsAffected()
}

// ListWallpaper ...
func ListWallpaper(conn *sqlx.DB, page int, limit int) ([]*Wallpaper, error) {
	_sql := "select `id`,`name`,`r`,`g`,`b`,`tags`,`ctime`,`zang`,`default_set`,`extra`,`publish` from `wallpaper` where `publish` = 1 order by `ctime` desc limit ?,? "
	rows, err := conn.Query(_sql, page, limit)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	papers := make([]*Wallpaper, 0)
	for rows.Next() {
		var name = new(sql.NullString)
		var r = new(sql.NullString)
		var g = new(sql.NullString)
		var b = new(sql.NullString)
		var tags = new(sql.NullString)
		var defaultSet = new(sql.NullString)
		var extra = new(sql.NullString)
		var id = new(sql.NullInt64)
		var zang = new(sql.NullInt64)
		var ctime = new(sql.NullInt64)
		var publish = new(sql.NullBool)

		rows.Scan(&id, &name, &r, &g, &b, &tags, &ctime, &zang, &defaultSet, &extra, &publish)

		paper := Wallpaper{}
		var getInt64 = func(i *sql.NullInt64) int64 {
			val, _ := i.Value()
			if val != nil {
				if v, ok := val.(int64); ok {
					return int64(v)
				}
			}
			return int64(0)
		}

		var getString = func(i *sql.NullString) string {
			val, _ := i.Value()
			if val != nil {
				if v, ok := val.(string); ok {
					return string(v)
				}
			}
			return ""
		}

		var getInt = func(i *sql.NullInt64) int {
			val, _ := i.Value()
			if val != nil {
				if v, ok := val.(int64); ok {
					return int(v)
				}
			}
			return int(0)
		}

		var getBool = func(i *sql.NullBool) bool {
			val, _ := i.Value()
			if val != nil {
				if v, ok := val.(bool); ok {
					return bool(v)
				}
			}
			return false
		}

		paper.ID = getInt64(id)
		paper.Name = getString(name)
		paper.R = getString(r)
		paper.G = getString(g)
		paper.B = getString(b)
		paper.Tags = getString(tags)
		paper.Ctime = getInt64(ctime)
		paper.Zang = getInt(zang)
		paper.DefaultSet = getString(defaultSet)
		paper.Extra = getString(extra)
		paper.Publish = getBool(publish)

		papers = append(papers, &paper)
	}
	rows.Close()
	return papers, nil
}
