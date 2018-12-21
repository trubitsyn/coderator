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

package main

import (
	"flag"
	"fmt"
	"github.com/trubitsyn/coderator/coderator"
)

func main() {
	fmt.Println("coderator started.")

	port := flag.Int("port", 8080, "Port")
	useDatabase := flag.Bool("database", false, "Use relational database")

	flag.Parse()

	var dataSource coderator.Repository

	if *useDatabase {
		db, err := coderator.NewDatabase("coderator")

		if err != nil {
			panic(err)
		}

		dataSource = db
	} else {
		dataSource = coderator.Config{}
	}

	coderator.Serve(dataSource, *port)
}
