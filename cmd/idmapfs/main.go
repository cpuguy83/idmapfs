package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/cpuguy83/idmapfs"
	"github.com/cpuguy83/idmapfs/idtools"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

func main() {
	var (
		err              error
		mapUIDs, mapGIDs []string
		debug            bool
	)

	flags := pflag.NewFlagSet("idmapfs", pflag.ContinueOnError)
	flags.StringSliceVar(&mapUIDs, "map-uids", nil, "specify UID ranges to map")
	flags.StringSliceVar(&mapGIDs, "map-gids", nil, "specify GID ranges to map")
	flags.BoolVar(&debug, "debug", false, "enable debug logging")

	if err := flags.Parse(os.Args); err != nil {
		errorOut(err)
	}

	mapping, err := idMap(mapUIDs, mapGIDs)
	if err != nil {
		errorOut(errors.Wrap(err, "invalid format for uid mapping"))
	}

	fs := idmapfs.NewNodefs(nodefs.NewMemNodeFSRoot(flags.Arg(1)), mapping, os.Stderr)

	opts := nodefs.NewOptions()
	opts.Owner = nil
	opts.Debug = true

	srv, _, err := nodefs.MountRoot(flags.Arg(2), fs, opts)
	if err != nil {
		errorOut(err)
	}

	shutdown := func() {
		if err := srv.Unmount(); err != nil {
			fmt.Fprintln(os.Stderr, "error unmounting filesystem on shutdown:", err)
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for range c {
			shutdown()
		}
	}()

	srv.Serve()
}

func idMap(uSpecs, gSpecs []string) (*idtools.IdentityMapping, error) {
	uidMaps := make([]idtools.IDMap, 0, len(uSpecs))
	gidMaps := make([]idtools.IDMap, 0, len(gSpecs))

	for _, s := range uSpecs {
		split := strings.SplitN(s, ":", 3)
		switch len(split) {
		case 3:
			id, err := parseMappingSpec(split[0], split[1], split[2])
			if err != nil {
				return nil, err
			}
			uidMaps = append(uidMaps, id)
		default:
			return nil, fmt.Errorf("bad format for id map, expect `<host ID>:<container ID>:<size>:` %s", s)
		}
	}

	for _, s := range gSpecs {
		split := strings.SplitN(s, ":", 3)
		switch len(split) {
		case 3:
			id, err := parseMappingSpec(split[0], split[1], split[2])
			if err != nil {
				return nil, err
			}
			gidMaps = append(gidMaps, id)
		default:
			return nil, fmt.Errorf("bad format for id map, expect `<host ID>:<container ID>:<size>:` %s", s)
		}
	}

	return idtools.NewIDMappingsFromMaps(uidMaps, gidMaps), nil
}

func parseMappingSpec(h, c, s string) (idtools.IDMap, error) {
	var m idtools.IDMap

	mapID, err := strconv.Atoi(h)
	if err != nil {
		return m, errors.Wrap(err, "could not read host id")
	}

	startID, err := strconv.Atoi(c)
	if err != nil {
		return m, errors.Wrap(err, "could not read mapping start id")
	}

	size, err := strconv.Atoi(s)
	if err != nil {
		return m, errors.Wrap(err, "could not read mapping size")
	}

	m.HostID = mapID
	m.ContainerID = startID
	m.Size = size
	return m, nil
}

func errorOut(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
