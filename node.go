package idmapfs

import (
	"fmt"
	"io"
	"time"

	"github.com/cpuguy83/idmapfs/idtools"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

func NewNodefs(node nodefs.Node, m *idtools.IdentityMapping, logger io.Writer) nodefs.Node {
	return &mappedNode{
		Node:   node,
		l:      logger,
		mapper: mapper{m: m, l: logger},
	}
}

type mappedNode struct {
	nodefs.Node
	mapper
	l io.Writer
}

func (n *mappedNode) mapInodeResult(ino *nodefs.Inode, status fuse.Status) (*nodefs.Inode, fuse.Status) {
	// TODO: map this
	return ino, status
}

func (n *mappedNode) Lookup(a *fuse.Attr, name string, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.unmapContext(c)
	ino, status := n.Node.Lookup(a, name, c)
	n.mapAttr(a)
	return n.mapInodeResult(ino, status)
}

func (n *mappedNode) Access(mode uint32, c *fuse.Context) (code fuse.Status) {
	n.unmapContext(c)
	return n.Node.Access(mode, c)
}

func (n *mappedNode) Readlink(c *fuse.Context) ([]byte, fuse.Status) {
	n.unmapContext(c)
	return n.Node.Readlink(c)
}

func (n *mappedNode) Mknod(name string, mode uint32, dev uint32, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.unmapContext(c)
	return n.mapInodeResult(n.Node.Mknod(name, mode, dev, c))
}

func (n *mappedNode) Mkdir(name string, mode uint32, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.unmapContext(c)
	return n.mapInodeResult(n.Node.Mkdir(name, mode, c))
}

func (n *mappedNode) Unlink(name string, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.Unlink(name, c)
}

func (n *mappedNode) Rmdir(name string, c *fuse.Context) (code fuse.Status) {
	n.unmapContext(c)
	return n.Node.Rmdir(name, c)
}

func (n *mappedNode) Symlink(name string, content string, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.unmapContext(c)
	return n.mapInodeResult(n.Node.Symlink(name, content, c))
}

func (n *mappedNode) Rename(oldName string, newParent nodefs.Node, newName string, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.Rename(oldName, newParent, newName, c)
}

func (n *mappedNode) Link(name string, existing nodefs.Node, c *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.unmapContext(c)
	return n.mapInodeResult(n.Node.Link(name, existing, c))
}

func (n *mappedNode) Create(name string, flags uint32, mode uint32, c *fuse.Context) (nodefs.File, *nodefs.Inode, fuse.Status) {
	n.unmapContext(c)

	f, ino, status := n.Node.Create(name, flags, mode, c)
	ino, status = n.mapInodeResult(ino, status)
	return &mappedFile{File: f, m: n.m}, ino, status
}

func (n *mappedNode) Open(flags uint32, c *fuse.Context) (nodefs.File, fuse.Status) {
	n.unmapContext(c)

	f, status := n.Node.Open(flags, c)
	return &mappedFile{File: f, m: n.m}, status
}

func (n *mappedNode) OpenDir(c *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	n.unmapContext(c)
	return n.Node.OpenDir(c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Read(f nodefs.File, dest []byte, off int64, c *fuse.Context) (fuse.ReadResult, fuse.Status) {
	n.unmapContext(c)
	return n.Node.Read(f, dest, off, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Write(f nodefs.File, data []byte, off int64, c *fuse.Context) (uint32, fuse.Status) {
	n.unmapContext(c)
	return n.Node.Write(f, data, off, c)
}

func (n *mappedNode) GetXAttr(attr string, c *fuse.Context) ([]byte, fuse.Status) {
	n.unmapContext(c)
	return n.Node.GetXAttr(attr, c)
}

func (n *mappedNode) RemoveXAttr(attr string, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.RemoveXAttr(attr, c)
}

func (n *mappedNode) SetXAttr(attr string, data []byte, flags int, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.SetXAttr(attr, data, flags, c)
}

func (n *mappedNode) ListXAttr(c *fuse.Context) (attrs []string, code fuse.Status) {
	n.unmapContext(c)
	return n.Node.ListXAttr(c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) GetLk(f nodefs.File, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.GetLk(f, owner, lk, flags, out, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) SetLk(f nodefs.File, owner uint64, lk *fuse.FileLock, flags uint32, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.SetLk(f, owner, lk, flags, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) SetLkw(f nodefs.File, owner uint64, lk *fuse.FileLock, flags uint32, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.SetLkw(f, owner, lk, flags, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) GetAttr(out *fuse.Attr, f nodefs.File, c *fuse.Context) (code fuse.Status) {
	n.unmapContext(c)
	status := n.Node.GetAttr(out, f, c)
	n.mapAttr(out)
	return status
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Chmod(f nodefs.File, perms uint32, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.Chmod(f, perms, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Chown(f nodefs.File, uid uint32, gid uint32, c *fuse.Context) fuse.Status {
	n.unmapContext(c)

	id, err := n.m.ToHost(idtools.Identity{UID: int(uid), GID: int(gid)})
	if err != nil {
		fmt.Fprintf(n.l, "Chown: no mapping for %d:%d, keeping original uid:gid", uid, gid)
		return n.Node.Chown(f, uid, gid, c)
	}

	return n.Node.Chown(f, uint32(id.UID), uint32(id.GID), c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Truncate(f nodefs.File, size uint64, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Node.Truncate(f, size, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Utimens(f nodefs.File, atime *time.Time, mtime *time.Time, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Utimens(f, atime, mtime, c)
}

// TODO: Does the file need mapping/unmapping?
func (n *mappedNode) Fallocate(f nodefs.File, off uint64, size uint64, mode uint32, c *fuse.Context) fuse.Status {
	n.unmapContext(c)
	return n.Fallocate(f, off, size, mode, c)

}

func (n *mappedNode) StatFs() *fuse.StatfsOut {
	return n.Node.StatFs()
}
