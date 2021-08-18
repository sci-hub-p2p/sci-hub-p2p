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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	torrent2 "github.com/anacrolix/torrent"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/client"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/persist"
)

type handler struct {
	torrentDB *bbolt.DB
	indexesDB *bbolt.DB
	btClient  *torrent2.Client
	m         *sync.Mutex
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
	raw := c.Request().Body()
	if len(raw) == 0 {
		return c.Status(fiber.StatusPaymentRequired).JSON(Error{
			Message: "error",
			Status:  "request body are empty",
		})
	}

	t, err := torrent.ParseRaw(raw)
	if err != nil {
		if errors.Is(err, torrent.ErrEncoding) {
			return fiber.NewError(fiber.StatusPermanentRedirect, "file content is not Bencode encoded")
		}

		if errors.Is(err, torrent.ErrNotValidTorrent) {
			return fiber.NewError(fiber.StatusPermanentRedirect, "file content is not valid torrent")
		}

		return fiber.NewError(fiber.StatusInternalServerError, "failed to parse torrent content")
	}

	var existed bool
	err = h.torrentDB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(consts.TorrentBucket())
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in the database")
		}

		existed = b.Get(t.RawInfoHash()) != nil

		return b.Put(t.RawInfoHash(), raw)
	})

	if err != nil {
		return errors.Wrapf(err, "failed to add torrent to database")
	}

	if !existed {
		c.Status(fiber.StatusCreated)
	}

	return nil
}

func (h *handler) indexesUpload(c *fiber.Ctx) error {
	raw := c.Request().Body()
	if len(raw) == 0 {
		return c.Status(fiber.StatusPaymentRequired).JSON(Error{
			Message: "error",
			Status:  "request body are empty",
		})
	}
	var count int
	err := h.indexesDB.Batch(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(consts.IndexBucketName())
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in the database")
		}
		count, err = indexes.LoadIndexRaw(b, raw)

		return errors.Wrapf(err, "failed to add indexes file")
	})

	if err == nil {
		return c.JSON(fiber.Map{
			"count": count,
		})
	}

	if errors.Is(err, &json.MarshalerError{}) {
		return c.Status(fiber.StatusPaymentRequired).JSON(Error{
			Message: "error",
			Status:  "body content is not valid jsonlines file",
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(Error{
		Message: "un-expected error:" + err.Error(),
		Status:  "error",
	})
}

func (h *handler) getPaper(doi string, c *fiber.Ctx) error {
	h.m.Lock()
	defer h.m.Unlock()

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

	if t == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrWithData{
			Status:  "error",
			Message: "missing torrent",
			Data: D{
				"info_hash": r.HexInfoHash(),
			},
		})
	}

	p, err := r.Build(doi, t)
	if err != nil {
		return errors.Wrap(err, "failed to detect offset of PDF file")
	}

	b, err := client.Fetch(h.btClient, p, t.Raw())
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

func (h *handler) torrentGet(c *fiber.Ctx) error {
	torrents := make([]*torrent.Torrent, 0)

	tx, err := h.torrentDB.Begin(false)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to open database")
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error("err close TX", zap.Error(err))
		}
	}()

	b := tx.Bucket(consts.TorrentBucket())
	if b == nil {
		return c.Status(fiber.StatusNotFound).JSON(WithData{torrents})
	}

	cur := b.Cursor()

	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		t, err := torrent.ParseRaw(v)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("failed to parse torrent %s, please consider re-add it", hex.EncodeToString(k)))
		}

		torrents = append(torrents, t)
	}

	return c.JSON(WithData{torrents})
}
