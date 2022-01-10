package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/JakWai01/sile-fystem/pkg/filesystem"
	"github.com/JakWai01/sile-fystem/pkg/helpers"
	"github.com/alphahorizonio/libentangle/internal/logging"
	"github.com/jacobsa/fuse"
	"github.com/pion/webrtc/v3"
	"github.com/pojntfx/stfs/pkg/cache"
	"github.com/pojntfx/stfs/pkg/config"
	"github.com/pojntfx/stfs/pkg/fs"
	"github.com/pojntfx/stfs/pkg/operations"
	"github.com/pojntfx/stfs/pkg/persisters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/volatiletech/sqlboiler/v4/boil"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizonio/libentangle/pkg/networking"
)

const (
	mountpointFlag = "mountpoint"
	recordSizeFlag = "recordSize"
	writeCacheFlag = "writeCache"
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Mount a folder on a given path",
	RunE: func(cmd *cobra.Command, args []string) error {

		l := logging.NewJSONLogger(viper.GetInt(verboseFlag))
		boil.DebugMode = true
		boil.DebugWriter = os.Stderr
		// tm := tape.NewTapeManager(
		// 	viper.GetString(driveFlag),
		// 	viper.GetInt(recordSizeFlag),
		// 	false,
		// )

		rmFile := networking.NewRemoteFile()

		onOpen := make(chan struct{})

		go networking.Connect("test", func(msg webrtc.DataChannelMessage) {
			log.Println(string(msg.Data))

			var v api.Message

			if err := json.Unmarshal(msg.Data, &v); err != nil {
				panic(err)
			}

			switch v.Opcode {
			// This literally is the open handler. We can't open before
			case api.OpcodeOpenResponse:
				var openOpResponse api.OpenOpResponse
				if err := json.Unmarshal(msg.Data, &openOpResponse); err != nil {
					panic(err)
				}

				go func() {
					rmFile.OpenCh <- openOpResponse
				}()

				break
			case api.OpcodeCloseResponse:
				var closeOpResponse api.CloseOpResponse
				if err := json.Unmarshal(msg.Data, &closeOpResponse); err != nil {
					panic(err)
				}

				rmFile.CloseCh <- closeOpResponse

				break
			case api.OpcodeReadResponse:
				var readOpResponse api.ReadOpResponse
				if err := json.Unmarshal(msg.Data, &readOpResponse); err != nil {
					panic(err)
				}

				rmFile.ReadCh <- readOpResponse

				break
			case api.OpcodeWriteResponse:
				var writeOpResponse api.WriteOpResponse
				if err := json.Unmarshal(msg.Data, &writeOpResponse); err != nil {
					panic(err)
				}

				rmFile.WriteCh <- writeOpResponse

				break
			case api.OpcodeSeekResponse:
				var seekOpResponse api.SeekOpResponse
				if err := json.Unmarshal(msg.Data, &seekOpResponse); err != nil {
					panic(err)
				}

				rmFile.SeekCh <- seekOpResponse

				break
			}
		}, func() {
			onOpen <- struct{}{}
		})

		<-onOpen

		metadataPersister := persisters.NewMetadataPersister(viper.GetString(metadataFlag))
		if err := metadataPersister.Open(); err != nil {
			panic(err)
		}

		metadataConfig := config.MetadataConfig{
			Metadata: metadataPersister,
		}
		pipeConfig := config.PipeConfig{
			Compression: config.NoneKey,
			Encryption:  config.NoneKey,
			Signature:   config.NoneKey,
			RecordSize:  viper.GetInt(recordSizeFlag),
		}
		backendConfig := config.BackendConfig{
			GetWriter: func() (config.DriveWriterConfig, error) {
				if err := rmFile.Open(false); err != nil {
					return config.DriveWriterConfig{}, err
				}

				return config.DriveWriterConfig{
					DriveIsRegular: true,
					Drive:          rmFile,
				}, nil
			},
			CloseWriter: rmFile.Close,

			GetReader: func() (config.DriveReaderConfig, error) {
				if err := rmFile.Open(true); err != nil {
					return config.DriveReaderConfig{}, err
				}

				return config.DriveReaderConfig{
					DriveIsRegular: true,
					Drive:          rmFile,
				}, nil
			},
			CloseReader: rmFile.Close,

			GetDrive: func() (config.DriveConfig, error) {
				if err := rmFile.Open(true); err != nil {
					return config.DriveConfig{}, err
				}

				return config.DriveConfig{
					DriveIsRegular: true,
					Drive:          rmFile,
				}, nil
			},
			CloseDrive: rmFile.Close,
		}
		readCryptoConfig := config.CryptoConfig{}

		readOps := operations.NewOperations(
			backendConfig,
			metadataConfig,
			pipeConfig,
			readCryptoConfig,

			func(event *config.HeaderEvent) {
				l.Debug("Header read", event)
			},
		)
		writeOps := operations.NewOperations(
			backendConfig,
			metadataConfig,

			pipeConfig,
			config.CryptoConfig{},

			func(event *config.HeaderEvent) {
				l.Debug("Header write", event)
			},
		)

		stfs := fs.NewSTFS(
			readOps,
			writeOps,

			config.MetadataConfig{
				Metadata: metadataPersister,
			},
			config.CompressionLevelFastestKey,
			func() (cache.WriteCache, func() error, error) {
				return cache.NewCacheWrite(
					viper.GetString(writeCacheFlag),
					config.WriteCacheTypeFile,
				)
			},
			false,

			func(hdr *config.Header) {
				l.Trace("Header transform", hdr)
			},
			l,
		)

		root, err := stfs.Initialize("/", os.ModePerm)
		if err != nil {
			panic(err)
		}

		fs, err := cache.NewCacheFilesystem(
			stfs,
			root,
			config.NoneKey,
			0,
			"",
		)
		if err != nil {
			panic(err)
		}

		serve := filesystem.NewFileSystem(helpers.CurrentUid(), helpers.CurrentGid(), viper.GetString(mountpointFlag), root, l, fs)
		cfg := &fuse.MountConfig{}

		mfs, err := fuse.Mount(viper.GetString(mountpointFlag), serve, cfg)
		if err != nil {
			log.Fatalf("Mount: %v", err)
		}

		if err := mfs.Join(context.Background()); err != nil {
			log.Fatalf("Join %v", err)
		}

		return nil
	},
}

func init() {
	exampleCmd.PersistentFlags().String(mountpointFlag, "/tmp/mount", "Mountpoint to use for FUSE")

	exampleCmd.PersistentFlags().Int(recordSizeFlag, 20, "Amount of 512-bit blocks per second")
	exampleCmd.PersistentFlags().String(writeCacheFlag, filepath.Join(os.TempDir(), "stfs-write-cache"), "Directory to use for write cache")

	// Bind env variables
	if err := viper.BindPFlags(exampleCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("sile-fystem")
	viper.AutomaticEnv()

	rootCmd.AddCommand(exampleCmd)
}
