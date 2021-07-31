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
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/client"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/persist"
)

type handler struct {
	torrentDB *bbolt.DB
	indexesDB *bbolt.DB
}

func newHandler(tDB, iDB *bbolt.DB) *handler {
	return &handler{torrentDB: tDB, indexesDB: iDB}
}

func (h *handler) index(c *fiber.Ctx) error {
	var s = struct {
		TorrentDB bool
		IndexesDB bool
	}{}

	err := h.torrentDB.View(func(tx *bbolt.Tx) error {
		if tx.Bucket(consts.TorrentBucket()) != nil {
			s.TorrentDB = true
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "can't open torrent database")
	}

	err = h.indexesDB.View(func(tx *bbolt.Tx) error {
		if tx.Bucket(consts.IndexBucketName()) != nil {
			s.IndexesDB = true
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "can't open indexes database")
	}

	return c.Render("index", s)
}

func (h *handler) torrentUpload(c *fiber.Ctx) error {
	mh, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	files, ok := mh.File["files"]
	if !ok {
		return c.SendString("can't find any uploaded file")
	}

	err = h.torrentDB.Batch(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(consts.TorrentBucket())
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in the database")
		}
		for _, file := range files {
			err := func() error {
				f, err := file.Open()
				if err != nil {
					return errors.Wrap(err, "failed to read uploaded file content")
				}
				defer f.Close()

				raw, err := io.ReadAll(f)
				if err != nil {
					return errors.Wrap(err, "failed to read uploaded file content")
				}

				return errors.Wrapf(persist.SaveTorrent(b, raw), "failed to add torrent %s", file.Filename)
			}()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return c.SendString(fmt.Sprintf("%d torrent uploaded", len(files)))
}

func (h *handler) indexesUpload(c *fiber.Ctx) error {
	mh, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	files, ok := mh.File["files"]
	if !ok {
		return c.SendString("can't find any uploaded file")
	}

	err = h.indexesDB.Batch(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(consts.IndexBucketName())
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in the database")
		}
		for _, file := range files {
			err := func() error {
				f, err := file.Open()
				if err != nil {
					return errors.Wrap(err, "failed to read uploaded file")
				}
				defer f.Close()

				_, err = indexes.LoadIndexReader(b, f)

				return errors.Wrapf(err, "failed to add indexes %s", file.Filename)
			}()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return c.SendString(fmt.Sprintf("%d indexes uploaded", len(files)))
}

func (h handler) getPaper(doi string, c *fiber.Ctx) error {
	r, err := persist.GetIndexRecordDB(h.indexesDB, []byte(doi))
	if err != nil {
		if errors.Is(err, persist.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "failed to find index in the database")
		}

		return errors.Wrap(err, "failed to find index in the database")
	}

	t, err := persist.GetTorrentDB(h.torrentDB, r.InfoHash[:])
	if err != nil {
		return errors.Wrapf(err, "failed to get torrent data from Database, torrent infohash %s", r.HexInfoHash())
	}

	p, err := r.Build(doi, t)
	if err != nil {
		return errors.Wrap(err, "failed to detect offset of PDF file")
	}

	b, err := client.Fetch(p, t.Raw())
	if err != nil {
		return errors.Wrap(err, "failed to fetch paper")
	}

	c.Response().Header.SetContentType("application/pdf")

	return c.Send(b)
}

func (h *handler) paperQuery(c *fiber.Ctx) error {
	doi := c.Query("doi")
	if doi == "" {
		return errors.New("doi can't be empty string")
	}

	return h.getPaper(doi, c)
}
