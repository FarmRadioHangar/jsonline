package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gernest/blue"
	"github.com/urfave/cli"
)

const version = "0.1.1"

func main() {
	app := cli.NewApp()
	app.Name = "jsonline"
	app.Usage = "translates json objects to influxdb line protocol"
	app.Version = version
	app.Commands = []cli.Command{
		{
			Name:    "stream",
			Aliases: []string{"s"},
			Usage:   "streams from  stdin",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "The name of the measurement",
				},
			},
			Action: streamLine,
		},
	}
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

func streamLine(ctx *cli.Context) error {
	cfg := &Config{}
	name := ctx.String("name")
	if name == "" {
		return errors.New("missng name of the measurement use the --name flag")
	}
	cfg.Measurement = name
	cfg.In = os.Stdin
	cfg.Out = os.Stdout
	return streamJSON(cfg)
}
