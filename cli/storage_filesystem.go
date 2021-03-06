package cli

import (
	"context"
	"os"
	"strconv"

	"github.com/kopia/kopia/repo/storage"
	"github.com/kopia/kopia/repo/storage/filesystem"
	"gopkg.in/alecthomas/kingpin.v2"
)

var options filesystem.Options

var (
	connectOwnerUID string
	connectOwnerGID string
	connectFileMode string
	connectDirMode  string
	connectFlat     bool
)

func connect(ctx context.Context) (storage.Storage, error) {
	fso := options
	if v := connectOwnerUID; v != "" {
		fso.FileUID = getIntPtrValue(v, 10)
	}
	if v := connectOwnerGID; v != "" {
		fso.FileGID = getIntPtrValue(v, 10)
	}
	if v := connectFileMode; v != "" {
		fso.FileMode = getFileModeValue(v, 8)
	}
	if v := connectDirMode; v != "" {
		fso.DirectoryMode = getFileModeValue(v, 8)
	}

	if connectFlat {
		fso.DirectoryShards = []int{}
	}

	return filesystem.New(ctx, &fso)
}

func init() {
	RegisterStorageConnectFlags(
		"filesystem",
		"a filesystem",
		func(cmd *kingpin.CmdClause) {
			cmd.Flag("path", "Path to the repository").Required().StringVar(&options.Path)
			cmd.Flag("owner-uid", "User ID owning newly created files").PlaceHolder("USER").StringVar(&connectOwnerUID)
			cmd.Flag("owner-gid", "Group ID owning newly created files").PlaceHolder("GROUP").StringVar(&connectOwnerGID)
			cmd.Flag("file-mode", "File mode for newly created files (0600)").PlaceHolder("MODE").StringVar(&connectFileMode)
			cmd.Flag("dir-mode", "Mode of newly directory files (0700)").PlaceHolder("MODE").StringVar(&connectDirMode)
			cmd.Flag("flat", "Use flat directory structure").BoolVar(&connectFlat)
		},
		connect)
}

func getIntPtrValue(value string, base int) *int {
	if int64Val, err := strconv.ParseInt(value, base, 32); err == nil {
		intVal := int(int64Val)
		return &intVal
	}

	return nil
}

func getFileModeValue(value string, def os.FileMode) os.FileMode {
	if uint32Val, err := strconv.ParseUint(value, 8, 32); err == nil {
		return os.FileMode(uint32Val)
	}

	return def
}
