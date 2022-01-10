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
	serverFlag     = "server"
)

var entangleCmd = &cobra.Command{
	Use:   "entangle",
	Short: "Mount a folder on a given path",
	RunE: func(cmd *cobra.Command, args []string) error {

		if viper.GetBool(serverFlag) {
			var file *os.File

			onOpen := make(chan struct{})
			dir, err := os.MkdirTemp(os.TempDir(), "serverfiles-*")
			if err != nil {
				panic(err)
			}

			myFile := filepath.Join(dir, "serverfile.tar")

			networking.Connect("test", func(msg webrtc.DataChannelMessage) {

				log.Println(string(msg.Data))

				var v api.Message

				if err := json.Unmarshal(msg.Data, &v); err != nil {
					panic(err)
				}

				switch v.Opcode {
				case api.OpcodeOpen:
					var openOp api.OpenOp
					if err := json.Unmarshal(msg.Data, &openOp); err != nil {
						panic(err)
					}
					log.Println(openOp.ReadOnly)
					if openOp.ReadOnly == true {
						file, err = os.Open(myFile)
						if err != nil {
							msg, err := json.Marshal(api.NewOpenOpResponse(err.Error()))
							if err != nil {
								panic(err)
							}

							networking.WriteToDataChannel(msg)
						} else {
							msg, err := json.Marshal(api.NewOpenOpResponse(""))
							if err != nil {
								panic(err)
							}

							networking.WriteToDataChannel(msg)
						}
					} else {
						file, err = os.OpenFile(myFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
						if err != nil {
							msg, err := json.Marshal(api.NewOpenOpResponse(err.Error()))
							if err != nil {
								panic(err)
							}

							networking.WriteToDataChannel(msg)
						} else {
							msg, err := json.Marshal(api.NewOpenOpResponse(""))
							if err != nil {
								panic(err)
							}

							networking.WriteToDataChannel(msg)
						}
					}

					break
				case api.OpcodeClose:
					var closeOp api.CloseOp
					if err := json.Unmarshal(msg.Data, &closeOp); err != nil {
						panic(err)
					}

					file.Close()

					msg, err := json.Marshal(api.NewCloseOpResponse(""))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)

					break
				case api.OpcodeRead:
					var readOp api.ReadOp
					if err := json.Unmarshal(msg.Data, &readOp); err != nil {
						panic(err)
					}

					buf := make([]byte, readOp.Length)

					n, err := file.Read(buf)
					if err != nil {
						msg, err := json.Marshal(api.NewReadOpResponse(buf, int64(n), err.Error()))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					} else {
						msg, err := json.Marshal(api.NewReadOpResponse(buf, int64(n), ""))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					}

					break
				case api.OpcodeWrite:
					var writeOp api.WriteOp
					if err := json.Unmarshal(msg.Data, &writeOp); err != nil {
						panic(err)
					}

					n, err := file.Write(writeOp.Payload)
					if err != nil {
						msg, err := json.Marshal(api.NewWriteOpResponse(int64(n), err.Error()))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					} else {
						msg, err := json.Marshal(api.NewWriteOpResponse(int64(n), ""))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					}

					break
				case api.OpcodeSeek:
					var seekOp api.SeekOp
					if err := json.Unmarshal(msg.Data, &seekOp); err != nil {
						panic(err)
					}

					offset, err := file.Seek(seekOp.Offset, seekOp.Whence)
					if err != nil {
						msg, err := json.Marshal(api.NewSeekOpResponse(offset, err.Error()))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					} else {
						msg, err := json.Marshal(api.NewSeekOpResponse(offset, ""))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					}

					break
				}
			}, func() {
				onOpen <- struct{}{}
			})

			<-onOpen

			select {}
		} else {
			l := logging.NewJSONLogger(viper.GetInt(verboseFlag))
			boil.DebugMode = true
			boil.DebugWriter = os.Stderr

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
		}

	},
}

func init() {
	entangleCmd.PersistentFlags().String(mountpointFlag, "/tmp/mount", "Mountpoint to use for FUSE")
	entangleCmd.PersistentFlags().Int(recordSizeFlag, 20, "Amount of 512-bit blocks per second")
	entangleCmd.PersistentFlags().String(writeCacheFlag, filepath.Join(os.TempDir(), "stfs-write-cache"), "Directory to use for write cache")
	entangleCmd.PersistentFlags().Bool(serverFlag, false, "Set true to run as server")

	if err := viper.BindPFlags(entangleCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("sile-fystem")
	viper.AutomaticEnv()

	rootCmd.AddCommand(entangleCmd)
}
