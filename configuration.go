package main

import "flag"

type ServiceConfiguration struct {
	Port       string
	Directory  string
	Lifetime   int64
	SizeLimit  int64
	TrustProxy bool
}

func parseConfiguration() *ServiceConfiguration {
	// parse
	port := flag.String("port", "2069", "Port to use for HTTP")
	directory := flag.String("dir", "data", "The directory where to store the files")
	sizeLimit := flag.Int64("size-limit", 5e+6, "The file upload size limit")
	lifetime := flag.Int64("lifetime", 5*60*1000, "The lifetime of the stored files")
	trustProxy := flag.Bool("trust-proxy", false, "Determines if we should trust the X-Forwarded-For header")
	flag.Parse()

	// to struct
	config := ServiceConfiguration{}
	config.Port = *port
	config.Directory = *directory
	config.SizeLimit = *sizeLimit
	config.Lifetime = *lifetime
	config.TrustProxy = *trustProxy
	return &config
}
