package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/block"
	"github.com/kopia/kopia/repo/storage"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	connectCommand              = repositoryCommands.Command("connect", "Connect to a repository.")
	connectPersistCredentials   bool
	connectCacheDirectory       string
	connectMaxCacheSizeMB       int64
	connectMaxListCacheDuration time.Duration
)

func setupConnectOptions(cmd *kingpin.CmdClause) {
	// Set up flags shared between 'create' and 'connect'. Note that because those flags are used by both command
	// we must use *Var() methods, otherwise one of the commands would always get default flag values.
	cmd.Flag("persist-credentials", "Persist credentials").Default("true").BoolVar(&connectPersistCredentials)
	cmd.Flag("cache-directory", "Cache directory").PlaceHolder("PATH").StringVar(&connectCacheDirectory)
	cmd.Flag("cache-size-mb", "Size of local cache").PlaceHolder("MB").Default("500").Int64Var(&connectMaxCacheSizeMB)
	cmd.Flag("max-list-cache-duration", "Duration of index cache").Default("600s").Hidden().DurationVar(&connectMaxListCacheDuration)
}

func connectOptions() repo.ConnectOptions {
	return repo.ConnectOptions{
		CachingOptions: block.CachingOptions{
			CacheDirectory:          connectCacheDirectory,
			MaxCacheSizeBytes:       connectMaxCacheSizeMB << 20,
			MaxListCacheDurationSec: int(connectMaxListCacheDuration.Seconds()),
		},
	}
}

func init() {
	setupConnectOptions(connectCommand)
}

func runConnectCommandWithStorage(ctx context.Context, st storage.Storage) error {
	password := mustGetPasswordFromFlags(false, false)
	return runConnectCommandWithStorageAndPassword(ctx, st, password)
}

func runConnectCommandWithStorageAndPassword(ctx context.Context, st storage.Storage, password string) error {
	configFile := repositoryConfigFileName()
	if err := repo.Connect(ctx, configFile, st, password, connectOptions()); err != nil {
		return err
	}

	if connectPersistCredentials {
		if err := persistPassword(configFile, getUserName(), password); err != nil {
			return fmt.Errorf("unable to persist password: %v", err)
		}
	} else {
		deletePassword(configFile, getUserName())
	}

	printStderr("Connected to repository.\n")
	promptForAnalyticsConsent()

	return nil
}
