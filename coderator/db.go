/*
 * Copyright (C) 2018 Nikola Trubitsyn
 *
 * This file is part of coderator.
 *
 * coderator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * coderator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with coderator.  If not, see <https://www.gnu.org/licenses/>.
 */

package coderator

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

func NewDatabase(dataSourceName string) (*Database, error) {
	connStr := "user=postgres dbname=tasks sslmode=disable " + dataSourceName
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Database{db}, nil
}

func (db *Database) FindTaskById(id uint64) (*Task, error) {
	statement := "SELECT title, text FROM tasks WHERE id = $1"
	var task Task
	row := db.QueryRow(statement, id)
	err := row.Scan(id, &task.Title, &task.Text)
	return &task, err
}

func (db *Database) AddTask(task Task) {
	statement := "INSERT INTO tasks (title, text) VALUES ($1, $2)"
	_, err := db.Exec(statement, task.Title, task.Text)
	if err != nil {
		panic(err)
	}
}

func (db *Database) AllTasks() ([]Task, error) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		task := new(Task)
		err := rows.Scan(&task.Title, &task.Text)

		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (db *Database) FindTestsByTaskId(taskId uint64) ([]Test, error) {
	statement := "SELECT * from tests WHERE taskId = $1"
	rows, err := db.Query(statement, taskId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tests := make([]Test, 0)
	for rows.Next() {
		test := new(Test)
		err := rows.Scan(&test.Id, &test.Input, &test.Output)
		if err != nil {
			return nil, err
		}
		tests = append(tests, *test)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tests, nil
}
