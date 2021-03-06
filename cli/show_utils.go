package cli

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	commonIndentJSON bool
	commonUnzip      bool
)

func setupShowCommand(cmd *kingpin.CmdClause) {
	cmd.Flag("json", "Pretty-print JSON content").Short('j').BoolVar(&commonIndentJSON)
	cmd.Flag("unzip", "Transparently unzip the content").Short('z').BoolVar(&commonUnzip)
}

func showContent(rd io.Reader) error {
	return showContentWithFlags(rd, commonUnzip, commonIndentJSON)
}

func showContentWithFlags(rd io.Reader, unzip, indentJSON bool) error {
	if unzip {
		gz, err := gzip.NewReader(rd)
		if err != nil {
			return fmt.Errorf("unable to open gzip stream: %v", err)
		}

		rd = gz
	}

	var buf1, buf2 bytes.Buffer
	if indentJSON {
		if _, err := io.Copy(&buf1, rd); err != nil {
			return err
		}

		if err := json.Indent(&buf2, buf1.Bytes(), "", "  "); err != nil {
			return err
		}

		rd = ioutil.NopCloser(&buf2)
	}

	if _, err := io.Copy(os.Stdout, rd); err != nil {
		return err
	}

	return nil
}
