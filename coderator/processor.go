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
	"os/exec"
	"fmt"
)

type LanguageProcessor struct {
	Name    string
	Version string
	Path    string
	Exec    string
}

func (processor LanguageProcessor) RunFile(args ...string) (string, error) {
	out, err := exec.Command(processor.Path, args...).Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(out), nil
}
