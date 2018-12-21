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
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var queue = make(map[uint64]bool)
var verificationResults = make(map[uint64][]bool)

type VerificationResult int

const (
	Success       VerificationResult = 200
	TestFailed    VerificationResult = 300
	BadSource     VerificationResult = 400
	InternalError VerificationResult = 500
)

func Queue(task Task) {
	queue[task.Id] = true
}

func Dequeue(task Task) {
	delete(queue, task.Id)
}

func IsVerificationQueued(task Task) bool {
	_, exists := queue[task.Id]
	return exists
}

func HasVerificationCompleted(task Task) bool {
	return GetVerificationResults(task) != nil
}

func GetVerificationResults(task Task) []bool {
	return verificationResults[task.Id]
}

func GetTempFileName(task Task) string {
	return "tmp" + fmt.Sprint(task.Id)
}

func SaveSolution(task Task, file multipart.File) (string, error) {
	tmpFileName := GetTempFileName(task)
	tmpFile, err := os.OpenFile(tmpFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	io.Copy(tmpFile, file)
	dir, err := filepath.Abs(filepath.Dir(tmpFile.Name()))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	path := dir + string(filepath.Separator) + tmpFile.Name()
	return path, nil
}

func RemoveTempFiles(task Task) {
	tmpFileName := GetTempFileName(task)
	err := os.Remove(tmpFileName)
	if err != nil {
		fmt.Println(err)
	}
}

func VerifyTaskSolution(task Task, filepath string) VerificationResult {
	defer RemoveTempFiles(task)
	Queue(task)
	defer Dequeue(task)

	tests, err := database.FindTestsByTaskId(task.Id)
	if err != nil {
		fmt.Println(err)
		verificationResults[task.Id] = nil
		return InternalError
	}

	config := Config{}
	appConfig, err := config.ApplicationConfig()
	if err != nil {
		return InternalError
	}

	processor := appConfig.FindProcessorByName(task.Processor)
	if processor == nil {
		return InternalError
	}

	sourceValidator := FindSourceValidatorByProcessor(*processor)
	if !sourceValidator.Valid() {
		return BadSource
	}

	tester := Tester{}
	results := tester.RunTests(*processor, filepath, tests)
	verificationResults[task.Id] = results

	for _, result := range results {
		if !result {
			return TestFailed
		}
	}
	return Success
}
