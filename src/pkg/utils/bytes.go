// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors
// forked from https://www.socketloop.com/tutorials/golang-byte-format-example

// Package utils provides generic utility functions.
package utils

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/defenseunicorns/zarf/src/pkg/message"
)

// RoundUp rounds a float64 to the given number of decimal places.
func RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

// ByteFormat formats a number of bytes into a human readable string.
func ByteFormat(inputNum float64, precision int) string {
	if precision <= 0 {
		precision = 1
	}

	var unit string
	var returnVal float64

	// https://www.techtarget.com/searchstorage/definition/mebibyte-MiB
	if inputNum >= 1000000000 {
		returnVal = RoundUp(inputNum/1000000000, precision)
		unit = " GB" // gigabyte
	} else if inputNum >= 1000000 {
		returnVal = RoundUp(inputNum/1000000, precision)
		unit = " MB" // megabyte
	} else if inputNum >= 1000 {
		returnVal = RoundUp(inputNum/1000, precision)
		unit = " KB" // kilobyte
	} else {
		returnVal = inputNum
		unit = " Byte" // byte
	}

	if returnVal > 1 {
		unit += "s"
	}

	return strconv.FormatFloat(returnVal, 'f', precision, 64) + unit
}

// RenderProgressBarForLocalDirWrite creates a progress bar that continuously tracks the progress of writing files to a local directory and all of its subdirectories.
// NOTE: This function runs infinitely until either completeChan or errChan is triggered, this function should be run in a goroutine while a different thread/process is writing to the directory.
func RenderProgressBarForLocalDirWrite(filepath string, expectedTotal int64, wg *sync.WaitGroup, completeChan chan int, errChan chan int, updateText string, successText string) {

	// Create a progress bar
	title := fmt.Sprintf("%s (%s of %s)", updateText, ByteFormat(float64(0), 2), ByteFormat(float64(expectedTotal), 2))
	progressBar := message.NewProgressBar(expectedTotal, title)

	for {
		// Could play around with only one
		select {
		case <-completeChan:
			// Send success message
			progressBar.Successf("%s (%s)", successText, ByteFormat(float64(expectedTotal), 2))
			wg.Done()
			return

		case <-errChan:
			progressBar.Stop()
			wg.Done()
			return

		default:
			// Read the directory size
			currentBytes, dirErr := GetDirSize(filepath)
			if dirErr != nil {
				message.Debugf("unable to get updated progress: %s", dirErr.Error())
				time.Sleep(200 * time.Millisecond)
				continue
			}

			// Update the progress bar with the current size
			title := fmt.Sprintf("%s (%s of %s)", updateText, ByteFormat(float64(currentBytes), 2), ByteFormat(float64(expectedTotal), 2))
			progressBar.Update(currentBytes, title)
			time.Sleep(200 * time.Millisecond)
		}
	}
}
