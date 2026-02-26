module github.com/centretown/avweb

go 1.24.0

//replace github.com/centretown/avcamx => /home/dave/src/avcamx

require (
	github.com/centretown/avcamx v0.0.0-20260225003450-0be477aee52b
	github.com/gorilla/websocket v1.5.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/mattn/go-sqlite3 v1.14.28
	golang.org/x/text v0.3.3
	gopkg.in/yaml.v2 v2.2.8
)

require (
	github.com/aws/aws-sdk-go v1.38.20 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/korandiz/v4l v1.1.0 // indirect
	github.com/mattn/go-mjpeg v0.0.3 // indirect
	github.com/u2takey/ffmpeg-go v0.5.0 // indirect
	github.com/u2takey/go-utils v0.3.1 // indirect
)
