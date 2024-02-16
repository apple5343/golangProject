package db

import (
	"database/sql"
	"log"
	"server/config"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqlDB struct {
	database *sql.DB
}

func NewDatabase(config config.Config) (*SqlDB, error) {
	database, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		return nil, err
	}
	db := SqlDB{database: database}
	return &db, nil
}

func (db *SqlDB) AddTask(task string, created time.Time) (int, error) {
	statement, err := db.database.Prepare("INSERT INTO tasks (expression, status, result, created, lastPing, lastStep) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer statement.Close()
	res, err := statement.Exec(task, "processing", "", created.Format("2006-01-02 15:04:05"), "", task)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastInsertId), nil
}

func (db *SqlDB) SetResult(taskId int, result string) error {
	statement, err := db.database.Prepare("UPDATE tasks SET result = ? WHERE id = ?")
	defer statement.Close()
	_, err = statement.Exec(result, taskId)
	if err != nil {
		return err
	}
	return nil
}

func (db *SqlDB) SetStatus(taskId int, newStatus string) error {
	statement, err := db.database.Prepare("UPDATE tasks SET status = ? WHERE id = ?")
	defer statement.Close()
	_, err = statement.Exec(newStatus, taskId)
	if err != nil {
		return err
	}
	return nil
}

func (db *SqlDB) GetAllTasks() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	rows, err := db.database.Query("SELECT * FROM tasks")
	if err != nil {
		return results, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return results, err
	}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			log.Fatal(err)
		}
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := pointers[i].(*interface{})
			rowMap[colName] = *val
		}

		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func (db *SqlDB) GetDelays() (map[string]int, error) {
	result := make(map[string]int)
	rows, err := db.database.Query("SELECT * FROM delays")
	if err != nil {
		return result, err
	}
	defer rows.Close()
	var op string
	var delay int
	for rows.Next() {
		err := rows.Scan(&op, &delay)
		if err != nil {
			return result, err
		}
		result[op] = delay
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func (db *SqlDB) UpdateDelays(newDelays map[string]int) error {
	for k, v := range newDelays {
		statement, err := db.database.Prepare("UPDATE delays SET delay = ? WHERE operation = ?")
		defer statement.Close()
		_, err = statement.Exec(v, k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *SqlDB) AddSubtask(value string, tim time.Time, parentId int, updated string) error {
	statement, err := db.database.Prepare("INSERT INTO subtasks (value, time, parentId, result) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()
	_, err = statement.Exec(value, tim.Format("2006-01-02 15:04:05"), parentId, updated)
	if err != nil {
		return err
	}
	return err
}

func (db *SqlDB) GetSubtasks(parentId int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	rows, err := db.database.Query("SELECT * FROM subtasks WHERE parentId = ?", parentId)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return results, err
	}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			log.Fatal(err)
		}
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := pointers[i].(*interface{})
			rowMap[colName] = *val
		}

		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func (db *SqlDB) GetTaskById(taskId int) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	var id int
	var expression, status, result, created, lastPing, lastStep string
	row := db.database.QueryRow("SELECT * FROM tasks WHERE id = ?", taskId)
	row.Scan(&id, &expression, &status, &result, &created, &lastPing, &lastStep)
	res["id"] = id
	res["expression"] = expression
	res["status"] = status
	res["result"] = result
	res["created"] = created
	res["lastPing"] = lastPing
	res["lastStep"] = lastStep
	subtasks, err := db.GetSubtasks(taskId)
	if err != nil {
		return res, err
	}
	res["subtasks"] = subtasks
	return res, nil
}

func (db *SqlDB) UpdatePing(taskId int, tim time.Time) error {
	statement, err := db.database.Prepare("UPDATE tasks SET lastPing = ? WHERE id = ?")
	defer statement.Close()
	_, err = statement.Exec(tim.Format("2006-01-02 15:04:05"), taskId)
	if err != nil {
		return err
	}
	return nil
}

func (db *SqlDB) GetInterruptedTasks() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	rows, err := db.database.Query("SELECT * FROM tasks WHERE status = ?", "processing")
	if err != nil {
		return results, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return results, err
	}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			log.Fatal(err)
		}
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := pointers[i].(*interface{})
			rowMap[colName] = *val
		}

		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func (db *SqlDB) UpdateLastStep(id int, step string) error {
	statement, err := db.database.Prepare("UPDATE tasks SET lastStep = ? WHERE id = ?")
	defer statement.Close()
	_, err = statement.Exec(step, id)
	if err != nil {
		return err
	}
	return nil
}
