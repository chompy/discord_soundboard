package sound

import (
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
)

type FileInfo struct {
	Size int64
	Path string
	Hash string
}

type Client struct {
	storagePath string
	logger      *zerolog.Logger
}

func New(storagePath string, logger *zerolog.Logger) *Client {
	return &Client{storagePath: storagePath, logger: logger}
}

func (c *Client) Save(reader io.Reader) (FileInfo, error) {
	file, err := os.CreateTemp(c.storagePath, "_*")
	if err != nil {
		return FileInfo{}, err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	hash, size, err := WriteOpusFrames(reader, file)
	if err != nil {
		return FileInfo{}, err
	}
	if err := file.Close(); err != nil {
		return FileInfo{}, err
	}

	savePath := c.getHashPath(hash)
	c.logger.Info().Str("fileHash", hash).Int64("fileSize", size).Msgf("Save sound file to %s", savePath)

	if err := os.Rename(file.Name(), savePath); err != nil {
		return FileInfo{}, err
	}

	return FileInfo{Size: size, Hash: hash, Path: savePath}, nil
}

func (c *Client) Load(hash string) (*os.File, error) {
	filePath := c.getHashPath(hash)
	c.logger.Info().Str("fileHash", hash).Msgf("Load sound file from %s", filePath)
	return os.Open(filePath)
}

func (c *Client) Delete(hash string) error {
	filePath := c.getHashPath(hash)
	c.logger.Info().Str("fileHash", hash).Msgf("Delete sound file at %s", filePath)
	return os.Remove(filePath)
}

func (c *Client) getHashPath(hash string) string {
	return path.Join(c.storagePath, hash+".dat")
}
