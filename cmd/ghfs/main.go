package main

import (
	"encoding/base64"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

func main() {
	log.SetFlags(0)

	// Parse arguments and require that we have the path.
	token := flag.String("token", "", "personal access token")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("path required")
	}
	log.Printf("mounting to: %s", flag.Arg(0))

	// Create FUSE connection.
	conn, err := fuse.Mount(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	// Create OAuth transport.
	var c *http.Client
	if *token != "" {
		c = (&oauth.Transport{Token: &oauth.Token{AccessToken: *token}}).Client()
	}

	// Create filesystem.
	filesys := &FS{Client: github.NewClient(c)}
	if err := fs.Serve(conn, filesys); err != nil {
		log.Fatal(err)
	}

	// Wait until the mount is unmounted or there is an error.
	<-conn.Ready
	if err := conn.MountError; err != nil {
		log.Fatal(err)
	}
}

// FS represents the
type FS struct {
	Client *github.Client
}

// Root returns the root filesystem node.
func (f *FS) Root() (fs.Node, error) {
	return &Root{FS: f}, nil
}

type Root struct {
	FS *FS
}

func (r *Root) Attr() fuse.Attr {
	return fuse.Attr{Mode: os.ModeDir | 0755}
}

func (r *Root) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if strings.HasPrefix(req.Name, ".") {
		return nil, fuse.ENOENT
	}

	u, _, err := r.FS.Client.Users.Get(req.Name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return &User{FS: r.FS, User: u}, nil
}

type User struct {
	*github.User
	FS *FS
}

func (u *User) Attr() fuse.Attr {
	return fuse.Attr{Mode: os.ModeDir | 0755}
}

func (u *User) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if strings.HasPrefix(req.Name, ".") {
		return nil, fuse.ENOENT
	}

	r, _, err := u.FS.Client.Repositories.Get(*u.Login, req.Name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return &Repository{FS: u.FS, Repository: r}, nil
}

type Repository struct {
	*github.Repository
	FS *FS
}

var _ = fs.HandleReadDirAller(&Repository{})

func (r *Repository) Attr() fuse.Attr {
	return fuse.Attr{Mode: os.ModeDir | 0755}
}

func (r *Repository) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if strings.HasPrefix(req.Name, ".") {
		return nil, fuse.ENOENT
	}

	fileContent, directoryContent, _, err := r.FS.Client.Repositories.GetContents(*r.Owner.Login, *r.Name, req.Name, nil)
	if err != nil {
		return nil, fuse.ENOENT
	}
	if fileContent != nil {
		return &File{FS: r.FS, Content: fileContent}, nil
	}
	return &Dir{FS: r.FS, Contents: directoryContent}, nil
}

func (r *Repository) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	_, directoryContent, _, err := r.FS.Client.Repositories.GetContents(*r.Owner.Login, *r.Name, "", nil)
	if err != nil {
		return nil, fuse.ENOENT
	}

	var entries []fuse.Dirent
	for _, f := range directoryContent {
		entries = append(entries, fuse.Dirent{Name: *f.Name})
	}
	return entries, nil
}

type File struct {
	Content *github.RepositoryContent
	FS      *FS
}

func (f *File) Attr() fuse.Attr {
	return fuse.Attr{
		Size: uint64(*f.Content.Size),
		Mode: 0755,
	}
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags |= fuse.OpenNonSeekable
	return &FileHandle{r: base64.NewDecoder(base64.StdEncoding, strings.NewReader(*f.Content.Content))}, nil
}

type FileHandle struct {
	r io.Reader
}

var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	resp.Data = buf[:n]
	return err
}

type Dir struct {
	Contents []*github.RepositoryContent
	FS       *FS
}

func (d *Dir) Attr() fuse.Attr {
	return fuse.Attr{
		Mode: os.ModeDir | 0755,
	}
}
