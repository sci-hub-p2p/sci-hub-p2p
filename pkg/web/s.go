// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.

package web

type Error struct {
	Message *string `json:"message,omitempty"`
	Status  string  `json:"status"`
}

type RequestEmptyBody Error

type RequestWrongBodyEncoding Error

type JSON409 struct {
	Data struct {
		InfoHash *string `json:"info_hash,omitempty"`
	} `json:"data,omitempty"`
	Error `yaml:",inline"`
}
