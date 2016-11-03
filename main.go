package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gernest/blue"
	"github.com/urfave/cli"
)

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
		cli.StringFlag{
			Name:  "out",
			Usage: "the output file",
		},
	}
	app.Action = line
	app.Run(os.Args)
}

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

func (c *Config) IsTag(key string) bool {
	for _, v := range c.Tags {
		if v == key {
			return true
		}
	}
	return blue.IsTag(key)
}
func (c *Config) IsField(key string) bool {
	for _, v := range c.Fields {
		if v == key {
			return true
		}
	}
	return blue.IsField(key)
}

func (c *Config) Ismeasurement(key string) bool {
	if c.Measurement != "" {
		return c.Measurement == key
	}
	return blue.IsMeasurement(key)
}

func line(ctx *cli.Context) error {
	var conf *Config
	outFile := ctx.String("out")
	cfg := ctx.String("config")
	append := ctx.Bool("append")
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

	if conf.Append {
		if conf.OutFile == "" {
			return errors.New("you must specify -out flag to to use append")
		}
		f, err := os.OpenFile(conf.OutFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		conf.Out = f
	} else {
		if conf.OutFile != "" {
			f, err := os.OpenFile(conf.OutFile, os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			conf.Out = f
		}
	}
	in := ctx.String("input")
	if in == "" {
		if len(os.Args) > 1 {
			o := os.Args[1]
			conf.In = strings.NewReader(o)
		} else {
			txt, err := readIn()
			if err != nil {
				return err
			}
			conf.In = strings.NewReader(txt)
		}
	} else {
		conf.In = strings.NewReader(in)
	}
	if conf.In != nil {
		o, err := blue.Line(conf.In, blue.Options{
			IsTag:         conf.IsTag,
			IsField:       conf.IsField,
			IsMeasurement: conf.Ismeasurement,
		})
		_, err = conf.Out.Write([]byte(o))
		return err
	}
	return errors.New("missing input")
}

func readIn() (string, error) {
	r := bufio.NewReader(os.Stdin)
	return readJSON(r)
}

func readJSON(r *bufio.Reader) (string, error) {
	var buf bytes.Buffer
	open := 0
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if buf.Len() > 0 && open == 0 {
			break
		}
		switch ch {
		case '{':
			buf.WriteRune(ch)
			open++
			continue
		case '}':
			buf.WriteRune(ch)
			open--
			continue
		default:
			if open == 0 && buf.Len() == 0 {
				continue
			}
			buf.WriteRune(ch)
		}
	}
	if open != 0 {
		return "", errors.New("failed to find json string")
	}
	return buf.String(), nil
}
