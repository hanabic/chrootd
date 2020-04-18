package utils

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/pkg/xattr"
)

func ExtractTar(tarball, dst string) (e error) {
	fd, err := os.Open(tarball)
	if err != nil {
		e = errors.Wrapf(err, "can not open tarball")
		return
	}
	defer fd.Close()
	defer func() {
		if e != nil {
			os.RemoveAll(dst)
		}
	}()

	tarfd := tar.NewReader(fd)
	for {
		hdr, err := tarfd.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			e = errors.Wrapf(err, "can not untar tarball")
			return
		}

		dst := filepath.Join(dst, hdr.Name)
		mode := hdr.FileInfo().Mode()

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dst, mode); err != nil {
				e = errors.Wrapf(err, "can not install dir")
				return
			}
		case tar.TypeReg, tar.TypeRegA:
			dstfd, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, mode)
			if err != nil {
				e = errors.Wrapf(err, "can not install file")
				return
			}

			_, err = io.Copy(dstfd, tarfd)
			if err != nil {
				dstfd.Close()
				e = errors.Wrapf(err, "can not fill file")
				return
			}

			dstfd.Close()
		case tar.TypeSymlink:
			os.Remove(dst)
			if err := os.Symlink(hdr.Linkname, dst); err != nil {
				e = errors.Wrapf(err, "can not overwrite the soft link")
				return
			}
		case tar.TypeLink:
			if err := os.Link(hdr.Linkname, dst); err != nil {
				e = errors.Wrapf(err, "can not overwrite the hard link")
				return
			}
		default:
			e = errors.Errorf("%s: unsupported type flag: %c", hdr.Name, hdr.Typeflag)
			return
		}

		switch hdr.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeRegA:
			dstfd, err := os.Open(dst)
			if err != nil {
				e = errors.Wrapf(err, "%s: can not open file %s", hdr.Name, dst)
				return
			}

			for attr, attrval := range hdr.PAXRecords {
				if strings.HasPrefix(attr, "SCHILY.xattr.") {
					attr = strings.TrimLeft(attr, "SCHILY.xattr.")

					if err := xattr.FSet(dstfd, attr, []byte(attrval)); err != nil {
						dstfd.Close()
						e = errors.Wrapf(err, "%s: can not set capability [%s] for file %s", hdr.Name, attr, dst)
						return
					}
				}
			}

			dstfd.Close()
		}

	}

	return
}
