package idmapfs

import (
	"fmt"
	"io"
	"runtime"

	"github.com/cpuguy83/idmapfs/idtools"
	"github.com/hanwen/go-fuse/fuse"
)

type mapper struct {
	m     *idtools.IdentityMapping
	debug bool
	l     io.Writer
}

func (m *mapper) mapAttr(a *fuse.Attr) {
	uid, gid, err := m.m.ToContainer(idFromOwner(&a.Owner))
	if err != nil {
		if m.debug {
			fmt.Fprintf(m.l, "no mapping for host attr owner %d:%d\n", a.Owner.Uid, a.Owner.Gid)
		}
		return
	}
	if m.debug {
		fmt.Fprintf(m.l, "mapping host attr owner %d:%d to container %d:%d\n", a.Owner.Uid, a.Owner.Gid, uid, gid)
	}
	a.Owner.Uid = uint32(uid)
	a.Owner.Gid = uint32(gid)
}

func (m *mapper) unmapAttr(a *fuse.Attr) {
	id, err := m.m.ToHost(idFromOwner(&a.Owner))
	if err != nil {
		if m.debug {
			fmt.Fprintf(m.l, "no mapping for host attr owner %d:%d\n", a.Owner.Uid, a.Owner.Gid)
		}
		return
	}
	if m.debug {
		fmt.Fprintf(m.l, "mapping host attr owner %d:%d to container %d:%d\n", a.Owner.Uid, a.Owner.Gid, id.UID, id.GID)
	}
	a.Owner.Uid = uint32(id.UID)
	a.Owner.Gid = uint32(id.GID)

}

func (m *mapper) unmapContext(c *fuse.Context) {
	var caller string
	if m.debug {
		_, file, line, _ := runtime.Caller(1)
		caller = fmt.Sprintf("%s#%d", file, line)
	}

	id, err := m.m.ToHost(idFromOwner(&c.Owner))
	if err != nil {
		if m.debug {
			fmt.Fprintf(m.l, "no mapping for user context %d:%d, caller: %s\n", c.Owner.Uid, c.Owner.Gid, caller)
		}
		return
	}
	if m.debug {
		fmt.Fprintf(m.l, "mapping user context %d:%d to container context %d:%d, caller: %s\n", c.Owner.Uid, c.Owner.Gid, id.UID, id.GID, caller)
	}
	c.Owner.Uid = uint32(id.UID)
	c.Owner.Gid = uint32(id.GID)
}

func idFromOwner(o *fuse.Owner) idtools.Identity {
	return idtools.Identity{UID: int(o.Uid), GID: int(o.Gid)}
}
