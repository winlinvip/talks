// This file is part of sophon.
// Copyright alibaba-inc.com

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	oe "github.com/ossrs/go-oryx-lib/errors"
	oh "github.com/ossrs/go-oryx-lib/http"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

func guess(html string) (string, error) {
	root := html
	if _, err := os.Stat(root); err == nil {
		return root, nil
	}

	root = path.Join(path.Dir(os.Args[0]), html)
	if _, err := os.Stat(root); err == nil {
		return root, nil
	}

	return "", os.ErrNotExist
}

type RTCConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	RegionEndpoint  string `json:"region_endpoint"`
	Region          string `json:"region"`
	GSLB            string `json:"gslb"`
}

type Config struct {
	Listen string     `json:"listen"`
	HTML   string     `json:"html"`
	RTC    *RTCConfig `json:"rtc"`
}

func run(ctx context.Context) error {
	cl := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var conf string
	cl.StringVar(&conf, "c", "", "The config file")
	cl.StringVar(&conf, "conf", "", "The file to load config from")

	var product string
	cl.StringVar(&product, "p", "", "The conf product")
	cl.StringVar(&product, "product", "", "The product section for config file")

	var version bool
	cl.BoolVar(&version, "v", false, "The version of Talks")
	cl.BoolVar(&version, "version", false, "The version of Talks")

	cl.Usage = func() {
		fmt.Println(fmt.Sprintf("Usage: %v -conf|-h|-v", os.Args[0]))
		fmt.Println(fmt.Sprintf("	-conf           The config file path."))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v -conf talks.conf", os.Args[0]))
	}

	if err := cl.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		return err
	}

	if version {
		fmt.Fprintf(os.Stderr, "Version %v\n", Version())
		os.Exit(0)
	}

	oh.Server = fmt.Sprintf("%v/%v", Signature(), Version())
	fmt.Println(fmt.Sprintf("Talks of %v/%v system", Signature(), Version()))

	if conf == "" {
		cl.Usage()
		os.Exit(-1)
	}
	ol.Tf(ctx, "Parse config %v", conf)

	f, err := os.Open(conf)
	if err != nil {
		return oe.Wrapf(err, "open %v", conf)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return oe.Wrapf(err, "read %v", conf)
	}

	c := Config{}
	if err := json.Unmarshal(b, &c); err != nil {
		return oe.Wrapf(err, "parse %v %s", conf, b)
	}

	root, err := guess(c.HTML)
	if err != nil {
		return oe.Wrapf(err, "guess %v", c.HTML)
	}
	ol.Tf(ctx, "Listen at %v, html is %v, root is %v", c.Listen, c.HTML, root)

	pattern := "/talks/v1/versions"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		oh.WriteVersion(w, r, Version())
	})

	pattern = "/"
	ol.Tf(ctx, "Handle %v", pattern)
	fs := http.FileServer(http.Dir(root))
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		oh.SetHeader(w)
		fs.ServeHTTP(w, r)
	})

	pattern = "/talks/v1/collect"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		oh.SetHeader(w)
		w.Header().Set("Content-Type", "image/gif")
		w.Write([]byte{
			0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0xff, 0x00, 0xff, 0xff, 0xff,
			0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
			0x01, 0x00, 0x3b,
		})
	})

	pattern = "/talks/v1/iceconfig"
	ol.Tf(ctx, "Handle %v", pattern)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		oh.SetHeader(w)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(
`{
	"lifetimeDuration": "86400s",
	"iceServers": [
		{
			"urls": [
				"stun:173.194.199.127:19302",
				"stun:[2607:f8b0:4003:c0c::7f]:19302"
			]
		},
		{
			"urls": [
				"turn:64.233.169.127:19305?transport=udp",
				"turn:[2607:f8b0:4003:c08::7f]:19305?transport=udp",
				"turn:64.233.169.127:19305?transport=tcp",
				"turn:[2607:f8b0:4003:c08::7f]:19305?transport=tcp"
			],
			"username": "CIPDgd0FEgb1etlNQ8QYqvGggqMKIICjBQ",
			"credential": "EUM/Cz+vcwPw8WgDqVDiboREIGY=",
			"maxRateKbps": "8000"
		}
	],
	"blockStatus": "NOT_BLOCKED",
	"iceTransportPolicy": "all"
}`,
		))
	})

	if err := http.ListenAndServe(c.Listen, nil); err != nil {
		return oe.Wrapf(err, "serve")
	}

	return nil
}

func main() {
	ctx := ol.WithContext(context.Background())

	if err := run(ctx); err != nil {
		ol.Ef(ctx, "run err %+v", err)
		os.Exit(-1)
	}
}
