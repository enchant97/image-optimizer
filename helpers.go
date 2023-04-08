package main

import "log"

// Return the minimum of two integers
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Panic on error and log the error
func panicOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
