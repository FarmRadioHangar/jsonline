package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gernest/blue"
	"github.com/urfave/cli"
)

const version = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "jsonline"
	app.Usage = "translates jsonobjects to influxdb line protocol"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "cofniguration file",
		},
		cli.StringFlag{
			Name:  "input",
			Usage: "json input",
		},
		cli.BoolFlag{
			Name:  "append",
			Usage: "appends to the output file, you need to speficy output file",
		},
		cli.BoolFlag{
			Name:  "stream",
			Usage: "reads from input stream",
		},
		cli.StringFlag{
			Name:  "out",
			Usage: "the output file",
		},
	}
	app.Action = line
	app.Version = version
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("jsonline: %v", err)
	}
}

//Config is the configuration setting for jsonline. This can be decoded from a
//json string.
type Config struct {
	Out         io.Writer
	In          io.Reader
	Tags        []string `json:"tags"`
	Fields      []string `json:"fields"`
	Measurement string   `json:"metric"`
	OutFile     string   `json:"output_file"`
	Append      bool     `json:"append"`
}

func defaultConfig() *Config {
	return &Config{
		Out: os.Stdout,
	}
}

//IsTag implements the Tag filetring function.
func (c *Config) IsTag(key string) bool {
	return false
}

//IsField implements field filtern function
func (c *Config) IsField(key string) bool {
	s := strings.Split(key, "_")
	if len(s) > 0 {
		if s[0] != "values" {
			return false
		}
		return true
	}
	return false
}

//IsMeasurement implements Measurement filtering function. This function is used
//to determine measurement name if the name is not provided yet.
func (c *Config) IsMeasurement(key string, value interface{}) (string, bool) {
	if key == c.Measurement {
		return "", false
	}
	s := strings.Split(key, "_")
	if len(s) > 1 {
		if s[0] != "values" {
			return "", false
		}
		if s[1] != c.Measurement {
			return "", false
		}
		return s[1], true
	}
	return "", false
}

func line(ctx *cli.Context) error {
	var conf *Config
	outFile := ctx.String("out")
	cfg := ctx.String("config")
	append := ctx.Bool("append")
	stream := ctx.Bool("stream")
	if cfg != "" {
		d, err := ioutil.ReadFile(cfg)
		if err != nil {
			return err
		}
		c := &Config{}
		err = json.Unmarshal(d, c)
		if err != nil {
			return err
		}

		conf = c
	} else {
		conf = defaultConfig()
		conf.Append = append
	}
	if conf.OutFile == "" {
		if outFile != "" {
			conf.OutFile = outFile
		}
	}
	conf.Measurement = "websocket"

	if conf.Append {
		if conf.OutFile == "" {
			return errors.New("you must specify -out flag to to use append")
		}
		f, err := os.OpenFile(conf.OutFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		conf.Out = f
	} else {
		if conf.OutFile != "" {
			f, err := os.OpenFile(conf.OutFile, os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			defer func() {
				_ = f.Close()
			}()
			conf.Out = f
		}
	}
	in := ctx.String("input")
	switch in {
	case "stdin":
		conf.In = os.Stdin
	case "file":
	case "":
		if ctx.NArg() > 0 {
			args := ctx.Args()
			conf.In = strings.NewReader(args.First())
		} else {
			conf.In = os.Stdin
		}

	}
	if conf.In != nil {
		if stream {
			return streamJSON(conf)
		}
		return renderJSON(conf)
	}
	return errors.New("missing input")
}

func streamJSON(conf *Config) error {
	r := bufio.NewReader(conf.In)
	for {
		txt, rerr := readJSON(r)
		if rerr != nil && txt == "" {
			return rerr
		}
		o, err := blue.Line(strings.NewReader(txt), blue.Options{
			IsTag:         conf.IsTag,
			IsField:       conf.IsField,
			IsMeasurement: conf.IsMeasurement,
		})
		if err != nil {
			return err
		}
		fmt.Fprintln(conf.Out, o)
		if rerr == io.EOF {
			break
		}
	}
	return nil
}

func renderJSON(conf *Config) error {
	r := bufio.NewReader(conf.In)
	txt, err := readJSON(r)
	if err != nil && txt == "" {
		return err
	}
	o, err := blue.Line(strings.NewReader(txt), blue.Options{
		IsTag:         conf.IsTag,
		IsField:       conf.IsField,
		IsMeasurement: conf.IsMeasurement,
	})
	fmt.Fprintln(conf.Out, o)
	return nil
}

func readJSON(r *bufio.Reader) (string, error) {
	var buf bytes.Buffer
	var rerr error
	open := 0
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				rerr = err
				break
			}
			return "", err
		}
		if buf.Len() > 0 && open == 0 {
			break
		}
		switch ch {
		case '{':
			_, _ = buf.WriteRune(ch)
			open++
			continue
		case '}':
			_, _ = buf.WriteRune(ch)
			open--
			continue
		default:
			if open == 0 && buf.Len() == 0 {
				continue
			}
			_, _ = buf.WriteRune(ch)
		}
	}
	if open != 0 {
		return "", errors.New("failed to find json string")
	}
	return buf.String(), rerr
}
