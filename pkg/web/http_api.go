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

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/client"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/vars"
)

const MB512 = 512 * 1024 * 1024

func Start(port int) error {
	tDB, err := bbolt.Open(vars.TorrentDBPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return errors.Wrap(err, "failed to open torrent database")
	}
	defer tDB.Close()

	iDB, err := bbolt.Open(vars.IndexesBoltPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return errors.Wrap(err, "failed to open torrent database")
	}
	defer iDB.Close()

	c, err := client.GetClient()
	if err != nil {
		return errors.Wrap(err, "failed to start BitTorrent client")
	}
	defer c.Close()

	fmt.Printf("Web-UI running on http://127.0.0.1:%d/\n", port)
	err = New(tDB, iDB, c).Listen(":" + strconv.Itoa(port))

	return errors.Wrap(err, "failed to start http server")
}

func New(tDB, iDB *bbolt.DB, c *torrent.Client) *fiber.App {
	app := fiber.New(
		fiber.Config{
			// Views:          engine,
			DisableStartupMessage: true,
			ReadBufferSize:        MB512,
			BodyLimit:             MB512,
			ErrorHandler:          errorHandler,
		})

	setupRouter(app, &handler{torrentDB: tDB, indexesDB: iDB, btClient: c, m: &sync.Mutex{}})
	app.Use("/", filesystem.New(filesystem.Config{
		Root: pkger.Dir("/frontend/dist/"),
	}))
	app.Use("*", func(c *fiber.Ctx) error {
		f, err := pkger.Open("/frontend/dist/index.html")
		if err != nil {
			return errors.Wrap(err, "failed to open index.html")
		}
		defer f.Close()

		return c.Type("html").SendStream(f)
	})

	return app
}

func setupRouter(app *fiber.App, h *handler) {
	router := app.Group("/api/v0")
	router.Get("/debug", func(c *fiber.Ctx) error {
		return c.JSON(DebugInfo{
			Version:   vars.Ref,
			Commit:    vars.Commit,
			Builder:   vars.Builder,
			BuildTime: vars.BuildTime,
			Os:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			BaseDir:   vars.GetAppBaseDir(),
		})
	})
	router.Post("/", h.index)
	router.Get("/torrent", h.torrentGet)
	router.Put("/torrent", h.torrentUpload)
	router.Put("/index", h.indexesUpload)
	router.Get("/paper", h.paperQuery)
}

func errorHandler(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// Default 500 StatusCode
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if ok := errors.Is(err, e); ok {
		// Override status code if fiber.Error type
		code = e.Code //nolint:govet
	}
	// Set Content-Type: application/json; charset=utf-8
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

	// Return StatusCode with error message
	return c.Status(code).JSON(Error{Status: "error", Message: err.Error()})
}
