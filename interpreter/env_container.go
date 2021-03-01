package interpreter

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/containers/image/docker"
	it "github.com/containers/image/types"
	"github.com/containers/libpod/libpod/events"
	"github.com/containers/libpod/libpod/image"
	"github.com/containers/libpod/pkg/util"
	"github.com/containers/storage"
	"go.starlark.net/starlark"
)

func containerBuiltins(s *Script) starlark.StringDict {
	return starlark.StringDict{
		"pull": starlark.NewBuiltin("pull", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var ref, arch starlark.String
			var fs *FSMountProxy
			if err := starlark.UnpackArgs("pull", args, kwargs, "fs", &fs, "ref", &ref, "arch?", &arch); err != nil {
				return starlark.None, err
			}

			if string(arch) == "" {
				arch = "arm64"
			}

			ref2, err := docker.ParseReference(string(ref))
			if err != nil {
				return nil, fmt.Errorf("parsing reference: %v", err)
			}

			ctx := context.Background()
			is, err := ref2.NewImageSource(ctx, &it.SystemContext{
				ArchitectureChoice: string(arch),
			})
			if err != nil {
				return nil, fmt.Errorf("creating image source: %v", err)
			}
			defer is.Close()

			runtime, err := image.NewImageRuntimeFromOptions(storage.StoreOptions{
				RunRoot:         path.Join(string(fs.fs.Mountpoint()), "run/containers/storage"),
				GraphRoot:       path.Join(string(fs.fs.Mountpoint()), "var/lib/containers/storage"),
				GraphDriverName: "overlay",
			})
			if err != nil {
				return nil, fmt.Errorf("creating image runtime: %v", err)
			}
			defer runtime.Shutdown(true)
			runtime.Eventer, err = events.NewEventer(events.EventerOptions{
				EventerType: "none",
			})
			if err != nil {
				return nil, fmt.Errorf("creating eventer: %v", err)
			}

			newImage, err := runtime.New(ctx, ref2.DockerReference().Name(), path.Join(string(fs.fs.Mountpoint()), "etc/containers/policy.json"), "", os.Stderr, &image.DockerRegistryOptions{
				ArchitectureChoice: string(arch),
			}, image.SigningOptions{}, nil, util.PullImageAlways)
			if err != nil {
				return nil, fmt.Errorf("pulling image: %v", err)
			}

			return starlark.String(newImage.ID()), nil
		}),
	}
}
