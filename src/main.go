package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/afero"
)

func main() {
	// input/outputディレクトリの定義
	inputDir, err := filepath.Abs("../files/input/")
	if err != nil {
		log.Fatalln(err)
	}
	outputDir, err := filepath.Abs("../files/output/")
	if err != nil {
		log.Fatalln(err)
	}
	os.Mkdir(outputDir, 0777)

	// 本処理
	insertAll(inputDir, outputDir)

	// 本処理（Afero ver.）
	var appFs = afero.NewOsFs()
	insertAllWithAfero(appFs, inputDir, outputDir)
}

//inputディレクトリ内のファイル全てに一行追加しoutputディレクトリに別名保存
func insertAll(inputDir string, outputDir string) {
	files := dirwalk(inputDir)
	for i, f := range files {

		// outputファイル名に便宜上連番を付与
		outputFile := filepath.Join(outputDir, ("testFile" + strconv.Itoa(i) + ".txt"))

		r, err := os.Open(f)
		if err != nil {
			log.Fatalln(err)
		}
		defer r.Close()
		w, err := os.Create(outputFile)
		if err != nil {
			log.Fatalln(err)
		}
		defer w.Close()

		insert(r, w)
	}
}

//ファイルをコピーして一行追加
func insert(r io.Reader, w io.Writer) {
	_, err := io.Copy(w, r)
	if err != nil {
		log.Fatalln(err)
	}
	w.Write([]byte("\nbar"))
}

// 指定したディレクトリ内のファイル一覧を取得
func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}
	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func insertAllWithAfero(appFs afero.Fs, inputDir string, outputDir string) afero.Fs {
	i := 0
	if err := afero.Walk(appFs, inputDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			outputFile := filepath.Join(outputDir, ("testFile" + strconv.Itoa(i) + ".txt"))
			r, err := appFs.Open(path)
			if err != nil {
				log.Fatalln(err)
				return err
			}
			defer r.Close()
			w, err := appFs.Create(outputFile)
			if err != nil {
				log.Fatalln(err)
				return err
			}
			insert(r, w)
			i++
		}
		return nil
	}); err != nil {
		log.Fatalln(err)
	}
	return appFs
}
