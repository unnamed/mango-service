package main

import "fmt"

func main() {
	// parse configuration from command line arguments
	config := *parseConfiguration()

	// create local file store
	store := &LocalFileStore{directory: config.Directory}
	fmt.Printf("[info] starting mango-service...\n")

	// delete existent files
	files, err := store.List()
	if err != nil {
		fmt.Printf("[error] failed to delete stored files: %v\n", err)
	} else {
		for _, file := range files {
			_, err = store.Delete(file)
			if err != nil {
				fmt.Printf("[error] failed to delete file '%s': %v\n", file, err)
			}
		}
	}

	// start http server
	_ = listenRequests(config.Port, store, &config)
}
