module github.com/twitchyliquid64/raspberry-box

// github.com/containers/buildah => github.com/containers/buildah v1.27.0
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191219165747-a9416c67da9f

require (
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/libpod v1.9.3
	github.com/containers/storage v1.18.2
	github.com/dsoprea/go-ext4 v0.0.0-20190528173430-c13b09fc0ff8
	github.com/dsoprea/go-logging v0.0.0-20190624164917-c4f10aab7696 // indirect
	github.com/freddierice/go-losetup v0.0.0-20170407175016-fc9adea44124
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/rekby/mbr v0.0.0-20190325193910-2b19b9cdeebc
	github.com/sirupsen/logrus v1.5.0
	github.com/tredoe/osutil v0.0.0-20161130133508-7d3ee1afa71c
	go.starlark.net v0.0.0-20190712141925-d6561f809f31
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

go 1.13
