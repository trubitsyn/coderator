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
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// FIXME: Performance of concurrent access heavily depends on disk I/O?

const (
	TasksDir  = "tasks"
	Extension = ".yml"
)

type Config struct {
}

type TaskConfig struct {
	Task  Task
	Tests []Test
}

type ApplicationConfig struct {
	Processors []LanguageProcessor
}

func (config Config) ApplicationConfig() (*ApplicationConfig, error) {
	data, err := ioutil.ReadFile("serve" + Extension)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	applicationConfig := ApplicationConfig{}
	if err = yaml.Unmarshal(data, &applicationConfig); err != nil {
		return nil, err
	}
	return &applicationConfig, nil
}

func (config Config) filenameForTask(task Task) string {
	lower := strings.ToLower(task.Title)
	underscores := strings.Replace(lower, " ", "_", -1)
	extension := underscores + Extension
	return extension
}

func (config Config) AllTasks() ([]Task, error) {
	files, err := ioutil.ReadDir(TasksDir)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0)
	for _, file := range files {
		taskConfig := TaskConfig{}
		data, err := ioutil.ReadFile(TasksDir + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		if err = yaml.Unmarshal(data, &taskConfig); err != nil {
			fmt.Println(err)
			continue
		}
		tasks = append(tasks, taskConfig.Task)
	}
	return tasks, nil
}

func (config Config) FindTaskById(id uint64) (*Task, error) {
	tasks, err := config.AllTasks()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.Id == id {
			return &task, nil
		}
	}
	return nil, errors.New("No task found")
}

func (config Config) AddTask(task Task) {
	panic(errors.New("Not implemented"))
}

func (config Config) FindTestsByTaskId(id uint64) ([]Test, error) {
	files, err := ioutil.ReadDir(TasksDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		taskConfig := TaskConfig{}
		data, err := ioutil.ReadFile(TasksDir + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		if err = yaml.Unmarshal(data, &taskConfig); err != nil {
			fmt.Println(err)
			continue
		}

		if taskConfig.Task.Id == id {
			tests := taskConfig.Tests
			return tests, nil
		}
	}
	return nil, errors.New("No tests found")
}

func (config ApplicationConfig) FindProcessorByName(name string) *LanguageProcessor {
	for _, processor := range config.Processors {
		if processor.Name+processor.Version == name {
			return &processor
		}
	}
	return nil
}
