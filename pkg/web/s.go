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
	Message string `json:"message,omitempty"`
	Status  string `json:"status"`
}

type RequestEmptyBody Error

type RequestWrongBodyEncoding Error

type DebugInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Builder   string `json:"builder"`
	BuildTime string `json:"build_time"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
	BaseDir   string `json:"base_dir"`
}

type WithData struct {
	Data interface{} `json:"data"`
}

type ErrWithData struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
	Status  string      `json:"status"`
}

type D = map[string]string
