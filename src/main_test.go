package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/spf13/afero"
)

// その1：ファイルシステムの操作をスコープから外したテスト
func TestInsert(t *testing.T) {

	// input/outputを初期化
	in := bytes.NewBufferString("foo")
	out := new(bytes.Buffer)

	// テスト対象の処理を実行
	insert(in, out)

	// 出力ファイルが期待通りかチェック
	expected := []byte("foo\nbar")
	if bytes.Compare(expected, out.Bytes()) != 0 {
		t.Fatalf("not matched. expected: %s, actual: %s", expected, out.Bytes())
	}
}

// その2：ファイルシステムの操作もスコープに入れたテスト
func TestInsertAll(t *testing.T) {

	// input/outputディレクトリを定義
	inputDir, err := filepath.Abs("../files/ut/input/")
	if err != nil {
		log.Fatalln(err)
	}
	outputDir, err := filepath.Abs("../files/ut/output/")
	if err != nil {
		log.Fatalln(err)
	}
	os.Mkdir(outputDir, 0777)

	// テスト対象の処理を実行
	insertAll(inputDir, outputDir)

	//各出力ファイルが期待通りかチェック
	files := dirwalk(outputDir)
	for i, f := range files {
		file, err := os.Open(f)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
		out, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalln(err)
		}

		// ファイルパスとファイル名が期待通りかテスト
		expFileName := filepath.Join(outputDir, ("testFile" + strconv.Itoa(i) + ".txt"))
		if expFileName != f {
			t.Fatalf("path or name not matched. expected: %s, actual: %s", expFileName, f)
		}

		// ファイルの中身が期待通りかテスト
		expContent := []byte("testFile" + strconv.Itoa(i) + "\nfoo\nbar")
		if bytes.Compare(expContent, out) != 0 {
			t.Fatalf("fileContent not matched. expected: %s, actual: %s", expContent, out)
		}

	}

	// テスト用に出力したファイルを全削除
	if err := os.RemoveAll(outputDir); err != nil {
		log.Fatalln("outputDir could not be deleted. dir path :", outputDir)
	}
}

// その3：/spf13/aferoの仮想ファイルシステムを用いたテスト
func TestInsertAllWithAfero(t *testing.T) {

	// mockを定義
	appFs := afero.NewMemMapFs()

	// inputファイル/outputディレクトリをmock内に作成
	inputDir := "../files/input/"
	outputDir := "../files/output/"
	afero.WriteFile(appFs, filepath.Join(inputDir, "testFile0.txt"), []byte("testFile0\nfoo"), 0644)
	afero.WriteFile(appFs, filepath.Join(inputDir, "testFile1.txt"), []byte("testFile1\nfoo"), 0644)
	afero.WriteFile(appFs, filepath.Join(inputDir, "testFile2.txt"), []byte("testFile2\nfoo"), 0644)
	appFs.Mkdir(outputDir, 0777)

	// テスト対象の処理を実行
	appFs = insertAllWithAfero(appFs, inputDir, outputDir)

	// 実行結果をactualに格納
	var actualFileNames []string
	var actualContents [][]byte
	if err := afero.Walk(appFs, outputDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			actualFileNames = append(actualFileNames, path)
			f, err := appFs.Open(path)
			if err != nil {
				log.Fatalln(err)
			}
			defer f.Close()
			actualContent, err := afero.ReadAll(f)
			if err != nil {
				log.Fatalln(err)
			}
			actualContents = append(actualContents, actualContent)
		}
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	// ファイルパスとファイル名が期待通りかテスト
	var expFileNames []string
	expFileNames = append(expFileNames, filepath.Join(outputDir, ("testFile0.txt")))
	expFileNames = append(expFileNames, filepath.Join(outputDir, ("testFile1.txt")))
	expFileNames = append(expFileNames, filepath.Join(outputDir, ("testFile2.txt")))
	if !reflect.DeepEqual(expFileNames, actualFileNames) {
		t.Fatalf("path or name not matched. expected: %s, actual: %s", expFileNames, actualFileNames)
	}

	// ファイルの中身が期待通りかテスト
	var expContents [][]byte
	expContents = append(expContents, []byte("testFile0\nfoo\nbar"))
	expContents = append(expContents, []byte("testFile1\nfoo\nbar"))
	expContents = append(expContents, []byte("testFile2\nfoo\nbar"))
	if !reflect.DeepEqual(expContents, actualContents) {
		t.Fatalf("fileContent not matched. expected: %s, actual: %s", expContents, actualContents)
	}
}
