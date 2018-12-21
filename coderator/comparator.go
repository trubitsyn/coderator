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
	"math"
	"os/exec"
	"strconv"
)

type Comparator interface {
	Compare(a string, b string) bool
}

type ExternalComparator struct {
	Command string
	Comparator
}

func (c ExternalComparator) Compare(a string, b string) bool {
	cmd := exec.Command(c.Command, a, b)
	if err := cmd.Run(); err != nil {
		fmt.Println(errors.New("Could not run command!"))
	}

	if err := cmd.Wait(); err != nil {
		panic("Error while waiting!")
	}

	state := cmd.ProcessState
	if state != nil {
		return state.Success()
	}
	fmt.Println(errors.New("Could not execute command!"))
	return false
}

type ExactComparator struct {
	Comparator
}

func (c ExactComparator) Compare(a string, b string) bool {
	return a == b
}

type ApproximateComparator struct {
	Accuracy float64
	Comparator
}

func (c ApproximateComparator) Compare(a string, b string) bool {
	ai, aerr := strconv.ParseFloat(a, 64)
	bi, berr := strconv.ParseFloat(b, 64)

	if aerr != nil || berr != nil {
		return false
	}
	return math.Abs(ai-bi) <= c.Accuracy
}
