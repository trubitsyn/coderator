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

package api

import (
	"fmt"
)

type Test struct {
	Id         uint64
	Input      string
	Output     string
	Comparator Comparator
}

type TestResult struct {
	Successful bool
}

type Tester struct {
	// TODO
}

func (t Tester) RunTest(processor LanguageProcessor, filepath string, test Test) bool {
	out, err := processor.RunFile(filepath, test.Input)

	if err != nil {
		fmt.Println(err)
		return false
	}
	return test.Comparator.Compare(string(out), test.Output)
}

func (t Tester) RunTests(processor LanguageProcessor, filepath string, tests []Test) []bool {
	results := make([]bool, 0)
	for _, test := range tests {
		func(test Test) {
			results = append(results, t.RunTest(processor, filepath, test))
		}(test)
	}
	return results
}

type TimesTester struct {
	Times uint64
	Tester
}

func (t TimesTester) RunTests(processor LanguageProcessor, filepath string, tests[]Test) []bool {
	results := make([]bool, 0)
	for _, test := range tests {
		func(test Test) {
			var i uint64
			for i = 0; i < t.Times; i++ {
				if !t.RunTest(processor, filepath, test) {
					results = append(results, false)
				}
			}
			results = append(results, true)
		}(test)
	}
	return results
}