module github.com/skilld-labs/go-ocis-api/rest

go 1.19

require (
	github.com/owncloud/libre-graph-api-go v1.0.5-0.20250217093259-fa3804be6c27
	github.com/studio-b12/gowebdav v0.10.0
)

require golang.org/x/net v0.41.0 // indirect

replace github.com/studio-b12/gowebdav => github.com/kobergj/gowebdav v0.0.0-20250102091030-aa65266db202
