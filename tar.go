package common

import (
	"archive/tar"
	"bufio"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

//Untar 将tar格式的镜像压缩包，解压到指定的目录下
func Untar(tarball string, dstPath string) error {
	hardLinks := make(map[string]string)

	//1.打开文件
	file, err := os.Open(tarball)
	if err != nil {
		return errors.New("os.Open failed")
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	//bufio.NewReader 可提升性能
	bufReader := bufio.NewReader(file)

	//2.读取文件中的每一行内容
	reader := tar.NewReader(bufReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			log.Println("tar file reach end of file EOF!")
			break
		} else if err != nil {
			return err
		}

		fileInfo := header.FileInfo()
		//dstFilePath 可能是文件的路径或文件夹的路径
		dstFilePath := filepath.Join(dstPath, header.Name)

		//判断镜像tar包中内容类型，文件和文件夹
		switch header.Typeflag {
		case tar.TypeDir: //目录
			//以什么权限来创建目录
			err = os.MkdirAll(dstFilePath, fileInfo.Mode())
			if err != nil {
				log.Println("os.MkdirAll error", err)
			}

			//log.Println("tar.TypeDir")
		case tar.TypeReg: //常规文件
			//拷贝普通文件到目的地址
			CopyRegFile(reader, header, dstFilePath)

			//log.Println("tar.TypeReg")
		case tar.TypeLink: //hard link
			//store details of hard links,process it finally
			linkPath := filepath.Join(dstPath, header.Linkname)
			linkPath2 := filepath.Join(dstPath, header.Name)
			hardLinks[linkPath2] = linkPath

			log.Println("tar.TypeLink")
		case tar.TypeSymlink: //Symbolic link
			err = os.Symlink(header.Linkname, dstFilePath)
			if os.IsExist(err) {
				continue
			}
			log.Println("tar.TypeSymlink")
		}
		//output header.Name
		log.Printf("name:%s\n", header.Name)
	}

	//4.要创建硬链接，目标必须存在，所以最后处理
	for k, v := range hardLinks {
		if err = os.Link(v, k); err != nil {
			return err
		}
	}

	//5.关闭打开的文件
	return nil
}
func CopyRegFile(reader io.Reader, hdr *tar.Header, dstFilePath string) {
	var (
		err  error
		file *os.File
	)

	//1.创建文件的父目录
	dstPath := filepath.Dir(dstFilePath)
	if _, err = os.Stat(dstPath); os.IsNotExist(err) {
		//父目录不存在，则创建目录
		err = os.MkdirAll(dstPath, os.FileMode(hdr.Mode))
		if err != nil {
			log.Println("os.MkdirAll failed", err)
			return
		}
	}

	//2.创建目标文件
	file, err = os.Create(dstFilePath)
	if err != nil {
		log.Println("os.Create failed", err)
		return
	}

	//3.拷贝源文件内容到目标文件
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Println("io.Copy failed", err)
		return
	}
}
