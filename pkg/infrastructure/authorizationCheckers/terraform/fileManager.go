//     MIT License
//
//     Copyright (c) Microsoft Corporation.
//
//     Permission is hereby granted, free of charge, to any person obtaining a copy
//     of this software and associated documentation files (the "Software"), to deal
//     in the Software without restriction, including without limitation the rights
//     to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//     copies of the Software, and to permit persons to whom the Software is
//     furnished to do so, subject to the following conditions:
//
//     The above copyright notice and this permission notice shall be included in all
//     copies or substantial portions of the Software.
//
//     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//     IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//     AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//     LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//     OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//     SOFTWARE

package terraform

import (
	"encoding/json"
	"io"
	"os"

	"github.com/manisbindra/az-mpf/pkg/domain"
	log "github.com/sirupsen/logrus"
)

func DoesTFFileExist(workingDir string, fileName string) bool {
	filePath := workingDir + "/" + fileName

	if _, err := os.Stat(filePath); err == nil {
		log.Infof("%s file exists \n", filePath)
		return true
	}
	log.Infof("%s file does not exist \n", filePath)
	return false
}

func CreateTFFile(workingDir string, fileName string) error {
	filePath := workingDir + "/" + fileName

	if _, err := os.Stat(filePath); err == nil {
		log.Infof("%s file already exists \n", filePath)
		return nil
	}

	_, err := os.Create(filePath)
	if err != nil {
		log.Warnf("error creating %s file: %s", filePath, err)
		return err
	}
	log.Infof("%s created file \n", filePath)
	return nil
}

func DeleteTFFile(workingDir string, fileName string) error {
	filePath := workingDir + "/" + fileName

	if _, err := os.Stat(filePath); err != nil {
		log.Infof("%s file does not exist \n", filePath)
		return nil
	}

	err := os.Remove(filePath)
	if err != nil {
		log.Warnf("error deleting %s file: %s", filePath, err)
		return err
	}
	log.Infof("%s deleted file \n", filePath)
	return nil
}

func doesEnteredDestroyPhaseStateFileExist(workingDir string, fileName string) bool {
	return DoesTFFileExist(workingDir, fileName)
}

func createEnteredDestroyPhaseStateFile(workingDir string, fileName string) error {
	return CreateTFFile(workingDir, fileName)
}

func deleteEnteredDestroyPhaseStateFile(workingDir string, fileName string) error {
	return DeleteTFFile(workingDir, fileName)
}

func saveResultAsJSON(rw io.ReadWriter, mpfResult domain.MPFResult) error {
	// serialize mpfREsult to json
	return json.NewEncoder(rw).Encode(mpfResult)
}

func loadResultFromJSON(r io.Reader) (*domain.MPFResult, error) {
	// deserialize json to mpfResult
	var mpfResult domain.MPFResult
	err := json.NewDecoder(r).Decode(&mpfResult)
	return &mpfResult, err
}

func LoadMPFResultFromFile(workingDir string, filename string) (*domain.MPFResult, error) {
	filePath := workingDir + "/" + filename
	file, err := os.Open(filePath)
	if err != nil {
		log.Warnf("error opening file for found permissions from failed run: %s", err)
		return nil, err
	}
	defer file.Close()

	return loadResultFromJSON(file)
}

// called for failed runs to save permissions to a file
func SaveMPFResultsToFile(workingDir string, filename string, mpfResult domain.MPFResult) error {
	filePath := workingDir + "/" + filename
	file, err := os.Create(filePath)
	if err != nil {
		log.Warnf("error creating file for found permissions from failed run: %s", err)
		return err
	}
	defer file.Close()

	return saveResultAsJSON(file, mpfResult)

}
